package roDriver

import (
	"context"

	"github.com/containerd/containerd/mount"
)

type Opt func()

type RoDriver interface {
	// ActiveMount is used for preparing mount struct used for running process like container rootfs
	ActiveMount(ctx context.Context, keyDir string, id string, parentDir string, parentPaths []string, opts ...Opt) ([]mount.Mount, error)
	// PrepareMount is used for preparing mount struct used for preparing content like applying layers
	PrepareMount(ctx context.Context, keyDir string, parents []string, opts ...Opt) ([]mount.Mount, error)
	// Cleanup is used for cleaning up mount
	Cleanup(ctx context.Context, id string) error
	// GetMount is used for getting mount for current directory
	GetMount(ctx context.Context, keyDir string) ([]mount.Mount, error)
	// Commit is used for committing a ro layer
	Commit(ctx context.Context, keyDir string) error
	// PreProcess is used for preprocessing layer and judge skip fetch
	PreProcess(ctx context.Context, keyDir, parentDir string, parent string, labels map[string]string) (skipFetch bool, err error)
}
