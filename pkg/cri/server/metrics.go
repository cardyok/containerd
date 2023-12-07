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

package server

import (
	"strings"
	"time"

	metrics "github.com/docker/go-metrics"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"

	containerstore "github.com/containerd/containerd/pkg/cri/store/container"
	imagestore "github.com/containerd/containerd/pkg/cri/store/image"
	sandboxstore "github.com/containerd/containerd/pkg/cri/store/sandbox"
)

var (
	sandboxListTimer                   metrics.Timer
	sandboxCreateNetworkTimer          metrics.Timer
	sandboxCreateNetworkTimeoutCounter metrics.Counter
	sandboxCreateNetworkFailureCounter metrics.Counter
	sandboxDeleteNetwork               metrics.Timer

	sandboxRuntimeCreateTimer metrics.LabeledTimer
	sandboxRuntimeStopTimer   metrics.LabeledTimer
	sandboxRemoveTimer        metrics.LabeledTimer
	sandboxCountReady         metrics.LabeledGauge
	sandboxCountNotReady      metrics.LabeledGauge

	containerListTimer       metrics.Timer
	containerRemoveTimer     metrics.LabeledTimer
	containerCreateTimer     metrics.LabeledTimer
	containerStopTimer       metrics.LabeledTimer
	containerStartTimer      metrics.LabeledTimer
	containerExecsyncTimeout metrics.Counter
	containerCountNotRunning metrics.LabeledGauge
	containerCountRunning    metrics.LabeledGauge

	imageHostCount     metrics.LabeledGauge
	imageSnapshotCount metrics.LabeledGauge

	networkPluginOperations        metrics.LabeledCounter
	networkPluginOperationsErrors  metrics.LabeledCounter
	networkPluginOperationsLatency metrics.LabeledTimer
)

func init() {
	// these CRI metrics record latencies for successful operations around a sandbox and container's lifecycle.
	ns := metrics.NewNamespace("containerd", "cri", nil)

	sandboxListTimer = ns.NewTimer("sandbox_list", "time to list sandboxes")
	sandboxCreateNetworkTimer = ns.NewTimer("sandbox_create_network", "time to create the network for a sandbox")
	sandboxDeleteNetwork = ns.NewTimer("sandbox_delete_network", "time to delete a sandbox's network")

	sandboxRuntimeCreateTimer = ns.NewLabeledTimer("sandbox_runtime_create", "time to create a sandbox in the runtime", "runtime")
	sandboxRuntimeStopTimer = ns.NewLabeledTimer("sandbox_runtime_stop", "time to stop a sandbox", "runtime")
	sandboxRemoveTimer = ns.NewLabeledTimer("sandbox_remove", "time to remove a sandbox", "runtime")

	containerListTimer = ns.NewTimer("container_list", "time to list containers")
	containerRemoveTimer = ns.NewLabeledTimer("container_remove", "time to remove a container", "runtime")
	containerCreateTimer = ns.NewLabeledTimer("container_create", "time to create a container", "runtime")
	containerStopTimer = ns.NewLabeledTimer("container_stop", "time to stop a container", "runtime")
	containerStartTimer = ns.NewLabeledTimer("container_start", "time to start a container", "runtime")

	networkPluginOperations = ns.NewLabeledCounter("network_plugin_operations_total", "cumulative number of network plugin operations by operation type", "operation_type")
	networkPluginOperationsErrors = ns.NewLabeledCounter("network_plugin_operations_errors_total", "cumulative number of network plugin operations by operation type", "operation_type")
	networkPluginOperationsLatency = ns.NewLabeledTimer("network_plugin_operations_duration_seconds", "latency in seconds of network plugin operations. Broken down by operation type", "operation_type")

	sandboxCreateNetworkTimeoutCounter = ns.NewCounter("sandbox_create_network_timeout", "netns create timeout counts")
	sandboxCreateNetworkFailureCounter = ns.NewCounter("sandbox_create_network_failure", "netns create failure counts")

	containerExecsyncTimeout = ns.NewCounter("container_execsync_timeout_count", "container exec timeout counts")
	imageHostCount = ns.NewLabeledGauge("image_host_count", "image host count on this node", metrics.Total, "image_host")
	imageSnapshotCount = ns.NewLabeledGauge("image_snapshot_count", "image snapshot count on this node", metrics.Total, "snapshotter")
	sandboxCountNotReady = ns.NewLabeledGauge("sandbox_count_not_ready", "sandbox count not ready", metrics.Total)
	sandboxCountReady = ns.NewLabeledGauge("sandbox_count_ready", "sandbox count ready", metrics.Total)
	containerCountNotRunning = ns.NewLabeledGauge("container_count_not_running", "container count not running", metrics.Total)
	containerCountRunning = ns.NewLabeledGauge("container_count_running", "container count running", metrics.Total)
	metrics.Register(ns)
}

// for backwards compatibility with kubelet/dockershim metrics
// https://github.com/containerd/containerd/issues/7801
const (
	networkStatusOp   = "get_pod_network_status"
	networkSetUpOp    = "set_up_pod"
	networkTearDownOp = "tear_down_pod"
)

// metricsSyncer syncs metrics periodically.
type metricsSyncer struct {
	containerStore *containerstore.Store
	sandboxStore   *sandboxstore.Store
	imageStore     *imagestore.Store
	syncPeriod     time.Duration
}

// newMetricsSyncer creates a metrics syncer.
func newMetricsSyncer(containerStore *containerstore.Store, sandboxStore *sandboxstore.Store, imageStore *imagestore.Store, period time.Duration) *metricsSyncer {
	return &metricsSyncer{
		containerStore: containerStore,
		sandboxStore:   sandboxStore,
		imageStore:     imageStore,
		syncPeriod:     period,
	}
}

func (ms *metricsSyncer) start() {
	go func() {
		for range time.Tick(ms.syncPeriod) {
			ms.sync()
		}
	}()
}

func (ms *metricsSyncer) sync() {
	// sandbox Metrics
	sandboxes := ms.sandboxStore.List()
	notReadySandbox := 0
	readySandbox := 0
	for _, s := range sandboxes {
		if s.Status.Get().State != sandboxstore.StateReady {
			notReadySandbox += 1
		} else {
			readySandbox += 1
		}
	}

	// container Metrics
	containers := ms.containerStore.List()
	notRunningContainer := 0
	runningContainer := 0
	for _, c := range containers {
		if c.Status.Get().State() != runtime.ContainerState_CONTAINER_RUNNING {
			notRunningContainer += 1
		} else {
			runningContainer += 1
		}
	}

	// image Metrics
	snapshotCounts := make(map[string]int)
	hostCounts := make(map[string]int)
	images := ms.imageStore.List()
	for _, i := range images {
		for _, sn := range i.Snapshotters {
			snapshotCounts[sn] += 1
		}
		repoTags, repoDigests := parseImageReferences(i.References)
		imageName, _ := normalizeRepoDigest(repoDigests)
		repoTagPairs := normalizeRepoTagPair(repoTags, imageName)
		for _, repoTag := range repoTagPairs {
			hostCounts[strings.Split(repoTag[0], "/")[0]] += 1
		}
	}

	containerCountRunning.WithValues().Set(float64(runningContainer))
	containerCountNotRunning.WithValues().Set(float64(notRunningContainer))
	sandboxCountReady.WithValues().Set(float64(readySandbox))
	sandboxCountNotReady.WithValues().Set(float64(notReadySandbox))
	for snapshotter, count := range snapshotCounts {
		imageSnapshotCount.WithValues(snapshotter).Set(float64(count))
	}
	for host, count := range hostCounts {
		imageHostCount.WithValues(host).Set(float64(count))
	}
}

// Ideally repo tag should always be image:tag.
// The repoTags is nil when pulling image by repoDigest,Then we will show image name instead.
func normalizeRepoTagPair(repoTags []string, imageName string) (repoTagPairs [][]string) {
	const none = "<none>"
	if len(repoTags) == 0 {
		repoTagPairs = append(repoTagPairs, []string{imageName, none})
		return
	}
	for _, repoTag := range repoTags {
		idx := strings.LastIndex(repoTag, ":")
		if idx == -1 {
			repoTagPairs = append(repoTagPairs, []string{"errorRepoTag", "errorRepoTag"})
			continue
		}
		name := repoTag[:idx]
		if name == none {
			name = imageName
		}
		repoTagPairs = append(repoTagPairs, []string{name, repoTag[idx+1:]})
	}
	return
}

func normalizeRepoDigest(repoDigests []string) (string, string) {
	if len(repoDigests) == 0 {
		return "<none>", "<none>"
	}
	repoDigestPair := strings.Split(repoDigests[0], "@")
	if len(repoDigestPair) != 2 {
		return "errorName", "errorRepoDigest"
	}
	return repoDigestPair[0], repoDigestPair[1]
}
