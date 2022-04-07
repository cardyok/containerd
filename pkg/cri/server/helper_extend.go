package server

import (
	"github.com/pkg/errors"
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
		return "", errors.Wrapf(errdefs.ErrUnavailable, "pod config file empty")
	}
	if val, ok := config.Annotations[snapshotterAnno]; ok {
		if found := checkStringSlice(snapshotters, val); !found {
			return "", errors.Wrapf(errdefs.ErrNotFound, "snapshotter %s specified in annotation not supported", val)
		}
		return val, nil
	}
	if val, ok := config.Labels[snapshotterLabel]; ok {
		if found := checkStringSlice(snapshotters, val); !found {
			return "", errors.Wrapf(errdefs.ErrNotFound, "snapshotter %s specified in label not supported", val)
		}
		return val, nil
	}
	return "", errdefs.ErrNotImplemented
}
