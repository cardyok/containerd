package imagegcplugin

import (
	"context"
	"time"
)

// Status represents status of image gc plugin, but it is not real handler.
type Status interface {
	// Enabled indicates we should run image gc or not.
	Enabled() bool
	// Config returns config of plugin.
	Config() Config
}

// GarbageCollect handles image garbage collect.
type GarbageCollect interface {
	// GarbageCollect starts to handle garbage collect.
	GarbageCollect(ctx context.Context) error
}

// CRIRuntime is wrapper level for CRI-API as util interface.
type CRIRuntime interface {
	ListImages(ctx context.Context) ([]Image, error)
	RemoveImage(ctx context.Context, idOrRef string) error
	ListContainers(ctx context.Context) ([]Container, error)
}

// Image represents image from CRI-API.
type Image struct {
	ID          string
	RepoTags    []string
	RepoDigests []string
	SizeBytes   uint64
}

// Container represents container from CRI-API.
type Container struct {
	ID      string
	Name    string
	Image   string
	ImageID string
}

// fsStats contains data about filesystem usage.
type fsStats struct {
	// availableBytes represents the storage space available bytes of filesystem.
	availableBytes uint64
	// capacityBytes represents the total capacity bytes of the filesystems.
	capacityBytes uint64
	// inodesFree represents the free inodes in the filesystem.
	inodesFree uint64
	// inodes represents the total inodes in the filesystem.
	inodes uint64
}

// GcPolicy is used to make decision when to garbage collect.
type GcPolicy struct {
	// Allow to higher than lowThresholdPercent.
	LowThresholdPercent int
	// Any usage below highThresholdPercent will never triger garbage collect.
	HighThresholdPercent int
	// MinAge is minimum age at which an image can be garbage collected.
	MinAge time.Duration
	// Whitelist is to keep images which can't be removed.
	Whitelist []string
	// WhitelistGoRegex is to keep images matched by regex.
	WhitelistGoRegex string
}

// imageRecord is simple record for image.
type imageRecord struct {
	// image ID
	id string
	// image tags
	tags []string
	// image digests
	digests []string
	// detected represents when the image is seen
	detected time.Time
	// lastUsed represents when the image is seen
	lastUsed time.Time
	// sizeBytes precedence image size in bytes
	sizeBytes uint64
}

type byLastUsedAndDetected []imageRecord

func (imgs byLastUsedAndDetected) Len() int {
	return len(imgs)
}

func (imgs byLastUsedAndDetected) Swap(i, j int) {
	imgs[i], imgs[j] = imgs[j], imgs[i]
}

func (imgs byLastUsedAndDetected) Less(i, j int) bool {
	if imgs[i].lastUsed.Equal(imgs[j].lastUsed) {
		return imgs[i].detected.Before(imgs[j].detected)
	}
	return imgs[i].lastUsed.Before(imgs[j].lastUsed)
}
