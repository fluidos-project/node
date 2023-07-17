/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rearmanager

import (
	"context"

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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"sigs.k8s.io/controller-runtime/pkg/predicate"

	advertisementv1alpha1 "fluidos.eu/node/api/advertisement/v1alpha1"
	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	reservationv1alpha1 "fluidos.eu/node/api/reservation/v1alpha1"
	"fluidos.eu/node/pkg/utils/common"
	"fluidos.eu/node/pkg/utils/flags"
	"fluidos.eu/node/pkg/utils/namings"
	"fluidos.eu/node/pkg/utils/resourceforge"
	"fluidos.eu/node/pkg/utils/tools"
)

// SolverReconciler reconciles a Solver object
type SolverReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=solvers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=solvers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=solvers/finalizers,verbs=update

func (r *SolverReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "solver", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)
	//_ = log.FromContext(ctx)

	var solver nodecorev1alpha1.Solver
	if err := r.Get(ctx, req.NamespacedName, &solver); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting Solver %s before reconcile: %s", req.NamespacedName, err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("Solver %s not found, probably deleted", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	if solver.Status.SolverPhase.Phase != nodecorev1alpha1.PhaseSolved &&
		solver.Status.SolverPhase.Phase != nodecorev1alpha1.PhaseTimeout &&
		solver.Status.SolverPhase.Phase != nodecorev1alpha1.PhaseFailed &&
		solver.Status.SolverPhase.Phase != nodecorev1alpha1.PhaseRunning &&
		solver.Status.SolverPhase.Phase != nodecorev1alpha1.PhaseIdle {

		solver.SetPhase(nodecorev1alpha1.PhaseIdle, "Solver initialized")

		if err := r.updateSolverStatus(ctx, &solver); err != nil {
			klog.Errorf("Error when updating Solver %s status before reconcile: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	klog.Infof("Reconciling Solver %s", req.NamespacedName)

	findCandidateStatus := solver.Status.FindCandidate
	reserveAndBuyStatus := solver.Status.ReserveAndBuy

	// Check if the Solver has expired or failed, in this case do nothing and return
	if solver.Status.SolverPhase.Phase == nodecorev1alpha1.PhaseFailed || solver.Status.SolverPhase.Phase == nodecorev1alpha1.PhaseTimeout {
		return ctrl.Result{}, nil
	}
	if solver.Spec.FindCandidate {
		switch findCandidateStatus {
		case nodecorev1alpha1.PhaseIdle:
			// Search a matching PeeringCandidate if available
			pc, err := r.searchPeeringCandidates(ctx, &solver)
			if client.IgnoreNotFound(err) != nil {
				klog.Errorf("Error when searching and booking a candidate for Solver %s: %s", req.NamespacedName.Name, err)
				return ctrl.Result{}, err
			}

			// If some PeeringCandidates are available, select one and book it
			if len(pc) > 0 {
				// If some PeeringCandidates are available, select one and book it
				selectedPc, err := r.selectAndBookPeeringCandidate(ctx, &solver, pc)
				if err != nil {
					klog.Errorf("Error when selecting and booking a candidate for Solver %s: %s", req.NamespacedName.Name, err)
					return ctrl.Result{}, err
				}
				klog.Infof("Solver %s has selected and booked candidate %s", req.NamespacedName.Name, selectedPc.Name)
				solver.SetFindCandidateStatus(nodecorev1alpha1.PhaseSolved)
				solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Solver has found a candidate")
				if err := r.updateSolverStatus(ctx, &solver); err != nil {
					klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, nil
			}

			// If no PeeringCandidate is available, Create a Discovery
			klog.Infof("Solver %s has not found any candidate. Trying a Discovery", req.NamespacedName.Name)
			solver.SetFindCandidateStatus(nodecorev1alpha1.PhaseRunning)
			solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Solver is trying a Discovery")

			// Update the Solver status
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		case nodecorev1alpha1.PhaseRunning:
			// Check solver expiration
			if tools.CheckExpiration(solver.Status.SolverPhase.LastChangeTime, flags.EXPIRATION_PHASE_RUNNING) {
				klog.Infof("Solver %s has expired", req.NamespacedName.Name)

				solver.SetPhase(nodecorev1alpha1.PhaseTimeout, "Solver has expired before finding a candidate")

				if err := r.updateSolverStatus(ctx, &solver); err != nil {
					klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, nil
			}

			klog.Infof("Getting or creating Discovery for Solver %s", req.NamespacedName.Name)
			discovery, err := r.createOrGetDiscovery(ctx, &solver)
			if err != nil {
				klog.Errorf("Error when creating or getting Discovery for Solver %s: %s", req.NamespacedName.Name, err)
				return ctrl.Result{}, err
			}

			common.DiscoveryStatusCheck(&solver, discovery)

			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		case nodecorev1alpha1.PhaseFailed:
			klog.Infof("Solver %s has not found any candidate", req.NamespacedName.Name)
			solver.SetPhase(nodecorev1alpha1.PhaseFailed, "Solver has not found any candidate")
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		case nodecorev1alpha1.PhaseSolved:
			klog.Infof("Solver %s has found a candidate", req.NamespacedName.Name)
			break
		default:
			solver.SetFindCandidateStatus(nodecorev1alpha1.PhaseIdle)
			solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Solver is running")
			// Update the Solver status
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	if solver.Spec.ReserveAndBuy && findCandidateStatus == nodecorev1alpha1.PhaseSolved {
		switch reserveAndBuyStatus {
		case nodecorev1alpha1.PhaseIdle:
			klog.Infof("Creating the Reservation", req.NamespacedName.Name)
			// Create the Reservation
			var pc advertisementv1alpha1.PeeringCandidate
			pcNamespaceName := types.NamespacedName{Name: solver.Status.PeeringCandidate.Name, Namespace: solver.Status.PeeringCandidate.Namespace}

			// Get the PeeringCandidate from the Solver
			if err := r.Get(ctx, pcNamespaceName, &pc); err != nil {
				klog.Errorf("Error when getting PeeringCandidate %s: %s", solver.Status.PeeringCandidate.Name, err)
				return ctrl.Result{}, err
			}

			// Forge the Reservation
			reservation := resourceforge.ForgeReservation(pc)
			if err := r.Client.Create(ctx, reservation); err != nil {
				klog.Errorf("Error when creating Reservation for Solver %s: %s", solver.Name, err)
				return ctrl.Result{}, err
			}

			klog.Infof("Reservation %s created", reservation.Name)

			solver.SetReserveAndBuyStatus(nodecorev1alpha1.PhaseRunning)
			solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Reservation created")
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		case nodecorev1alpha1.PhaseRunning:
			// Check solver expiration
			if tools.CheckExpiration(solver.Status.SolverPhase.LastChangeTime, flags.EXPIRATION_PHASE_RUNNING) {
				klog.Infof("Solver %s has expired", req.NamespacedName.Name)
				solver.SetPhase(nodecorev1alpha1.PhaseTimeout, "Solver has expired before reserving the resources")

				if err := r.updateSolverStatus(ctx, &solver); err != nil {
					klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, nil
			}

			reservation := &reservationv1alpha1.Reservation{}
			flavourID := namings.RetrieveFlavourNameFromPC(solver.Status.PeeringCandidate.Name)
			resNamespaceName := types.NamespacedName{Name: namings.ForgeReservationName(flavourID), Namespace: flags.RESERVATION_DEFAULT_NAMESPACE}

			// Get the Reservation
			err := r.Get(ctx, resNamespaceName, reservation)
			if client.IgnoreNotFound(err) != nil {
				klog.Errorf("Error when getting Reservation for Solver %s: %s", solver.Name, err)
				return ctrl.Result{}, err
			}

			common.ReservationStatusCheck(&solver, reservation)

			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil

		case nodecorev1alpha1.PhaseFailed:
			klog.Infof("Solver %s has failed to reserve and buy the resources", req.NamespacedName.Name)
			solver.SetPhase(nodecorev1alpha1.PhaseFailed, "Solver has failed to reserve the resources")
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		case nodecorev1alpha1.PhaseSolved:
			klog.Infof("Solver %s has reserved and purchased the resources", req.NamespacedName.Name)
			break
		default:
			solver.SetReserveAndBuyStatus(nodecorev1alpha1.PhaseIdle)
			// Update the Solver status
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		switch reserveAndBuyStatus {
		}
	}

	if solver.Spec.EnstablishPeering && reserveAndBuyStatus == nodecorev1alpha1.PhaseSolved {
		// Peering phase to be implemented
	}

	return ctrl.Result{}, nil
}

func (r *SolverReconciler) searchPeeringCandidates(ctx context.Context, solver *nodecorev1alpha1.Solver) ([]advertisementv1alpha1.PeeringCandidate, error) {
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

	// Filter the reserved PeeringCandidates
	filtered := []advertisementv1alpha1.PeeringCandidate{}
	for _, p := range pc.Items {
		if p.Spec.Reserved == false && p.Spec.SolverID == "" {
			filtered = append(filtered, p)
		}
	}

	// Filter the list of PeeringCandidates based on the Flavour Selector
	for _, p := range filtered {
		res := common.FilterPeeringCandidate(&selector, &p)
		if res {
			result = append(result, p)
		}
	}

	return result, nil
}

// TODO: unify this logic with the one of the discovery controller
func (r *SolverReconciler) selectAndBookPeeringCandidate(ctx context.Context, solver *nodecorev1alpha1.Solver, pcList []advertisementv1alpha1.PeeringCandidate) (*advertisementv1alpha1.PeeringCandidate, error) {
	// Select the first PeeringCandidate

	var selected *advertisementv1alpha1.PeeringCandidate

	for _, pc := range pcList {
		// Select the first PeeringCandidate that is not reserved
		if pc.Spec.Reserved == false && pc.Spec.SolverID == "" {
			// Book the PeeringCandidate
			pc.Spec.Reserved = true
			pc.Spec.SolverID = solver.Name

			// Update the PeeringCandidate
			if err := r.Update(ctx, &pc); err != nil {
				klog.Errorf("Error when updating PeeringCandidate %s: %s", selected.Name, err)
				continue
			}

			// Getting the just updated PeeringCandidate
			if err := r.Get(ctx, types.NamespacedName{Name: pc.Name, Namespace: pc.Namespace}, selected); err != nil {
				klog.Errorf("Error when getting the reserved PeeringCandidate %s: %s", selected.Name, err)
				continue
			}

			// Check if the PeeringCandidate has been reserved correctly
			if pc.Spec.Reserved == false || pc.Spec.SolverID != solver.Name {
				klog.Errorf("Error when reserving PeeringCandidate %s. Trying with another one", selected.Name)
				continue
			}

			break
		}
	}

	// check if a PeeringCandidate has been selected
	if selected == nil || selected.Name == "" {
		klog.Infof("No PeeringCandidate selected")
		return nil, errors.NewNotFound(schema.GroupResource{Group: "advertisement", Resource: "PeeringCandidate"}, "PeeringCandidate")
	}

	return selected, nil
}

/* func (r *SolverReconciler) createOrUpdateDiscovery(ctx context.Context, solver *nodecorev1alpha1.Solver) (*nodecorev1alpha1.GenericRef, error) {
	pc := &nodecorev1alpha1.GenericRef{}
	// Create the Discovery
	discovery := resourceforge.ForgeDiscovery(solver.Spec.Selector, solver.Name)
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, discovery, func() error {
		if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseSolved {
			// Get the PeeringCandidate
			pc = &discovery.Spec.PeeringCandidate
			return nil
		}
		if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseTimeout || discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseIdle {
			return nil
		}
		if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseRunning {
			// Check if the phase has expired
			if common.CheckExpiration(discovery.Status.Phase.LastChangeTime, flags.EXPIRATION_PHASE_RUNNING) {
				discovery.Status.Phase.Phase = nodecorev1alpha1.PhaseTimeout
				discovery.Status.Phase.LastChangeTime = common.GetTimeNow()
				return nil
			}
			return nil
		}
		pc = nil
		return nil
	}); err != nil {
		klog.Errorf("Error when creating Discovery for Solver %s: %s", solver.Name, err)
		return nil, err
	}
	if pc != nil {
		return pc, nil
	}
	return nil, nil
} */

func (r *SolverReconciler) createOrGetDiscovery(ctx context.Context, solver *nodecorev1alpha1.Solver) (*advertisementv1alpha1.Discovery, error) {
	discovery := &advertisementv1alpha1.Discovery{}

	// Get the Discovery
	if err := r.Get(ctx, types.NamespacedName{Name: namings.ForgeDiscoveryName(solver.Name), Namespace: flags.DISCOVERY_DEFAULT_NAMESPACE}, discovery); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting Discovery for Solver %s: %s", solver.Name, err)
		return nil, err
	} else if err != nil {
		// Create the Discovery
		discovery := resourceforge.ForgeDiscovery(solver.Spec.Selector, solver.Name)
		if err := r.Client.Create(ctx, discovery); err != nil {
			klog.Errorf("Error when creating Discovery for Solver %s: %s", solver.Name, err)
			return nil, err
		}
	}
	return discovery, nil
}

/* func (r *SolverReconciler) createOrGetReservation(ctx context.Context, solver *nodecorev1alpha1.Solver) (*reservationv1alpha1.Reservation, error) {
	reservation := &reservationv1alpha1.Reservation{}
	flavourID := namings.ForgeFlavourNameFromPC(solver.Status.PeeringCandidate.Name)
	resNamespaceName := types.NamespacedName{Name: namings.ForgeReservationName(flavourID), Namespace: flags.RESERVATION_DEFAULT_NAMESPACE}
	pcNamespaceName := types.NamespacedName{Name: solver.Status.PeeringCandidate.Name, Namespace: solver.Status.PeeringCandidate.Namespace}

	// Get the Reservation
	if err := r.Get(ctx, resNamespaceName, reservation); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting Reservation for Solver %s: %s", solver.Name, err)
		return nil, err
	} else if err != nil {
		// Create the Reservation
		var pc *advertisementv1alpha1.PeeringCandidate

		// Get the PeeringCandidate from the Solver
		if err := r.Get(ctx, pcNamespaceName, pc); err != nil {
			klog.Errorf("Error when getting PeeringCandidate %s: %s", solver.Status.PeeringCandidate.Name, err)
			return nil, err
		}

		// Forge the Reservation
		reservation := resourceforge.ForgeReservationCustomResource(*pc)
		if err := r.Client.Create(ctx, reservation); err != nil {
			klog.Errorf("Error when creating Reservation for Solver %s: %s", solver.Name, err)
			return nil, err
		}
	}
	return reservation, nil
} */

func (r *SolverReconciler) updateSolverStatus(ctx context.Context, solver *nodecorev1alpha1.Solver) error {
	return r.Status().Update(ctx, solver)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SolverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nodecorev1alpha1.Solver{}).
		//Owns(&advertisementv1alpha1.Discovery{}).
		//Owns(&reservationv1alpha1.Contract{}, builder.WithPredicates()).
		/* Watches(&source.Kind{Type: &advertisementv1alpha1.Discovery{}}, &handler.EnqueueRequestForObject{}, builder.WithPredicates(predicate.Funcs{})).WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				return e.ObjectNew.(*advertisementv1alpha1.Discovery).Status.Phase.Phase == nodecorev1alpha1.PhaseSolved
			},
		}). */
		// Here we want to watch Discovery resources which status is Solved
		// This is because we want to create a contract only when a Discovery has found a candidate
		Watches(&advertisementv1alpha1.Discovery{}, handler.EnqueueRequestsFromMapFunc(
			r.discoveryToSolver,
		), builder.WithPredicates(discoveryPredicate())).
		Watches(&reservationv1alpha1.Reservation{}, handler.EnqueueRequestsFromMapFunc(
			r.reservationToSolver,
		), builder.WithPredicates(reservationPredicate())).
		// Here we want to watch Contract resources which status is Solved

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

func (r *SolverReconciler) discoveryToSolver(ctx context.Context, o client.Object) []reconcile.Request {
	solverName := namings.RetrieveSolverNameFromDiscovery(o.GetName())
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      solverName,
				Namespace: flags.SOLVER_DEFAULT_NAMESPACE,
			},
		},
	}
}

func (r *SolverReconciler) reservationToSolver(ctx context.Context, o client.Object) []reconcile.Request {
	solverName := namings.RetrieveSolverNameFromReservation(o.GetName())
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      solverName,
				Namespace: flags.SOLVER_DEFAULT_NAMESPACE,
			},
		},
	}
}
