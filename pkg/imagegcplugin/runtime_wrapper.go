package imagegcplugin

import (
	"context"

	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// GRpcCRIService is part of CRI API interface.
type GRpcCRIService interface {
	ListImages(context.Context, *runtime.ListImagesRequest) (*runtime.ListImagesResponse, error)
	ListContainers(context.Context, *runtime.ListContainersRequest) (*runtime.ListContainersResponse, error)
	RemoveImage(ctx context.Context, r *runtime.RemoveImageRequest) (*runtime.RemoveImageResponse, error)
}

// NewCRIRuntime is wrapper of grpcCRIService.
//
// NOTE: Make unit test easier.
func NewCRIRuntime(endpoint GRpcCRIService) CRIRuntime {
	return &internalCRIRuntime{
		endpoint: endpoint,
	}
}

// internalCRIRuntime implements criRuntime interface.
type internalCRIRuntime struct {
	endpoint GRpcCRIService
}

func (internal *internalCRIRuntime) ListImages(ctx context.Context) ([]Image, error) {
	// ignore filter and get all of them
	resp, err := internal.endpoint.ListImages(ctx, nil)
	if err != nil {
		return nil, err
	}

	res := make([]Image, 0, len(resp.Images))
	for _, img := range resp.Images {
		if img == nil {
			continue
		}
		res = append(res, Image{
			ID:          img.GetId(),
			RepoTags:    img.GetRepoTags(),
			RepoDigests: img.GetRepoDigests(),
			SizeBytes:   img.GetSize_(),
		})
	}
	return res, nil
}

func (internal *internalCRIRuntime) ListContainers(ctx context.Context) ([]Container, error) {
	// ignore filter and get all of them
	resp, err := internal.endpoint.ListContainers(ctx, nil)
	if err != nil {
		return nil, err
	}

	res := make([]Container, 0, len(resp.Containers))
	for _, cntr := range resp.Containers {
		if cntr == nil {
			continue
		}
		res = append(res, Container{
			ID:      cntr.GetId(),
			Name:    cntr.GetMetadata().GetName(),
			Image:   cntr.GetImage().GetImage(),
			ImageID: cntr.GetImageRef(),
		})
	}
	return res, nil
}

func (internal *internalCRIRuntime) RemoveImage(ctx context.Context, idOrRef string) error {
	_, err := internal.endpoint.RemoveImage(ctx, &runtime.RemoveImageRequest{
		Image: &runtime.ImageSpec{
			Image: idOrRef,
		},
	})
	return err
}
