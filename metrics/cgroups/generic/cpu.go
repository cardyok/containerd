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
	"strconv"

	metrics "github.com/docker/go-metrics"
	"github.com/prometheus/client_golang/prometheus"

	v1 "github.com/containerd/containerd/metrics/types/v1"
	v2 "github.com/containerd/containerd/metrics/types/v2"
)

var cpuMetrics = []*Metric{
	// v1 metrics
	{
		name: "cpu_total",
		help: "The total cpu time",
		unit: metrics.Nanoseconds,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.CPU == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.CPU.Usage.Total),
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
		name: "cpu_kernel",
		help: "The total kernel cpu time",
		unit: metrics.Nanoseconds,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.CPU == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.CPU.Usage.Kernel),
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
		name: "cpu_user",
		help: "The total user cpu time",
		unit: metrics.Nanoseconds,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.CPU == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.CPU.Usage.User),
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
		name:   "per_cpu",
		help:   "The total cpu time per cpu",
		unit:   metrics.Nanoseconds,
		vt:     prometheus.GaugeValue,
		labels: []string{"cpu"},
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.CPU == nil {
					return nil
				}
				var out []value
				for i, v := range s.CPU.Usage.PerCPU {
					out = append(out, value{
						v: float64(v),
						l: []string{strconv.Itoa(i)},
					})
				}
				return out
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name: "cpu_throttle_periods",
		help: "The total cpu throttle periods",
		unit: metrics.Total,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.CPU == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.CPU.Throttling.Periods),
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
		name: "cpu_throttled_periods",
		help: "The total cpu throttled periods",
		unit: metrics.Total,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.CPU == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.CPU.Throttling.ThrottledPeriods),
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
		name: "cpu_throttled_time",
		help: "The total cpu throttled time",
		unit: metrics.Nanoseconds,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.CPU == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.CPU.Throttling.ThrottledTime),
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
		name: "cpu_usage_usec",
		help: "Total cpu usage (cgroup v2)",
		unit: metrics.Unit("microseconds"),
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.CPU == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.CPU.UsageUsec),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "cpu_user_usec",
		help: "Current cpu usage in user space (cgroup v2)",
		unit: metrics.Unit("microseconds"),
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.CPU == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.CPU.UserUsec),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "cpu_system_usec",
		help: "Current cpu usage in kernel space (cgroup v2)",
		unit: metrics.Unit("microseconds"),
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.CPU == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.CPU.SystemUsec),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "cpu_nr_periods",
		help: "Current cpu number of periods (only if controller is enabled)",
		unit: metrics.Total,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.CPU == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.CPU.NrPeriods),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "cpu_nr_throttled",
		help: "Total number of times tasks have been throttled (only if controller is enabled)",
		unit: metrics.Total,
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.CPU == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.CPU.NrThrottled),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "cpu_throttled_usec",
		help: "Total time duration for which tasks have been throttled. (only if controller is enabled)",
		unit: metrics.Unit("microseconds"),
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.CPU == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.CPU.ThrottledUsec),
					},
				}
			default:
				return nil
			}
		},
	},
}
