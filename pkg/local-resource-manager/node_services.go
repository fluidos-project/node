// Copyright 2022-2023 FLUIDOS Project
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
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/fluidos-project/node/pkg/utils/models"
)

// GetNodesResources retrieves the metrics from all the worker nodes in the cluster.
func GetNodesResources(ctx context.Context, cl client.Client) ([]models.NodeInfo, error) {
	// Set a label selector to filter worker nodes
	labelSelector := labels.Set{flags.ResourceNodeLabel: "true"}.AsSelector()

	// Get a list of nodes
	nodes := &corev1.NodeList{}
	err := cl.List(ctx, nodes, &client.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		klog.Errorf("Error when listing nodes: %s", err)
		return nil, err
	}

	// Get a list of nodes metrics
	nodesMetrics := &metricsv1beta1.NodeMetricsList{}
	err = cl.List(ctx, nodesMetrics, &client.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		klog.Errorf("Error when listing nodes metrics: %s", err)
		return nil, err
	}

	var nodesInfo []models.NodeInfo
	// Print the name of each node
	for n := range nodes.Items {
		for m := range nodesMetrics.Items {
			node := nodes.Items[n]
			metrics := nodesMetrics.Items[m]
			if nodes.Items[n].Name != nodesMetrics.Items[m].Name {
				// So that we can select just the nodes that we want
				continue
			}
			metricsStruct := forgeResourceMetrics(&metrics, &node)
			nodeInfo := forgeNodeInfo(&node, metricsStruct)
			nodesInfo = append(nodesInfo, *nodeInfo)
		}
	}

	return nodesInfo, nil
}

// forgeResourceMetrics creates from params a new ResourceMetrics Struct.
func forgeResourceMetrics(nodeMetrics *metricsv1beta1.NodeMetrics, node *corev1.Node) *models.ResourceMetrics {
	// Get the total and used resources
	cpuTotal := node.Status.Allocatable.Cpu()
	cpuUsed := nodeMetrics.Usage.Cpu()
	memoryTotal := node.Status.Allocatable.Memory()
	memoryUsed := nodeMetrics.Usage.Memory()
	podsTotal := node.Status.Allocatable.Pods()
	podsUsed := nodeMetrics.Usage.Pods()
	ephemeralStorage := nodeMetrics.Usage.StorageEphemeral()

	// Compute the available resources
	cpuAvail := cpuTotal.DeepCopy()
	memAvail := memoryTotal.DeepCopy()
	podsAvail := podsTotal.DeepCopy()
	cpuAvail.Sub(*cpuUsed)
	memAvail.Sub(*memoryUsed)
	podsAvail.Sub(*podsUsed)

	return &models.ResourceMetrics{
		CPUTotal:         *cpuTotal,
		CPUAvailable:     cpuAvail,
		MemoryTotal:      *memoryTotal,
		MemoryAvailable:  memAvail,
		PodsTotal:        *podsTotal,
		PodsAvailable:    podsAvail,
		EphemeralStorage: *ephemeralStorage,
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
