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

package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"

	"github.com/containerd/containerd/plugin"
)

type TestStruct struct {
	IntField1   int      `toml:"int1"`
	IntField2   int      `toml:"int2"`
	IntField3   int      `toml:"int3"`
	IntField4   int      `toml:"int4"`
	IntField5   int      `toml:"int5"`
	BoolField1  bool     `toml:"bool1"`
	BoolField2  bool     `toml:"bool2"`
	BoolField3  bool     `toml:"bool3"`
	StrField1   string   `toml:"str1"`
	StrField2   string   `toml:"str2"`
	StrField3   string   `toml:"str3"`
	SliceField1 []string `toml:"slice1"`
	SliceField2 []string `toml:"slice2"`
	SliceField3 []string `toml:"slice3"`
	SliceField4 []string `toml:"slice4"`
	SliceField5 []string `toml:"slice5"`
}

func TestMergeConfig2(t *testing.T) {
	for desc, test := range map[string]struct {
		to             *TestStruct
		from           string
		expectedResult *TestStruct
	}{
		"no map no slice": {
			to: &TestStruct{
				IntField3:   2,
				IntField5:   2,
				BoolField1:  false,
				BoolField2:  true,
				BoolField3:  true,
				StrField1:   "",
				StrField2:   "a",
				StrField3:   "a",
				SliceField2: []string{"a"},
				SliceField3: []string{"a"},
				SliceField4: []string{"a"},
			},
			from: `
int1=0
int4=2
int5=0
bool1=true
bool2=false
str1="a"
str2=""
slice1=["a"]
slice2=["b"]
slice4=[]
`,
			expectedResult: &TestStruct{
				IntField1:   0,
				IntField2:   0,
				IntField3:   2,
				IntField4:   2,
				IntField5:   0,
				BoolField1:  true,
				BoolField2:  false,
				BoolField3:  true,
				StrField1:   "a",
				StrField2:   "",
				StrField3:   "a",
				SliceField1: []string{"a"},
				SliceField2: []string{"b"},
				SliceField3: []string{"a"},
				SliceField4: []string{},
				SliceField5: nil,
			},
		},
	} {
		t.Run(desc, func(t *testing.T) {
			err := mergeConfigTest(test.to, test.from)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResult, test.to)
		})
	}
}

func TestValidateConfig(t *testing.T) {
	for desc, test := range map[string]struct {
		config      *PluginConfig
		expectedErr string
		expected    *PluginConfig
	}{
		"deprecated untrusted_workload_runtime": {
			config: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					UntrustedWorkloadRuntime: Runtime{
						Type: "untrusted",
					},
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Type: "default",
						},
					},
				},
			},
			expected: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					UntrustedWorkloadRuntime: Runtime{
						Type: "untrusted",
					},
					Runtimes: map[string]Runtime{
						RuntimeUntrusted: {
							Type: "untrusted",
						},
						RuntimeDefault: {
							Type: "default",
						},
					},
				},
			},
		},
		"both untrusted_workload_runtime and runtime[untrusted]": {
			config: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					UntrustedWorkloadRuntime: Runtime{
						Type: "untrusted-1",
					},
					Runtimes: map[string]Runtime{
						RuntimeUntrusted: {
							Type: "untrusted-2",
						},
						RuntimeDefault: {
							Type: "default",
						},
					},
				},
			},
			expectedErr: fmt.Sprintf("conflicting definitions: configuration includes both `untrusted_workload_runtime` and `runtimes[%q]`", RuntimeUntrusted),
		},
		"deprecated default_runtime": {
			config: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntime: Runtime{
						Type: "default",
					},
				},
			},
			expected: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntime: Runtime{
						Type: "default",
					},
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Type: "default",
						},
					},
				},
			},
		},
		"no default_runtime_name": {
			config:      &PluginConfig{},
			expectedErr: "`default_runtime_name` is empty",
		},
		"no runtime[default_runtime_name]": {
			config: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
				},
			},
			expectedErr: "no corresponding runtime configured in `containerd.runtimes` for `containerd` `default_runtime_name = \"default\"",
		},
		"deprecated systemd_cgroup for v1 runtime": {
			config: &PluginConfig{
				SystemdCgroup: true,
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Type: plugin.RuntimeLinuxV1,
						},
					},
				},
			},
			expected: &PluginConfig{
				SystemdCgroup: true,
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Type: plugin.RuntimeLinuxV1,
						},
					},
				},
			},
		},
		"deprecated systemd_cgroup for v2 runtime": {
			config: &PluginConfig{
				SystemdCgroup: true,
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Type: plugin.RuntimeRuncV1,
						},
					},
				},
			},
			expectedErr: fmt.Sprintf("`systemd_cgroup` only works for runtime %s", plugin.RuntimeLinuxV1),
		},
		"no_pivot for v1 runtime": {
			config: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					NoPivot:            true,
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Type: plugin.RuntimeLinuxV1,
						},
					},
				},
			},
			expected: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					NoPivot:            true,
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Type: plugin.RuntimeLinuxV1,
						},
					},
				},
			},
		},
		"no_pivot for v2 runtime": {
			config: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					NoPivot:            true,
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Type: plugin.RuntimeRuncV1,
						},
					},
				},
			},
			expectedErr: fmt.Sprintf("`no_pivot` only works for runtime %s", plugin.RuntimeLinuxV1),
		},
		"deprecated runtime_engine for v1 runtime": {
			config: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Engine: "runc",
							Type:   plugin.RuntimeLinuxV1,
						},
					},
				},
			},
			expected: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Engine: "runc",
							Type:   plugin.RuntimeLinuxV1,
						},
					},
				},
			},
		},
		"deprecated runtime_engine for v2 runtime": {
			config: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Engine: "runc",
							Type:   plugin.RuntimeRuncV1,
						},
					},
				},
			},
			expectedErr: fmt.Sprintf("`runtime_engine` only works for runtime %s", plugin.RuntimeLinuxV1),
		},
		"deprecated runtime_root for v1 runtime": {
			config: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Root: "/run/containerd/runc",
							Type: plugin.RuntimeLinuxV1,
						},
					},
				},
			},
			expected: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Root: "/run/containerd/runc",
							Type: plugin.RuntimeLinuxV1,
						},
					},
				},
			},
		},
		"deprecated runtime_root for v2 runtime": {
			config: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Root: "/run/containerd/runc",
							Type: plugin.RuntimeRuncV1,
						},
					},
				},
			},
			expectedErr: fmt.Sprintf("`runtime_root` only works for runtime %s", plugin.RuntimeLinuxV1),
		},
		"deprecated auths": {
			config: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Type: plugin.RuntimeRuncV1,
						},
					},
				},
				Registry: Registry{
					Auths: map[string]AuthConfig{
						"https://gcr.io": {Username: "test"},
					},
				},
			},
			expected: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Type: plugin.RuntimeRuncV1,
						},
					},
				},
				Registry: Registry{
					Configs: map[string]RegistryConfig{
						"gcr.io": {
							Auth: &AuthConfig{
								Username: "test",
							},
						},
					},
					Auths: map[string]AuthConfig{
						"https://gcr.io": {Username: "test"},
					},
				},
			},
		},
		"invalid stream_idle_timeout": {
			config: &PluginConfig{
				StreamIdleTimeout: "invalid",
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Type: "default",
						},
					},
				},
			},
			expectedErr: "invalid stream idle timeout",
		},
		"conflicting mirror registry config": {
			config: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Type: "default",
						},
					},
				},
				Registry: Registry{
					ConfigPath: "/etc/containerd/conf.d",
					Mirrors: map[string]Mirror{
						"something.io": {},
					},
				},
			},
			expectedErr: "`mirrors` cannot be set when `config_path` is provided",
		},
		"conflicting tls registry config": {
			config: &PluginConfig{
				ContainerdConfig: ContainerdConfig{
					DefaultRuntimeName: RuntimeDefault,
					Runtimes: map[string]Runtime{
						RuntimeDefault: {
							Type: "default",
						},
					},
				},
				Registry: Registry{
					ConfigPath: "/etc/containerd/conf.d",
					Configs: map[string]RegistryConfig{
						"something.io": {
							TLS: &TLSConfig{},
						},
					},
				},
			},
			expectedErr: "`configs.tls` cannot be set when `config_path` is provided",
		},
	} {
		t.Run(desc, func(t *testing.T) {
			err := ValidatePluginConfig(context.Background(), test.config)
			if test.expectedErr != "" {
				assert.Contains(t, err.Error(), test.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, test.config)
			}
		})
	}
}

func mergeConfigTest(to *TestStruct, from string) error {

	file, err := toml.Load(from)
	if err != nil {
		return err
	}
	if err := file.Unmarshal(to); err != nil {
		return err
	}
	return nil
}
