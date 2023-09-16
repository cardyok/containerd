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

var ioMetrics = []*Metric{
	{
		name:   "io_rbytes",
		help:   "IO bytes read",
		unit:   metrics.Bytes,
		vt:     prometheus.GaugeValue,
		labels: []string{"major", "minor"},
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Io == nil {
					return nil
				}
				var out []value
				for _, e := range s.Io.Usage {
					out = append(out, value{
						v: float64(e.Rbytes),
						l: []string{strconv.FormatUint(e.Major, 10), strconv.FormatUint(e.Minor, 10)},
					})
				}
				return out
			default:
				return nil
			}
		},
	},
	{
		name:   "io_wbytes",
		help:   "IO bytes written",
		unit:   metrics.Bytes,
		vt:     prometheus.GaugeValue,
		labels: []string{"major", "minor"},
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Io == nil {
					return nil
				}
				var out []value
				for _, e := range s.Io.Usage {
					out = append(out, value{
						v: float64(e.Wbytes),
						l: []string{strconv.FormatUint(e.Major, 10), strconv.FormatUint(e.Minor, 10)},
					})
				}
				return out
			default:
				return nil
			}
		},
	},
	{
		name:   "io_rios",
		help:   "Number of read IOs",
		unit:   metrics.Total,
		vt:     prometheus.GaugeValue,
		labels: []string{"major", "minor"},
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Io == nil {
					return nil
				}
				var out []value
				for _, e := range s.Io.Usage {
					out = append(out, value{
						v: float64(e.Rios),
						l: []string{strconv.FormatUint(e.Major, 10), strconv.FormatUint(e.Minor, 10)},
					})
				}
				return out
			default:
				return nil
			}
		},
	},
	{
		name:   "io_wios",
		help:   "Number of write IOs",
		unit:   metrics.Total,
		vt:     prometheus.GaugeValue,
		labels: []string{"major", "minor"},
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				return nil
			case *v2.Metrics:
				s, _ := stats.(*v2.Metrics)
				if s.Io == nil {
					return nil
				}
				var out []value
				for _, e := range s.Io.Usage {
					out = append(out, value{
						v: float64(e.Wios),
						l: []string{strconv.FormatUint(e.Major, 10), strconv.FormatUint(e.Minor, 10)},
					})
				}
				return out
			default:
				return nil
			}
		},
	},
}
