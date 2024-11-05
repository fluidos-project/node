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
	"fmt"

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
			result, err := r.handleFindCandidate(ctx, req, &solver)
			if err != nil {
				solver.SetPhase(nodecorev1alpha1.PhaseFailed, "Error when finding a candidate: "+err.Error())
				if err := r.updateSolverStatus(ctx, &solver); err != nil {
					klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
			}
			return result, err
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
			result, err := r.handleReserveAndBuy(ctx, req, &solver)
			if err != nil {
				solver.SetPhase(nodecorev1alpha1.PhaseFailed, "Error when reserving and buying the resources: "+err.Error())
				if err := r.updateSolverStatus(ctx, &solver); err != nil {
					klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
			}
			return result, err
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
			pc, err := r.searchReservedPeeringCandidate(ctx, &solver)
			if err != nil {
				klog.Errorf("Error when searching and booking a candidate for Solver %s: %s", req.NamespacedName.Name, err)
				return ctrl.Result{}, err
			}

			// Get Contract by PeeringCandidate
			contract, err := r.getContractByPeeringCandidate(ctx, pc)
			if err != nil {
				klog.Errorf("Error when getting Contract for Solver %s: %s", req.NamespacedName.Name, err)
				return ctrl.Result{}, err
			}

			result, err := r.handlePeering(ctx, req, &solver, contract)
			if err != nil {
				solver.SetPhase(nodecorev1alpha1.PhaseFailed, "Error when establishing a peering: "+err.Error())
				if err := r.updateSolverStatus(ctx, &solver); err != nil {
					klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
			}
			return result, err
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

	// Check if all the phases are solved
	if solver.Status.FindCandidate == nodecorev1alpha1.PhaseSolved &&
		solver.Status.ReserveAndBuy == nodecorev1alpha1.PhaseSolved &&
		solver.Status.Peering == nodecorev1alpha1.PhaseSolved {
		solver.SetPhase(nodecorev1alpha1.PhaseSolved, "Solver has completed all the phases")
		if err := r.updateSolverStatus(ctx, &solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
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
			selectedPc, err := r.selectAvaiablePeeringCandidate(pc)
			if err != nil {
				klog.Errorf("Error when selecting and booking a candidate for Solver %s: %s", req.NamespacedName.Name, err)
				return ctrl.Result{}, err
			}

			// Update the PeeringCandidate InterestedSolverIDs
			selectedPc.Spec.InterestedSolverIDs = append(selectedPc.Spec.InterestedSolverIDs, solver.Name)
			if err := r.Client.Update(ctx, selectedPc); err != nil {
				klog.Errorf("Error when updating PeeringCandidate %s for Solver %s: %s", selectedPc.Name, solver.Name, err)
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

func (r *SolverReconciler) getContractByPeeringCandidate(ctx context.Context, pc *advertisementv1alpha1.PeeringCandidate) (
	*reservationv1alpha1.Contract, error) {
	contract := &reservationv1alpha1.Contract{}
	// Get the contract list
	contractList := &reservationv1alpha1.ContractList{}
	if err := r.List(ctx, contractList, client.InNamespace(pc.Namespace)); err != nil {
		return nil, err
	}

	flavorName := pc.Spec.Flavor.Name
	for i := 0; i < len(contractList.Items); i++ {
		c := contractList.Items[i]
		if c.Spec.Flavor.Name == flavorName {
			contract = &c
			break
		}
	}

	if contract.Name == "" {
		return nil, errors.NewNotFound(schema.GroupResource{
			Group:    reservationv1alpha1.GroupVersion.Group,
			Resource: "contracts",
		}, pc.Name)
	}

	return contract, nil
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
				if !contains(pc.Spec.InterestedSolverIDs, solver.Name) {
					pc.Spec.InterestedSolverIDs = append(pc.Spec.InterestedSolverIDs, solver.Name)
					if err := r.Client.Update(ctx, &pc); err != nil {
						klog.Errorf("Error when updating PeeringCandidate %s for Solver %s: %s", pc.Name, solver.Name, err)
						return ctrl.Result{}, err
					}
				}
				// TODO: Logic is not so strong: implementing workaround to select first peering candidate
				break
			}
		}

		if pc.Name == "" {
			klog.Errorf("No PeeringCandidate found for Solver %s:", solver.Name)
			return ctrl.Result{}, nil
		}

		if solver.Spec.Selector != nil {
			// Parse Solver Selector
			solverTypeIdentifier, solverTypeData, err := nodecorev1alpha1.ParseSolverSelector(solver.Spec.Selector)
			if err != nil {
				klog.Errorf("Error when parsing Solver Selector for Solver %s: %s", solver.Name, err)
				return ctrl.Result{}, err
			}
			switch solverTypeIdentifier {
			case nodecorev1alpha1.TypeK8Slice:
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

					k8SliceConfiguration := resourceforge.ForgeK8SliceConfiguration(k8sliceSelector, &k8slice)
					// Convert the K8SlicePartition to JSON
					k8SliceConfigurationJSON, err := json.Marshal(k8SliceConfiguration)
					if err != nil {
						klog.Errorf("Error when marshaling K8SlicePartition for Solver %s: %s", solver.Name, err)
						return ctrl.Result{}, err
					}
					configuration = &nodecorev1alpha1.Configuration{
						ConfigurationTypeIdentifier: solverTypeIdentifier,
						// Partition Data is the raw data of the k8slicepartion
						ConfigurationData: runtime.RawExtension{Raw: k8SliceConfigurationJSON},
					}
				}
			case nodecorev1alpha1.TypeVM:
				// TODO(VM): Implement the VM configuration forging if possible
				klog.Warningf("Solver %s has a VM selector, but the VM configuration forging is not implemented yet", solver.Name)
			case nodecorev1alpha1.TypeService:

				// Parse flavor from PeeringCandidate
				flavorTypeIdentifier, flavorData, err := nodecorev1alpha1.ParseFlavorType(&pc.Spec.Flavor)
				if err != nil {
					klog.Errorf("Error when parsing Flavor for Solver %s: %s", solver.Name, err)
					return ctrl.Result{}, err
				}

				if flavorTypeIdentifier != nodecorev1alpha1.TypeService {
					klog.Errorf("Flavor type is different from Service as expected by the Solver selector for Solver %s", solver.Name)
					return ctrl.Result{}, fmt.Errorf("flavor type is different from Service as expected by the Solver selector for Solver %s", solver.Name)
				}

				// Force casting
				service := flavorData.(nodecorev1alpha1.ServiceFlavor)

				// Get the default service configuration based on the service category
				serviceConfiguration, err := resourceforge.ForgeDefaultServiceConfiguration(&service)
				if err != nil {
					klog.Errorf("Error when forging the default Service configuration for Solver %s: %s", solver.Name, err)
					return ctrl.Result{}, err
				}

				// Convert the ServiceConfiguration to JSON
				serviceConfigurationJSON, err := json.Marshal(serviceConfiguration)
				if err != nil {
					klog.Errorf("Error when marshaling ServiceConfiguration for Solver %s: %s", solver.Name, err)
					return ctrl.Result{}, err
				}
				configuration = &nodecorev1alpha1.Configuration{
					ConfigurationTypeIdentifier: solverTypeIdentifier,
					// Configuration Data is the raw data of the service configuration
					ConfigurationData: runtime.RawExtension{Raw: serviceConfigurationJSON},
				}
				// TODO(Service): Implement the Service configuration forging if possible
			case nodecorev1alpha1.TypeSensor:
				// TODO(Sensor): Implement the Sensor configuration forging if possible
				klog.Warningf("Solver %s has a Sensor selector, but the Sensor configuration forging is not implemented yet", solver.Name)
			default:
				klog.Errorf("Solver %s has an unsupported selector type", solver.Name)
			}
		}

		// Get the NodeIdentity
		nodeIdentity := getters.GetNodeIdentity(ctx, r.Client)

		// Forge the Reservation
		reservation := resourceforge.ForgeReservation(&pc, configuration, *nodeIdentity, solver.Name)
		if err := r.Client.Create(ctx, reservation); err != nil {
			klog.Errorf("Error when creating Reservation for Solver %s: %s", solver.Name, err)
			return ctrl.Result{}, err
		}

		klog.Infof("Reservation %s created", reservation.Name)

		solver.SetReserveAndBuyStatus(nodecorev1alpha1.PhaseRunning)
		klog.Infof("Solver set ReserveAndBuy status to %s", solver.Status.ReserveAndBuy)
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Reservation created")
		klog.Infof("Solver set phase to %s", solver.Status.SolverPhase.Phase)
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

		klog.Infof("Solver status changing to %s", solver.Status.SolverPhase.Phase)
		return ctrl.Result{}, nil
	case nodecorev1alpha1.PhaseAllocating:
		klog.Infof("Solver %s has reserved and purchased the resources, creating the Allocation", req.NamespacedName.Name)
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

func contains(slice []string, s string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}
	return false
}

func (r *SolverReconciler) handlePeering(
	ctx context.Context,
	req ctrl.Request,
	solver *nodecorev1alpha1.Solver,
	contract *reservationv1alpha1.Contract,
) (ctrl.Result, error) {
	peeringStatus := solver.Status.Peering
	switch peeringStatus {
	case nodecorev1alpha1.PhaseIdle:
		// Allocation creation phase
		klog.Infof("Solver %s is trying to establish a peering", req.NamespacedName.Name)

		contractNamespaceName := types.NamespacedName{Name: contract.Name, Namespace: contract.Namespace}
		contract := reservationv1alpha1.Contract{}
		err := r.Client.Get(ctx, contractNamespaceName, &contract)
		if err != nil {
			klog.Errorf("Error when getting Contract for Solver %s: %s", solver.Name, err)
			return ctrl.Result{}, err
		}
		// VirtualNode Name is not used in the current implementation
		vnName := namings.ForgeVirtualNodeName(contract.Spec.PeeringTargetCredentials.ClusterName)
		klog.Infof("Virtual Node Name: %s", vnName)

		allocation := resourceforge.ForgeAllocation(&contract)
		if err := r.Client.Create(ctx, allocation); err != nil {
			klog.Errorf("Error when creating Allocation for Solver %s: %s", solver.Name, err)
			return ctrl.Result{}, err
		}
		klog.Infof("Allocation %s created", allocation.Name)

		solver.SetPeeringStatus(nodecorev1alpha1.PhaseRunning)
		klog.Infof("Solver set peering status to %s", solver.Status.Peering)
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Allocation created")
		klog.Infof("Solver set phase to %s", solver.Status.SolverPhase.Phase)
		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}

		klog.Infof("Solver %s has created the Allocation %s, moving to %s phase", req.NamespacedName.Name, allocation.Name, nodecorev1alpha1.PhaseRunning)

		return ctrl.Result{}, nil
	case nodecorev1alpha1.PhaseRunning:
		klog.Info("Checking Allocation status")

		// Check solver expiration
		if tools.CheckExpirationSinceTime(solver.Status.SolverPhase.LastChangeTime, flags.ExpirationSolver) {
			klog.Infof("Solver %s has expired", req.NamespacedName.Name)
			solver.SetPhase(nodecorev1alpha1.PhaseTimeout, "Solver has expired before reserving the resources")
			solver.SetPeeringStatus(nodecorev1alpha1.PhaseFailed)
			if err := r.updateSolverStatus(ctx, solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		// Get the Reservation with spec SolverID == Solver.Name
		reservationList := reservationv1alpha1.ReservationList{}
		if err := r.List(ctx, &reservationList); err != nil {
			klog.Errorf("Error when listing Reservations: %s", err)
			return ctrl.Result{}, err
		}

		reservation := &reservationv1alpha1.Reservation{}
		for i := range reservationList.Items {
			if reservationList.Items[i].Spec.SolverID == solver.Name {
				reservation = &reservationList.Items[i]
				break
			}
		}

		if reservation.Name == "" {
			klog.Infof("No Reservation found for Solver %s", solver.Name)
			return ctrl.Result{}, errors.NewNotFound(schema.GroupResource{Group: "reservation", Resource: "Reservation"}, "Reservation")
		}

		// Get the Allocations
		allocationList := nodecorev1alpha1.AllocationList{}
		if err := r.List(ctx, &allocationList); err != nil {
			klog.Errorf("Error when listing Allocations: %s", err)
			return ctrl.Result{}, err
		}

		allocation := &nodecorev1alpha1.Allocation{}
		for i := range allocationList.Items {
			if allocationList.Items[i].Spec.Contract.Name == reservation.Status.Contract.Name &&
				allocationList.Items[i].Spec.Contract.Namespace == reservation.Status.Contract.Namespace {
				allocation = &allocationList.Items[i]
				break
			}
		}

		common.AllocationStatusCheck(solver, allocation)

		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
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

func (r *SolverReconciler) searchReservedPeeringCandidate(ctx context.Context,
	solver *nodecorev1alpha1.Solver) (*advertisementv1alpha1.PeeringCandidate, error) {
	var reservationList reservationv1alpha1.ReservationList

	// Get the Reservation from the Solver
	reservation := &reservationv1alpha1.Reservation{}

	// Get reservation list and check the one that is related to the Solver
	if err := r.List(ctx, &reservationList); err != nil {
		klog.Errorf("Error when listing Reservations: %s", err)
		return nil, err
	}

	for i := range reservationList.Items {
		if reservationList.Items[i].Spec.SolverID == solver.Name {
			reservation = &reservationList.Items[i]
			break
		}
	}

	if reservation.Name == "" {
		klog.Infof("No Reservation found for Solver %s", solver.Name)
		return nil, errors.NewNotFound(schema.GroupResource{Group: "reservation", Resource: "Reservation"}, "Reservation")
	}

	// Get the PeeringCandidate from the Reservation
	pc := &advertisementv1alpha1.PeeringCandidate{}
	pcNamespaceName := types.NamespacedName{Name: reservation.Spec.PeeringCandidate.Name, Namespace: reservation.Namespace}
	if err := r.Get(ctx, pcNamespaceName, pc); err != nil {
		klog.Errorf("Error when getting PeeringCandidate for Solver %s: %s", solver.Name, err)
		return nil, err
	}

	return pc, nil
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
		// PeeringCandidate needs to be available
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

func (r *SolverReconciler) selectAvaiablePeeringCandidate(
	pcList []advertisementv1alpha1.PeeringCandidate) (*advertisementv1alpha1.PeeringCandidate,
	error) {
	var pc *advertisementv1alpha1.PeeringCandidate

	for i := range pcList {
		pc = &pcList[i]

		if pc.Spec.Available {
			return pc, nil
		}
	}

	klog.Infof("No PeeringCandidate selected")
	return nil, errors.NewNotFound(schema.GroupResource{Group: "advertisement", Resource: "PeeringCandidate"}, "PeeringCandidate")
}

func (r *SolverReconciler) createOrGetDiscovery(ctx context.Context,
	solver *nodecorev1alpha1.Solver) (*advertisementv1alpha1.Discovery, error) {
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
	klog.Infof("Updating Solver %s status", solver.Name)
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
				!IsProvider(context.Background(), e.ObjectNew.(*nodecorev1alpha1.Allocation), c)
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
	var solverName string
	contractName := o.(*nodecorev1alpha1.Allocation).Spec.Contract.Name
	// Get the reservation with status.Contract.Name == solverName
	reservationList := &reservationv1alpha1.ReservationList{}
	if err := r.Client.List(context.Background(), reservationList); err != nil {
		klog.Errorf("Error when listing Reservations: %s", err)
		return nil
	}

	for i := range reservationList.Items {
		if reservationList.Items[i].Status.Contract.Name == contractName {
			solverName = reservationList.Items[i].Spec.SolverID
			break
		}
	}

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
