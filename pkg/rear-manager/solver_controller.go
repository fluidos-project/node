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

	fcutils "github.com/liqotech/liqo/pkg/utils/foreignCluster"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	advertisementv1alpha1 "github.com/fluidos-project/node/apis/advertisement/v1alpha1"
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/common"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/namings"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
	"github.com/fluidos-project/node/pkg/utils/tools"
)

// SolverReconciler reconciles a Solver object.
type SolverReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// clusterRole
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=solvers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=solvers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=solvers/finalizers,verbs=update
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavors/finalizers,verbs=update
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations/finalizers,verbs=update
// +kubebuilder:rbac:groups=advertisement.fluidos.eu,resources=peeringcandidates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=advertisement.fluidos.eu,resources=peeringcandidates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=advertisement.fluidos.eu,resources=peeringcandidates/finalizers,verbs=update
// +kubebuilder:rbac:groups=advertisement.fluidos.eu,resources=discoveries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=advertisement.fluidos.eu,resources=discoveries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=advertisement.fluidos.eu,resources=discoveries/finalizers,verbs=update
// +kubebuilder:rbac:groups=reservation.fluidos.eu,resources=reservations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=reservation.fluidos.eu,resources=reservations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=reservation.fluidos.eu,resources=reservations/finalizers,verbs=update
// +kubebuilder:rbac:groups=reservation.fluidos.eu,resources=contracts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=reservation.fluidos.eu,resources=contracts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=reservation.fluidos.eu,resources=contracts/finalizers,verbs=update
// +kubebuilder:rbac:groups=discovery.liqo.io,resources=foreignclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.liqo.io,resources=foreignclusters/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=discovery.liqo.io,resources=foreignclusters/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=*,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *SolverReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "solver", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	var solver nodecorev1alpha1.Solver
	if err := r.Get(ctx, req.NamespacedName, &solver); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting Solver %s before reconcile: %s", req.NamespacedName, err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("Solver %s not found, probably deleted", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	if checkInitialStatus(&solver) {
		if err := r.updateSolverStatus(ctx, &solver); err != nil {
			klog.Errorf("Error when updating Solver %s status before reconcile: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	klog.Infof("Reconciling Solver %s", req.NamespacedName)

	findCandidateStatus := solver.Status.FindCandidate
	reserveAndBuyStatus := solver.Status.ReserveAndBuy
	peeringStatus := solver.Status.Peering

	// Check if the Solver has expired or failed, in this case do nothing and return
	if solver.Status.SolverPhase.Phase == nodecorev1alpha1.PhaseFailed ||
		solver.Status.SolverPhase.Phase == nodecorev1alpha1.PhaseTimeout {
		return ctrl.Result{}, nil
	}

	if solver.Spec.FindCandidate {
		if findCandidateStatus != nodecorev1alpha1.PhaseSolved {
			return r.handleFindCandidate(ctx, req, &solver)
		}
	} else {
		klog.Infof("Solver %s Solved : No need to find a candidate", req.NamespacedName.Name)
		solver.SetPhase(nodecorev1alpha1.PhaseSolved, "No need to find a candidate")
		err := r.updateSolverStatus(ctx, &solver)
		if err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
	}

	if solver.Spec.ReserveAndBuy {
		if (findCandidateStatus == nodecorev1alpha1.PhaseSolved || !solver.Spec.FindCandidate) &&
			reserveAndBuyStatus != nodecorev1alpha1.PhaseSolved {
			return r.handleReserveAndBuy(ctx, req, &solver)
		}
	} else {
		klog.Infof("Solver %s Solved : No need to reserve and buy the resources", req.NamespacedName.Name)
		solver.SetPhase(nodecorev1alpha1.PhaseSolved, "No need to reserve and buy the resources")
		err := r.updateSolverStatus(ctx, &solver)
		if err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if solver.Spec.EstablishPeering {
		if reserveAndBuyStatus == nodecorev1alpha1.PhaseSolved && findCandidateStatus == nodecorev1alpha1.PhaseSolved &&
			peeringStatus != nodecorev1alpha1.PhaseSolved {
			return r.handlePeering(ctx, req, &solver)
		}
	} else {
		klog.Infof("Solver %s Solved : No need to establish a peering", req.NamespacedName.Name)
		solver.SetPhase(nodecorev1alpha1.PhaseSolved, "No need to establish a peering")
		err := r.updateSolverStatus(ctx, &solver)
		if err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func checkInitialStatus(solver *nodecorev1alpha1.Solver) bool {
	if solver.Status.SolverPhase.Phase != nodecorev1alpha1.PhaseSolved &&
		solver.Status.SolverPhase.Phase != nodecorev1alpha1.PhaseTimeout &&
		solver.Status.SolverPhase.Phase != nodecorev1alpha1.PhaseFailed &&
		solver.Status.SolverPhase.Phase != nodecorev1alpha1.PhaseRunning &&
		solver.Status.SolverPhase.Phase != nodecorev1alpha1.PhaseIdle {
		solver.SetPhase(nodecorev1alpha1.PhaseIdle, "Solver initialized")
		return true
	}
	return false
}

func (r *SolverReconciler) handleFindCandidate(ctx context.Context, req ctrl.Request, solver *nodecorev1alpha1.Solver) (ctrl.Result, error) {
	findCandidateStatus := solver.Status.FindCandidate
	switch findCandidateStatus {
	case nodecorev1alpha1.PhaseIdle:
		// Search a matching PeeringCandidate if available
		pc, err := r.searchPeeringCandidates(ctx, solver)
		if client.IgnoreNotFound(err) != nil {
			klog.Errorf("Error when searching and booking a candidate for Solver %s: %s", req.NamespacedName.Name, err)
			return ctrl.Result{}, err
		}

		// If some PeeringCandidates are available, select one and book it
		if len(pc) > 0 {
			// If some PeeringCandidates are available, select one and book it
			selectedPc, err := r.selectAndBookPeeringCandidate(pc)
			if err != nil {
				klog.Errorf("Error when selecting and booking a candidate for Solver %s: %s", req.NamespacedName.Name, err)
				return ctrl.Result{}, err
			}
			klog.Infof("Solver %s has selected and booked candidate %s", req.NamespacedName.Name, selectedPc.Name)
			solver.SetFindCandidateStatus(nodecorev1alpha1.PhaseSolved)
			solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Solver has found a candidate")
			if err := r.updateSolverStatus(ctx, solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		// If no PeeringCandidate is available, Create a Discovery
		klog.Infof("Solver %s has not found any candidate. Trying a Discovery", req.NamespacedName.Name)

		// Change the status of the Solver to Running
		solver.SetFindCandidateStatus(nodecorev1alpha1.PhaseRunning)
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Solver is trying a Discovery")

		// Update the Solver status
		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil

	case nodecorev1alpha1.PhaseRunning:
		// Check solver expiration
		if tools.CheckExpirationSinceTime(solver.Status.SolverPhase.LastChangeTime, flags.ExpirationPhaseRunning) {
			klog.Infof("Solver %s has expired", req.NamespacedName.Name)

			solver.SetPhase(nodecorev1alpha1.PhaseTimeout, "Solver has expired before finding a candidate")

			if err := r.updateSolverStatus(ctx, solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		klog.Infof("Getting or creating Discovery for Solver %s", req.NamespacedName.Name)
		discovery, err := r.createOrGetDiscovery(ctx, solver)
		if err != nil {
			klog.Errorf("Error when creating or getting Discovery for Solver %s: %s", req.NamespacedName.Name, err)
			return ctrl.Result{}, err
		}

		common.DiscoveryStatusCheck(solver, discovery)

		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	case nodecorev1alpha1.PhaseFailed:
		klog.Infof("Solver %s has not found any candidate", req.NamespacedName.Name)
		solver.SetPhase(nodecorev1alpha1.PhaseFailed, "Solver has not found any candidate")
		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	default:
		solver.SetFindCandidateStatus(nodecorev1alpha1.PhaseIdle)
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Solver is running")
		// Update the Solver status
		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
}

func (r *SolverReconciler) handleReserveAndBuy(ctx context.Context, req ctrl.Request, solver *nodecorev1alpha1.Solver) (ctrl.Result, error) {
	reserveAndBuyStatus := solver.Status.ReserveAndBuy
	switch reserveAndBuyStatus {
	case nodecorev1alpha1.PhaseIdle:
		var configuration *nodecorev1alpha1.Configuration
		klog.Infof("Creating the Reservation %s", req.NamespacedName.Name)
		var pcList []advertisementv1alpha1.PeeringCandidate
		pc := advertisementv1alpha1.PeeringCandidate{}
		// Create the Reservation

		pcList, err := r.searchPeeringCandidates(ctx, solver)
		if err != nil {
			klog.Errorf("Error when searching and booking a candidate for Solver %s: %s", req.NamespacedName.Name, err)
			return ctrl.Result{}, err
		}

		// Filter PeeringCandidate by SolverID
		for i := range pcList {
			if pcList[i].Spec.Available {
				pc = pcList[i]
				// Update the SolverID in the PeeringCandidate selected if it is different from the current Solver
				if pc.Spec.SolverID != solver.Name {
					pc.Spec.SolverID = solver.Name
					if err := r.Client.Update(ctx, &pc); err != nil {
						klog.Errorf("Error when updating PeeringCandidate %s for Solver %s: %s", pc.Name, solver.Name, err)
						return ctrl.Result{}, err
					}
				}
				break
			}
		}

		if pc.Name == "" {
			klog.Errorf("No PeeringCandidate found for Solver %s:", solver.Name)
			return ctrl.Result{}, nil
		}

		if solver.Spec.Selector != nil {
			// Forge the Partition

			// Parse Solver Selector
			solverTypeIdentifier, solverTypeData, err := nodecorev1alpha1.ParseSolverSelector(solver.Spec.Selector)
			if err != nil {
				klog.Errorf("Error when parsing Solver Selector for Solver %s: %s", solver.Name, err)
				return ctrl.Result{}, err
			}
			if solverTypeIdentifier == nodecorev1alpha1.TypeK8Slice {
				// Check if the SolverTypeData is not nil, so if the Solver specifies in the selector even filters
				if solverTypeData != nil {
					// Forge partition
					k8sliceSelector := solverTypeData.(nodecorev1alpha1.K8SliceSelector)
					klog.Infof("K8S Slice Selector: %v", k8sliceSelector)

					// Parse flavor from PeeringCandidate
					flavorTypeIdentifier, flavorData, err := nodecorev1alpha1.ParseFlavorType(&pc.Spec.Flavor)
					if err != nil {
						klog.Errorf("Error when parsing Flavor for Solver %s: %s", solver.Name, err)
						return ctrl.Result{}, err
					}
					if flavorTypeIdentifier != nodecorev1alpha1.TypeK8Slice {
						klog.Errorf("Flavor type is different from K8Slice as expected by the Solver selector for Solver %s", solver.Name)
					}

					// Force casting
					k8slice := flavorData.(nodecorev1alpha1.K8Slice)

					k8slicepartition := resourceforge.ForgeK8SliceConfiguration(k8sliceSelector, &k8slice)
					// Convert the K8SlicePartition to JSON
					k8slicepartitionJSON, err := json.Marshal(k8slicepartition)
					if err != nil {
						klog.Errorf("Error when marshaling K8SlicePartition for Solver %s: %s", solver.Name, err)
						return ctrl.Result{}, err
					}
					configuration = &nodecorev1alpha1.Configuration{
						ConfigurationTypeIdentifier: solverTypeIdentifier,
						// Partition Data is the raw data of the k8slicepartion
						ConfigurationData: runtime.RawExtension{Raw: k8slicepartitionJSON},
					}
				}
			}
			// TODO: If not K8Slice, implement the other types partitions if possible
		}

		// Get the NodeIdentity
		nodeIdentity := getters.GetNodeIdentity(ctx, r.Client)

		// Forge the Reservation
		reservation := resourceforge.ForgeReservation(&pc, configuration, *nodeIdentity)
		if err := r.Client.Create(ctx, reservation); err != nil {
			klog.Errorf("Error when creating Reservation for Solver %s: %s", solver.Name, err)
			return ctrl.Result{}, err
		}

		klog.Infof("Reservation %s created", reservation.Name)

		solver.SetReserveAndBuyStatus(nodecorev1alpha1.PhaseRunning)
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Reservation created")
		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	case nodecorev1alpha1.PhaseRunning:
		// Check solver expiration
		if tools.CheckExpirationSinceTime(solver.Status.SolverPhase.LastChangeTime, flags.ExpirationPhaseRunning) {
			klog.Infof("Solver %s has expired", req.NamespacedName.Name)
			solver.SetPhase(nodecorev1alpha1.PhaseTimeout, "Solver has expired before reserving the resources")

			if err := r.updateSolverStatus(ctx, solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		reservation := &reservationv1alpha1.Reservation{}
		resNamespaceName := types.NamespacedName{Name: namings.ForgeReservationName(solver.Name), Namespace: flags.FluidosNamespace}

		// Get the Reservation
		err := r.Get(ctx, resNamespaceName, reservation)
		if client.IgnoreNotFound(err) != nil {
			klog.Errorf("Error when getting Reservation for Solver %s: %s", solver.Name, err)
			return ctrl.Result{}, err
		}

		common.ReservationStatusCheck(solver, reservation)

		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	case nodecorev1alpha1.PhaseAllocating:
		klog.Infof("Solver %s has reserved and purchased the resources, creating the Allocation", req.NamespacedName.Name)
		// Create the Allocation
		//TODO: delete allocating phase
		return ctrl.Result{}, nil
	case nodecorev1alpha1.PhaseFailed:
		klog.Infof("Solver %s has failed to reserve and buy the resources", req.NamespacedName.Name)
		solver.SetPhase(nodecorev1alpha1.PhaseFailed, "Solver has failed to reserve the resources")
		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	default:
		solver.SetReserveAndBuyStatus(nodecorev1alpha1.PhaseIdle)
		// Update the Solver status
		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
}

func (r *SolverReconciler) handlePeering(ctx context.Context, req ctrl.Request, solver *nodecorev1alpha1.Solver) (ctrl.Result, error) {
	peeringStatus := solver.Status.Peering
	switch peeringStatus {
	case nodecorev1alpha1.PhaseIdle:
		klog.Infof("Solver %s is trying to establish a peering", req.NamespacedName.Name)

		contractNamespaceName := types.NamespacedName{Name: solver.Status.Contract.Name, Namespace: solver.Status.Contract.Namespace}
		contract := reservationv1alpha1.Contract{}
		err := r.Client.Get(ctx, contractNamespaceName, &contract)
		if err != nil {
			klog.Errorf("Error when getting Contract for Solver %s: %s", solver.Name, err)
			return ctrl.Result{}, err
		}
		// VirtualNode Name is not used in the current implementation
		vnName := namings.ForgeVirtualNodeName(contract.Spec.PeeringTargetCredentials.ClusterName)
		klog.Infof("Virtual Node Name: %s", vnName)

		allocation := resourceforge.ForgeAllocation(&contract, solver.Name)
		if err := r.Client.Create(ctx, allocation); err != nil {
			klog.Errorf("Error when creating Allocation for Solver %s: %s", solver.Name, err)
			return ctrl.Result{}, err
		}
		klog.Infof("Allocation %s created", allocation.Name)
		solver.Status.Allocation = nodecorev1alpha1.GenericRef{
			Name:      allocation.Name,
			Namespace: allocation.Namespace,
		}
		solver.SetPeeringStatus(nodecorev1alpha1.PhaseRunning)
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Allocation created")
		solver.Status.Credentials = contract.Spec.PeeringTargetCredentials
		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil

	case nodecorev1alpha1.PhaseRunning:
		klog.Info("Checking peering status")
		fc, err := fcutils.GetForeignClusterByID(ctx, r.Client, solver.Status.Credentials.ClusterID)
		if err != nil {
			if errors.IsNotFound(err) {
				klog.Infof("ForeignCluster %s not found", solver.Status.Credentials.ClusterID)
				// Retry later
				return ctrl.Result{Requeue: true}, nil
			}
			klog.Errorf("Error when getting ForeignCluster %s: %v", solver.Status.Credentials.ClusterID, err)
			solver.SetPeeringStatus(nodecorev1alpha1.PhaseFailed)
			if err := r.updateSolverStatus(ctx, solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		klog.Infof("ForeignCluster %s found", solver.Status.Credentials.ClusterID)
		if fcutils.IsOutgoingJoined(fc) &&
			fcutils.IsAuthenticated(fc) &&
			fcutils.IsNetworkingEstablishedOrExternal(fc) &&
			!fcutils.IsUnpeered(fc) {
			klog.Infof("Solver %s has established a peering", req.NamespacedName.Name)
			solver.SetPeeringStatus(nodecorev1alpha1.PhaseSolved)
			solver.SetPhase(nodecorev1alpha1.PhaseSolved, "Solver has established a peering")
			if err := r.updateSolverStatus(ctx, solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		klog.Infof("Solver %s is still peering", req.NamespacedName.Name)
		// Check solver expiration
		if tools.CheckExpirationSinceTime(solver.Status.SolverPhase.LastChangeTime, flags.ExpirationPhaseRunning) {
			klog.Infof("Solver %s has expired", req.NamespacedName.Name)
			solver.SetPhase(nodecorev1alpha1.PhaseTimeout, "Solver has expired before reserving the resources")
			solver.SetPeeringStatus(nodecorev1alpha1.PhaseFailed)
			if err := r.updateSolverStatus(ctx, solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{Requeue: true}, nil
	case nodecorev1alpha1.PhaseFailed:
		klog.Infof("Solver %s has failed to establish a peering", req.NamespacedName.Name)
		solver.SetPhase(nodecorev1alpha1.PhaseFailed, "Solver has failed to establish a peering")
		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	case nodecorev1alpha1.PhaseSolved:
		klog.Infof("Solver %s has established a peering", req.NamespacedName.Name)
		return ctrl.Result{}, nil
	default:
		solver.SetPeeringStatus(nodecorev1alpha1.PhaseIdle)
		// Update the Solver status
		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
}

func (r *SolverReconciler) searchPeeringCandidates(ctx context.Context,
	solver *nodecorev1alpha1.Solver) ([]advertisementv1alpha1.PeeringCandidate, error) {
	pc := advertisementv1alpha1.PeeringCandidateList{}
	result := []advertisementv1alpha1.PeeringCandidate{}

	// Get the Flavour Selector from the Solver
	selector := solver.Spec.Selector

	// Get the list of PeeringCandidates
	if err := r.List(ctx, &pc); err != nil {
		klog.Errorf("Error when listing PeeringCandidates: %s", err)
		return nil, err
	}

	// TODO: Maybe not needed
	if len(pc.Items) == 0 {
		klog.Infof("No PeeringCandidates found")
		return nil, errors.NewNotFound(schema.GroupResource{Group: "advertisement", Resource: "PeeringCandidate"}, "PeeringCandidate")
	}

	// Filter out the PeeringCandidates that are not available
	filtered := []advertisementv1alpha1.PeeringCandidate{}
	for i := range pc.Items {
		p := pc.Items[i]
		if p.Spec.Available {
			filtered = append(filtered, p)
		}
	}

	// Filter the list of PeeringCandidates based on the Flavour Selector
	for i := range filtered {
		p := filtered[i]
		res := common.FilterPeeringCandidate(selector, &p)
		if res {
			result = append(result, p)
		}
	}

	return result, nil
}

// TODO: unify this logic with the one of the discovery controller.
func (r *SolverReconciler) selectAndBookPeeringCandidate(
	pcList []advertisementv1alpha1.PeeringCandidate) (*advertisementv1alpha1.PeeringCandidate, error) {
	// Select the first PeeringCandidate

	var pc *advertisementv1alpha1.PeeringCandidate

	for i := range pcList {
		pc = &pcList[i]
		// Select the first PeeringCandidate that is not reserved
		// TODO: A better default logic should be implemented.
		if pc.Spec.Available {
			return pc, nil
		}
	}

	klog.Infof("No PeeringCandidate selected")
	return nil, errors.NewNotFound(schema.GroupResource{Group: "advertisement", Resource: "PeeringCandidate"}, "PeeringCandidate")
}

func (r *SolverReconciler) createOrGetDiscovery(ctx context.Context, solver *nodecorev1alpha1.Solver) (*advertisementv1alpha1.Discovery, error) {
	discovery := &advertisementv1alpha1.Discovery{}

	// Get the Discovery
	if err := r.Get(ctx, types.NamespacedName{Name: namings.ForgeDiscoveryName(solver.Name),
		Namespace: flags.FluidosNamespace}, discovery); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting Discovery for Solver %s: %s", solver.Name, err)
		return nil, err
	} else if err != nil {
		// Create the Discovery
		discovery = resourceforge.ForgeDiscovery(solver.Spec.Selector, solver.Name)
		if err := r.Client.Create(ctx, discovery); err != nil {
			klog.Errorf("Error when creating Discovery for Solver %s: %s", solver.Name, err)
			return nil, err
		}
	}
	return discovery, nil
}

func (r *SolverReconciler) updateSolverStatus(ctx context.Context, solver *nodecorev1alpha1.Solver) error {
	return r.Status().Update(ctx, solver)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SolverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nodecorev1alpha1.Solver{}).
		Watches(&advertisementv1alpha1.Discovery{}, handler.EnqueueRequestsFromMapFunc(
			r.discoveryToSolver,
		), builder.WithPredicates(discoveryPredicate())).
		Watches(&reservationv1alpha1.Reservation{}, handler.EnqueueRequestsFromMapFunc(
			r.reservationToSolver,
		), builder.WithPredicates(reservationPredicate())).
		Watches(&nodecorev1alpha1.Allocation{}, handler.EnqueueRequestsFromMapFunc(
			r.allocationToSolver,
		), builder.WithPredicates(allocationPredicate(mgr.GetClient()))).
		Complete(r)
}

func discoveryPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectNew.(*advertisementv1alpha1.Discovery).Status.Phase.Phase == nodecorev1alpha1.PhaseSolved ||
				e.ObjectNew.(*advertisementv1alpha1.Discovery).Status.Phase.Phase == nodecorev1alpha1.PhaseFailed
		},
	}
}

func reservationPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectNew.(*reservationv1alpha1.Reservation).Status.Phase.Phase == nodecorev1alpha1.PhaseSolved ||
				e.ObjectNew.(*reservationv1alpha1.Reservation).Status.Phase.Phase == nodecorev1alpha1.PhaseFailed
		},
	}
}

func allocationPredicate(c client.Client) predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return (e.ObjectNew.(*nodecorev1alpha1.Allocation).Status.Status == nodecorev1alpha1.Active ||
				e.ObjectNew.(*nodecorev1alpha1.Allocation).Status.Status == nodecorev1alpha1.Released) &&
				!IsProvider(context.Background(), e.ObjectNew.(*nodecorev1alpha1.Allocation), c) &&
				e.ObjectNew.(*nodecorev1alpha1.Allocation).Spec.IntentID != ""
		},
	}
}

func (r *SolverReconciler) discoveryToSolver(_ context.Context, o client.Object) []reconcile.Request {
	solverName := namings.RetrieveSolverNameFromDiscovery(o.GetName())
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      solverName,
				Namespace: flags.FluidosNamespace,
			},
		},
	}
}

func (r *SolverReconciler) reservationToSolver(_ context.Context, o client.Object) []reconcile.Request {
	solverName := namings.RetrieveSolverNameFromReservation(o.GetName())
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      solverName,
				Namespace: flags.FluidosNamespace,
			},
		},
	}
}

func (r *SolverReconciler) allocationToSolver(_ context.Context, o client.Object) []reconcile.Request {
	solverName := o.(*nodecorev1alpha1.Allocation).Spec.IntentID
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      solverName,
				Namespace: flags.FluidosNamespace,
			},
		},
	}
}

// IsProvider returns true if the allocation is a provider allocation base on the contract reference.
func IsProvider(ctx context.Context, allocation *nodecorev1alpha1.Allocation, c client.Client) bool {
	// Get the contract reference
	contractRef := allocation.Spec.Contract
	// Get the contract
	contract := &reservationv1alpha1.Contract{}
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: contractRef.Namespace,
		Name:      contractRef.Name,
	}, contract); err != nil {
		return false
	}

	// Get the seller information
	seller := contract.Spec.Seller

	ni := getters.GetNodeIdentity(ctx, c)

	// Check if the seller is the same as the local node
	if seller.NodeID != ni.NodeID || seller.IP != ni.IP || seller.Domain != ni.Domain {
		return false
	}

	return true
}
