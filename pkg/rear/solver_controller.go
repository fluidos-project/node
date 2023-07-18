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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	//"sigs.k8s.io/controller-runtime/pkg/log"

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

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Solver object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
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
		solver.Status.SolverPhase.Phase != nodecorev1alpha1.PhaseReady &&
		solver.Status.SolverPhase.Phase != nodecorev1alpha1.PhaseRunning &&
		solver.Status.SolverPhase.Phase != nodecorev1alpha1.PhaseIdle {
		t := time.Now().String()
		solver.Status.SolverPhase.Phase = nodecorev1alpha1.PhaseIdle
		/* solver.Status.CandidatesPhase = nodecorev1alpha1.PhaseIdle
		solver.Status.DiscoveryPhase = nodecorev1alpha1.PhaseIdle
		solver.Status.ReservationPhase = nodecorev1alpha1.PhaseIdle
		solver.Status.PurchasingPhase = nodecorev1alpha1.PhaseIdle
		solver.Status.ConsumePhase = nodecorev1alpha1.PhaseIdle */
		solver.Status.SolverPhase.StartTime = t
		solver.Status.SolverPhase.LastChangeTime = t
		if err := r.updateSolverStatus(ctx, &solver); err != nil {
			klog.Errorf("Error when updating Solver %s status before reconcile: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	klog.Infof("Reconciling Solver %s", req.NamespacedName)

	if solver.Spec.FindCandidate {
		candidatePhase := solver.Status.CandidatesPhase
		switch candidatePhase {
		case nodecorev1alpha1.PhaseIdle:

			// Search a matching PeeringCandidate if available
			pc, err := r.searchPeeringCandidates(ctx, &solver)
			if client.IgnoreNotFound(err) != nil {
				klog.Errorf("Error when searching and booking a candidate for Solver %s: %s", req.NamespacedName.Name, err)
				return ctrl.Result{}, err
			}

			// If some PeeringCandidates are available, select one and book it
			if len(pc) > 0 {
				selectedPc, err := r.selectAndBookPeeringCandidate(ctx, &solver, pc)
				if err != nil {
					klog.Errorf("Error when selecting and booking a candidate for Solver %s: %s", req.NamespacedName.Name, err)
					return ctrl.Result{}, err
				}
				klog.Infof("Solver %s has selected and booked candidate %s", req.NamespacedName.Name, selectedPc.Name)
				solver.Status.CandidatesPhase = nodecorev1alpha1.PhaseSolved
				solver.Status.SolverPhase.LastChangeTime = time.Now().String()
			} else {
				klog.Infof("Solver %s has not found any candidate.Trying a Discovery", req.NamespacedName.Name)
				solver.Status.CandidatesPhase = nodecorev1alpha1.PhaseRunning
				solver.Status.SolverPhase.LastChangeTime = time.Now().String()
			}
			solver.Status.SolverPhase.Phase = nodecorev1alpha1.PhaseRunning
			// Update the Solver status
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}

		case nodecorev1alpha1.PhaseRunning:
			// Check solver expiration
			if r.checkExpiration(&solver) {
				solver.Status.SolverPhase.Phase = nodecorev1alpha1.PhaseTimeout
				solver.Status.SolverPhase.LastChangeTime = time.Now().String()
				if err := r.updateSolverStatus(ctx, &solver); err != nil {
					klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, nil
			}

			discovery, err := r.createOrGetDiscovery(ctx, &solver)
			if err != nil {
				klog.Errorf("Error when creating or getting Discovery for Solver %s: %s", req.NamespacedName.Name, err)
				return ctrl.Result{}, err
			}
			if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseSolved {
				klog.Infof("Solver %s has found a candidate", req.NamespacedName.Name)
				solver.Status.CandidatesPhase = nodecorev1alpha1.PhaseSolved
				solver.Status.DiscoveryPhase = nodecorev1alpha1.PhaseSolved
				solver.Status.SolverPhase.LastChangeTime = time.Now().String()
				solver.Status.PeeringCandidate = discovery.Spec.PeeringCandidate
				if err := r.updateSolverStatus(ctx, &solver); err != nil {
					klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
			}
			return ctrl.Result{}, nil
		case nodecorev1alpha1.PhaseFailed:
			t := time.Now().String()
			solver.Status.SolverPhase.Phase = nodecorev1alpha1.PhaseFailed
			solver.Status.SolverPhase.LastChangeTime = t
			solver.Status.SolverPhase.Message = "Solver failed to find a candidate"
			solver.Status.SolverPhase.EndTime = t
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		case nodecorev1alpha1.PhaseSolved:
			break
		default:
			t := time.Now().String()
			solver.Status.CandidatesPhase = nodecorev1alpha1.PhaseIdle
			solver.Status.SolverPhase.LastChangeTime = t

			// Update the Solver status
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	if solver.Spec.EnstablishPeering {

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SolverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nodecorev1alpha1.Solver{}).
		Complete(r)
}

func (r *SolverReconciler) updateSolverStatus(ctx context.Context, solver *nodecorev1alpha1.Solver) error {
	return r.Status().Update(ctx, solver)
}

func (r *SolverReconciler) createOrUpdatePeering(ctx context.Context, solver *nodecorev1alpha1.Solver) error {
	return nil
}

func (r *SolverReconciler) createOrUpdateContract(ctx context.Context, solver *nodecorev1alpha1.Solver) error {
	return nil
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
		res := filterPeeringCandidate(&selector, &p)
		if res {
			result = append(result, p)
		}
	}

	return result, nil
}

func (r *SolverReconciler) selectAndBookPeeringCandidate(ctx context.Context, solver *nodecorev1alpha1.Solver, pc []advertisementv1alpha1.PeeringCandidate) (*advertisementv1alpha1.PeeringCandidate, error) {
	// Select the first PeeringCandidate
	selected := pc[0]

	// Book the PeeringCandidate
	selected.Spec.Reserved = true
	selected.Spec.SolverID = solver.Name

	// Update the PeeringCandidate
	if err := r.Update(ctx, &selected); err != nil {
		klog.Errorf("Error when updating PeeringCandidate %s: %s", selected.Name, err)
		return nil, err
	}

	return &selected, nil
}

func filterPeeringCandidate(selector *nodecorev1alpha1.FlavourSelector, pc *advertisementv1alpha1.PeeringCandidate) bool {
	if selector.Cpu.Cmp(pc.Spec.Flavour.Spec.Characteristics.Cpu) > 0 {
		return false
	}
	if selector.Memory.Cmp(pc.Spec.Flavour.Spec.Characteristics.Memory) > 0 {
		return false
	}
	if !selector.EphemeralStorage.IsZero() && selector.EphemeralStorage.Cmp(pc.Spec.Flavour.Spec.Characteristics.EphemeralStorage) > 0 {
		return false
	}
	if !selector.Gpu.IsZero() && selector.Gpu.Cmp(pc.Spec.Flavour.Spec.Characteristics.Gpu) > 0 {
		return false
	}
	if !selector.Storage.IsZero() && selector.Storage.Cmp(pc.Spec.Flavour.Spec.Characteristics.PersistentStorage) > 0 {
		return false
	}
	return true
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
				discovery.Status.Phase.LastChangeTime = time.Now().Format(time.RFC3339)
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
		solver.Status.DiscoveryPhase = nodecorev1alpha1.PhaseRunning
		if err := r.updateSolverStatus(ctx, solver); err != nil {
			klog.Errorf("Error when updating Solver %s: %s", solver.Name, err)
			return nil, err
		}
	}
	return discovery, nil
}

/* func (r *SolverReconciler) checkDiscoveryStatus(discovery *advertisementv1alpha1.Discovery, solver *nodecorev1alpha1.Solver) bool {
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseSolved {
		solver.Status.DiscoveryPhase = discovery.Status.Phase
		return true
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseTimeout || discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseIdle {

		return true
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseRunning {
		// Check if the phase has expired
		if checkPhaseExpiration(discovery.Status.Phase) {
			solver.Status.DiscoveryPhase.Phase = nodecorev1alpha1.PhaseTimeout
			solver.Status.DiscoveryPhase.LastChangeTime = time.Now().Format(time.RFC3339)
			return true
		}
		return false
	}
	return false
} */

func checkPhaseExpiration(phase nodecorev1alpha1.PhaseStatus) bool {
	// Check if the phase has expired
	t, err := time.Parse(time.RFC3339, phase.LastChangeTime)
	if err != nil {
		klog.Errorf("Error when parsing time %s: %s", phase.LastChangeTime, err)
		return true
	}
	if t.Add(EXPIRATION_PHASE_RUNNING).After(time.Now()) {
		return true
	}
	return false
}

func (r *SolverReconciler) checkExpiration(solver *nodecorev1alpha1.Solver) bool {
	phase := solver.Status.SolverPhase
	// Check if the phase has expired
	t, err := time.Parse(time.RFC3339, phase.LastChangeTime)
	if err != nil {
		klog.Errorf("Error when parsing time %s: %s", phase.LastChangeTime, err)
		return true
	}
	if t.Add(EXPIRATION_SOLVER).After(time.Now()) {
		return true
	}
	return false
}
