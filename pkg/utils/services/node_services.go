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

package services

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/consts"
	"github.com/fluidos-project/node/pkg/utils/models"
)

var (
	scheme = runtime.NewScheme()
	uids   []string
)

func init() {
	_ = metricsv1beta1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = nodecorev1alpha1.AddToScheme(scheme)
}

// GetKClient creates a kubernetes API client and returns it.
func GetKClient(ctx context.Context) (client.Client, error) {
	config := ctrl.GetConfigOrDie()

	cl, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		klog.Fatalf("error creating manager: %", err)
	}

	return cl, nil
}

// GetNodesResources retrieves the metrics from all the worker nodes in the cluster
func GetNodesResources(ctx context.Context, cl client.Client) (*[]models.NodeInfo, error) {
	// Set a label selector to filter worker nodes
	labelSelector := labels.Set{consts.WORKER_LABEL_KEY: "true"}.AsSelector()

	// Get a list of nodes
	nodes := &corev1.NodeList{}
	err := cl.List(ctx, nodes, &client.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	// Get a list of nodes metrics
	nodesMetrics := &metricsv1beta1.NodeMetricsList{}
	err = cl.List(ctx, nodesMetrics, &client.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	var nodesInfo []models.NodeInfo
	// Print the name of each node
	for _, node := range nodes.Items {
		for _, metrics := range nodesMetrics.Items {
			if node.Name != metrics.Name {
				// So that we can select just the nodes that we want
				continue
			}

			metricsStruct := getNodeResourceMetrics(&metrics, &node)
			nodeInfo := getNodeInfo(&node, metricsStruct)
			nodesInfo = append(nodesInfo, *nodeInfo)
			uids = append(uids, string(node.UID))
		}
	}

	return &nodesInfo, nil
}

// getNodeResourceMetrics gets a ResourceMetrics struct
func getNodeResourceMetrics(nodeMetrics *metricsv1beta1.NodeMetrics, node *corev1.Node) *models.ResourceMetrics {
	cpuTotal := node.Status.Allocatable.Cpu()
	cpuUsed := nodeMetrics.Usage.Cpu()
	memoryTotal := node.Status.Allocatable.Memory()
	memoryUsed := nodeMetrics.Usage.Memory()
	ephemeralStorage := nodeMetrics.Usage.StorageEphemeral()
	return fromResourceMetrics(*cpuTotal, *cpuUsed, *memoryTotal, *memoryUsed, *ephemeralStorage)
}

// getNodeInfo gets a NodeInfo struct
func getNodeInfo(node *corev1.Node, metrics *models.ResourceMetrics) *models.NodeInfo {
	return fromNodeInfo(string(node.UID), node.Name, node.Status.NodeInfo.Architecture, node.Status.NodeInfo.OperatingSystem, *metrics)
}

// fromNodeInfo creates from params a new NodeInfo Struct
func fromNodeInfo(uid, name, arch, os string, metrics models.ResourceMetrics) *models.NodeInfo {
	return &models.NodeInfo{
		UID:             uid,
		Name:            name,
		Architecture:    arch,
		OperatingSystem: os,
		ResourceMetrics: metrics,
	}
}

// fromResourceMetrics creates from params a new ResourceMetrics Struct
func fromResourceMetrics(cpuTotal, cpuUsed, memoryTotal, memoryUsed, ephStorage resource.Quantity) *models.ResourceMetrics {
	cpuAvail := cpuTotal.DeepCopy()
	memAvail := memoryTotal.DeepCopy()

	cpuAvail.Sub(cpuUsed)
	memAvail.Sub(memoryUsed)

	return &models.ResourceMetrics{
		CPUTotal:         cpuTotal,
		CPUAvailable:     cpuAvail,
		MemoryTotal:      memoryTotal,
		MemoryAvailable:  memAvail,
		EphemeralStorage: ephStorage,
	}
}

// compareByCPUAvailable is a custom comparison function based on CPUAvailable
func compareByCPUAvailable(node1, node2 models.NodeInfo) bool {
	cmpResult := node1.ResourceMetrics.CPUAvailable.Cmp(node2.ResourceMetrics.CPUAvailable)
	if cmpResult == 1 {
		return true
	} else {
		return false
	}
}

// maxNode find the node with the maximum value based on the provided comparison function
func maxNode(nodes []models.NodeInfo, compareFunc func(models.NodeInfo, models.NodeInfo) bool) models.NodeInfo {
	if len(nodes) == 0 {
		panic("Empty node list")
	}

	maxNode := nodes[0]
	for _, node := range nodes {
		if compareFunc(node, maxNode) {
			maxNode = node
		}
	}
	return maxNode
}
