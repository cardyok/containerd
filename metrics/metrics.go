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

package metrics

import (
	"github.com/containerd/containerd/version"
	goMetrics "github.com/docker/go-metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ContainerdVersion   goMetrics.LabeledCounter
	ImagePulls          goMetrics.LabeledCounter
	ImagePulledSize     goMetrics.LabeledCounter
	HostPulledSize      goMetrics.LabeledCounter
	InProgressPulls     goMetrics.LabeledGauge
	ImageResolveFailure goMetrics.LabeledCounter
	ImagePullSpeed      *prometheus.HistogramVec
)

func init() {
	ns := goMetrics.NewNamespace("containerd", "containerd", nil)

	ContainerdVersion = ns.NewLabeledCounter("containerd_version", "containerd version summary", "containerd_version")

	ImagePulledSize = ns.NewLabeledCounter("proxy_throughput_summary", "traffic summary for each proxy", "proxy")
	HostPulledSize = ns.NewLabeledCounter("host_throughput_summary", "traffic summary for each host", "host")
	ImagePulls = ns.NewLabeledCounter("image_pulls", "succeeded and failed counters", "status", "error", "host")
	ImageResolveFailure = ns.NewLabeledCounter("image_resolve_failure", "image resolve failure count", "registry", "path", "code")
	InProgressPulls = ns.NewLabeledGauge("in_progress_pull", "in progress image pulls", goMetrics.Total, "registry", "path")
	ImagePullSpeed = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Buckets:   []float64{0.1, 1, 5, 10, 30},
		Namespace: "containerd",
		Subsystem: "containerd",
		Name:      "image_pull_time",
		Help:      "average time to pull 1MB image",
	}, []string{"hosts"})
	ns.Add(ImagePullSpeed)

	c := ns.NewLabeledCounter("build_info", "containerd build information", "version", "revision")
	c.WithValues(version.Version, version.Revision).Inc()
	goMetrics.Register(ns)
}
