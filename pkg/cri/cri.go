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

package cri

import (
	"flag"
	"fmt"
	"path/filepath"

	imagespec "github.com/opencontainers/image-spec/specs-go/v1"
	"k8s.io/klog/v2"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/api/services/containers/v1"
	"github.com/containerd/containerd/api/services/diff/v1"
	"github.com/containerd/containerd/api/services/images/v1"
	introspectionapi "github.com/containerd/containerd/api/services/introspection/v1"
	"github.com/containerd/containerd/api/services/namespaces/v1"
	"github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/pkg/netns"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/plugin"
	"github.com/containerd/containerd/services"
	"github.com/containerd/containerd/snapshots"

	criconfig "github.com/containerd/containerd/pkg/cri/config"
	"github.com/containerd/containerd/pkg/cri/constants"
	"github.com/containerd/containerd/pkg/cri/server"
	"github.com/containerd/containerd/pkg/imagegcplugin"
)

// Register CRI service plugin
func init() {
	config := criconfig.DefaultConfig()
	plugin.Register(&plugin.Registration{
		Type:   plugin.GRPCPlugin,
		ID:     "cri",
		Config: &config,
		Requires: []plugin.Type{
			plugin.EventPlugin,
			plugin.ServicePlugin,
			plugin.InternalPlugin,
		},
		InitFn: initCRIService,
	})
}

func initCRIService(ic *plugin.InitContext) (interface{}, error) {
	ready := ic.RegisterReadiness()
	ic.Meta.Platforms = []imagespec.Platform{platforms.DefaultSpec()}
	ic.Meta.Exports = map[string]string{"CRIVersion": constants.CRIVersion, "CRIVersionAlpha": constants.CRIVersionAlpha}
	ctx := ic.Context
	pluginConfig := ic.Config.(*criconfig.PluginConfig)
	if err := criconfig.ValidatePluginConfig(ctx, pluginConfig); err != nil {
		return nil, fmt.Errorf("invalid plugin config: %w", err)
	}

	c := criconfig.Config{
		PluginConfig:       *pluginConfig,
		ContainerdRootDir:  filepath.Dir(ic.Root),
		ContainerdEndpoint: ic.Address,
		RootDir:            ic.Root,
		StateDir:           ic.State,
	}
	log.G(ctx).Infof("Start cri plugin with config %+v", c)

	if err := setGLogLevel(); err != nil {
		return nil, fmt.Errorf("failed to set glog level: %w", err)
	}

	servicesOpts, err := getServicesOpts(ic)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	log.G(ctx).Info("Connect containerd service")
	client, err := containerd.New(
		"",
		containerd.WithDefaultNamespace(constants.K8sContainerdNamespace),
		containerd.WithDefaultPlatform(platforms.Default()),
		containerd.WithServices(servicesOpts...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}

	// creating a guard netns before cri service starting to avoid a kernel bug:
	// concurrently creating the first netns after rebooting would lead to creation failure later.
	// Debian Bug report logs - #949235 ip netns add race condition can wreak havoc in mount points
	// link: https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=949235
	log.G(ctx).Info("Create place holder netns to avoid kernel bug")
	if err := netns.RecoverNetNS("/var/run/netns/CONTAINERD_PLACE_HOLDER"); err != nil {
		return nil, fmt.Errorf("failed to create place holder netns: %w", err)
	}

	s, err := server.NewCRIService(c, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRI service: %w", err)
	}

	if err := initImageGCForCri(ic, s.(initializeImageGC)); err != nil {
		return nil, fmt.Errorf("failed to init image gc for cri: %w", err)
	}

	go func() {
		if err := s.Run(ready); err != nil {
			log.G(ctx).WithError(err).Fatal("Failed to run CRI service")
		}
		// TODO(random-liu): Whether and how we can stop containerd.
	}()
	return s, nil
}

// getServicesOpts get service options from plugin context.
func getServicesOpts(ic *plugin.InitContext) ([]containerd.ServicesOpt, error) {
	plugins, err := ic.GetByType(plugin.ServicePlugin)
	if err != nil {
		return nil, fmt.Errorf("failed to get service plugin: %w", err)
	}

	ep, err := ic.Get(plugin.EventPlugin)
	if err != nil {
		return nil, fmt.Errorf("failed to get event plugin: %w", err)
	}

	opts := []containerd.ServicesOpt{
		containerd.WithEventService(ep.(containerd.EventService)),
	}
	for s, fn := range map[string]func(interface{}) containerd.ServicesOpt{
		services.ContentService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithContentStore(s.(content.Store))
		},
		services.ImagesService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithImageClient(s.(images.ImagesClient))
		},
		services.SnapshotsService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithSnapshotters(s.(map[string]snapshots.Snapshotter))
		},
		services.ContainersService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithContainerClient(s.(containers.ContainersClient))
		},
		services.TasksService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithTaskClient(s.(tasks.TasksClient))
		},
		services.DiffService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithDiffClient(s.(diff.DiffClient))
		},
		services.NamespacesService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithNamespaceClient(s.(namespaces.NamespacesClient))
		},
		services.LeasesService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithLeasesService(s.(leases.Manager))
		},
		services.IntrospectionService: func(s interface{}) containerd.ServicesOpt {
			return containerd.WithIntrospectionClient(s.(introspectionapi.IntrospectionClient))
		},
	} {
		p := plugins[s]
		if p == nil {
			return nil, fmt.Errorf("service %q not found", s)
		}
		i, err := p.Instance()
		if err != nil {
			return nil, fmt.Errorf("failed to get instance of service %q: %w", s, err)
		}
		if i == nil {
			return nil, fmt.Errorf("instance of service %q not found", s)
		}
		opts = append(opts, fn(i))
	}
	return opts, nil
}

// Set glog level.
func setGLogLevel() error {
	l := log.GetLevel()
	fs := flag.NewFlagSet("klog", flag.PanicOnError)
	klog.InitFlags(fs)
	if err := fs.Set("logtostderr", "true"); err != nil {
		return err
	}
	switch l {
	case log.TraceLevel:
		return fs.Set("v", "5")
	case log.DebugLevel:
		return fs.Set("v", "4")
	case log.InfoLevel:
		return fs.Set("v", "2")
	default:
		// glog doesn't support other filters. Defaults to v=0.
	}
	return nil
}

type initializeImageGC interface {
	InitImageGC(imagegcplugin.Status) error
}

func initImageGCForCri(ic *plugin.InitContext, handler initializeImageGC) error {
	// init image gc
	internalPlugins, err := ic.GetByType(plugin.InternalPlugin)
	if err != nil {
		return fmt.Errorf("failed to get internal plugins: %w", err)
	}

	imageGCSwitch := internalPlugins[imagegcplugin.SwitchPluginName]
	if imageGCSwitch == nil {
		return fmt.Errorf("failed to get internal plugin for image gc switch: %w", err)
	}

	gcSwitchInstance, err := imageGCSwitch.Instance()
	if err != nil {
		return fmt.Errorf("failed to get image gc switch instance: %w", err)
	}

	if err := handler.InitImageGC(gcSwitchInstance.(imagegcplugin.Status)); err != nil {
		return fmt.Errorf("failed to init image gc: %w", err)
	}
	return nil
}
