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
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/fluidos-project/node/pkg/utils/getters"
	models "github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
)

// ClusterRole
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
// +kubebuilder:rbac:groups=metrics.k8s.io,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=metrics.k8s.io,resources=nodes,verbs=get;list;watch

// NodeReconciler reconciles a Node object and creates Flavor objects.
type NodeReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	EnableAutoDiscovery bool
	WebhookServer       webhook.Server
}

// Reconcile reconciles a Node object to create Flavor objects.
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "node", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	// Check if AutoDiscovery is enabled
	if !r.EnableAutoDiscovery {
		klog.Info("AutoDiscovery is disabled")
		return ctrl.Result{}, nil
	}

	// Check if the webhook server is running
	if err := r.WebhookServer.StartedChecker()(nil); err != nil {
		klog.Info("Webhook server not started yet, requeuing the request")
		return ctrl.Result{Requeue: true}, nil
	}

	// Set for labels over the node
	labelSelector := labels.Set{flags.ResourceNodeLabel: "true"}.AsSelector()

	// Fetch the Node instance
	var node corev1.Node
	if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Info("Node not found")
			return ctrl.Result{}, nil
		}
	}

	// Check if the node has the label
	if !labelSelector.Matches(labels.Set(node.GetLabels())) {
		klog.Infof("Node %s does not have the label %s", node.Name, flags.ResourceNodeLabel)
		return ctrl.Result{}, nil
	}

	nodeMetrics := &metricsv1beta1.NodeMetrics{}

	// Get the node metrics referred to the node
	if err := r.Client.Get(ctx, client.ObjectKey{Name: node.Name}, nodeMetrics); err != nil {
		klog.Errorf("Error getting NodeMetrics: %v", err)
		return ctrl.Result{}, err
	}

	// Get the NodeInfo struct for the node and its metrics
	nodeInfo, err := GetNodeInfos(&node, nodeMetrics)
	if err != nil {
		klog.Errorf("Error getting NodeInfo: %v", err)
		return ctrl.Result{}, err
	}
	klog.Infof("NodeInfo created: %s", nodeInfo.Name)

	// Get NodeIdentity
	nodeIdentity := getters.GetNodeIdentity(ctx, r.Client)
	if nodeIdentity == nil {
		klog.Error("Error getting FLUIDOS Node identity")
		return ctrl.Result{}, nil
	}

	// Create ownerReferences with only the current node under examination
	ownerReferences := []metav1.OwnerReference{
		{
			APIVersion: nodecorev1alpha1.GroupVersion.String(),
			Kind:       "Node",
			Name:       node.Name,
			UID:        node.UID,
		},
	}

	// Get all the Flavors owned by this node as kubernetes ownership
	flavorsList := &nodecorev1alpha1.FlavorList{}
	err = r.List(ctx, flavorsList)
	if err != nil {
		klog.Errorf("Error listing Flavors: %v", err)
		return ctrl.Result{}, nil
	}

	var matchFlavors []nodecorev1alpha1.Flavor

	// Filter the Flavors by the owner reference
	for i := range flavorsList.Items {
		flavor := &flavorsList.Items[i]
		// Check if the node is one of the owners
		for _, owner := range flavor.OwnerReferences {
			if owner.Name == node.Name {
				// Add the Flavor to the list
				matchFlavors = append(matchFlavors, *flavor)
			}
		}
	}

	// Check if you have found any Flavor
	if len(matchFlavors) > 0 {
		klog.Infof("Found %d flavors for node %s", len(matchFlavors), node.Name)
		// TODO: Check if the Flavors are consistent with the NodeInfo
		// TODO: Update the Flavors if necessary
		return ctrl.Result{}, nil
	}

	// No Flavor found, create a new one
	flavor, err := r.createFlavor(ctx, nodeInfo, *nodeIdentity, ownerReferences)
	if err != nil {
		klog.Errorf("Error creating Flavor: %v", err)
		return ctrl.Result{Requeue: true}, nil
	}
	klog.Infof("Flavor created: %s", flavor.Name)

	return ctrl.Result{}, nil
}

func (r *NodeReconciler) createFlavor(ctx context.Context, nodeInfo *models.NodeInfo,
	nodeIdentity nodecorev1alpha1.NodeIdentity, ownerReferences []metav1.OwnerReference) (flavor *nodecorev1alpha1.Flavor, err error) {
	// Forge the Flavor from the NodeInfo and NodeIdentity
	flavorResult := resourceforge.ForgeK8SliceFlavorFromMetrics(nodeInfo, nodeIdentity, ownerReferences)

	// Create the Flavor
	err = r.Create(ctx, flavorResult)
	if err != nil {
		return nil, err
	}
	klog.Infof("Flavor created: %s", flavorResult.Name)

	return flavorResult, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Watches(&nodecorev1alpha1.Flavor{}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
