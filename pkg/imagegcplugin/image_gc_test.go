package imagegcplugin

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	assertlib "github.com/stretchr/testify/assert"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func TestDetechImagesLogic(t *testing.T) {
	assert := assertlib.New(t)

	cwdPath, err := os.Getwd()
	assert.NoError(err, "failed to get cwdPath")

	var (
		pauseImage = makeImage("pause", size1k)
		abcImage   = makeImage("abc", size8k)
		defgImage  = makeImage("defg", size16k)
		hijkImage  = makeImage("hijk", size128k)
	)

	fake := &fakeCRIRuntime{
		images: []*runtime.Image{
			pauseImage,
			abcImage,
			defgImage,
			hijkImage,
		},
		containers: []*runtime.Container{
			makeContainer("abc", "abc"),
			makeContainer("defg", "defg"),
		},
	}

	gcinstance, err := NewImageGarbageCollect(NewCRIRuntime(fake), GcPolicy{
		HighThresholdPercent: 100,
		LowThresholdPercent:  1,
		MinAge:               10 * time.Second,
		Whitelist:            []string{imageTag("pause"), imageTag("abc")},
	}, cwdPath)
	assert.NoError(err, "failed to new image garbage collector")

	handler := gcinstance.(*imageGCHandler)

	// check usedImageSet result and detectd time
	{
		detectedTime := time.Now().Add(-10 * time.Second)
		resultSet, err := handler.detectImages(context.TODO(), detectedTime)
		assert.NoError(err, "failed to detect images")
		t.Logf("detected result: %v at %v", resultSet, detectedTime)

		for _, img := range []string{
			imageID("hijk"), imageTag("hijk"), imageDigest("hijk"),
			imageTag("pause"), imageDigest("pause"),
		} {
			assert.NotEqual(resultSet.contains(img), true, "should not contains %s in result", img)
		}

		assert.Equal(len(resultSet), 3, "should only have 3 image")
		for _, img := range []*runtime.Image{
			pauseImage, abcImage, defgImage,
		} {
			assert.Equal(resultSet.contains(img.Id), true, "should contains ID %s in result", img.Id)
			assert.NotNil(handler.images[img.Id], nil, "should detect image ID %s", img.Id)
			assert.Equal(handler.images[img.Id].sizeBytes, img.Size_, "image size should be %v", img.Size_)
			assert.Equal(handler.images[img.Id].detected, detectedTime, "image should be detected at %v", detectedTime)

			if img.Id == imageID("pause") {
				assert.Equal(handler.images[img.Id].lastUsed, time.Time{}, "image should not be used")
			} else {
				assert.Equal(handler.images[img.Id].lastUsed.After(detectedTime), true, "detected should before lastUsed(%v)", handler.images[img.Id].lastUsed)
			}
		}
	}

	// remove abc container but abc should be used list
	{
		fake.containers = fake.containers[1:]

		detectedTime := time.Now()
		resultSet, err := handler.detectImages(context.TODO(), detectedTime)
		assert.NoError(err, "failed to detect images")
		t.Logf("detected result: %v at %v", resultSet, detectedTime)

		assert.Equal(len(resultSet), 3, "should have 3 image")

		// abc image should not be updated for lastUsed
		assert.Equal(resultSet.contains(abcImage.Id), true, "should contains abc image in result")
		assert.Equal(handler.images[abcImage.Id].lastUsed.Before(detectedTime), true, "abc image should not be updated lastUsed")

		assert.Equal(handler.images[defgImage.Id].lastUsed.After(detectedTime), true, "defg image should be updated lastUsed")

		for _, img := range []*runtime.Image{
			pauseImage, abcImage, defgImage,
		} {
			assert.Equal(handler.images[img.Id].detected.Before(detectedTime), true, "%s image should not be updated detectedTime", img.Id)
		}
	}

	// remove whitelist but add whitelist regexp to keep pause and abc image
	{
		oldwhitelist := handler.policy.Whitelist
		handler.policy.Whitelist = nil
		handler.policy.WhitelistGoRegex = "pause*|abc*"

		goregexp, err := regexp.Compile(handler.policy.WhitelistGoRegex)
		assert.NoError(err, "failed to compile goregexp")
		handler.whitelistGoRegex = goregexp

		detectedTime := time.Now()
		resultSet, err := handler.detectImages(context.TODO(), detectedTime)
		assert.NoError(err, "failed to detect images")
		t.Logf("detected result: %v at %v", resultSet, detectedTime)

		assert.Equal(len(resultSet), 3, "should have 3 image")

		for _, img := range []*runtime.Image{
			pauseImage, abcImage, defgImage,
		} {
			assert.Equal(handler.images[img.Id].detected.Before(detectedTime), true, "%s image should not be updated detectedTime", img.Id)
		}

		handler.whitelistGoRegex = nil
		handler.policy.Whitelist = oldwhitelist
		handler.policy.WhitelistGoRegex = ""
	}

	// remove pause image and plugin record should remove it too
	{
		fake.images = fake.images[1:]

		detectedTime := time.Now()
		resultSet, err := handler.detectImages(context.TODO(), detectedTime)
		assert.NoError(err, "failed to detect images")
		t.Logf("detected result: %v at %v", resultSet, detectedTime)

		assert.Equal(len(resultSet), 2, "should have 2 image")

		for _, img := range []*runtime.Image{
			abcImage, defgImage,
		} {
			assert.Equal(resultSet.contains(img.Id), true, "should contains ID %s in result", img.Id)
		}
	}
}

func TestRemoveUnusedImage(t *testing.T) {
	assert := assertlib.New(t)

	cwdPath, err := os.Getwd()
	assert.NoError(err, "failed to get cwdPath")

	var (
		pauseImage = makeImage("pause", size1k)
		abcImage   = makeImage("abc", size8k)
		defgImage  = makeImage("defg", size16k)
		hijkImage  = makeImage("hijk", size128k)
	)

	fake := &fakeCRIRuntime{
		images: []*runtime.Image{
			pauseImage,
			abcImage,
			defgImage,
			hijkImage,
		},
		containers: []*runtime.Container{
			makeContainer("abc", "abc"),
			makeContainer("defg", "defg"),
		},
	}

	gcinstance, err := NewImageGarbageCollect(NewCRIRuntime(fake), GcPolicy{
		HighThresholdPercent: 100,
		LowThresholdPercent:  1,
		MinAge:               10 * time.Second,
		Whitelist:            []string{imageTag("pause")},
	}, cwdPath)
	assert.NoError(err, "failed to new image garbage collector")

	handler := gcinstance.(*imageGCHandler)

	detectedTime := time.Now().Add(-10 * time.Second)
	freedSize, err := handler.pruneImages(context.TODO(), size1M, detectedTime)
	assert.Nil(err, "first detected and do nothing")
	assert.Equal(freedSize, uint64(0), "first detected and do nothing")

	// after min age is 10s, detect time is now and it will remove one image
	freedSize, err = handler.pruneImages(context.TODO(), size1M, time.Now())
	assert.Nil(err, "no error during pruneImages")
	assert.Equal(freedSize, size128k, "second detected should remove hijk image")

	// remove container abc and since it has been detected 10s ago, safely remove it
	fake.containers = fake.containers[1:]
	freedSize, err = handler.pruneImages(context.TODO(), size1M, time.Now())
	assert.Nil(err, "no error during pruneImages")
	assert.Equal(freedSize, size8k, "second detected should remove abc image")

	// remove container defg but add new image hijk back
	fake.containers = nil
	fake.images = append(fake.images, hijkImage)
	freedSize, err = handler.pruneImages(context.TODO(), size1M, time.Now())
	assert.Nil(err, "no error during pruneImages")
	assert.Equal(freedSize, size16k, "third detected should remove defg image but keep hijk image")

	// change minAge to 10us and remove whitelist
	// remove pause image first because it is too old
	handler.policy.MinAge = 10 * time.Millisecond
	handler.policy.Whitelist = nil
	time.Sleep(handler.policy.MinAge)
	freedSize, err = handler.pruneImages(context.TODO(), size1k, time.Now())
	assert.Nil(err, "no error during pruneImages")
	assert.Equal(freedSize, size1k, "4th detected should remove pause image")
}

var (
	size1k   uint64 = 1024
	size8k          = size1k * 8
	size16k         = size1k * 16
	size128k        = size1k * 128
	size1M          = size1k * 1024
)

func makeImage(id string, size uint64) *runtime.Image {
	return &runtime.Image{
		Id:          imageID(id),
		RepoTags:    []string{imageTag(id)},
		RepoDigests: []string{imageDigest(id)},
		Size_:       size,
	}
}

func makeContainer(id string, imgID string) *runtime.Container {
	return &runtime.Container{
		Id: fmt.Sprintf("container-%s", id),
		Metadata: &runtime.ContainerMetadata{
			Name: fmt.Sprintf("container-name-%s", id),
		},
		Image: &runtime.ImageSpec{
			Image: imageID(imgID),
		},
		ImageRef: imageID(imgID),
	}
}

func imageID(id string) string {
	return "sha256:" + id
}

func imageTag(id string) string {
	return id + ":v1.0"
}

func imageDigest(id string) string {
	return id + "@sha256:xyz"
}
