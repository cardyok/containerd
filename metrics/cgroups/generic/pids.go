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

var pidMetrics = []*Metric{
	{
		name: "pids",
		help: "The limit to the number of pids allowed",
		unit: metrics.Unit("limit"),
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Pids == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Pids.Limit),
					},
				}
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Pids == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Pids.Limit),
					},
				}
			default:
				return nil
			}
		},
	},
	{
		name: "pids",
		help: "The current number of pids",
		unit: metrics.Unit("current"),
		vt:   prometheus.GaugeValue,
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Pids == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Pids.Current),
					},
				}
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Pids == nil {
					return nil
				}
				return []value{
					{
						v: float64(s.Pids.Current),
					},
				}
			default:
				return nil
			}
		},
	},
}
