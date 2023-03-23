package metahive

import (
	"context"

	"github.com/hashicorp/memberlist"

	"github.com/containerd/containerd/log"
)

func New(ctx context.Context, config *Config) (string, error) {
	memlist, err := memberlist.Create(memberlist.DefaultLANConfig())
	if err != nil {
		return "", err
	}
	log.G(ctx).Infof("Using Root Address %s\n", config.RootAddress)

	_, err = memlist.Join([]string{config.RootAddress})
	if err != nil {
		log.G(ctx).Infof("Failed to join cluster: " + err.Error())
	}
	return "", nil
}
