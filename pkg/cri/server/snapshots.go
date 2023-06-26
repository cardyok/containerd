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

package server

import (
	"context"
	"fmt"
	"github.com/containerd/containerd/log"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/containerd/containerd/errdefs"
	snapshot "github.com/containerd/containerd/snapshots"

	snapshotstore "github.com/containerd/containerd/pkg/cri/store/snapshot"
	ctrdutil "github.com/containerd/containerd/pkg/cri/util"
)

// snapshotsSyncer syncs snapshot stats periodically. imagefs info and container stats
// should both use cached result here.
// TODO(random-liu): Benchmark with high workload. We may need a statsSyncer instead if
// benchmark result shows that container cpu/memory stats also need to be cached.
type snapshotsSyncer struct {
	store           map[string]*snapshotstore.Store
	snapshotter     map[string]snapshot.Snapshotter
	snapshotterList []string
	syncPeriod      time.Duration
}

// newSnapshotsSyncer creates a snapshot syncer.
func newSnapshotsSyncer(store map[string]*snapshotstore.Store, snapshotterList []string, snapshotter map[string]snapshot.Snapshotter,
	period time.Duration) *snapshotsSyncer {
	return &snapshotsSyncer{
		store:           store,
		snapshotter:     snapshotter,
		snapshotterList: snapshotterList,
		syncPeriod:      period,
	}
}

// start starts the snapshots syncer. No stop function is needed because
// the syncer doesn't update any persistent states, it's fine to let it
// exit with the process.
func (s *snapshotsSyncer) start() {
	tick := time.NewTicker(s.syncPeriod)
	go func() {
		defer tick.Stop()
		// TODO(random-liu): This is expensive. We should do benchmark to
		// check the resource usage and optimize this.
		for {
			if err := s.sync(); err != nil {
				logrus.WithError(err).Error("Failed to sync snapshot stats")
			}
			<-tick.C
		}
	}()
}

// sync updates all snapshots stats.
func (s *snapshotsSyncer) sync() error {
	for _, snName := range s.snapshotterList {
		ctx := ctrdutil.NamespacedContext()
		start := time.Now().UnixNano()
		var snapshots []snapshot.Info
		// Do not call `Usage` directly in collect function, because
		// `Usage` takes time, we don't want `Walk` to hold read lock
		// of snapshot metadata store for too long time.
		// TODO(random-liu): Set timeout for the following 2 contexts.
		if err := s.snapshotter[snName].Walk(ctx, func(ctx context.Context, info snapshot.Info) error {
			snapshots = append(snapshots, info)
			return nil
		}); err != nil {
			return fmt.Errorf("walk all snapshots failed: %w", err)
		}
		for _, info := range snapshots {
			sn, err := s.store[snName].Get(info.Name)
			if err == nil {
				// Only update timestamp for non-active snapshot.
				if sn.Kind == info.Kind && sn.Kind != snapshot.KindActive {
					sn.Timestamp = time.Now().UnixNano()
					s.store[snName].Add(sn)
					continue
				}
			}
			// Get newest stats if the snapshot is new or active.
			sn = snapshotstore.Snapshot{
				Key:       info.Name,
				Kind:      info.Kind,
				Timestamp: time.Now().UnixNano(),
				RootPath:  "",
			}
			usage, err := s.snapshotter[snName].Usage(ctx, info.Name)
			if err != nil {
				if !errdefs.IsNotFound(err) {
					logrus.WithError(err).Errorf("Failed to get usage for snapshot %q", info.Name)
				}
				continue
			}
			sn.Size = uint64(usage.Size)
			sn.Inodes = uint64(usage.Inodes)
			if sn.RootPath == "" && info.Kind == snapshot.KindActive {
				snInfo, err := s.snapshotter[snName].Stat(ctx, info.Name)
				if err != nil || snInfo.Labels == nil {
					if !errdefs.IsNotFound(err) {
						logrus.WithError(err).Debugf("Failed to get stat for snapshot %q", info.Name)
					}
					continue
				}
				if rootPath, ok := snInfo.Labels["RootPath"]; ok {
					sn.RootPath = rootPath
				}
			}
			if sn.RootPath == "" {
				log.G(ctx).Warnf("Snapshot %v has no root path, WritableLayer field in cri will be skipped", info.Name)
			}
			s.store[snName].Add(sn)
		}
		for _, sn := range s.store[snName].List() {
			if sn.Timestamp >= start {
				continue
			}
			// Delete the snapshot stats if it's not updated this time.
			s.store[snName].Delete(sn.Key)
		}
	}
	return nil
}
