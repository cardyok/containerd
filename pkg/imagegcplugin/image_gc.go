package imagegcplugin

import (
	"context"
	"os"
	"regexp"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/mount"
)

var errUnexpectedFSCapacity = errors.Errorf("unexpected fs capacity")

type imageGCHandler struct {
	criruntime       CRIRuntime
	policy           GcPolicy
	whitelistGoRegex *regexp.Regexp

	imageFSPath       string
	imageFSMountpoint string

	lock   sync.Mutex
	images map[string]*imageRecord
}

// NewImageGarbageCollect returns GarbageCollect instance.
func NewImageGarbageCollect(criruntime CRIRuntime, policy GcPolicy, imageFSPath map[string]string, snapshotter string) (GarbageCollect, error) {
	var err error
	if err = validateGCPolicy(policy); err != nil {
		return nil, err
	}

	var whitelistGoRegex *regexp.Regexp
	if policy.WhitelistGoRegex != "" {
		whitelistGoRegex, err = regexp.Compile(policy.WhitelistGoRegex)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to compile whitelist go regexp: %s", policy.WhitelistGoRegex)
		}
	}

	info, err := mount.Lookup(imageFSPath[snapshotter])
	if err != nil {
		return nil, errors.Wrap(err, "failed to lookup image fs path in mountinfo")
	}

	return &imageGCHandler{
		criruntime:        criruntime,
		policy:            policy,
		whitelistGoRegex:  whitelistGoRegex,
		imageFSPath:       imageFSPath[snapshotter],
		imageFSMountpoint: info.Mountpoint,
		images:            make(map[string]*imageRecord),
	}, nil
}

func validateGCPolicy(policy GcPolicy) error {
	if policy.HighThresholdPercent <= 0 || policy.HighThresholdPercent > 100 {
		return errors.Errorf("invalid highThresholdPercent %d, must be in range [0-100]", policy.HighThresholdPercent)
	}
	if policy.LowThresholdPercent <= 0 || policy.LowThresholdPercent > 100 {
		return errors.Errorf("invalid LowThresholdPercent %d, must be in range [0-100]", policy.LowThresholdPercent)
	}
	if policy.LowThresholdPercent >= policy.HighThresholdPercent {
		return errors.Errorf("lowThresholdPercent %d can not be >= HighThresholdPercent %d", policy.LowThresholdPercent, policy.HighThresholdPercent)
	}
	if policy.MinAge == 0 {
		return errors.Errorf("image min age can not be zero duration")
	}
	if len(policy.Whitelist) == 0 {
		return errors.Errorf("whitelist can not be empty, at least there is one pause infra image")
	}
	return nil

}

// GarbageCollect removed unused images.
func (handler *imageGCHandler) GarbageCollect(ctx context.Context) error {
	totalSizes, err := handler.calculateFreeSizes(ctx)
	if err != nil {
		return nil
	}

	if totalSizes == 0 {
		log.G(ctx).Infof("no disk pressure on image fs")
		return nil
	}

	_, err = handler.pruneImages(ctx, totalSizes, time.Now())
	return err
}

// calculateFreeSizes returns totalSizes needs to free.
//
// NOTE: nil error and zero value means no need to free.
func (handler *imageGCHandler) calculateFreeSizes(ctx context.Context) (uint64, error) {
	fsstatinfo, err := handler.imageFSStats(ctx)
	if err != nil {
		return 0, err
	}

	if fsstatinfo.capacityBytes == 0 {
		return 0, errors.Wrapf(errUnexpectedFSCapacity, "0 bytes filesystem for image fs path %s", handler.imageFSPath)
	}

	if fsstatinfo.availableBytes > fsstatinfo.capacityBytes {
		return 0, errors.Wrapf(errUnexpectedFSCapacity, "freesize > capacity on filesystem for image fs path %s", handler.imageFSPath)
	}

	usagePercent := 100 - int(fsstatinfo.availableBytes*100/fsstatinfo.capacityBytes)
	if usagePercent < handler.policy.HighThresholdPercent {
		return 0, nil
	}
	return fsstatinfo.capacityBytes*uint64(100-handler.policy.LowThresholdPercent)/100 - fsstatinfo.availableBytes, nil
}

// detectImages will updated detected/lastUsed/sizeBytes for image record and
// return used image ID set.
//
// NOTE: Must be called with Lock to protect map data race.
func (handler *imageGCHandler) detectImages(ctx context.Context, detected time.Time) (stringSet, error) {
	usedImageSet := newStringSet()

	imagesFromCRI, err := handler.criruntime.ListImages(ctx)
	if err != nil {
		return usedImageSet, errors.Wrap(err, "failed to list images from CRI")
	}

	containersFromCRI, err := handler.criruntime.ListContainers(ctx)
	if err != nil {
		return usedImageSet, errors.Wrap(err, "failed to list containers from CRI")
	}

	// init used image id set
	for _, cntr := range containersFromCRI {
		log.G(ctx).Debugf("container %s(id=%s) is using image ID %s", cntr.Name, cntr.ID, cntr.ImageID)
		usedImageSet.inserts(cntr.ImageID)
	}

	now := time.Now()
	currentImages := newStringSet()
	tagDigestToID := make(map[string]string)
	for _, img := range imagesFromCRI {
		log.G(ctx).Debugf("found image ID %s (size=%v bytes) (tags=%v) (digest=%v)", img.ID, img.SizeBytes, img.RepoTags, img.RepoDigests)

		// NOTE: There is no way to have two same image tag but with
		// difference ID and Digest.
		for _, tag := range img.RepoTags {
			if handler.whitelistGoRegex != nil && handler.whitelistGoRegex.MatchString(tag) {
				log.G(ctx).Debugf("image %s matches whitelistGoRegex %s", tag, handler.policy.WhitelistGoRegex)
				usedImageSet.inserts(img.ID)
			}
			tagDigestToID[tag] = img.ID
		}
		for _, digest := range img.RepoDigests {
			tagDigestToID[digest] = img.ID
		}

		currentImages.inserts(img.ID)

		if _, ok := handler.images[img.ID]; !ok {
			log.G(ctx).Debugf("first detected image ID %s", img.ID)
			handler.images[img.ID] = &imageRecord{
				id:       img.ID,
				tags:     img.RepoTags,
				digests:  img.RepoDigests,
				detected: detected,
			}
		}

		if usedImageSet.contains(img.ID) {
			log.G(ctx).Debugf("updated last used time for image ID %s", img.ID)
			handler.images[img.ID].lastUsed = now
		}

		handler.images[img.ID].sizeBytes = img.SizeBytes
	}

	// NOTE:
	// 1. whitelist must contains pause infra image.
	// 2. whitelist can be imageTag, imageDigest, but usedImageSet must be imageID
	for _, protectedImg := range handler.policy.Whitelist {
		if theID, ok := tagDigestToID[protectedImg]; ok {
			usedImageSet.inserts(theID)
		}
	}

	for imgID := range handler.images {
		if !currentImages.contains(imgID) {
			log.G(ctx).Debugf("current image list doesn't contains image ID %s, removing it from image gc plugin", imgID)
			delete(handler.images, imgID)
		}
	}
	return usedImageSet, nil
}

// pruneImages will try to remove images to free totalSize space.
func (handler *imageGCHandler) pruneImages(ctx context.Context, totalSize uint64, gctime time.Time) (uint64, error) {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	// gctime can be used to be detected time
	//
	// NOTE: It is not used to be lastUsed! The lastUsed should be the date
	// after criruntime.ListContainers.
	usedImageSet, err := handler.detectImages(ctx, gctime)
	if err != nil {
		return 0, err
	}

	// if the image is not used right now, add it into removingImages list
	removingImages := make([]imageRecord, 0, len(handler.images))
	for _, img := range handler.images {
		if !usedImageSet.contains(img.id) {
			removingImages = append(removingImages, *img)
		}
	}
	sort.Sort(byLastUsedAndDetected(removingImages))

	var freedSize uint64
	var failedToRemoveImages []string
	for _, img := range removingImages {
		// gctime might be older than lastUsed
		if img.lastUsed.Equal(gctime) || img.lastUsed.After(gctime) {
			log.G(ctx).Infof("Image ID %s has lastUsed=%v which is >= gcTime=%v, keep it", img.id, img.lastUsed, gctime)
			continue
		}

		if age := gctime.Sub(img.detected); age < handler.policy.MinAge {
			log.G(ctx).Infof("Image ID %s has short age=%v < minAge=%v, keep it", img.id, age, handler.policy.MinAge)
			continue
		}

		if err := handler.criruntime.RemoveImage(ctx, img.id); err != nil {
			failedToRemoveImages = append(failedToRemoveImages, img.id)
			log.G(ctx).WithError(err).Warnf("failed to remove image ID: %v", img.id)
			continue
		}
		log.G(ctx).Infof("Removed image ID %s (tags=%v) (digests=%v) ", img.id, img.tags, img.digests)

		delete(handler.images, img.id)
		freedSize += img.sizeBytes
		if freedSize >= totalSize {
			break
		}
	}

	if len(failedToRemoveImages) > 0 {
		return freedSize, errors.Errorf("freed %v bytes but failed to images: %v", freedSize, failedToRemoveImages)
	}
	return freedSize, nil
}

// imageFSStats restrieves image fs stats.
func (handler *imageGCHandler) imageFSStats(ctx context.Context) (fsStats, error) {
	log.G(ctx).Debugf("image fs path %v is in mount point %v", handler.imageFSPath, handler.imageFSMountpoint)

	f, err := os.Open(handler.imageFSMountpoint)
	if err != nil {
		return fsStats{}, err
	}
	defer f.Close()

	var statfs syscall.Statfs_t
	if err := syscall.Fstatfs(int(f.Fd()), &statfs); err != nil {
		return fsStats{}, errors.Wrapf(err, "failed to calculate statfs for mountinfo %v", handler.imageFSMountpoint)
	}

	return fsStats{
		availableBytes: uint64(statfs.Bsize) * statfs.Bfree,
		capacityBytes:  uint64(statfs.Bsize) * statfs.Blocks,
		inodesFree:     statfs.Ffree,
		inodes:         statfs.Files,
	}, nil
}
