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
	"strings"

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

		setSolverPhase(&solver, nodecorev1alpha1.PhaseIdle, "Solver initialized")

		if err := r.updateSolverStatus(ctx, &solver); err != nil {
			klog.Errorf("Error when updating Solver %s status before reconcile: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	klog.Infof("Reconciling Solver %s", req.NamespacedName)

	if solver.Spec.FindCandidate {
		candidatePhase := solver.Status.CandidatesPhase
		klog.Infof("Candidate phase is %s", candidatePhase)

		switch candidatePhase {
		case nodecorev1alpha1.PhaseIdle:

			// Search a matching PeeringCandidate if available
			pc, err := r.searchPeeringCandidates(ctx, &solver)
			if client.IgnoreNotFound(err) != nil {
				klog.Errorf("Error when searching and booking a candidate for Solver %s: %s", req.NamespacedName.Name, err)
				return ctrl.Result{}, err
			}

			if len(pc) > 0 {
				// If some PeeringCandidates are available, select one and book it
				selectedPc, err := r.selectAndBookPeeringCandidate(ctx, &solver, pc)
				if err != nil {
					klog.Errorf("Error when selecting and booking a candidate for Solver %s: %s", req.NamespacedName.Name, err)
					return ctrl.Result{}, err
				}
				klog.Infof("Solver %s has selected and booked candidate %s", req.NamespacedName.Name, selectedPc.Name)
				setCandidatePhase(&solver, nodecorev1alpha1.PhaseSolved)
			} else {

				// If no PeeringCandidate is available, Create a Discovery
				klog.Infof("Solver %s has not found any candidate. Trying a Discovery", req.NamespacedName.Name)
				setCandidatePhase(&solver, nodecorev1alpha1.PhaseRunning)
			}

			setSolverPhase(&solver, nodecorev1alpha1.PhaseRunning, "Solver is running")

			// Update the Solver status
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
		case nodecorev1alpha1.PhaseRunning:
			// Check solver expiration
			if r.checkExpiration(&solver) {
				klog.Infof("Solver %s has expired", req.NamespacedName.Name)

				setSolverPhase(&solver, nodecorev1alpha1.PhaseTimeout, "Solver has expired before finding a candidate")

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

			discoveryStatusCheck(&solver, discovery)

			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		case nodecorev1alpha1.PhaseFailed:
			klog.Infof("Solver %s has not found any candidate", req.NamespacedName.Name)
			setSolverPhase(&solver, nodecorev1alpha1.PhaseFailed, "Solver has not found any candidate")
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		case nodecorev1alpha1.PhaseSolved:
			// Create contract
			break
		default:
			setCandidatePhase(&solver, nodecorev1alpha1.PhaseIdle)
			// Update the Solver status
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	if solver.Spec.EnstablishPeering && solver.Status.PurchasingPhase == nodecorev1alpha1.PhaseSolved {
		// Peering phase to be implemented
	}

	return ctrl.Result{}, nil
}

/* func (r *SolverReconciler) discoveryEnrolled(o client.Object) []reconcile.Request {
	requests := []reconcile.Request{}
	discovery := o.(*advertisementv1alpha1.Discovery)
	solver := nodecorev1alpha1.Solver{}
	if err := r.Get(context.Background(), types.NamespacedName{Name: discovery.Spec.SolverName, Namespace: discovery.Namespace}, &solver); err != nil {
		klog.Errorf("Error when getting Solver %s: %s", discovery.Spec.SolverName, err)
		return nil
	}
	requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: solver.Name, Namespace: solver.Namespace}})
	return requests
} */

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
		res := filterPeeringCandidate(&selector, &p)
		if res {
			result = append(result, p)
		}
	}

	return result, nil
}

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

func (r *SolverReconciler) createOrUpdateDiscovery(ctx context.Context, solver *nodecorev1alpha1.Solver) (*nodecorev1alpha1.GenericRef, error) {
	pc := &nodecorev1alpha1.GenericRef{}
	// Create the Discovery
	discovery := forgeDiscovery(solver.Spec.Selector, solver.Name)
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
			if checkPhaseExpiration(discovery.Status.Phase) {
				discovery.Status.Phase.Phase = nodecorev1alpha1.PhaseTimeout
				discovery.Status.Phase.LastChangeTime = getTimeNow()
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
}

func (r *SolverReconciler) createOrGetDiscovery(ctx context.Context, solver *nodecorev1alpha1.Solver) (*advertisementv1alpha1.Discovery, error) {
	discovery := &advertisementv1alpha1.Discovery{}

	// Get the Discovery
	if err := r.Get(ctx, types.NamespacedName{Name: forgeDiscoveryName(solver.Name), Namespace: discoveryNamespace}, discovery); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting Discovery for Solver %s: %s", solver.Name, err)
		return nil, err
	} else if err != nil {
		// Create the Discovery
		discovery := forgeDiscovery(solver.Spec.Selector, solver.Name)
		if err := r.Client.Create(ctx, discovery); err != nil {
			klog.Errorf("Error when creating Discovery for Solver %s: %s", solver.Name, err)
			return nil, err
		}
		/* solver.Status.DiscoveryPhase = nodecorev1alpha1.PhaseRunning
		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s: %s", solver.Name, err)
			return nil, err
		} */
	}
	return discovery, nil
}

func (r *SolverReconciler) checkExpiration(solver *nodecorev1alpha1.Solver) bool {
	phase := solver.Status.SolverPhase
	// Check if the phase has expired
	klog.Infof("Checking solver expiration")
	return checkPhaseExpiration(phase)
}

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

func contractPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// TODO:
			return false
		},
	}
}

func (r *SolverReconciler) discoveryToSolver(ctx context.Context, o client.Object) []reconcile.Request {
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      strings.TrimSuffix(o.GetName(), "-discovery"),
				Namespace: o.GetNamespace(),
			},
		},
	}
}
