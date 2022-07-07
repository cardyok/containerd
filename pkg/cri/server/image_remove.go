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

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/log"

	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// RemoveImage removes the image.
// TODO(random-liu): Update CRI to pass image reference instead of ImageSpec. (See
// kubernetes/kubernetes#46255)
// TODO(random-liu): We should change CRI to distinguish image id and image spec.
// Remove the whole image no matter the it's image id or reference. This is the
// semantic defined in CRI now.
func (c *criService) RemoveImage(ctx context.Context, r *runtime.RemoveImageRequest) (*runtime.RemoveImageResponse, error) {
	image, err := c.localResolve(r.GetImage().GetImage())
	if err != nil {
		if errdefs.IsNotFound(err) {
			// return empty without error when image not found.
			return &runtime.RemoveImageResponse{}, nil
		}
		return nil, fmt.Errorf("can not resolve %q locally: %w", r.GetImage().GetImage(), err)
	}

	// Remove all image references.
	for _, ref := range image.References {
		var opts []leases.DeleteOpt
		opts = []leases.DeleteOpt{leases.SynchronousDelete}
		leaseService := c.client.LeasesService()
		filter := fmt.Sprintf("id==%s", ref)
		lease, err := leaseService.List(ctx, filter)
		if err != nil || len(lease) != 1 {
			log.G(ctx).WithError(err).Warnf("Failed to get lease for img %q", ref)
		} else {
			if err := leaseService.Delete(ctx, lease[0], opts...); err != nil {
				return nil, fmt.Errorf("failed to delete lease for img %q: %w", ref, err)
			}
		}

		// Update image store to reflect the newest state in containerd.
		if err := c.imageStore.Update(ctx, ref); err != nil {
			return nil, fmt.Errorf("failed to update image reference %q for %q: %w", ref, image.ID, err)
		}
	}
	return &runtime.RemoveImageResponse{}, nil
}
