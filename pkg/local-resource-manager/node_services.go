// Copyright 2022-2024 FLUIDOS Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package localresourcemanager

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"github.com/fluidos-project/node/pkg/utils/models"
)

// GetNodeInfos returns the NodeInfo struct for a given node and its metrics.
func GetNodeInfos(node *corev1.Node, nodeMetrics *metricsv1beta1.NodeMetrics) (*models.NodeInfo, error) {
	// Check if the node and the node metrics match
	if node.Name != nodeMetrics.Name {
		klog.Info("Node and NodeMetrics do not match")
		return nil, fmt.Errorf("node and node metrics do not match")
	}

	metricsStruct := forgeResourceMetrics(nodeMetrics, node)
	nodeInfo := forgeNodeInfo(node, metricsStruct)

	return nodeInfo, nil
}

// forgeResourceMetrics creates from params a new ResourceMetrics Struct.
func forgeResourceMetrics(nodeMetrics *metricsv1beta1.NodeMetrics, node *corev1.Node) *models.ResourceMetrics {
	// Get the total and used resources
	cpuTotal := node.Status.Allocatable.Cpu().DeepCopy()
	cpuUsed := nodeMetrics.Usage.Cpu().DeepCopy()
	memoryTotal := node.Status.Allocatable.Memory().DeepCopy()
	memoryUsed := nodeMetrics.Usage.Memory().DeepCopy()
	podsTotal := node.Status.Allocatable.Pods().DeepCopy()
	podsUsed := nodeMetrics.Usage.Pods().DeepCopy()
	ephemeralStorage := nodeMetrics.Usage.StorageEphemeral().DeepCopy()

	// Compute the available resources
	cpuAvail := cpuTotal.DeepCopy()
	memAvail := memoryTotal.DeepCopy()
	podsAvail := podsTotal.DeepCopy()
	cpuAvail.Sub(cpuUsed)
	memAvail.Sub(memoryUsed)
	podsAvail.Sub(podsUsed)

	return &models.ResourceMetrics{
		CPUTotal:         cpuTotal,
		CPUAvailable:     cpuAvail,
		MemoryTotal:      memoryTotal,
		MemoryAvailable:  memAvail,
		PodsTotal:        podsTotal,
		PodsAvailable:    podsAvail,
		EphemeralStorage: ephemeralStorage,
	}
}

// forgeNodeInfo creates from params a new NodeInfo struct.
func forgeNodeInfo(node *corev1.Node, metrics *models.ResourceMetrics) *models.NodeInfo {
	return &models.NodeInfo{
		UID:             string(node.UID),
		Name:            node.Name,
		Architecture:    node.Status.NodeInfo.Architecture,
		OperatingSystem: node.Status.NodeInfo.OperatingSystem,
		ResourceMetrics: *metrics,
	}
}
