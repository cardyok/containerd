package server

import (
	"fmt"

	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"

	"github.com/containerd/containerd/errdefs"
)

const (
	// snapshotterAnno is a annotation which contains default snapshotter selected for pod and its containers
	snapshotterAnno = "alibabacloud.com/cri.snapshotter"

	// snapshotterLabel is a label which contains default snapshotter selected for pod and its containers
	snapshotterLabel = "containerd.io/snapshot.cri.snapshotter"
)

// checkStringSlice checks whether target exists in src string slice.
func checkStringSlice(src []string, target string) bool {
	for _, i := range src {
		if i == target {
			return true
		}
	}
	return false
}

// getPodSnapshotter tries to retrieve specified snapshotter from pod config annotation and label
func getPodSnapshotter(config *runtime.PodSandboxConfig, snapshotters []string) (string, error) {
	if config == nil {
		return "", fmt.Errorf("pod config file empty: %w", errdefs.ErrUnavailable)
	}
	if val, ok := config.Annotations[snapshotterAnno]; ok {
		if found := checkStringSlice(snapshotters, val); !found {
			return "", fmt.Errorf("snapshotter %s specified in annotation not supported: %w", val, errdefs.ErrNotFound)
		}
		return val, nil
	}
	if val, ok := config.Labels[snapshotterLabel]; ok {
		if found := checkStringSlice(snapshotters, val); !found {
			return "", fmt.Errorf("snapshotter %s specified in label not supported: %w", val, errdefs.ErrNotFound)
		}
		return val, nil
	}
	return "", errdefs.ErrNotImplemented
}
