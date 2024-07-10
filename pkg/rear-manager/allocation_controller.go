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

package rearmanager

import (
	"context"
	"encoding/json"

	liqodiscovery "github.com/liqotech/liqo/apis/discovery/v1alpha1"
	discovery "github.com/liqotech/liqo/pkg/discovery"
	fcutils "github.com/liqotech/liqo/pkg/utils/foreignCluster"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavors,verbs=get;list;watch;create;update;patch;delete
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

	// Different actions based on the type of Flavor related to the contract of the Allocation
	// Get the contract related to the Allocation
	contract := &reservation.Contract{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Name:      allocation.Spec.Contract.Name,
		Namespace: allocation.Spec.Contract.Namespace,
	}, contract); err != nil {
		klog.Errorf("Error when getting Contract %s: %v", allocation.Spec.Contract.Name, err)
		allocation.SetStatus(nodecorev1alpha1.Error, "Error when getting Contract")
		if err := r.updateAllocationStatus(ctx, &allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Parse Flavor type and data
	flavorTypeIdentifier, _, err := nodecorev1alpha1.ParseFlavorType(&contract.Spec.Flavor)
	if err != nil {
		klog.Errorf("Error when parsing Flavor %s: %v", contract.Spec.Flavor.Name, err)
		allocation.SetStatus(nodecorev1alpha1.Error, "Error when parsing Flavor")
		if err := r.updateAllocationStatus(ctx, &allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	switch flavorTypeIdentifier {
	case nodecorev1alpha1.TypeK8Slice:
		// We are handling a K8Slice type
		// Check if we are the provider of the Allocation
		if IsProvider(ctx, &allocation, r.Client) {
			// We need to handle the Allocation as a provider
			return r.handleK8SliceProviderAllocation(ctx, req, &allocation, contract)
		}
		// We need to handle the Allocation as a consumer
		return r.handleK8SliceConsumerAllocation(ctx, req, &allocation, contract)
		// TODO: handle other types of Flavor in the Contract of the Allocation
	default:
		// We are handling a different type
		klog.Errorf("Flavor %s is not a K8Slice type", contract.Spec.Flavor.Name)
		allocation.SetStatus(nodecorev1alpha1.Error, "Flavor is not a K8Slice type")
		if err := r.updateAllocationStatus(ctx, &allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
}

// checkInitialStatus checks if the Allocation is in an initial status, if not it sets it to Inactive.
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

func (r *AllocationReconciler) handleK8SliceProviderAllocation(ctx context.Context,
	req ctrl.Request, allocation *nodecorev1alpha1.Allocation, contract *reservation.Contract) (ctrl.Result, error) {
	allocStatus := allocation.Status.Status
	// Get the contract related to the Allocation
	switch allocStatus {
	case nodecorev1alpha1.Active:
		// We need to check if the ForeignCluster is still ready
		// If the ForeignCluster is not ready we need to set the Allocation to Released
		klog.Infof("Allocation %s is active", req.NamespacedName)
		return ctrl.Result{}, nil
	case nodecorev1alpha1.Reserved:
		// We need to check the status of the ForeignCluster
		// If the ForeignCluster is Ready the Allocation can be set to Active
		// else we need to wait for the ForeignCluster to be Ready
		klog.Infof("Allocation %s is reserved", req.NamespacedName)
		fc, err := fcutils.GetForeignClusterByID(ctx, r.Client, contract.Spec.PeeringTargetCredentials.ClusterID)
		// check if not found
		if err != nil {
			if apierrors.IsNotFound(err) {
				klog.Infof("ForeignCluster %s not found", contract.Spec.PeeringTargetCredentials.ClusterID)
				allocation.SetStatus(nodecorev1alpha1.Reserved, "ForeignCluster not found, peering not yet started")
			} else {
				klog.Errorf("Error when getting ForeignCluster %s: %v", contract.Spec.PeeringTargetCredentials.ClusterID, err)
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
			klog.Infof("ForeignCluster %s is ready, incoming peering enstablished", contract.Spec.PeeringTargetCredentials.ClusterID)
			allocation.SetStatus(nodecorev1alpha1.Active, "Incoming peering ready, Allocation is now Active")
		} else {
			klog.Infof("ForeignCluster %s is not ready yet", contract.Spec.PeeringTargetCredentials.ClusterID)
			allocation.SetStatus(nodecorev1alpha1.Reserved, "Incoming peering not yet ready, Allocation is still Reserved")
		}
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
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
		// Allocation is performed by the provider, so we need to invalidate the Flavor
		// and eventually create a new one detaching the right Partition from the old one
		klog.Infof("Allocation %s is inactive", req.NamespacedName)

		// Get the Flavor related to the Allocation by the Contract
		flavor, err := services.GetFlavorByID(contract.Spec.Flavor.Name, r.Client)
		if err != nil {
			klog.Errorf("Error when getting Flavor %s: %v", contract.Spec.Flavor.Name, err)
			allocation.SetStatus(nodecorev1alpha1.Error, "Error when getting Flavor")
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		// Reduce the availability of the Flavor
		if err := reduceFlavorAvailability(ctx, flavor, contract, r.Client); err != nil {
			klog.Errorf("Error when reducing Flavor %s availability: %v", contract.Spec.Flavor.Name, err)
			allocation.SetStatus(nodecorev1alpha1.Error, "Error when reducing Flavor availability")
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		// Switch to Reserved state
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

func (r *AllocationReconciler) handleK8SliceConsumerAllocation(ctx context.Context,
	req ctrl.Request, allocation *nodecorev1alpha1.Allocation, contract *reservation.Contract) (ctrl.Result, error) {
	allocStatus := allocation.Status.Status
	switch allocStatus {
	case nodecorev1alpha1.Active:
		// The Allocation is active,
		// We need to check if the ForeignCluster is ready

		// Get the foreign cluster related to the Allocation
		fc, err := fcutils.GetForeignClusterByID(ctx, r.Client, contract.Spec.PeeringTargetCredentials.ClusterID)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// The ForeignCluster is not found
				// We need to roll back to the establish peering phase
				klog.Infof("ForeignCluster %s not found", contract.Spec.PeeringTargetCredentials.ClusterID)
				// Change the status of the Allocation to Reserved
				allocation.SetStatus(nodecorev1alpha1.Reserved, "ForeignCluster not found, peering not yet started")
			} else {
				// Error when getting the ForeignCluster
				klog.Errorf("Error when getting ForeignCluster %s: %v", contract.Spec.PeeringTargetCredentials.ClusterID, err)
				// Change the status of the Allocation to Error
				allocation.SetStatus(nodecorev1alpha1.Error, "Error when getting ForeignCluster")
			}
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		// Check if the ForeignCluster is ready with Liqo checks
		if fcutils.IsOutgoingJoined(fc) &&
			fcutils.IsAuthenticated(fc) &&
			fcutils.IsNetworkingEstablishedOrExternal(fc) &&
			!fcutils.IsUnpeered(fc) {
			// The ForeignCluster is ready
			klog.Infof("ForeignCluster %s is ready, outgoing peering established", contract.Spec.PeeringTargetCredentials.ClusterID)
			// Change the status of the Allocation to Active
			allocation.SetStatus(nodecorev1alpha1.Active, "Outgoing peering ready, Allocation is Active")
		} else {
			// The ForeignCluster is not ready
			klog.Infof("ForeignCluster %s is not ready yet", contract.Spec.PeeringTargetCredentials.ClusterID)
			// Change the status of the Allocation to Reserved
			allocation.SetStatus(nodecorev1alpha1.Active, "Outgoing peering not yet ready, Allocation is Active")
		}

		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	case nodecorev1alpha1.Reserved:
		// The Allocation is reserved,
		// We need to establish the peering with the ForeignCluster

		klog.Infof("Allocation %s is reserved", req.NamespacedName)

		klog.Infof("Allocation %s is trying to establish a peering", req.NamespacedName.Name)

		klog.InfofDepth(1, "Allocation %s is retrieving credentials", req.NamespacedName)

		klog.InfofDepth(1, "Allocation %s has retrieved contract %s from namespace %s", req.NamespacedName, contract.Name, contract.Namespace)

		// Get the Liqo credentials for the peering target cluster, that in this scenario is the provider
		credentials := contract.Spec.PeeringTargetCredentials

		// Establish peering
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
		// Peering established
		klog.Infof("Allocation %s has started the peering with cluster %s", req.NamespacedName.Name, credentials.ClusterName)

		// Change the status of the Allocation to Active
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

		// Change the status of the Allocation to Reserved
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

// computeK8SliceCharacteristics computes the new K8SliceCharacteristics from the origin and the partition obtained by the K8SliceConfiguration.
func computeK8SliceCharacteristics(origin *nodecorev1alpha1.K8SliceCharacteristics,
	part *nodecorev1alpha1.K8SliceConfiguration) *nodecorev1alpha1.K8SliceCharacteristics {
	// Compulsory fields
	newCPU := origin.CPU.DeepCopy()
	newMemory := origin.Memory.DeepCopy()
	newPods := origin.Pods.DeepCopy()
	newCPU.Sub(part.CPU)
	newMemory.Sub(part.Memory)
	newPods.Sub(part.Pods)

	return &nodecorev1alpha1.K8SliceCharacteristics{
		Architecture: origin.Architecture,
		CPU:          newCPU,
		Memory:       newMemory,
		Pods:         newPods,
		Gpu: func() *nodecorev1alpha1.GPU {
			switch {
			case part.Gpu != nil && origin != nil:
				// Gpu is present in the origin and in the partition
				newGpuCores := origin.Gpu.Cores.DeepCopy()
				newGpuCores.Sub(part.Gpu.Cores)
				newGpuMemory := origin.Gpu.Memory.DeepCopy()
				newGpuMemory.Sub(part.Gpu.Memory)
				return &nodecorev1alpha1.GPU{
					Model:  origin.Gpu.Model,
					Cores:  newGpuCores,
					Memory: newGpuMemory,
				}
			case part.Gpu == nil && origin.Gpu != nil:
				// Gpu is present in the origin but not in the partition
				return origin.Gpu.DeepCopy()
			default:
				// Gpu is not present in the origin and in the partition
				return nil
			}
		}(),
		Storage: func() *resource.Quantity {
			switch {
			case part.Storage != nil && origin.Storage != nil:
				// Storage is present in the origin and in the partition
				newStorage := origin.Storage.DeepCopy()
				newStorage.Sub(*part.Storage)
				return &newStorage
			case part.Storage == nil && origin.Storage != nil:
				// Storage is present in the origin but not in the partition
				newStorage := origin.Storage.DeepCopy()
				return &newStorage
			default:
				// Storage is not present in the origin and in the partition
				return nil
			}
		}(),
	}
}

// reduceFlavorAvailability reduces the availability of a Flavor and manage the creation fo new one in case of specific configuration.
func reduceFlavorAvailability(ctx context.Context, flavor *nodecorev1alpha1.Flavor, contract *reservation.Contract, r client.Client) error {
	flavor.Spec.Availability = false
	klog.Infof("Updating Flavor %s: Availability %t", flavor.Name, flavor.Spec.Availability)

	// TODO: check update is actually working
	if err := r.Update(ctx, flavor); err != nil {
		klog.Errorf("Error when updating Flavor %s: %v", flavor.Name, err)
		return err
	}

	klog.Info("Flavor availability updated")

	if contract.Spec.Configuration != nil {
		// Depending on the flavor type we may need to perform different action
		// Parse Flavor to get Flavor type
		flavorTypeIdentifier, flavorTypeData, err := nodecorev1alpha1.ParseFlavorType(flavor)
		if err != nil {
			klog.Errorf("Error when parsing Flavor %s: %v", flavor.Name, err)
			return err
		}

		switch flavorTypeIdentifier {
		case nodecorev1alpha1.TypeK8Slice:
			// We are handling a K8Slice type
			// Force casting
			k8SliceFlavor := flavorTypeData.(nodecorev1alpha1.K8Slice)
			// Get configuration from the contract
			configurationTypeIdentifier, configurationData, err := nodecorev1alpha1.ParseConfiguration(contract.Spec.Configuration)
			if err != nil {
				klog.Errorf("Error when parsing Configuration %s: %v", contract.Spec.Configuration.ConfigurationTypeIdentifier, err)
				return err
			}
			if configurationTypeIdentifier != nodecorev1alpha1.TypeK8Slice {
				klog.Errorf("Configuration %s is not a K8Slice type", contract.Spec.Configuration.ConfigurationTypeIdentifier, err)
				return err
			}
			// Force casting
			k8SliceConfiguration := configurationData.(nodecorev1alpha1.K8SliceConfiguration)

			// Compute new Flavor that will be created based on the configuration
			// Get new K8SliceCharacteristics from the origin and the partition obtained by the K8SliceConfiguration
			newCharacteristics := computeK8SliceCharacteristics(&k8SliceFlavor.Characteristics, &k8SliceConfiguration)

			// Forge new flavor type
			newK8Slice := &nodecorev1alpha1.K8Slice{
				Characteristics: *newCharacteristics,
				Policies:        *k8SliceFlavor.Policies.DeepCopy(),
				Properties:      *k8SliceFlavor.Properties.DeepCopy(),
			}

			// Serialize new K8Slice to fit in the FlavorType->Data field
			newK8SliceBytes, err := json.Marshal(newK8Slice)
			if err != nil {
				klog.Errorf("Error when marshaling new K8Slice %s: %v", contract.Spec.Flavor.Name, err)
				return err
			}

			// Create new FlavorType
			newFlavorType := &nodecorev1alpha1.FlavorType{
				TypeIdentifier: nodecorev1alpha1.TypeK8Slice,
				TypeData:       runtime.RawExtension{Raw: newK8SliceBytes},
			}

			newFlavor := resourceforge.ForgeFlavorFromRef(flavor, newFlavorType)

			// Create new Flavor
			if err := r.Create(ctx, newFlavor); err != nil {
				klog.Errorf("Error when creating Flavor %s: %v", newFlavor.Name, err)
				return err
			}

			klog.Infof("Flavor %s created", newFlavor.Name)
		default:
			klog.Errorf("Flavor type %s is not supported", flavorTypeIdentifier)
			return nil
		}
	}

	klog.Info("No configuration found, no further action needed")
	return nil
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
				Namespace: flags.FluidosNamespace,
			},
		},
	}
}
