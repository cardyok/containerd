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

var blkioMetrics = []*Metric{
	{
		name:   "blkio_io_merged_recursive",
		help:   "The blkio io merged recursive",
		unit:   metrics.Total,
		vt:     prometheus.GaugeValue,
		labels: []string{"op", "device", "major", "minor"},
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Blkio == nil {
					return nil
				}
				return blkioValues(s.Blkio.IoMergedRecursive)
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name:   "blkio_io_queued_recursive",
		help:   "The blkio io queued recursive",
		unit:   metrics.Total,
		vt:     prometheus.GaugeValue,
		labels: []string{"op", "device", "major", "minor"},
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Blkio == nil {
					return nil
				}
				return blkioValues(s.Blkio.IoQueuedRecursive)
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name:   "blkio_io_service_bytes_recursive",
		help:   "The blkio io service bytes recursive",
		unit:   metrics.Bytes,
		vt:     prometheus.GaugeValue,
		labels: []string{"op", "device", "major", "minor"},
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Blkio == nil {
					return nil
				}
				return blkioValues(s.Blkio.IoServiceBytesRecursive)
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name:   "blkio_io_service_time_recursive",
		help:   "The blkio io service time recursive",
		unit:   metrics.Total,
		vt:     prometheus.GaugeValue,
		labels: []string{"op", "device", "major", "minor"},
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Blkio == nil {
					return nil
				}
				return blkioValues(s.Blkio.IoServiceTimeRecursive)
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name:   "blkio_io_serviced_recursive",
		help:   "The blkio io serviced recursive",
		unit:   metrics.Total,
		vt:     prometheus.GaugeValue,
		labels: []string{"op", "device", "major", "minor"},
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Blkio == nil {
					return nil
				}
				return blkioValues(s.Blkio.IoServicedRecursive)
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name:   "blkio_io_time_recursive",
		help:   "The blkio io time recursive",
		unit:   metrics.Total,
		vt:     prometheus.GaugeValue,
		labels: []string{"op", "device", "major", "minor"},
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Blkio == nil {
					return nil
				}
				return blkioValues(s.Blkio.IoTimeRecursive)
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
	{
		name:   "blkio_sectors_recursive",
		help:   "The blkio sectors recursive",
		unit:   metrics.Total,
		vt:     prometheus.GaugeValue,
		labels: []string{"op", "device", "major", "minor"},
		getValues: func(stats interface{}) []value {
			switch stats.(type) {
			case *v1.Metrics:
				s, _ := stats.(*v1.Metrics)
				if s.Blkio == nil {
					return nil
				}
				return blkioValues(s.Blkio.SectorsRecursive)
			case *v2.Metrics:
				return nil
			default:
				return nil
			}
		},
	},
}

func blkioValues(l []*v1.BlkIOEntry) []value {
	var out []value
	for _, e := range l {
		out = append(out, value{
			v: float64(e.Value),
			l: []string{e.Op, e.Device, strconv.FormatUint(e.Major, 10), strconv.FormatUint(e.Minor, 10)},
		})
	}
	return out
}
