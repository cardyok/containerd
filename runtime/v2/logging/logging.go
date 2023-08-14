/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package logging

import (
	"context"
	"io"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/containerd/containerd/log"
)

var (
	// shimLoggingPath describes the path for shim logs
	mu            sync.Mutex
	shimLogPath   = make(map[string]log.LoggerConfig)
	shimLoggers   = make(map[string]loggerStruct)
	DefaultLogger io.Writer
)

type loggerStruct struct {
	writer io.Writer
	closer func() error
}

// Config of the container logs
type Config struct {
	ID        string
	Namespace string
	Stdout    io.Reader
	Stderr    io.Reader
}

// LoggerFunc is implemented by custom v2 logging binaries
type LoggerFunc func(context.Context, *Config, func() error) error

func GetShimLogger(ctx context.Context, runtime string, ns string, id string) (writer io.Writer, closer func() error) {
	mu.Lock()
	if logger, ok := shimLoggers[runtime]; ok {
		mu.Unlock()
		return logger.writer, logger.closer
	}

	loggerConfig, ok := shimLogPath[runtime]
	mu.Unlock()

	defer func() {
		mu.Lock()
		if _, ok := shimLoggers[runtime]; !ok {
			shimLoggers[runtime] = loggerStruct{
				writer: writer,
				closer: closer,
			}
		}
		mu.Unlock()
	}()

	log.G(ctx).Infof("[RuntimeV2] Got shim logger wit path \"%v\" for container %v with runtime %v", loggerConfig.LogPath, id, runtime)
	if !ok || loggerConfig.LogPath == "" {
		return DefaultLogger, func() error { return nil }
	}

	shimLogExporter := &lumberjack.Logger{
		Filename:         loggerConfig.LogPath,
		MaxSize:          500,
		MaxBackups:       3,
		Compress:         true,
		CompressRate:     10 * 1024 * 1024,
		CompressCapacity: 10 * 1024 * 1024,
	}
	if loggerConfig.LogReplica != 0 {
		shimLogExporter.MaxBackups = loggerConfig.LogReplica
	}
	if loggerConfig.LogSize != 0 {
		shimLogExporter.MaxSize = loggerConfig.LogSize
	}
	if loggerConfig.NoCompress {
		shimLogExporter.Compress = false
	}

	return shimLogExporter, func() error {
		return shimLogExporter.Close()
	}
}

func SetShimLogger(ctx context.Context, logger log.LoggerConfig, runtime string) error {
	mu.Lock()
	shimLogPath[runtime] = logger
	mu.Unlock()
	log.G(ctx).Infof("[RuntimeV2] Set shim logger wit path \"%v\" for runtime %v", logger.LogPath, runtime)
	return nil
}
