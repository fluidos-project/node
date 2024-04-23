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

package rearmanager

import (
	"context"

	liqodiscovery "github.com/liqotech/liqo/apis/discovery/v1alpha1"
	discovery "github.com/liqotech/liqo/pkg/discovery"
	fcutils "github.com/liqotech/liqo/pkg/utils/foreignCluster"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservation "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
	"github.com/fluidos-project/node/pkg/utils/services"
	virtualfabricmanager "github.com/fluidos-project/node/pkg/virtual-fabric-manager"
)

// clusterRole
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations/finalizers,verbs=update
// +kubebuilder:rbac:groups=discovery.liqo.io,resources=foreignclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.liqo.io,resources=foreignclusters/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.liqo.io,resources=foreignclusters/finalizers,verbs=get;update;patch

// AllocationReconciler reconciles a Allocation object.
type AllocationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *AllocationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "allocation", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	var allocation nodecorev1alpha1.Allocation
	if err := r.Get(ctx, req.NamespacedName, &allocation); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting Allocation %s before reconcile: %v", req.NamespacedName, err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("Allocation %s not found, probably deleted", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	if r.checkInitialStatus(&allocation) {
		if err := r.updateAllocationStatus(ctx, &allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	klog.Infof("Reconciling Allocation %s", req.NamespacedName)

	if allocation.Status.Status == nodecorev1alpha1.Error {
		klog.Infof("Allocation %s is in error state", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	if allocation.Spec.Type == nodecorev1alpha1.Node {
		return r.handleNodeAllocation(ctx, req, &allocation)
	}

	if allocation.Spec.Type == nodecorev1alpha1.VirtualNode && allocation.Spec.Destination == nodecorev1alpha1.Local {
		return r.handleVirtualNodeAllocation(ctx, req, &allocation)
	}

	return ctrl.Result{}, nil
}

func (r *AllocationReconciler) checkInitialStatus(allocation *nodecorev1alpha1.Allocation) bool {
	if allocation.Status.Status != nodecorev1alpha1.Active &&
		allocation.Status.Status != nodecorev1alpha1.Reserved &&
		allocation.Status.Status != nodecorev1alpha1.Released &&
		allocation.Status.Status != nodecorev1alpha1.Inactive {
		allocation.SetStatus(nodecorev1alpha1.Inactive, "Allocation has been set to Inactive")
		return true
	}
	return false
}

func (r *AllocationReconciler) handleNodeAllocation(ctx context.Context,
	req ctrl.Request, allocation *nodecorev1alpha1.Allocation) (ctrl.Result, error) {
	allocStatus := allocation.Status.Status
	switch allocStatus {
	case nodecorev1alpha1.Active:
		// We need to check if the ForeignCluster is still ready
		// If the ForeignCluster is not ready we need to set the Allocation to Released
		klog.Infof("Allocation %s is active", req.NamespacedName)
		return ctrl.Result{}, nil
	case nodecorev1alpha1.Reserved:
		if allocation.Spec.Destination == nodecorev1alpha1.Remote {
			// We need to check the status of the ForeignCluster
			// If the ForeignCluster is Ready the Allocation can be set to Active
			// else we need to wait for the ForeignCluster to be Ready
			klog.Infof("Allocation %s is reserved", req.NamespacedName)
			fc, err := fcutils.GetForeignClusterByID(ctx, r.Client, allocation.Spec.RemoteClusterID)
			// check if not found
			if err != nil {
				if apierrors.IsNotFound(err) {
					klog.Infof("ForeignCluster %s not found", allocation.Spec.RemoteClusterID)
					allocation.SetStatus(nodecorev1alpha1.Reserved, "ForeignCluster not found, peering not yet started")
				} else {
					klog.Errorf("Error when getting ForeignCluster %s: %v", allocation.Spec.RemoteClusterID, err)
					allocation.SetStatus(nodecorev1alpha1.Error, "Error when getting ForeignCluster")
				}
				if err := r.updateAllocationStatus(ctx, allocation); err != nil {
					klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, nil
			}
			if fcutils.IsIncomingJoined(fc) &&
				fcutils.IsNetworkingEstablishedOrExternal(fc) &&
				fcutils.IsAuthenticated(fc) &&
				!fcutils.IsUnpeered(fc) {
				klog.Infof("ForeignCluster %s is ready, incoming peering enstablished", allocation.Spec.RemoteClusterID)
				allocation.SetStatus(nodecorev1alpha1.Active, "Incoming peering ready, Allocation is now Active")
			} else {
				klog.Infof("ForeignCluster %s is not ready yet", allocation.Spec.RemoteClusterID)
				allocation.SetStatus(nodecorev1alpha1.Reserved, "Incoming peering not yet ready, Allocation is still Reserved")
			}
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		// We can set the Allocation to Active
		klog.Infof("Allocation will be used locally, we can put it in 'Active' State", req.NamespacedName)
		allocation.SetStatus(nodecorev1alpha1.Active, "Allocation ready, will be used locally")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	case nodecorev1alpha1.Released:
		// The Allocation is released,
		klog.Infof("Allocation %s is released", req.NamespacedName)
		// We need to check if the ForeignCluster is again ready
		return ctrl.Result{}, nil
	case nodecorev1alpha1.Inactive:
		// Alloction Type is Node, so we need to invalidate the Flavour
		// and eventually create a new one detaching the right Partition from the old one
		klog.Infof("Allocation %s is inactive", req.NamespacedName)

		// Get the contract related to the Allocation
		contract := &reservation.Contract{}
		if err := r.Client.Get(ctx, types.NamespacedName{
			Name:      allocation.Spec.Contract.Name,
			Namespace: allocation.Spec.Contract.Namespace,
		}, contract); err != nil {
			klog.Errorf("Error when getting Contract %s: %v", allocation.Spec.Contract.Name, err)
			allocation.SetStatus(nodecorev1alpha1.Error, "Error when getting Contract")
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		flavour, err := services.GetFlavourByID(contract.Spec.Flavour.Name, r.Client)
		if err != nil {
			klog.Errorf("Error when getting Flavour %s: %v", contract.Spec.Flavour.Name, err)
			allocation.SetStatus(nodecorev1alpha1.Error, "Error when getting Flavour")
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		flavourCopy := flavour.DeepCopy()
		flavourCopy.Spec.OptionalFields.Availability = false
		klog.Infof("Updating Flavour %s: Availability %t", flavourCopy.Name, flavourCopy.Spec.OptionalFields.Availability)
		// TODO: Known issue, availability will be not updated
		if err := r.Client.Update(ctx, flavourCopy); err != nil {
			klog.Errorf("Error when updating Flavour %s: %v", flavourCopy.Name, err)
			allocation.SetStatus(nodecorev1alpha1.Error, "Error when updating Flavour")
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		if contract.Spec.Partition != nil {
			// We need to create a new Flavour with the right Partition
			flavourRes := contract.Spec.Flavour.Spec.Characteristics
			allocationRes := computeResources(contract)

			newCharacteristics := computeCharacteristics(&flavourRes, allocationRes)
			newFlavour := resourceforge.ForgeFlavourFromRef(flavour, newCharacteristics)

			if err := r.Client.Create(ctx, newFlavour); err != nil {
				klog.Errorf("Error when creating Flavour %s: %v", newFlavour.Name, err)
				allocation.SetStatus(nodecorev1alpha1.Error, "Error when creating Flavour")
				if err := r.updateAllocationStatus(ctx, allocation); err != nil {
					klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, nil
			}
		}

		allocation.SetStatus(nodecorev1alpha1.Reserved, "Resources reserved")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	default:
		klog.Infof("Allocation %s is in an unknown state", req.NamespacedName)
		return ctrl.Result{}, nil
	}
}

func (r *AllocationReconciler) handleVirtualNodeAllocation(ctx context.Context,
	req ctrl.Request, allocation *nodecorev1alpha1.Allocation) (ctrl.Result, error) {
	allocStatus := allocation.Status.Status
	switch allocStatus {
	case nodecorev1alpha1.Active:
		// We need to check if the ForeignCluster is ready

		fc, err := fcutils.GetForeignClusterByID(ctx, r.Client, allocation.Spec.RemoteClusterID)
		// check if not found
		if err != nil {
			if apierrors.IsNotFound(err) {
				klog.Infof("ForeignCluster %s not found", allocation.Spec.RemoteClusterID)
				allocation.SetStatus(nodecorev1alpha1.Reserved, "ForeignCluster not found, peering not yet started")
			} else {
				klog.Errorf("Error when getting ForeignCluster %s: %v", allocation.Spec.RemoteClusterID, err)
				allocation.SetStatus(nodecorev1alpha1.Error, "Error when getting ForeignCluster")
			}
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		if fcutils.IsOutgoingJoined(fc) &&
			fcutils.IsAuthenticated(fc) &&
			fcutils.IsNetworkingEstablishedOrExternal(fc) &&
			!fcutils.IsUnpeered(fc) {
			klog.Infof("ForeignCluster %s is ready, outgoing peering enstablished", allocation.Spec.RemoteClusterID)
			allocation.SetStatus(nodecorev1alpha1.Active, "Outgoing peering ready, Allocation is Active")
		} else {
			klog.Infof("ForeignCluster %s is not ready yet", allocation.Spec.RemoteClusterID)
			allocation.SetStatus(nodecorev1alpha1.Active, "Outgoing peering not yet ready, Allocation is Active")
		}

		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	case nodecorev1alpha1.Reserved:

		klog.Infof("Allocation %s is reserved", req.NamespacedName)

		klog.Infof("Allocation %s is trying to enstablish a peering", req.NamespacedName.Name)

		klog.InfofDepth(1, "Allocation %s is retrieving credentials", req.NamespacedName)
		// Get the contract from the allocation
		contract := &reservation.Contract{}
		if err := r.Client.Get(ctx, types.NamespacedName{
			Name:      allocation.Spec.Contract.Name,
			Namespace: allocation.Spec.Contract.Namespace,
		}, contract); err != nil {
			klog.Errorf("Error when getting Contract %s: %v", allocation.Spec.Contract.Name, err)
			allocation.SetStatus(nodecorev1alpha1.Error, "Error when getting Contract")
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		klog.InfofDepth(1, "Allocation %s has retrieved contract %s from namespace %s", req.NamespacedName, contract.Name, contract.Namespace)

		credentials := contract.Spec.SellerCredentials

		klog.InfofDepth(1, "Allocation %s is peering with cluster %s", req.NamespacedName, credentials.ClusterName)
		_, err := virtualfabricmanager.PeerWithCluster(ctx, r.Client, credentials.ClusterID,
			credentials.ClusterName, credentials.Endpoint, credentials.Token)
		if err != nil {
			klog.Errorf("Error when peering with cluster %s: %s", credentials.ClusterName, err)
			allocation.SetStatus(nodecorev1alpha1.Error, "Error when peering with cluster "+credentials.ClusterName)
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
		klog.Infof("Allocation %s has started the peering with cluster %s", req.NamespacedName.Name, credentials.ClusterName)
		allocation.SetStatus(nodecorev1alpha1.Active, "Allocation is now Active")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	case nodecorev1alpha1.Released:
		// The Allocation is released,
		// We need to check if the ForeignCluster is again ready
		return ctrl.Result{}, nil
	case nodecorev1alpha1.Inactive:
		klog.Infof("Allocation %s is inactive", req.NamespacedName)
		allocation.SetStatus(nodecorev1alpha1.Reserved, "Allocation is now Reserved")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	default:
		klog.Infof("Allocation %s is in an unknown state", req.NamespacedName)
		return ctrl.Result{}, nil
	}
}

func computeResources(contract *reservation.Contract) *nodecorev1alpha1.Characteristics {
	if contract.Spec.Partition != nil {
		return &nodecorev1alpha1.Characteristics{
			Architecture:      contract.Spec.Partition.Architecture,
			Cpu:               contract.Spec.Partition.CPU,
			Memory:            contract.Spec.Partition.Memory,
			Pods:              contract.Spec.Partition.Pods,
			EphemeralStorage:  contract.Spec.Partition.EphemeralStorage,
			Gpu:               contract.Spec.Partition.Gpu,
			PersistentStorage: contract.Spec.Partition.Storage,
		}
	}
	return contract.Spec.Flavour.Spec.Characteristics.DeepCopy()
}

func computeCharacteristics(origin, part *nodecorev1alpha1.Characteristics) *nodecorev1alpha1.Characteristics {
	newCPU := origin.Cpu.DeepCopy()
	newMemory := origin.Memory.DeepCopy()
	newStorage := origin.PersistentStorage.DeepCopy()
	newGpu := origin.Gpu.DeepCopy()
	newEphemeralStorage := origin.EphemeralStorage.DeepCopy()
	newCPU.Sub(part.Cpu)
	newMemory.Sub(part.Memory)
	newStorage.Sub(part.PersistentStorage)
	newGpu.Sub(part.Gpu)
	newEphemeralStorage.Sub(part.EphemeralStorage)
	return &nodecorev1alpha1.Characteristics{
		Architecture:      origin.Architecture,
		Cpu:               newCPU,
		Memory:            newMemory,
		Gpu:               newGpu,
		PersistentStorage: newStorage,
		EphemeralStorage:  newEphemeralStorage,
	}
}

func (r *AllocationReconciler) updateAllocationStatus(ctx context.Context, allocation *nodecorev1alpha1.Allocation) error {
	return r.Status().Update(ctx, allocation)
}

// SetupWithManager sets up the controller with the Manager.
func (r *AllocationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nodecorev1alpha1.Allocation{}).
		Watches(&liqodiscovery.ForeignCluster{}, handler.EnqueueRequestsFromMapFunc(r.fcToAllocation), builder.WithPredicates(foreignClusterPredicate())).
		Complete(r)
}

func foreignClusterPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return fcutils.IsOutgoingJoined(e.ObjectNew.(*liqodiscovery.ForeignCluster)) ||
				fcutils.IsIncomingJoined(e.ObjectNew.(*liqodiscovery.ForeignCluster))
		},
	}
}

func (r *AllocationReconciler) fcToAllocation(_ context.Context, o client.Object) []reconcile.Request {
	clusterID := o.GetLabels()[discovery.ClusterIDLabel]
	allocationName := getters.GetAllocationNameByClusterIDSpec(context.Background(), r.Client, clusterID)
	if allocationName == nil {
		klog.Infof("Allocation with clusterID %s not found", clusterID)
		return nil
	}
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      *allocationName,
				Namespace: flags.FluidoNamespace,
			},
		},
	}
}
