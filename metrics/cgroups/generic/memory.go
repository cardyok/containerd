//go:build linux
// +build linux

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

package generic

import (
	metrics "github.com/docker/go-metrics"
	"github.com/prometheus/client_golang/prometheus"

	v1 "github.com/containerd/containerd/metrics/types/v1"
	v2 "github.com/containerd/containerd/metrics/types/v2"
)

var memoryMetrics = []*Metric{
	// v1 metrics
	{
		name: "memory_cache",
		help: "The cache amount used",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Cache),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_rss",
		help: "The rss amount used",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.RSS),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_rss_huge",
		help: "The rss_huge amount used",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.RSSHuge),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_mapped_file",
		help: "The mapped_file amount used",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.MappedFile),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_dirty",
		help: "The dirty amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Dirty),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_writeback",
		help: "The writeback amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Writeback),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_pgpgin",
		help: "The pgpgin amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.PgPgIn),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_pgpgout",
		help: "The pgpgout amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.PgPgOut),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_pgfault",
		help: "The pgfault amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.PgFault),
					},
				}
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Pgfault),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_pgmajfault",
		help: "The pgmajfault amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.PgMajFault),
					},
				}
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Pgmajfault),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_inactive_anon",
		help: "The inactive_anon amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.InactiveAnon),
					},
				}
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.InactiveAnon),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_active_anon",
		help: "The active_anon amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.ActiveAnon),
					},
				}
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.ActiveAnon),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_inactive_file",
		help: "The inactive_file amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.InactiveFile),
					},
				}
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.InactiveFile),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_active_file",
		help: "The active_file amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.ActiveFile),
					},
				}
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.ActiveFile),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_unevictable",
		help: "The unevictable amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Unevictable),
					},
				}
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Unevictable),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_hierarchical_memory_limit",
		help: "The hierarchical_memory_limit amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.HierarchicalMemoryLimit),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_hierarchical_memsw_limit",
		help: "The hierarchical_memsw_limit amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.HierarchicalSwapLimit),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_cache",
		help: "The total_cache amount used",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalCache),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_rss",
		help: "The total_rss amount used",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalRSS),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_rss_huge",
		help: "The total_rss_huge amount used",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalRSSHuge),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_mapped_file",
		help: "The total_mapped_file amount used",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalMappedFile),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_dirty",
		help: "The total_dirty amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalDirty),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_writeback",
		help: "The total_writeback amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalWriteback),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_pgpgin",
		help: "The total_pgpgin amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalPgPgIn),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_pgpgout",
		help: "The total_pgpgout amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalPgPgOut),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_pgfault",
		help: "The total_pgfault amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalPgFault),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_pgmajfault",
		help: "The total_pgmajfault amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalPgMajFault),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_inactive_anon",
		help: "The total_inactive_anon amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalInactiveAnon),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_active_anon",
		help: "The total_active_anon amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalActiveAnon),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_inactive_file",
		help: "The total_inactive_file amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalInactiveFile),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_active_file",
		help: "The total_active_file amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalActiveFile),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_total_unevictable",
		help: "The total_unevictable amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.TotalUnevictable),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_usage_failcnt",
		help: "The usage failcnt",
		unit: metrics.Total,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.Usage == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Usage.Failcnt),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_usage_limit",
		help: "The memory limit",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.Usage == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Usage.Limit),
					},
				}
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.UsageLimit),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_usage_max",
		help: "The memory maximum usage",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.Usage == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Usage.Max),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_usage_usage",
		help: "The memory usage",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.Usage == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Usage.Usage),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_swap_failcnt",
		help: "The swap failcnt",
		unit: metrics.Total,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.Usage == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Usage.Failcnt),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_swap_limit",
		help: "The swap limit",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.Swap == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Swap.Limit),
					},
				}
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.SwapLimit),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_swap_max",
		help: "The swap maximum usage",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.Usage == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Usage.Failcnt),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_swap_usage",
		help: "The swap usage",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.Swap == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Swap.Usage),
					},
				}
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.SwapUsage),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_kernel_failcnt",
		help: "The kernel failcnt",
		unit: metrics.Total,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.Kernel == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Kernel.Failcnt),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_kernel_limit",
		help: "The kernel limit",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.Kernel == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Kernel.Limit),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_kernel_max",
		help: "The kernel maximum usage",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.Kernel == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Kernel.Max),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_kernel_usage",
		help: "The kernel usage",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.Kernel == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Kernel.Usage),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_kerneltcp_failcnt",
		help: "The kerneltcp failcnt",
		unit: metrics.Total,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.KernelTCP == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.KernelTCP.Failcnt),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_kerneltcp_limit",
		help: "The kerneltcp limit",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.KernelTCP == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.KernelTCP.Limit),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_kerneltcp_max",
		help: "The kerneltcp maximum usage",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.KernelTCP == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.KernelTCP.Max),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "memory_kerneltcp_usage",
		help: "The kerneltcp usage",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Memory == nil {
					return nil
				}
				if s.Memory.KernelTCP == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.KernelTCP.Usage),
					},
				}
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	// v2 metrics
	{
		name: "memory_usage",
		help: "Current memory usage (cgroup v2)",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Usage),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_file_mapped",
		help: "The file_mapped amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.FileMapped),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_file_dirty",
		help: "The file_dirty amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.FileDirty),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_file_writeback",
		help: "The file_writeback amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.FileWriteback),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_pgactivate",
		help: "The pgactivate amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Pgactivate),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_pgdeactivate",
		help: "The pgdeactivate amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Pgdeactivate),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_pglazyfree",
		help: "The pglazyfree amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Pglazyfree),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_pgrefill",
		help: "The pgrefill amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Pgrefill),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_pglazyfreed",
		help: "The pglazyfreed amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Pglazyfreed),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_pgscan",
		help: "The pgscan amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Pgscan),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_pgsteal",
		help: "The pgsteal amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Pgsteal),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_anon",
		help: "The anon amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Anon),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_file",
		help: "The file amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.File),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_kernel_stack",
		help: "The kernel_stack amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.KernelStack),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_slab",
		help: "The slab amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Slab),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_sock",
		help: "The sock amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Sock),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_shmem",
		help: "The shmem amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.Shmem),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_anon_thp",
		help: "The anon_thp amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.AnonThp),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_slab_reclaimable",
		help: "The slab_reclaimable amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.SlabReclaimable),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_slab_unreclaimable",
		help: "The slab_unreclaimable amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.SlabUnreclaimable),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_workingset_refault",
		help: "The workingset_refault amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.WorkingsetRefault),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_workingset_activate",
		help: "The workingset_activate amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.WorkingsetActivate),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_workingset_nodereclaim",
		help: "The workingset_nodereclaim amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.WorkingsetNodereclaim),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_thp_fault_alloc",
		help: "The thp_fault_alloc amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.ThpFaultAlloc),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_thp_collapse_alloc",
		help: "The thp_collapse_alloc amount",
		unit: metrics.Bytes,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Memory == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Memory.ThpCollapseAlloc),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "memory_oom",
		help: "The number of times a container has received an oom event",
		unit: metrics.Total,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.MemoryEvents == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.MemoryEvents.Oom),
					},
				}
			default:
				return nil
			}
		},
	},
}
