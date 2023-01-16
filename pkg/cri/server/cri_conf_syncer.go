package server

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	criconfig "github.com/containerd/containerd/pkg/cri/config"
)

// criConfSyncer is used to reload cri conf triggered by fs change
// events.
type criConfSyncer struct {
	// only used for lastSyncStatus
	sync.RWMutex
	lastSyncStatus error

	watcher  *fsnotify.Watcher
	confPath string
}

// newCRIConfSyncer creates cri conf syncer.
func (c *criService) newCRIConfSyncer(confPath string) (_ *criConfSyncer, retErr error) {
	if confPath == "" {
		return nil, nil
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create fsnotify watcher")
	}
	defer func() {
		if retErr != nil {
			watcher.Close()
		}
	}()

	if !filepath.IsAbs(confPath) {
		return nil, fmt.Errorf("criConfSyncer failed: dynamic config path must be abs path: %w", err)
	}

	confDir := filepath.Dir(confPath)
	if err := os.MkdirAll(confDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cri conf dir=%s for watch: %w", confDir, err)
	}

	if err := watcher.Add(confDir); err != nil {
		return nil, fmt.Errorf("failed to watch cri conf dir %s: %w", confDir, err)
	}

	syncer := &criConfSyncer{
		watcher:  watcher,
		confPath: confPath,
	}

	// chaofeng.tty: should we panic here?
	if err := syncer.update(&c.config); err != nil {
		logrus.WithError(err).Error("failed to load dynamic cri config during init, please check CRI plugin status")
		syncer.updateLastStatus(err)
	}

	return syncer, nil
}

// confSyncLoop monitors any fs change events from cri config file and tries to update config
func (c *criService) confSyncLoop() error {
	if c.criConfMonitor == nil {
		return nil
	}
	for {
		select {
		case event, ok := <-c.criConfMonitor.watcher.Events:
			if !ok {
				logrus.Debugf("cri conf watcher channel is closed")
				return nil
			}

			if event.Name != c.criConfMonitor.confPath {
				continue
			}
			// Only reload config when receiving write/rename/remove
			// events
			if event.Op&(fsnotify.Chmod|fsnotify.Rename) > 0 {
				continue
			}
			logrus.Infof("receiving change event from cri conf path: %s %s", event, c.criConfMonitor.confPath)

			lerr := c.criConfMonitor.update(&c.config)
			if lerr != nil {
				logrus.WithError(lerr).Errorf("failed to reload cri configuration after receiving fs change event(%s)", event)
			}

			c.criConfMonitor.updateLastStatus(lerr)

		case err := <-c.criConfMonitor.watcher.Errors:
			if err != nil {
				logrus.WithError(err).Error("failed to continue sync cri conf change")
				return err
			}
		}
	}
}

// update monitors any fs change events from cri conf dir and tries to update configuration.
func (syncer *criConfSyncer) update(config *criconfig.Config) error {
	if err := criconfig.LoadConfig(syncer.confPath, config); err != nil {
		return fmt.Errorf("failed to load dynamic cri config file: %w", err)
	}
	return nil
}

// lastStatus retrieves last sync status.
func (syncer *criConfSyncer) lastStatus() error {
	syncer.RLock()
	defer syncer.RUnlock()
	return syncer.lastSyncStatus
}

// updateLastStatus will be called after every single cri conf load.
func (syncer *criConfSyncer) updateLastStatus(err error) {
	syncer.Lock()
	defer syncer.Unlock()
	syncer.lastSyncStatus = err
}

// stop stops watcher in the syncLoop.
func (syncer *criConfSyncer) stop() error {
	return syncer.watcher.Close()
}
