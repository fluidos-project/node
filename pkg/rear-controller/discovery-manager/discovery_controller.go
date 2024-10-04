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

package discoverymanager

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	advertisementv1alpha1 "github.com/fluidos-project/node/apis/advertisement/v1alpha1"
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	gateway "github.com/fluidos-project/node/pkg/rear-controller/gateway"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
	"github.com/fluidos-project/node/pkg/utils/tools"
)

// DiscoveryReconciler reconciles a Discovery object.
type DiscoveryReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Gateway *gateway.Gateway
}

// clusterRole
//+kubebuilder:rbac:groups=advertisement.fluidos.eu,resources=peeringcandidates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=advertisement.fluidos.eu,resources=peeringcandidates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=advertisement.fluidos.eu,resources=peeringcandidates/finalizers,verbs=update
//+kubebuilder:rbac:groups=advertisement.fluidos.eu,resources=discoveries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=advertisement.fluidos.eu,resources=discoveries/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=advertisement.fluidos.eu,resources=discoveries/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *DiscoveryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "discovery", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	var peeringCandidate *advertisementv1alpha1.PeeringCandidate

	var discovery advertisementv1alpha1.Discovery
	if err := r.Get(ctx, req.NamespacedName, &discovery); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting Discovery %s before reconcile: %s", req.NamespacedName, err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("Discovery %s not found, probably deleted", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	klog.Infof("Discovery %s started", discovery.Name)

	if discovery.Status.Phase.Phase != nodecorev1alpha1.PhaseSolved &&
		discovery.Status.Phase.Phase != nodecorev1alpha1.PhaseTimeout &&
		discovery.Status.Phase.Phase != nodecorev1alpha1.PhaseFailed &&
		discovery.Status.Phase.Phase != nodecorev1alpha1.PhaseRunning &&
		discovery.Status.Phase.Phase != nodecorev1alpha1.PhaseIdle {
		discovery.Status.Phase.StartTime = tools.GetTimeNow()
		discovery.Status.PeeringCandidateList.Items = []advertisementv1alpha1.PeeringCandidate{}
		discovery.SetPhase(nodecorev1alpha1.PhaseRunning, "Discovery started")

		if err := r.updateDiscoveryStatus(ctx, &discovery); err != nil {
			klog.Errorf("Error when updating Discovery %s status before reconcile: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	//nolint:exhaustive // We don't need to handle all the cases
	switch discovery.Status.Phase.Phase {
	case nodecorev1alpha1.PhaseRunning:
		klog.Infof("Discovery %s running", discovery.Name)
		flavors, err := r.Gateway.DiscoverFlavors(ctx, discovery.Spec.Selector)
		if err != nil {
			klog.Errorf("Error when getting Flavor: %s", err)
			discovery.SetPhase(nodecorev1alpha1.PhaseFailed, "Error when getting Flavor")
			if err := r.updateDiscoveryStatus(ctx, &discovery); err != nil {
				klog.Errorf("Error when updating Discovery %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		if len(flavors) == 0 {
			klog.Infof("No Flavors found")
			discovery.SetPhase(nodecorev1alpha1.PhaseFailed, "No Flavors found")
			if err := r.updateDiscoveryStatus(ctx, &discovery); err != nil {
				klog.Errorf("Error when updating Discovery %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		klog.Infof("Flavors found: %d", len(flavors))

		for _, flavor := range flavors {
			peeringCandidate = resourceforge.ForgePeeringCandidate(flavor, discovery.Spec.SolverID, true)
			err = r.Create(context.Background(), peeringCandidate)
			if err != nil {
				klog.Infof("Discovery %s failed: error while creating Peering Candidate", discovery.Name)
				return ctrl.Result{}, err
			}
			peeringCandidate.Status.CreationTime = tools.GetTimeNow()
			if err := r.Status().Update(ctx, peeringCandidate); err != nil {
				klog.Errorf("Error when updating PeeringCandidate %s status before reconcile: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			// Append the PeeringCandidate to the list of PeeringCandidates found by the Discovery
			discovery.Status.PeeringCandidateList.Items = append(discovery.Status.PeeringCandidateList.Items, *peeringCandidate)
		}

		discovery.SetPhase(nodecorev1alpha1.PhaseSolved, "Discovery Solved: Peering Candidate found")
		if err := r.updateDiscoveryStatus(ctx, &discovery); err != nil {
			klog.Errorf("Error when updating Discovery %s: %s", discovery.Name, err)
			return ctrl.Result{}, err
		}
		klog.Infof("Discovery %s updated", discovery.Name)

		return ctrl.Result{}, nil

	case nodecorev1alpha1.PhaseSolved:
		klog.Infof("Discovery %s solved", discovery.Name)
	case nodecorev1alpha1.PhaseFailed:
		klog.Infof("Discovery %s failed", discovery.Name)
	}

	return ctrl.Result{}, nil
}

// updateDiscoveryStatus updates the status of the discovery.
func (r *DiscoveryReconciler) updateDiscoveryStatus(ctx context.Context, discovery *advertisementv1alpha1.Discovery) error {
	return r.Status().Update(ctx, discovery)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DiscoveryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&advertisementv1alpha1.Discovery{}).
		Complete(r)
}
