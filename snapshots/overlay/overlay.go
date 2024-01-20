//go:build linux
// +build linux

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

package overlay

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/containerd/continuity/fs"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/snapshots"
	"github.com/containerd/containerd/snapshots/overlay/overlaybd"
	"github.com/containerd/containerd/snapshots/overlay/overlayutils"
	"github.com/containerd/containerd/snapshots/overlay/quota"
	quotatypes "github.com/containerd/containerd/snapshots/overlay/quota/types"
	"github.com/containerd/containerd/snapshots/overlay/roDriver"
	"github.com/containerd/containerd/snapshots/storage"
)

// upperdirKey is a key of an optional lablel to each snapshot.
// This optional label of a snapshot contains the location of "upperdir" where
// the change set between this snapshot and its parent is stored.
const upperdirKey = "containerd.io/snapshot/overlay.upperdir"

// volatileOpt is a key of an optional lablel to each snapshot.
// If this optional label of a snapshot is specified, when mounted to rootdir
// this snapshot will include volatile option
const volatileOpt = "containerd.io/snapshot/overlay.volatile"

// activePath is a key of an optional label to active snapshot.
// If this label is specified, content of this active snapshot(and subsequent rootfs write)
// will be stored in the specified path.
const activePath = "containerd.io/snapshot/overlay.active.path"

// SnapshotterLabelOverlayActivePath is a key of an optional label to active snapshot.
// If this label is specified, content of this active snapshot(and subsequent rootfs write)
// will be stored in the specified path.
const SnapshotterLabelOverlayActivePath = "containerd.io/snapshot.overlay.active-path"

// SnapshotterLabelOverlayActiveQuota is a key of an optional label to active snapshot.
// If this label is specified, content of this active snapshot(and subsequent rootfs write)
// will be set file system size.
const SnapshotterLabelOverlayActiveQuota = "containerd.io/snapshot.overlay.active-quota"

// MaxActiveQuota is defined the max usage quota of active layer.
const MaxActiveQuota = 64 * 1024 * 1024 * 1024 * 1024

// MinActiveQuota is defined the min usage quota of active layer.
const MinActiveQuota = 1024 * 1024 * 1024

const SandBoxMetaFile = "pod_sandbox_meta"

const labelSnapshotRef = "containerd.io/snapshot.ref"

// SnapshotterConfig is used to configure the overlay snapshotter instance
type SnapshotterConfig struct {
	asyncRemove     bool
	upperdirLabel   bool
	quotaDriver     string
	defaultUpperDir string
}

// Opt is an option to configure the overlay snapshotter
type Opt func(config *SnapshotterConfig) error

// AsynchronousRemove defers removal of filesystem content until
// the Cleanup method is called. Removals will make the snapshot
// referred to by the key unavailable and make the key immediately
// available for re-use.
func AsynchronousRemove(config *SnapshotterConfig) error {
	config.asyncRemove = true
	return nil
}

// WithUpperdirLabel adds as an optional label
// "containerd.io/snapshot/overlay.upperdir". This stores the location
// of the upperdir that contains the changeset between the labelled
// snapshot and its parent.
func WithUpperdirLabel(config *SnapshotterConfig) error {
	config.upperdirLabel = true
	return nil
}

type snapshotter struct {
	root            string
	ms              *storage.MetaStore
	asyncRemove     bool
	upperdirLabel   bool
	defaultUpperDir string
	indexOff        bool
	userxattr       bool // whether to enable "userxattr" mount option
	quotaDriver     quotatypes.Quota
	roDriver        roDriver.RoDriver
}

func WithQuotaDriver(driver string) Opt {
	return func(config *SnapshotterConfig) error {
		config.quotaDriver = driver
		return nil
	}
}

// WithDefaultUpperDir adds as a global default upper dir for active layers
func WithDefaultUpperDir(upperDir string) Opt {
	return func(config *SnapshotterConfig) error {
		config.defaultUpperDir = upperDir
		return nil
	}
}

// NewSnapshotter returns a Snapshotter which uses overlayfs. The overlayfs
// diffs are stored under the provided root. A metadata file is stored under
// the root.
func NewSnapshotter(root string, opts ...Opt) (snapshots.Snapshotter, error) {
	var config SnapshotterConfig
	for _, opt := range opts {
		if err := opt(&config); err != nil {
			return nil, err
		}
	}

	if config.defaultUpperDir == "" {
		config.defaultUpperDir = root
	}

	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(config.defaultUpperDir, 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}
	supportsDType, err := fs.SupportsDType(root)
	if err != nil {
		return nil, err
	}
	if !supportsDType {
		return nil, fmt.Errorf("%s does not support d_type. If the backing filesystem is xfs, please reformat with ftype=1 to enable d_type support", root)
	}
	ms, err := storage.NewMetaStore(filepath.Join(root, "metadata.db"))
	if err != nil {
		return nil, err
	}

	if err := os.Mkdir(filepath.Join(root, "snapshots"), 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}
	if err := os.Mkdir(filepath.Join(config.defaultUpperDir, "snapshots"), 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}
	// figure out whether "userxattr" option is recognized by the kernel && needed
	userxattr, err := overlayutils.NeedsUserXAttr(root)
	if err != nil {
		logrus.WithError(err).Warnf("cannot detect whether \"userxattr\" option needs to be used, assuming to be %v", userxattr)
	}

	// init upper layer quota driver
	quotaDriver := quota.New(config.quotaDriver, nil)

	roDriverInit, err := overlaybd.New()
	if err != nil {
		return nil, fmt.Errorf("failed to start overlaybd driver: %v", err)
	}

	return &snapshotter{
		root:            root,
		ms:              ms,
		asyncRemove:     config.asyncRemove,
		upperdirLabel:   config.upperdirLabel,
		defaultUpperDir: config.defaultUpperDir,
		indexOff:        supportsIndex(),
		userxattr:       userxattr,
		quotaDriver:     quotaDriver,
		roDriver:        roDriverInit,
	}, nil
}

// Stat returns the info for an active or committed snapshot by name or
// key.
//
// Should be used for parent resolution, existence checks and to discern
// the kind of snapshot.
func (o *snapshotter) Stat(ctx context.Context, key string) (snapshots.Info, error) {
	ctx, t, err := o.ms.TransactionContext(ctx, false)
	if err != nil {
		return snapshots.Info{}, err
	}
	defer t.Rollback()
	id, info, size, err := storage.GetInfo(ctx, key)
	if err != nil {
		return snapshots.Info{}, err
	}
	if info.Labels == nil {
		info.Labels = map[string]string{}
	}
	info.Labels["Backend-id"] = id
	info.Labels["Backend-inode"] = strconv.FormatInt(size.Inodes, 10)
	info.Labels["Backend-size"] = ByteCountDecimal(size.Size)
	info.Labels["RootPath"] = o.upperPath(&info, id, key)

	if o.upperdirLabel {
		if info.Labels == nil {
			info.Labels = make(map[string]string)
		}
		info.Labels[upperdirKey] = o.upperPath(&info, id, key)
	}

	return info, nil
}

func (o *snapshotter) Update(ctx context.Context, info snapshots.Info, fieldpaths ...string) (snapshots.Info, error) {
	ctx, t, err := o.ms.TransactionContext(ctx, true)
	if err != nil {
		return snapshots.Info{}, err
	}

	info, err = storage.UpdateInfo(ctx, info, fieldpaths...)
	if err != nil {
		t.Rollback()
		return snapshots.Info{}, err
	}

	if o.upperdirLabel {
		id, _, _, err := storage.GetInfo(ctx, info.Name)
		if err != nil {
			return snapshots.Info{}, err
		}
		if info.Labels == nil {
			info.Labels = make(map[string]string)
		}
		info.Labels[upperdirKey] = o.upperPath(&info, id, info.Name)
	}

	if err := t.Commit(); err != nil {
		return snapshots.Info{}, err
	}

	return info, nil
}

// Usage returns the resources taken by the snapshot identified by key.
//
// For active snapshots, this will scan the usage of the overlay "diff" (aka
// "upper") directory and may take some time.
//
// For committed snapshots, the value is returned from the metadata database.
func (o *snapshotter) Usage(ctx context.Context, key string) (snapshots.Usage, error) {
	ctx, t, err := o.ms.TransactionContext(ctx, false)
	if err != nil {
		return snapshots.Usage{}, err
	}
	id, info, usage, err := storage.GetInfo(ctx, key)
	t.Rollback() // transaction no longer needed at this point.

	if err != nil {
		return snapshots.Usage{}, err
	}

	if info.Kind == snapshots.KindActive {
		upperPath := o.upperPath(&info, id, key)
		du, err := fs.DiskUsage(ctx, upperPath)
		if err != nil {
			// TODO(stevvooe): Consider not reporting an error in this case.
			return snapshots.Usage{}, err
		}

		usage = snapshots.Usage(du)
	}

	return usage, nil
}

func (o *snapshotter) Prepare(ctx context.Context, key, parent string, opts ...snapshots.Opt) (_ []mount.Mount, retErr error) {
	ctx, t, err := o.ms.TransactionContext(ctx, true)
	if err != nil {
		return nil, err
	}
	var base snapshots.Info
	for _, opt := range opts {
		if err := opt(&base); err != nil {
			return nil, err
		}
	}
	log.G(ctx).Infof("prepare for layer %v with parent %v labels %v", key, parent, base.Labels)
	s, parentDir, mounts, err := o.prepareLower(ctx, snapshots.KindActive, key, parent, false, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare: %w", err)
	}
	snapshotDir := o.fsPath(&base, s.ID, key)
	skipFetch, err := o.roDriver.PreProcess(ctx, snapshotDir, parentDir, parent, base.Labels)
	if err != nil {
		t.Rollback()
		return nil, fmt.Errorf("roDriver failed preparing layer %v: %w", key, err)
	}
	defer func() {
		if retErr != nil {
			t.Rollback()
		}
	}()
	if skipFetch {
		name := base.Labels[labelSnapshotRef]
		if err := o.Commit(ctx, name, key, opts...); err != nil {
			return nil, fmt.Errorf("failed to commit layer %v: %w", key, err)
		}
		// skip fetch means the layer is on demand layer
		return nil, errdefs.ErrAlreadyExists
	}
	if err = t.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit layer: %w", err)
	}
	log.G(ctx).Infof("prepare for layer %v gives mount %v, skip fetch: %v", key, mounts, skipFetch)
	return mounts, err
}

func (o *snapshotter) Active(ctx context.Context, key, parent string, opts ...snapshots.Opt) (_ []mount.Mount, retErr error) {
	ctx, t, err := o.ms.TransactionContext(ctx, true)
	if err != nil {
		return nil, err
	}
	if parent == "" {
		return nil, fmt.Errorf("active layer parent cannot be nil")
	}
	opts = append(opts, snapshots.WithLabels(map[string]string{"rwlayer": "true"}))

	var base snapshots.Info
	for _, opt := range opts {
		if err := opt(&base); err != nil {
			return nil, fmt.Errorf("failed to apply options: %w", err)
		}
	}

	s, _, lowerMount, err := o.prepareLower(ctx, snapshots.KindActive, key, parent, true, opts)
	if err != nil {
		t.Rollback()
		return lowerMount, err
	}
	defer func() {
		if retErr != nil {
			t.Rollback()
		}
	}()
	upperDir := o.upperPath(&base, s.ID, key)
	lowerDir := o.lowerPath(&base, s.ID, key)
	fsDir := filepath.Join(upperDir, "fs")
	workDir := filepath.Join(upperDir, "work")
	if err := o.prepareUpperDir(ctx, upperDir, fsDir, workDir, base); err != nil {
		o.roDriver.Cleanup(ctx, s.ID)
		return nil, fmt.Errorf("failed to prepare upper dir: %w", err)
	}
	if err := os.MkdirAll(lowerDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to prepare lower dir: %w", err)
	}
	if err := mount.All(lowerMount, lowerDir); err != nil {
		o.roDriver.Cleanup(ctx, s.ID)
		return nil, fmt.Errorf("failed to mount lower dir with %v: %w", lowerMount, err)
	}

	if err = t.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit layer: %w", err)
	}
	options := []string{"lowerdir=" + lowerDir, "upperdir=" + fsDir, "workdir=" + workDir}
	// set index=off when mount overlayfs
	if o.indexOff {
		options = append(options, "index=off")
	}

	if o.userxattr {
		options = append(options, "userxattr")
	}
	return []mount.Mount{
		{
			Source:  "overlay",
			Type:    "overlay",
			Options: options,
		},
	}, nil
}

func (o *snapshotter) View(ctx context.Context, key, parent string, opts ...snapshots.Opt) ([]mount.Mount, error) {
	opts = append(opts, snapshots.WithLabels(map[string]string{"rwlayer": "true"}))
	_, _, mounts, err := o.prepareLower(ctx, snapshots.KindView, key, parent, true, opts)
	return mounts, err
}

func (o *snapshotter) prepareUpperDir(ctx context.Context, target, fsDir, workDir string, base snapshots.Info) error {
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("failed to create upper dir: %w", err)
	}
	if err := os.MkdirAll(fsDir, 0755); err != nil {
		return fmt.Errorf("failed to create fs dir: %w", err)
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create work dir: %w", err)
	}

	activeQuota, err := o.getActiveQuota(&base)
	if err != nil && !errdefs.IsNotFound(err) {
		log.G(ctx).WithError(err).WithField("snapshotter", base.Name).Warn("activeQuota specified invalid")
	}
	if o.quotaDriver != nil && activeQuota > 0 {
		opts := map[string]string{
			"base":   target,
			"rwFlag": "rw",
		}
		err := o.quotaDriver.Setup(ctx, target, activeQuota, opts)
		if err != nil {
			log.G(ctx).WithError(err).Warn("failed to prepare quota size")
		}
	}

	if err := os.MkdirAll(fsDir, 0755); err != nil {
		return fmt.Errorf("failed to create fs dir: %w", err)
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create work dir: %w", err)
	}
	return nil
}

func (o *snapshotter) prepareLower(ctx context.Context, kind snapshots.Kind, key, parent string, isActive bool, opts []snapshots.Opt) (storage.Snapshot, string, []mount.Mount, error) {
	var base snapshots.Info
	for _, opt := range opts {
		if err := opt(&base); err != nil {
			return storage.Snapshot{}, "", nil, fmt.Errorf("failed to apply options: %w", err)
		}
	}
	var snapshotDir string
	if isActive {
		snapshotDir = filepath.Join(o.defaultUpperDir, "snapshots")
	} else {
		snapshotDir = filepath.Join(o.root, "snapshots")
	}

	if homePath, err := getActivePath(&base, key); err == nil {
		if _, err := os.Stat(homePath); err == nil {
			if err = os.RemoveAll(homePath); err != nil {
				logrus.WithError(err).Errorf("failed to cleanup %s created by last time", homePath)
			}
		}
		if err = os.MkdirAll(homePath, 0711); err != nil {
			return storage.Snapshot{}, "", nil, err
		}
		snapshotDir = homePath
	} else {
		if !errdefs.IsNotFound(err) {
			log.G(ctx).WithError(err).Warn("activePath specified invalid")
		}
	}
	td, err := os.MkdirTemp(snapshotDir, "new-")
	if err != nil {
		return storage.Snapshot{}, "", nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	var path string
	defer func() {
		if err != nil {
			if td != "" {
				if err1 := os.RemoveAll(td); err1 != nil {
					log.G(ctx).WithError(err1).Warn("failed to cleanup temp snapshot directory")
				}
			}
			if path != "" {
				if err1 := os.RemoveAll(path); err1 != nil {
					log.G(ctx).WithError(err1).WithField("path", path).Error("failed to reclaim snapshot directory, directory may need removal")
					err = fmt.Errorf("failed to remove path: %v: %w", err1, err)
				}
			}
		}
	}()

	var mounts []mount.Mount
	// 1. create snapshot metadata
	s, err := storage.CreateSnapshot(ctx, kind, key, parent, opts...)
	if err != nil {
		return storage.Snapshot{}, "", nil, fmt.Errorf("failed to create snapshot: %w", err)
	}
	path = o.fsPath(&base, s.ID, key)
	// 2. get parents
	parentDir := ""
	if parent != "" {
		pID, pInfo, _, err := storage.GetInfo(ctx, parent)
		if err != nil {
			return storage.Snapshot{}, "", nil, fmt.Errorf("failed to get parent snapshot: %w", err)
		}
		parentDir = o.fsPath(&pInfo, pID, parent)
	}

	if err = os.Rename(td, path); err != nil {
		return storage.Snapshot{}, "", nil, fmt.Errorf("failed to rename: %w", err)
	}
	if isActive {
		if data, ok := base.Labels["PodSandboxMetadata"]; ok {
			if err := ioutil.WriteFile(filepath.Join(path, SandBoxMetaFile), []byte(data), 0644); err != nil {
				log.G(ctx).Errorf("write sandbox meta failed. path: %s, err: %s", filepath.Join(path, SandBoxMetaFile), err.Error())
			}
		}
		mounts, err = o.roDriver.ActiveMount(ctx, path, s.ID, parentDir, o.idToDirectory(s.ParentIDs))
		if err != nil {
			return storage.Snapshot{}, "", nil, fmt.Errorf("failed to prepare active lower dir mount: %w", err)
		}
	} else {
		mounts, err = o.roDriver.PrepareMount(ctx, path, nil)
		if err != nil {
			return storage.Snapshot{}, "", nil, fmt.Errorf("failed to prepare readable lower dir mount: %w", err)
		}
	}
	defer func() {
		if err != nil {
			o.roDriver.Cleanup(ctx, s.ID)
		}
	}()
	parentPath := ""
	if parent != "" {
		if pID, pInfo, _, err := storage.GetInfo(ctx, parent); err == nil {
			parentPath = o.fsPath(&pInfo, pID, parent)
		}
	}
	return s, parentPath, mounts, nil
}

func (o *snapshotter) idToDirectory(ids []string) []string {
	ret := []string{}
	for _, v := range ids {
		ret = append(ret, filepath.Join(o.root, "snapshots", v))
	}
	return ret
}

// Mounts returns the mounts for the transaction identified by key. Can be
// called on an read-write or readonly transaction.
//
// This can be used to recover mounts after calling View or Prepare.
func (o *snapshotter) Mounts(ctx context.Context, key string) ([]mount.Mount, error) {
	ctx, t, err := o.ms.TransactionContext(ctx, false)
	if err != nil {
		return nil, err
	}
	_, info, _, err := storage.GetInfo(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get info for snapshot %s: %w", key, err)
	}
	s, err := storage.GetSnapshot(ctx, key)
	t.Rollback()
	if err != nil {
		return nil, fmt.Errorf("failed to get active mount: %w", err)
	}
	snapshotDir := o.fsPath(&info, s.ID, key)
	upperDir := o.upperPath(&info, s.ID, key)
	lowerDir := o.lowerPath(&info, s.ID, key)
	fsDir := filepath.Join(upperDir, "fs")
	workDir := filepath.Join(upperDir, "work")
	if _, err = os.Stat(upperDir); err == nil {
		options := []string{"lowerdir=" + lowerDir, "upperdir=" + fsDir, "workdir=" + workDir}
		// set index=off when mount overlayfs
		if o.indexOff {
			options = append(options, "index=off")
		}

		if o.userxattr {
			options = append(options, "userxattr")
		}
		return []mount.Mount{
			{
				Source:  "overlay",
				Type:    "overlay",
				Options: options,
			},
		}, nil
	}
	return o.roDriver.GetMount(ctx, snapshotDir)
}

func (o *snapshotter) Commit(ctx context.Context, name, key string, opts ...snapshots.Opt) error {
	ctx, t, err := o.ms.TransactionContext(ctx, true)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rerr := t.Rollback(); rerr != nil {
				log.G(ctx).WithError(rerr).Warn("failed to rollback transaction")
			}
		}
	}()

	// grab the existing id
	id, info, _, err := storage.GetInfo(ctx, key)
	if err != nil {
		return err
	}

	log.G(ctx).Infof("commit for layer %v  to %v ", key, name)
	usage, err := fs.DiskUsage(ctx, o.fsPath(&info, id, key))
	if err != nil {
		return err
	}
	if err := o.roDriver.Commit(ctx, o.fsPath(&info, id, key)); err != nil {
		log.G(ctx).Infof("commit for layer %v failed %v", key, err)
		return fmt.Errorf("failed to commit active mount: %w", err)
	}

	if _, err = storage.CommitActive(ctx, key, name, snapshots.Usage(usage), opts...); err != nil {
		return fmt.Errorf("failed to commit snapshot: %w", err)
	}
	return t.Commit()
}

// Remove abandons the snapshot identified by key. The snapshot will
// immediately become unavailable and unrecoverable. Disk space will
// be freed up on the next call to `Cleanup`.
func (o *snapshotter) Remove(ctx context.Context, key string) (err error) {
	ctx, t, err := o.ms.TransactionContext(ctx, true)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if rerr := t.Rollback(); rerr != nil {
				log.G(ctx).WithError(rerr).Warn("failed to rollback transaction")
			}
		}
	}()

	id, info, _, err := storage.GetInfo(ctx, key)
	if err != nil {
		if strings.Contains(err.Error(), "snapshot does not exist") {
			if rerr := t.Rollback(); rerr != nil {
				log.G(ctx).WithError(rerr).Warn("failed to rollback transaction")
			}
			return nil
		}
		return fmt.Errorf("failed to get snapshot info: %w", err)
	}

	upperDir := o.upperPath(&info, id, key)
	lowerDir := o.lowerPath(&info, id, key)
	// If lowerDir found, try to umount it
	if _, err := os.Stat(lowerDir); err == nil {
		if err := mount.Unmount(lowerDir, 0); err != nil {
			return fmt.Errorf("failed to umount lower dir: %w", err)
		}
	}
	// Always try to ask roDriver to clean up its resource.
	if err := o.roDriver.Cleanup(ctx, id); err != nil {
		return fmt.Errorf("failed to cleanup roDriver mount: %w", err)
	}
	activeQuota, err := o.getActiveQuota(&info)
	if err != nil && !errdefs.IsNotFound(err) {
		log.G(ctx).WithError(err).Warn("activeQuota specified invalid")
	}
	if o.quotaDriver != nil && activeQuota > 0 {
		err := o.quotaDriver.Remove(ctx, upperDir)
		if err != nil {
			log.G(ctx).WithError(err).Warn("failed to remove active quota")
			return fmt.Errorf("failed to remove active quota: %w", err)
		}
	}

	var home string
	if home, err = getActivePath(&info, key); err == nil {
		if err = removeActivePath(&info, key); err != nil {
			return fmt.Errorf("failed to remove directory %v: %w", home, err)
		}
	}

	_, _, err = storage.Remove(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to remove: %w", err)
	}

	if !o.asyncRemove {
		var removals []string
		removals, err = o.getCleanupDirectories(ctx, t)
		if err != nil {
			return fmt.Errorf("unable to get directories for removal: %w", err)
		}

		// Remove directories after the transaction is closed, failures must not
		// return error since the transaction is committed with the removal
		// key no longer available.
		defer func() {
			if err == nil {
				for _, dir := range removals {
					if err := os.RemoveAll(dir); err != nil {
						log.G(ctx).WithError(err).WithField("path", dir).Warn("failed to remove directory")
					}
				}
			}
		}()

	}

	return t.Commit()
}

// Walk the snapshots.
func (o *snapshotter) Walk(ctx context.Context, fn snapshots.WalkFunc, fs ...string) error {
	ctx, t, err := o.ms.TransactionContext(ctx, false)
	if err != nil {
		return err
	}
	defer t.Rollback()
	if o.upperdirLabel {
		return storage.WalkInfo(ctx, func(ctx context.Context, info snapshots.Info) error {
			id, _, _, err := storage.GetInfo(ctx, info.Name)
			if err != nil {
				return err
			}
			if info.Labels == nil {
				info.Labels = make(map[string]string)
			}
			info.Labels[upperdirKey] = o.upperPath(&info, id, info.Name)
			return fn(ctx, info)
		}, fs...)
	}
	return storage.WalkInfo(ctx, fn, fs...)
}

// Cleanup cleans up disk resources from removed or abandoned snapshots
func (o *snapshotter) Cleanup(ctx context.Context) error {
	cleanup, err := o.cleanupDirectories(ctx)
	if err != nil {
		return err
	}

	for _, dir := range cleanup {
		if err := os.RemoveAll(dir); err != nil {
			log.G(ctx).WithError(err).WithField("path", dir).Warn("failed to remove directory")
		}
	}

	return nil
}

func (o *snapshotter) cleanupDirectories(ctx context.Context) ([]string, error) {
	// Get a write transaction to ensure no other write transaction can be entered
	// while the cleanup is scanning.
	ctx, t, err := o.ms.TransactionContext(ctx, true)
	if err != nil {
		return nil, err
	}

	defer t.Rollback()
	return o.getCleanupDirectories(ctx, t)
}

func (o *snapshotter) getCleanupDirectories(ctx context.Context, t storage.Transactor) ([]string, error) {
	ids, err := storage.IDMap(ctx)
	if err != nil {
		return nil, err
	}

	snapshotDir := filepath.Join(o.root, "snapshots")
	fd, err := os.Open(snapshotDir)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	upperDir := filepath.Join(o.defaultUpperDir, "snapshots")
	fdu, uerr := os.Open(upperDir)
	if uerr != nil {
		return nil, uerr
	}
	defer fdu.Close()
	udirs, uerr := fdu.Readdirnames(0)
	if uerr != nil {
		return nil, uerr
	}

	dirs, err := fd.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	cleanup := []string{}
	for _, d := range dirs {
		if _, ok := ids[d]; ok {
			continue
		}

		cleanup = append(cleanup, filepath.Join(snapshotDir, d))
	}
	for _, d := range udirs {
		if _, ok := ids[d]; ok {
			continue
		}

		cleanup = append(cleanup, filepath.Join(upperDir, d))
	}
	log.G(ctx).Infof("cleaning up directories: %v", cleanup)

	return cleanup, nil
}

// getActivePath tries to obtail specified path for Active layer in annotation
func getActivePath(info *snapshots.Info, key string) (string, error) {
	if info.Labels != nil {
		if home, ok := info.Labels[activePath]; ok {
			if !filepath.IsAbs(home) {
				return "", fmt.Errorf("path for active layer must be an absolute path: %w", errdefs.ErrInvalidArgument)
			}
			if key == "" {
				return filepath.Join(home, ".rwlayer", info.Created.String()), nil
			}
			return filepath.Join(home, ".rwlayer", key), nil
		}
	}
	return "", fmt.Errorf("active snapshot path is not specified: %w", errdefs.ErrNotFound)
}

// removeActivePath removes the path created as rwlayer in specified path
func removeActivePath(info *snapshots.Info, key string) error {
	var removePath string

	if info.Labels == nil {
		return nil
	}

	if home, ok := info.Labels[activePath]; ok {
		// If user specified the Active Path label but didn't specify value, similar to logic in getActivePath
		// path indicated by creation timestamp should be removed
		if key == "" {
			removePath = filepath.Join(home, ".rwlayer", info.Created.String())
			return os.RemoveAll(removePath)
		}

		keyList := append([]string{".rwlayer"}, strings.Split(key, "/")...)
		if err := os.RemoveAll(filepath.Join(home, ".rwlayer", key)); err != nil {
			return err
		}

		for i := len(keyList) - 1; i > 0; i-- {
			dirPath := strings.Join(keyList[0:i], "/")
			if dir, _ := ioutil.ReadDir(filepath.Join(home, dirPath)); len(dir) == 0 {
				os.RemoveAll(filepath.Join(home, dirPath))
			}
		}
	}
	return nil
}

// getActiveQuota get the usage quota of active layer.
func (o *snapshotter) getActiveQuota(info *snapshots.Info) (int, error) {
	if info.Labels == nil {
		return -1, fmt.Errorf("snapshot label is nil: %w", errdefs.ErrNotFound)
	}

	quota, ok := info.Labels[SnapshotterLabelOverlayActiveQuota]
	if !ok {
		return -1, fmt.Errorf("active snapshot quota is not specified: %w", errdefs.ErrNotFound)
	}

	parse := resource.MustParse(quota)
	s := int(parse.Value())

	if s < MinActiveQuota || s > MaxActiveQuota {
		return -1, fmt.Errorf("active snapshot quota is invalid[%d, %d]: %d", MinActiveQuota, MaxActiveQuota, s)
	}

	return s, nil
}

func (o *snapshotter) fsPath(info *snapshots.Info, id string, key string) string {
	if info != nil {
		if home, err := getActivePath(info, key); err == nil {
			// return path combined by home
			return filepath.Join(home, id)
		}
		if _, ok := info.Labels["rwlayer"]; ok {
			return filepath.Join(o.defaultUpperDir, "snapshots", id)
		}
	}
	return filepath.Join(o.root, "snapshots", id)
}

func (o *snapshotter) upperPath(info *snapshots.Info, id string, key string) string {
	if info != nil {
		if home, err := getActivePath(info, key); err == nil {
			// return path combined by home
			return filepath.Join(home, id, "upper")
		}
		if _, ok := info.Labels["rwlayer"]; ok {
			return filepath.Join(o.defaultUpperDir, "snapshots", id, "upper")
		}
	}
	return filepath.Join(o.root, "snapshots", id, "upper")
}

func (o *snapshotter) lowerPath(info *snapshots.Info, id string, key string) string {
	if info != nil {
		if home, err := getActivePath(info, key); err == nil {
			// return path combined by home
			return filepath.Join(home, id, "lower")
		}
		if _, ok := info.Labels["rwlayer"]; ok {
			return filepath.Join(o.defaultUpperDir, "snapshots", id, "lower")
		}
	}
	return filepath.Join(o.root, "snapshots", id, "lower")
}

// Close closes the snapshotter
func (o *snapshotter) Close() error {
	return o.ms.Close()
}

// supportsIndex checks whether the "index=off" option is supported by the kernel.
func supportsIndex() bool {
	if _, err := os.Stat("/sys/module/overlay/parameters/index"); err == nil {
		return true
	}
	return false
}

func ByteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
