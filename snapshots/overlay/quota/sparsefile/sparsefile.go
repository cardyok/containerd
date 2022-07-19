package sparsefile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/fscrypt/filesystem"

	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/snapshots/overlay/quota/types"
)

const (
	QuotaName              = "sparsefile"
	defaultFsType          = "ext4"
	defaultSparsefileQuota = 10 * 1024 * 1024 * 1024
	sparseFileName         = "rw.img"
)

type sparseFileQuota struct {
	fsType       string
	defaultQuota int
}

func (q *sparseFileQuota) Prepare(ctx context.Context, target string, opts map[string]string) error {
	return nil
}

func (q *sparseFileQuota) Setup(ctx context.Context, target string, size int, opts map[string]string) error {
	location := opts["base"]
	if location == "" {
		location = filepath.Base(target)
	}

	rwFlag := opts["rwFlag"]
	if rwFlag == "" {
		rwFlag = "rw"
	}

	if size <= 0 {
		size = q.defaultQuota
	}

	if m, err := filesystem.FindMount(target); err == nil && m.Path == target {
		log.G(ctx).Debugf("get mountpoint %s", m.String())
		return nil
	}

	sparseFile := filepath.Join(location, sparseFileName)
	_, err := os.Lstat(sparseFile)
	if err != nil {
		log.G(ctx).Debugf("prepare sparsefile: %s", sparseFile)
		err := q.createImageFile(sparseFile, size)
		if err != nil {
			return fmt.Errorf("failed to create image file: %s, size %d, err: %w", sparseFile, size, err)
		}
		err = q.formatImageFile(sparseFile)
		if err != nil {
			return fmt.Errorf("failed to format image file: %s, err: %w", sparseFile, err)
		}
	}

	log.G(ctx).Debugf("mount %s to %s", sparseFile, target)
	err = q.mount(sparseFile, target, rwFlag)
	if err != nil {
		return fmt.Errorf("failed to mount image file: %s to %s, err: %w", sparseFile, target, err)
	}

	return nil
}

func (q *sparseFileQuota) Remove(ctx context.Context, target string) error {
	log.G(ctx).Debugf("remove target: %s", target)
	return q.umount(ctx, target)
}

func (q *sparseFileQuota) Get(ctx context.Context, target string) (int error) {
	// TODO:
	return nil
}

func New(config map[string]string) types.Quota {
	fsType := config["fs_type"]
	if fsType == "" {
		fsType = defaultFsType
	}

	defaultQuota, err := strconv.Atoi(config["default_quota"])
	if err != nil || defaultQuota <= 0 {
		defaultQuota = defaultSparsefileQuota
	}

	return &sparseFileQuota{
		fsType:       fsType,
		defaultQuota: defaultQuota,
	}
}
