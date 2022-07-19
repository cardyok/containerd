package quota

import (
	"github.com/containerd/containerd/snapshots/overlay/quota/sparsefile"
	"github.com/containerd/containerd/snapshots/overlay/quota/types"
)

func New(driver string, config map[string]string) types.Quota {
	switch driver {
	case sparsefile.QuotaName:
		return sparsefile.New(config)
	}
	return nil
}
