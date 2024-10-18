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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
)

// ClusterRole
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
// +kubebuilder:rbac:groups=metrics.k8s.io,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=metrics.k8s.io,resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=serviceblueprints,verbs=get;list;watch;create;update;patch;delete

// ServiceBlueprintReconciler reconciles a ServiceBlueprint object and creates Flavor objects.
type ServiceBlueprintReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	EnableAutoDiscovery bool
	WebhookServer       webhook.Server
}

// Reconcile reconciles a Node object to create Flavor objects.
func (r *ServiceBlueprintReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "serviceblueprint", req.NamespacedName)
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

	// Fetch the ServiceBlueprint instance
	serviceBlueprint := &nodecorev1alpha1.ServiceBlueprint{}
	if err := r.Get(ctx, req.NamespacedName, serviceBlueprint); err != nil {
		// If the ServiceBlueprint is not found, check if it's a Flavor
		flavor := &nodecorev1alpha1.Flavor{}
		err = r.Get(ctx, req.NamespacedName, flavor)
		if err != nil {
			klog.Errorf("Error getting ServiceBlueprint or Flavor: %v", err)
			return ctrl.Result{}, err
		}
		// If it's a Flavor, handle accordingly (e.g., update or requeue)
		klog.Infof("Flavor %s triggered the reconcile loop", flavor.Name)
		// Check the Flavor Type
		// Ignore if it's not a Service Flavor
		if flavor.Spec.FlavorType.TypeIdentifier != nodecorev1alpha1.TypeService {
			klog.Infof("Flavor %s is not a Service Flavor", flavor.Name)
			return ctrl.Result{}, nil
		}
		// The Flavor's owner is the ServiceBlueprint
		serviceBlueprint = &nodecorev1alpha1.ServiceBlueprint{}
		err = r.Get(ctx, client.ObjectKey{Name: flavor.OwnerReferences[0].Name, Namespace: flavor.Namespace}, serviceBlueprint)
		if err != nil {
			klog.Errorf("Error getting ServiceBlueprint: %v", err)
			return ctrl.Result{}, err
		}
		// TODO: Add logic to manage the Flavor associated with the ServiceBlueprint
		return ctrl.Result{}, nil
	}

	// Create ownerReferences with only the current node under examination
	ownerReferences := []metav1.OwnerReference{
		{
			APIVersion: nodecorev1alpha1.GroupVersion.String(),
			Kind:       "ServiceBlueprint",
			Name:       serviceBlueprint.Name,
			UID:        serviceBlueprint.UID,
		},
	}

	// Get all the Flavors owned by this node as kubernetes ownership
	flavorsList := &nodecorev1alpha1.FlavorList{}
	err := r.List(ctx, flavorsList)
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
			if owner.Name == serviceBlueprint.Name {
				// Add the Flavor to the list
				matchFlavors = append(matchFlavors, *flavor)
			}
		}
	}

	// Check if you have found any Flavor
	if len(matchFlavors) > 0 {
		klog.Infof("Found %d flavors for node %s", len(matchFlavors), serviceBlueprint.Name)
		// TODO: Check if the Flavors are consistent with the NodeInfo
		// TODO: Update the Flavors if necessary
		return ctrl.Result{}, nil
	}

	// Get NodeIdentity
	nodeIdentity := getters.GetNodeIdentity(ctx, r.Client)
	if nodeIdentity == nil {
		klog.Error("Error getting FLUIDOS Node identity")
		return ctrl.Result{}, nil
	}

	// No Flavor found, create a new one
	flavor, err := r.createFlavor(ctx, serviceBlueprint, nodeIdentity, ownerReferences)
	if err != nil {
		klog.Errorf("Error creating Flavor: %v", err)
		return ctrl.Result{Requeue: true}, nil
	}
	klog.Infof("Flavor created: %s", flavor.Name)

	return ctrl.Result{}, nil
}

func (r *ServiceBlueprintReconciler) createFlavor(ctx context.Context, serviceBlueprint *nodecorev1alpha1.ServiceBlueprint,
	nodeIdentity *nodecorev1alpha1.NodeIdentity,
	ownerReferences []metav1.OwnerReference) (flavor *nodecorev1alpha1.Flavor, err error) {
	// Forge the Flavor from the NodeInfo and NodeIdentity
	flavorResult := resourceforge.ForgeServiceFlavorFromBlueprint(serviceBlueprint, nodeIdentity, ownerReferences)

	// Create the Flavor
	err = r.Create(ctx, flavorResult)
	if err != nil {
		return nil, err
	}
	klog.Infof("Flavor created: %s", flavorResult.Name)

	return flavorResult, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceBlueprintReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nodecorev1alpha1.ServiceBlueprint{}).
		Watches(&nodecorev1alpha1.Flavor{}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
