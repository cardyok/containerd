package imagegcplugin

import (
	"os"
	"path"

	"github.com/pkg/errors"

	"github.com/containerd/containerd/plugin"
)

var turnoffFile = "turnoff"

// SwitchPluginName is for containerd plugin name
var SwitchPluginName = "imagegcswitch"

// imagegcpswitch is used to be switch of image gc handler.
//
// It is too heavy to change config and restart containerd. The imagegcpswitch
// plugin will read plugin root dir and indicates we should turn off image
// gc if there is one file named `turnoff` in plugin root dir.
//
// NOTE: If the config.HighThresholdPercent is 100, always indicates turn off
// image gc.
func init() {
	config := defaultConfig()
	plugin.Register(&plugin.Registration{
		Type:   plugin.InternalPlugin,
		ID:     SwitchPluginName,
		Config: &config,
		InitFn: initImageGCSwitchPlugin,
	})
}

func initImageGCSwitchPlugin(ic *plugin.InitContext) (interface{}, error) {
	ic.Meta.Exports["root"] = ic.Root
	pluginConfig := ic.Config.(*Config)

	if err := os.MkdirAll(ic.Root, 0755); err != nil {
		return nil, errors.Wrapf(err, "failed to make dir %s for imagegcswitch plugin", ic.Root)
	}
	return &imagegcswitch{
		config:     *pluginConfig,
		targetFile: path.Join(ic.Root, turnoffFile),
	}, nil
}

type imagegcswitch struct {
	config     Config
	targetFile string
}

func (internal *imagegcswitch) Enabled() bool {
	if internal.config.HighThresholdPercent == 100 {
		return false
	}

	if _, err := os.Stat(internal.targetFile); err == nil {
		return false
	}
	return true
}

func (internal *imagegcswitch) Config() Config {
	return internal.config
}
