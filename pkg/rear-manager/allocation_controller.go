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

	"github.com/ghodss/yaml"
	liqodiscovery "github.com/liqotech/liqo/apis/discovery/v1alpha1"
	discovery "github.com/liqotech/liqo/pkg/discovery"
	fcutils "github.com/liqotech/liqo/pkg/utils/foreignCluster"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservation "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/consts"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/namings"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
	"github.com/fluidos-project/node/pkg/utils/services"
	virtualfabricmanager "github.com/fluidos-project/node/pkg/virtual-fabric-manager"
)

// TODO(Service): Too many rbac rules to let the allocation controller create service resources, need to be redesigned

// clusterRole
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=serviceblueprints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=serviceblueprints/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations/finalizers,verbs=update
// +kubebuilder:rbac:groups=discovery.liqo.io,resources=foreignclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.liqo.io,resources=foreignclusters/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.liqo.io,resources=foreignclusters/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=offloading.liqo.io,resources=namespaceoffloadings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=offloading.liqo.io,resources=namespaceoffloadings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=create;delete;deletecollection;patch;update;get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;delete;deletecollection;patch;update;get;list;watch
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=create;delete;deletecollection;patch;update;get;list;watch
// +kubebuilder:rbac:groups="",resources=pods,verbs=create;delete;deletecollection;patch;update;get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=create;delete;deletecollection;patch;update;get;list;watch
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=create;delete;deletecollection;patch;update;get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=create;delete;deletecollection;patch;update;get;list;watch
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=create;delete;deletecollection;patch;update;get;list;watch
// +kubebuilder:rbac:groups="apps",resources=replicasets,verbs=create;delete;deletecollection;patch;update;get;list;watch
// +kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=create;delete;deletecollection;patch;update;get;list;watch
// +kubebuilder:rbac:groups="batch",resources=cronjobs,verbs=create;delete;deletecollection;patch;update;get;list;watch
// +kubebuilder:rbac:groups="batch",resources=jobs,verbs=create;delete;deletecollection;patch;update;get;list;watch
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=create;delete;deletecollection;patch;update;get;list;watch

// AllocationReconciler reconciles a Allocation object.
type AllocationReconciler struct {
	client.Client
	Manager manager.Manager
	Scheme  *runtime.Scheme
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

	klog.Infof("Reconciling Allocation %s", req.NamespacedName)

	klog.Infof("Allocation %s is in status %s", req.NamespacedName, allocation.Status.Status)

	if allocation.Status.Status == nodecorev1alpha1.Error {
		klog.Infof("Allocation %s is in error state", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	if r.checkInitialStatus(&allocation) {
		if err := r.updateAllocationStatus(ctx, &allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
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
	case nodecorev1alpha1.TypeVM:
		// TODO(VM): handle VM type allocation
		klog.Errorf("Flavor type %s not supported yet", flavorTypeIdentifier)
		allocation.SetStatus(nodecorev1alpha1.Error, "Flavor type VM not supported yet")
		if err := r.updateAllocationStatus(ctx, &allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	case nodecorev1alpha1.TypeService:
		// We are handling a Service type
		// Check if we are the provider of the Allocation
		if IsProvider(ctx, &allocation, r.Client) {
			// We need to handle the Allocation as a provider
			return r.handleServiceProviderAllocation(ctx, req, &allocation, contract)
		}
		// We need to handle the Allocation as a consumer
		return r.handleServiceConsumerAllocation(ctx, req, &allocation, contract)
	case nodecorev1alpha1.TypeSensor:
		// TODO(Sensor): handle Sensor type allocation
		klog.Errorf("Flavor type %s not supported yet", flavorTypeIdentifier)
		allocation.SetStatus(nodecorev1alpha1.Error, "Flavor type Sensor not supported yet")
		if err := r.updateAllocationStatus(ctx, &allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
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
		allocation.Status.Status != nodecorev1alpha1.Provisioning &&
		allocation.Status.Status != nodecorev1alpha1.Released &&
		allocation.Status.Status != nodecorev1alpha1.Peering &&
		allocation.Status.Status != nodecorev1alpha1.ResourceCreation &&
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
	case nodecorev1alpha1.Provisioning:
		// We need to check the status of the ForeignCluster
		// If the ForeignCluster is Ready the Allocation can be set to Active
		// else we need to wait for the ForeignCluster to be Ready
		klog.Infof("Allocation %s is provisioning", req.NamespacedName)
		fc, err := fcutils.GetForeignClusterByID(ctx, r.Client, contract.Spec.Buyer.AdditionalInformation.LiqoID)
		// check if not found
		if err != nil {
			if apierrors.IsNotFound(err) {
				klog.Infof("ForeignCluster %s not found", contract.Spec.PeeringTargetCredentials.ClusterID)
				allocation.SetStatus(nodecorev1alpha1.Provisioning, "ForeignCluster not found, peering not yet started")
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
			klog.Infof("ForeignCluster %s is ready, incoming peering established", contract.Spec.PeeringTargetCredentials.ClusterID)
			allocation.SetStatus(nodecorev1alpha1.Active, "Incoming peering ready, Allocation is now Active")
		} else {
			klog.Infof("ForeignCluster %s is not ready yet", contract.Spec.PeeringTargetCredentials.ClusterID)
			allocation.SetStatus(nodecorev1alpha1.Provisioning, "Incoming peering not yet ready, Allocation is still Reserved")
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

		// Switch to Provisioning state
		allocation.SetStatus(nodecorev1alpha1.Provisioning, "Resources reserved")
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
				allocation.SetStatus(nodecorev1alpha1.Provisioning, "ForeignCluster not found, peering not yet started")
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

		allocation.SetResourceRef(nodecorev1alpha1.GenericRef{
			Name:       fc.Name,
			Namespace:  fc.Namespace,
			Kind:       fc.Kind,
			APIVersion: fc.APIVersion,
		})

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
	case nodecorev1alpha1.Peering:
		// The Allocation is Peering,
		// We need to establish the peering with the ForeignCluster

		klog.Infof("Allocation %s is Peering", req.NamespacedName)

		klog.Infof("Allocation %s is trying to establish a peering", req.NamespacedName.Name)

		klog.InfofDepth(1, "Allocation %s is retrieving credentials", req.NamespacedName)

		klog.InfofDepth(1, "Allocation %s has retrieved contract %s from namespace %s", req.NamespacedName, contract.Name, contract.Namespace)

		// Get the Liqo credentials for the peering target cluster, that in this scenario is the provider
		credentials := contract.Spec.PeeringTargetCredentials
		// Check if a Liqo peering has been already established
		_, err := fcutils.GetForeignClusterByID(ctx, r.Client, credentials.ClusterID)
		if err != nil {
			if apierrors.IsNotFound(err) {
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
			} else {
				klog.Errorf("Error when getting ForeignCluster %s: %v", credentials.ClusterID, err)
				allocation.SetStatus(nodecorev1alpha1.Error, "Error when getting ForeignCluster")
				if err := r.updateAllocationStatus(ctx, allocation); err != nil {
					klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
			}
		} else {
			// Peering already established
			klog.Infof("Allocation %s has already peered with cluster %s", req.NamespacedName.Name, credentials.ClusterName)
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, nil
	case nodecorev1alpha1.Released:
		// The Allocation is released,
		// We need to check if the ForeignCluster is again ready
		return ctrl.Result{}, nil
	case nodecorev1alpha1.Inactive:
		klog.Infof("Allocation %s is inactive", req.NamespacedName)

		// Change the status of the Allocation to Reserved
		allocation.SetStatus(nodecorev1alpha1.Peering, "Allocation is now in Peering")
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

func (r *AllocationReconciler) handleServiceProviderAllocation(ctx context.Context,
	req ctrl.Request, allocation *nodecorev1alpha1.Allocation, contract *reservation.Contract) (ctrl.Result, error) {
	allocStatus := allocation.Status.Status
	// Get the contract related to the Allocation
	switch allocStatus {
	case nodecorev1alpha1.Active:
		// We need to check if the ForeignCluster is still ready
		// Get the foreign cluster related to the Allocation
		fc, err := fcutils.GetForeignClusterByID(ctx, r.Client, contract.Spec.PeeringTargetCredentials.ClusterID)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// The ForeignCluster is not found
				// We need to roll back to the establish peering phase
				klog.Infof("ForeignCluster %s not found", contract.Spec.PeeringTargetCredentials.ClusterID)
				// Change the status of the Allocation to Reserved
				allocation.SetStatus(nodecorev1alpha1.Provisioning, "ForeignCluster not found, peering not yet started")
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
			// allocation.SetStatus(nodecorev1alpha1.Active, "Outgoing peering ready, Allocation is Active")
		} else {
			// The ForeignCluster is not ready
			klog.Infof("ForeignCluster %s is not ready yet", contract.Spec.PeeringTargetCredentials.ClusterID)
			// Change the status of the Allocation to Active
			allocation.SetStatus(nodecorev1alpha1.Active, "Outgoing peering not yet ready, Allocation is Active")

			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
		}

		// TODO(Service): check if the service software applied is working correctly, maybe checking the pods deployed in the namespace offloaded.

		return ctrl.Result{}, nil
	case nodecorev1alpha1.ResourceCreation:
		// The Allocation is ResourceCreation,

		klog.Infof("Allocation %s is ResourceCreation", req.NamespacedName)

		namespaceName, err := r.createConsumerNamespace(ctx, req, contract, allocation)
		if err != nil {
			klog.Errorf("Error when creating consumer namespace: %v", err)
			return ctrl.Result{}, err
		}

		if err := r.createServiceResources(ctx, req, contract, allocation, namespaceName); err != nil {
			klog.Errorf("Error when creating service resources: %v", err)
			return ctrl.Result{}, err
		}

		// Change the status of the Allocation to Active
		allocation.SetStatus(nodecorev1alpha1.Active, "Allocation has created the namespace and applied the manifests, now is Active")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}

		klog.Infof(
			"Allocation %s has created the namespace %s and applied the manifests, now is %s",
			req.NamespacedName.Name,
			namespaceName, allocation.Status.Status,
		)

		return ctrl.Result{}, nil
	case nodecorev1alpha1.Provisioning:
		// The Allocation is provisioning,
		// We need to check the status of the ForeignCluster
		// If the ForeignCluster is Ready the Allocation can be set to ResourceCreation

		klog.Infof("Allocation %s is provisioning", req.NamespacedName)

		readiness, err := r.checkOutgoingForeignClusterReadiness(ctx, contract.Spec.Buyer.AdditionalInformation.LiqoID, allocation)
		if err != nil {
			klog.Errorf("Error when checking ForeignCluster readiness: %v", err)
			return ctrl.Result{}, err
		}

		if readiness {
			// The ForeignCluster is ready
			klog.Infof("ForeignCluster %s is ready, outgoing peering established", contract.Spec.Buyer.AdditionalInformation.LiqoID)
			// Change the status of the Allocation to ResourceCreation
			allocation.SetStatus(nodecorev1alpha1.ResourceCreation, "ForeignCluster ready, Allocation is now ResourceCreation")
		}

		// Update the status of the Allocation
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	case nodecorev1alpha1.Peering:
		// Create peering with the consumer
		klog.Infof("Allocation %s is peering", req.NamespacedName)

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
		allocation.SetStatus(nodecorev1alpha1.Provisioning, "Allocation is now provisioning")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
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

		klog.Infof("Flavor %s availability reduced", contract.Spec.Flavor.Name)

		// Switch to Peering state
		allocation.SetStatus(nodecorev1alpha1.Peering, "Starting peering")
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

func (r *AllocationReconciler) checkOutgoingForeignClusterReadiness(ctx context.Context,
	clusterID string, allocation *nodecorev1alpha1.Allocation) (ready bool, err error) {
	// Get the foreign cluster related to the Allocation
	fc, er := fcutils.GetForeignClusterByID(ctx, r.Client, clusterID)
	if er != nil {
		if apierrors.IsNotFound(er) {
			// The ForeignCluster is not found
			// We need to roll back to the establish peering phase
			klog.Infof("ForeignCluster %s not found", clusterID)
			// Change the status of the Allocation to Reserved
			allocation.SetStatus(nodecorev1alpha1.Peering, "ForeignCluster not found, peering not yet started")
		} else {
			// Error when getting the ForeignCluster
			klog.Errorf("Error when getting ForeignCluster %s: %v", clusterID, er)
			// Change the status of the Allocation to Error
			allocation.SetStatus(nodecorev1alpha1.Error, "Error when getting ForeignCluster")
			return false, er
		}
		return false, nil
	}

	// Check if the ForeignCluster is ready with Liqo checks
	if fcutils.IsOutgoingJoined(fc) &&
		fcutils.IsAuthenticated(fc) &&
		fcutils.IsNetworkingEstablishedOrExternal(fc) &&
		!fcutils.IsUnpeered(fc) {
		// The ForeignCluster is ready
		klog.Infof("ForeignCluster %s is ready, outgoing peering established", clusterID)
	} else {
		// The ForeignCluster is not ready
		klog.Infof("ForeignCluster %s is not ready yet", clusterID)
		// Change the status of the Allocation to Reserved
		allocation.SetStatus(nodecorev1alpha1.Provisioning, "Outgoing peering not yet ready, Allocation is still Reserved")
		return false, nil
	}

	return true, nil
}

func (r *AllocationReconciler) createConsumerNamespace(ctx context.Context, req ctrl.Request,
	contract *reservation.Contract, allocation *nodecorev1alpha1.Allocation) (nsName string, err error) {
	// Create namespace for the application software service to be provided
	// Create a namespace
	namespaceName := namings.ForgeNamespaceName(contract)
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	// Create the namespace
	if err := r.Client.Create(ctx, namespace); err != nil {
		klog.Errorf("Error when creating Namespace %s: %v", namespaceName, err)
		allocation.SetStatus(nodecorev1alpha1.Error, "Error when creating Namespace")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return "", err
		}
		return "", err
	}

	klog.Infof("Namespace %s created", namespaceName)

	hostingPolicy, err := resourceforge.ForgeHostingPolicyFromContract(contract, r.Client)
	if err != nil {
		klog.Errorf("Error when forging HostingPolicy: %v", err)
		allocation.SetStatus(nodecorev1alpha1.Error, "Error when forging HostingPolicy")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return "", err
		}
		return "", err
	}

	podOffloadingStrategy, err := resourceforge.ForgePodOffloadingStrategy(&hostingPolicy)
	if err != nil {
		klog.Errorf("Error when forging PodOffloadingStrategy: %v", err)
		allocation.SetStatus(nodecorev1alpha1.Error, "Error when forging PodOffloadingStrategy")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return "", err
		}
		return "", err
	}

	klog.Infof("PodOffloadingStrategy %s forged", podOffloadingStrategy)
	klog.Infof("PeeringTargetCredentials %s forged", contract.Spec.PeeringTargetCredentials)

	credentials := contract.Spec.PeeringTargetCredentials

	klog.Infof("Credentials clusterID: %s", credentials.ClusterID)

	no, err := virtualfabricmanager.OffloadNamespace(ctx, r.Client, namespaceName, podOffloadingStrategy, credentials.ClusterID)

	if err != nil {
		klog.Errorf("Error when offloading namespace %s: %v", namespaceName, err)
		allocation.SetStatus(nodecorev1alpha1.Error, "Error when offloading namespace")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return "", err
		}
		return "", err
	}

	klog.Infof("Namespace %s offloaded to cluster %s with OffloadingNamespace %s", namespaceName, credentials.ClusterName, no.Name)

	return namespaceName, nil
}

func (r *AllocationReconciler) createServiceResources(ctx context.Context, req ctrl.Request,
	contract *reservation.Contract, allocation *nodecorev1alpha1.Allocation, namespaceName string) (err error) {
	// Get manifests to apply to the namespace from the ServiceBlueprint in the contract
	manifests, err := resourceforge.ForgeServiceManifests(ctx, r.Client, contract)
	if err != nil {
		klog.Errorf("Error when forging ServiceManifests: %v", err)
		allocation.SetStatus(nodecorev1alpha1.Error, "Error when forging ServiceManifests")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return err
		}
		return err
	}

	// Create FLUIDOS Labels
	labels := map[string]string{
		consts.FluidosContractLabel: contract.Name,
	}

	var serviceEndpoint *corev1.Service

	// Apply the manifests to the namespace
	for _, manifest := range manifests {
		// Convert the manifest to an unstructured object
		obj := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(manifest), obj); err != nil {
			klog.Errorf("Error when unmarshalling manifest: %v", err)
			allocation.SetStatus(nodecorev1alpha1.Error, "Error when unmarshalling manifest")
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return err
			}
			return err
		}

		klog.Infof("Obtained manifest %s", obj.GetName())

		// Force namespace assignment
		obj.SetNamespace(namespaceName)
		// Append labels to original label of the manifest
		var newLabels = make(map[string]string)
		for k, v := range obj.GetLabels() {
			newLabels[k] = v
		}
		for k, v := range labels {
			newLabels[k] = v
		}
		obj.SetLabels(newLabels)

		// Apply the manifest
		if err := r.Client.Create(ctx, obj); err != nil {
			klog.Errorf("Error when creating manifest %s: %v", obj.GetName(), err)
			allocation.SetStatus(nodecorev1alpha1.Error, "Error when creating manifest")
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return err
			}
			return err
		}

		if obj.GetKind() == "Service" && obj.GetLabels()[consts.FluidosServiceEndpoint] == "true" {
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &serviceEndpoint)
			if err != nil {
				klog.Errorf("Error when converting unstructured to service: %v", err)
				allocation.SetStatus(nodecorev1alpha1.Error, "Error when converting unstructured to service")
				if err := r.updateAllocationStatus(ctx, allocation); err != nil {
					klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
					return err
				}
				return err
			}
			klog.Infof("ServiceEndpoint %s created", serviceEndpoint.Name)
		} else {
			klog.Infof("Manifest is kind %s", obj.GetKind())
			klog.Infof("Manifest labels %v", obj.GetLabels())
			klog.Infof("Manifest is not compatible with ServiceEndpoint")
		}

		klog.Infof("Manifest %s applied to namespace %s", obj.GetName(), namespaceName)
	}

	klog.Infof("Allocation %s has created the namespace %s and applied the manifests", req.NamespacedName.Name, namespaceName)

	// Create the secret for the credentials in the namespace
	secret, err := resourceforge.ForgeSecretForService(contract, serviceEndpoint)
	if err != nil {
		klog.Errorf("Error when forging Secret: %v", err)
		allocation.SetStatus(nodecorev1alpha1.Error, "Error when forging Secret")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return err
		}
		return err
	}

	// Force namespace assignment
	secret.SetNamespace(namespaceName)

	// Apply the secret to the namespace
	if err := r.Client.Create(ctx, secret); err != nil {
		klog.Errorf("Error when creating Secret %s: %v", secret.Name, err)
		allocation.SetStatus(nodecorev1alpha1.Error, "Error when creating Secret")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return err
		}
		return err
	}

	klog.Infof("Secret %s created", secret.Name)

	return nil
}

func (r *AllocationReconciler) handleServiceConsumerAllocation(ctx context.Context,
	req ctrl.Request, allocation *nodecorev1alpha1.Allocation, contract *reservation.Contract) (ctrl.Result, error) {
	allocStatus := allocation.Status.Status
	switch allocStatus {
	case nodecorev1alpha1.Active:
		// TODO(Service): implement Active Allocation Consumer side
		klog.Infof("Allocation %s is active", req.NamespacedName)
		return ctrl.Result{}, nil
	case nodecorev1alpha1.Provisioning:

		klog.Infof("Allocation %s is provisioning", req.NamespacedName)

		// The Allocation can look for incoming peering associated with the contract
		fc, err := fcutils.GetForeignClusterByID(ctx, r.Client, contract.Spec.Seller.AdditionalInformation.LiqoID)
		// check if not found
		if err != nil {
			if apierrors.IsNotFound(err) {
				klog.Infof("ForeignCluster %s not found", contract.Spec.Seller.AdditionalInformation.LiqoID)
				allocation.SetStatus(nodecorev1alpha1.Provisioning, "ForeignCluster not found, peering not yet started")
			} else {
				klog.Errorf("Error when getting ForeignCluster %s: %v", contract.Spec.Seller.AdditionalInformation.LiqoID, err)
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
			klog.Infof("ForeignCluster %s is ready, incoming peering established", contract.Spec.Seller.AdditionalInformation.LiqoID)
			allocation.SetStatus(nodecorev1alpha1.Active, "Incoming peering ready, Allocation is now Active")
		} else {
			klog.Infof("ForeignCluster %s is not ready yet", contract.Spec.PeeringTargetCredentials.ClusterID)
			allocation.SetStatus(nodecorev1alpha1.Provisioning, "Incoming peering not yet ready, Allocation is still provisioning")
		}

		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}

		// Looking for the namespace offloaded
		// Get namespaces labeled with the liqo.io/remote-cluster-id labelset to the contract.Spec.Seller.AdditionalInformation.LiqoID
		offloadedNamespaces := &corev1.NamespaceList{}
		if err := r.Client.List(
			ctx,
			offloadedNamespaces,
			client.MatchingLabels{
				consts.LiqoRemoteClusterIDLabel: contract.Spec.Seller.AdditionalInformation.LiqoID,
			},
		); err != nil {
			klog.Errorf("Error when listing offloaded namespaces: %v", err)
			allocation.SetStatus(nodecorev1alpha1.Error, "Error when listing offloaded namespaces")
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		if len(offloadedNamespaces.Items) == 0 {
			klog.Infof("No offloaded namespaces found")
			allocation.SetStatus(nodecorev1alpha1.Provisioning, "No offloaded namespaces found, still provisioning")
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		// Get the namespace from the list whose name is the same as the contract name
		offloadedNamespace := &corev1.Namespace{}
		for i := range offloadedNamespaces.Items {
			ns := &offloadedNamespaces.Items[i]
			if ns.Name == contract.Name {
				offloadedNamespace = ns
				break
			}
		}

		if offloadedNamespace.Name == "" {
			klog.Infof("Offloaded namespace %s not found", contract.Name)
			allocation.SetStatus(nodecorev1alpha1.Provisioning, "Namespace not found, still provisioning")
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		klog.Infof("Offloaded namespace %s found", offloadedNamespace.Name)

		// Switch to Active state
		allocation.SetStatus(nodecorev1alpha1.Active, "Namespace offloaded found")
		if err := r.updateAllocationStatus(ctx, allocation); err != nil {
			klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
			return ctrl.Result{}, err
		}

		// Get the credentials secret in the namespace labeled that matches the label const.FluidosServiceCredentials as true
		credentialsSecrets := &corev1.SecretList{}
		if err := r.Client.List(
			ctx,
			credentialsSecrets,
			client.InNamespace(offloadedNamespace.Name),
			client.MatchingLabels{consts.FluidosServiceCredentials: "true"},
		); err != nil {
			klog.Errorf("Error when listing credentials secret: %v", err)
			allocation.SetStatus(nodecorev1alpha1.Error, "Error when listing credentials secret")
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
		}

		if len(credentialsSecrets.Items) == 0 {
			klog.Infof("No credentials secret found, possibly not yet created")
			allocation.SetStatus(nodecorev1alpha1.Provisioning, "Credentials secret not found, still provisioning")
			if err := r.updateAllocationStatus(ctx, allocation); err != nil {
				klog.Errorf("Error when updating Allocation %s status: %v", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		// Get the first secret
		credentialsSecret := credentialsSecrets.Items[0]

		klog.Infof("Credentials secret %s found", credentialsSecret.Name)

		// Set the resource reference in the status
		allocation.SetResourceRef(nodecorev1alpha1.GenericRef{
			Name:       credentialsSecret.Name,
			Namespace:  credentialsSecret.Namespace,
			Kind:       credentialsSecret.Kind,
			APIVersion: credentialsSecret.APIVersion,
		})

		allocation.SetStatus(nodecorev1alpha1.Active, "Credentials secret found, Allocation is now Active")

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
		klog.Infof("Allocation %s is inactive", req.NamespacedName)

		// No particular action is needed, the Allocation can be set to Provisioning
		allocation.SetStatus(nodecorev1alpha1.Provisioning, "Resources provisioning")
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
	// Reduce the availability of the original Flavor
	flavor.Spec.Availability = false
	klog.Infof("Updating Flavor %s: Availability %t", flavor.Name, flavor.Spec.Availability)

	// TODO: check update is actually working
	if err := r.Update(ctx, flavor); err != nil {
		klog.Errorf("Error when updating Flavor %s: %v", flavor.Name, err)
		return err
	}

	klog.Info("Flavor availability updated")

	klog.Infof("Checking if configuration is present in the contract %s", contract.Name)
	klog.Infof("Configuration: %v", contract.Spec.Configuration)

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
			configurationTypeIdentifier, configurationData, err := nodecorev1alpha1.ParseConfiguration(contract.Spec.Configuration, &contract.Spec.Flavor)
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
		case nodecorev1alpha1.TypeVM:
			// TODO(VM): handle VM type allocation
			klog.Errorf("Flavor type %s not supported yet", flavorTypeIdentifier)
			return nil
		case nodecorev1alpha1.TypeService:
			// TODO(Service): evaluate current reduce flavor availability strategy
			// reduceFlavorAvailability current strategy: recreate exact copy of the Flavor, flagging the original as unavailable

			// Copy the original Flavor type, no changes applied
			newFlavorType := flavor.Spec.FlavorType.DeepCopy()

			// Create new FlavorType
			newFlavor := resourceforge.ForgeFlavorFromRef(flavor, newFlavorType)

			// Create new Flavor
			if err := r.Create(ctx, newFlavor); err != nil {
				klog.Errorf("Error when creating Flavor %s: %v", newFlavor.Name, err)
				return err
			}

			klog.Infof("Flavor %s created", newFlavor.Name)
		case nodecorev1alpha1.TypeSensor:
			// TODO(Sensor): handle Sensor type allocation
			klog.Errorf("Flavor type %s not supported yet", flavorTypeIdentifier)
			return nil
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
		Watches(&corev1.Namespace{}, handler.EnqueueRequestsFromMapFunc(r.nsToAllocation), builder.WithPredicates(namespacePredicate())).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(r.secretToAllocation), builder.WithPredicates(secretPredicate())).
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

func namespacePredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			_, ok := e.Object.GetLabels()[consts.FluidosServiceCredentials]
			return ok
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			_, ok := e.ObjectNew.GetLabels()[consts.FluidosServiceCredentials]
			return ok
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			_, ok := e.Object.GetLabels()[consts.FluidosServiceCredentials]
			return ok
		},
	}
}

func secretPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			_, ok := e.Object.GetLabels()[consts.FluidosContractLabel]
			return ok
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			_, ok := e.ObjectNew.GetLabels()[consts.FluidosContractLabel]
			return ok
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			_, ok := e.Object.GetLabels()[consts.FluidosContractLabel]
			return ok
		},
	}
}

func (r *AllocationReconciler) fcToAllocation(_ context.Context, o client.Object) []reconcile.Request {
	clusterID := o.GetLabels()[discovery.ClusterIDLabel]
	allocations, err := getters.GetAllocationByClusterIDSpec(context.Background(), r.Client, clusterID)
	if err != nil {
		klog.Errorf("Error when getting Allocation by clusterID %s: %v", clusterID, err)
		return nil
	}
	if len(allocations.Items) == 0 {
		klog.Infof("Allocation with clusterID %s not found", clusterID)
		return nil
	}
	// Filter the Allocation list the get only the ones not in Peering status
	var filteredAllocations []nodecorev1alpha1.Allocation
	for i := range allocations.Items {
		allocation := allocations.Items[i]
		if allocation.Status.Status != nodecorev1alpha1.Peering {
			filteredAllocations = append(filteredAllocations, allocation)
		}
	}
	if len(filteredAllocations) == 0 {
		klog.Infof("No Allocation found with clusterID %s not in Peering status", clusterID)
		return nil
	}
	var requests []reconcile.Request
	for i := range filteredAllocations {
		allocation := filteredAllocations[i]
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      allocation.Name,
				Namespace: allocation.Namespace,
			},
		})
	}
	return requests
}

func (r *AllocationReconciler) nsToAllocation(_ context.Context, o client.Object) []reconcile.Request {
	// Get the LiqoID from the Namespace labels
	liqoID := o.GetLabels()[consts.LiqoRemoteClusterIDLabel]
	// Get the Allocation related to the LiqoID
	allocations, err := getters.GetAllocationByClusterIDSpec(context.Background(), r.Client, liqoID)
	if err != nil {
		klog.Errorf("Error when getting Allocation by LiqoID %s: %v", liqoID, err)
		return nil
	}
	if len(allocations.Items) == 0 {
		klog.Infof("Allocation with LiqoID %s not found", liqoID)
		return nil
	}
	// Filter the allocation list to get only the ones generated for service flavors
	var filteredAllocations []nodecorev1alpha1.Allocation
	for i := range allocations.Items {
		allocation := allocations.Items[i]
		// Get the Contract related to the Allocation with the
		var contract reservation.Contract
		if err := r.Client.Get(context.Background(), types.NamespacedName{
			Name:      allocation.Spec.Contract.Name,
			Namespace: allocation.Spec.Contract.Namespace,
		}, &contract); err != nil {
			klog.Errorf("Error when getting Contract %s: %v", allocation.Spec.Contract, err)
			continue
		}
		// Check the Flavor type
		if contract.Spec.Flavor.Spec.FlavorType.TypeIdentifier == nodecorev1alpha1.TypeService {
			// Check the LiqoID is the same as the contract.spec.seller.additionalInformation.liqoID
			if contract.Spec.Seller.AdditionalInformation.LiqoID == liqoID {
				// Check the name of the namespace is the same as the contract
				if o.GetName() == namings.ForgeNamespaceName(&contract) {
					filteredAllocations = append(filteredAllocations, allocation)
				} else {
					klog.Infof("Namespace %s is not related to the Allocation %s", o.GetName(), allocation.Name)
				}
			} else {
				klog.Infof("Allocation %s is not related to the LiqoID %s", allocation.Name, liqoID)
			}
		} else {
			klog.Infof("Allocation %s is not related to a service flavor", allocation.Name)
		}
	}
	if len(filteredAllocations) == 0 {
		klog.Info("Filtered Allocation list is empty")
		return nil
	}
	var requests []reconcile.Request
	for i := range filteredAllocations {
		allocation := filteredAllocations[i]
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      allocation.Name,
				Namespace: allocation.Namespace,
			},
		})
	}
	return requests
}

func (r *AllocationReconciler) secretToAllocation(_ context.Context, o client.Object) []reconcile.Request {
	// Get contract name from the namespace name of the secret
	contractName := o.GetNamespace()
	// Get the Contract related to the Allocation with the
	var contract reservation.Contract
	if err := r.Client.Get(context.Background(), types.NamespacedName{
		Name:      contractName,
		Namespace: flags.FluidosNamespace,
	}, &contract); err != nil {
		klog.Errorf("Error when getting Contract %s: %v", contractName, err)
		return nil
	}
	// Get the Namespace of the secret
	namespaceName := o.GetNamespace()
	var namespace corev1.Namespace
	if err := r.Client.Get(context.Background(), types.NamespacedName{
		Name: namespaceName,
	}, &namespace); err != nil {
		klog.Errorf("Error when getting Namespace %s: %v", namespaceName, err)
		return nil
	}
	// Get the consts.LiqoRemoteClusterIDLabel value from the namespace labels
	remoteLiqoClusterID := namespace.GetLabels()[consts.LiqoRemoteClusterIDLabel]
	// Check if this secret is triggering in the provider or consumer side
	if contract.Spec.Buyer.AdditionalInformation.LiqoID == remoteLiqoClusterID {
		// This is a provider side secret, so we do not need to trigger the Allocation
		klog.Infof("Secret %s is related to the provider side", o.GetName())
		return nil
	}
	// Get the Allocation related to the Contract
	allocation, err := getters.GetAllocationByContractName(context.Background(), r.Client, contractName)
	if err != nil {
		klog.Errorf("Error when getting Allocation by ContractName %s: %v", contractName, err)
		return nil
	}

	if allocation == nil {
		klog.Infof("Allocation related to Contract %s not found", contractName)
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      allocation.Name,
				Namespace: allocation.Namespace,
			},
		},
	}
}
