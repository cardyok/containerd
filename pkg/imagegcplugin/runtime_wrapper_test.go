package imagegcplugin

import (
	"context"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

var _ CRIRuntime = &internalCRIRuntime{}

func TestBasicRuntimeWrapper(t *testing.T) {
	fake := &fakeCRIRuntime{
		images: []*runtime.Image{
			{
				Id:          "A",
				RepoTags:    []string{"A:1"},
				RepoDigests: []string{"A@sha256:xyz"},
				Size_:       1024,
			},
			{
				Id:          "B",
				RepoTags:    []string{"B:1"},
				RepoDigests: []string{"B@sha256:abc"},
				Size_:       6666,
			},
		},
		containers: []*runtime.Container{
			{
				Id: "C-A",
				Metadata: &runtime.ContainerMetadata{
					Name: "C-A-ME",
				},
				Image: &runtime.ImageSpec{
					Image: "A",
				},
				ImageRef: "A@sha256:xyz",
			},
			{
				Id: "C-B",
				Metadata: &runtime.ContainerMetadata{
					Name: "C-B-ME",
				},
				Image:    nil,
				ImageRef: "B@sha256:abc",
			},
		},
	}

	ctx := context.Background()
	wrapper := NewCRIRuntime(fake)

	// ListImages
	expectedListImages := []Image{
		{
			ID:          "A",
			RepoTags:    []string{"A:1"},
			RepoDigests: []string{"A@sha256:xyz"},
			SizeBytes:   1024,
		},
		{
			ID:          "B",
			RepoTags:    []string{"B:1"},
			RepoDigests: []string{"B@sha256:abc"},
			SizeBytes:   6666,
		},
	}
	gotImages, err := wrapper.ListImages(ctx)
	if err != nil {
		t.Errorf("unexpected error during list image")
	}
	if !reflect.DeepEqual(expectedListImages, gotImages) {
		t.Errorf("expected images=%v, but got=%v", expectedListImages, gotImages)
	}

	// ListContainers
	expectedListContainers := []Container{
		{
			ID:      "C-A",
			Name:    "C-A-ME",
			Image:   "A",
			ImageID: "A@sha256:xyz",
		},
		{
			ID:      "C-B",
			Name:    "C-B-ME",
			Image:   "",
			ImageID: "B@sha256:abc",
		},
	}
	gotContainers, err := wrapper.ListContainers(ctx)
	if err != nil {
		t.Errorf("unexpected error during list containers")
	}
	if !reflect.DeepEqual(expectedListContainers, gotContainers) {
		t.Errorf("expected containers=%v, but got=%v", expectedListContainers, gotContainers)
	}

	// RemoveImage
	if err := wrapper.RemoveImage(ctx, "A"); err != nil {
		t.Errorf("unexpected error during remove image A")
	}

	gotImages, err = wrapper.ListImages(ctx)
	if err != nil {
		t.Errorf("unexpected error during list image")
	}
	if len(gotImages) != 1 {
		t.Errorf("expected len images = 1, but got %v", len(gotImages))
	}
}

type fakeCRIRuntime struct {
	images     []*runtime.Image
	containers []*runtime.Container
}

func (f *fakeCRIRuntime) ListImages(_ context.Context, _ *runtime.ListImagesRequest) (*runtime.ListImagesResponse, error) {
	return &runtime.ListImagesResponse{
		Images: f.images,
	}, nil
}

func (f *fakeCRIRuntime) ListContainers(_ context.Context, _ *runtime.ListContainersRequest) (*runtime.ListContainersResponse, error) {
	return &runtime.ListContainersResponse{
		Containers: f.containers,
	}, nil
}

func (f *fakeCRIRuntime) RemoveImage(_ context.Context, r *runtime.RemoveImageRequest) (*runtime.RemoveImageResponse, error) {
	for idx, image := range f.images {
		if image.Id == r.GetImage().GetImage() {
			f.images = append(f.images[:idx], f.images[idx+1:]...)
			return &runtime.RemoveImageResponse{}, nil
		}
	}
	return nil, errors.Errorf("not found")
}
