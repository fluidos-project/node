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

	klog.Infof("Reconciling Solver %s", req.NamespacedName)

	if solver.Spec.FindCandidate {
		candidatePhase := solver.Status.CandidatesPhase
		if candidatePhase.Phase == nodecorev1alpha1.PhaseIdle {
			//candidatePhase.Phase = nodecorev1alpha1.PhaseRunning
			candidatePhase.LastChangeTime = time.Now().String()

			// Get the Flavour Selector from the Solver
			//selector := solver.Spec.Selector
			// HERE CREATE A CANDIDATE OBJECT TO START THE SEARCH OF A MATCHING FLAVOR

			pc, err := r.searchPeeringCandidates(ctx, &solver)
			if client.IgnoreNotFound(err) != nil {
				klog.Errorf("Error when searching and booking a candidate for Solver %s: %s", req.NamespacedName.Name, err)
				return ctrl.Result{}, err
			}

			if len(pc) > 0 {
				selectedPc, err := r.selectAndBookPeeringCandidate(ctx, &solver, pc)
				if err != nil {
					klog.Errorf("Error when selecting and booking a candidate for Solver %s: %s", req.NamespacedName.Name, err)
					return ctrl.Result{}, err
				}
				klog.Infof("Solver %s has selected and booked candidate %s", req.NamespacedName.Name, selectedPc.Name)
				candidatePhase.Phase = nodecorev1alpha1.PhaseSolved
				candidatePhase.LastChangeTime = time.Now().String()
			} else {
				klog.Infof("Solver %s has not found any candidate.Trying a Discovery", req.NamespacedName.Name)
				candidatePhase.Phase = nodecorev1alpha1.PhaseRunning
				candidatePhase.LastChangeTime = time.Now().String()
			}

			// Update the Solver status
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
		} else if candidatePhase.Phase == nodecorev1alpha1.PhaseRunning {
			t, err := time.Parse(time.RFC3339, candidatePhase.LastChangeTime)
			if err != nil {
				klog.Errorf("Error when parsing time %s: %s", candidatePhase.LastChangeTime, err)
				return ctrl.Result{}, err
			}
			if t.Add(EXPIRATION_PHASE_RUNNING).After(time.Now()) {
				candidatePhase.Phase = nodecorev1alpha1.PhaseTimeout

				// Update the Solver status
				if err := r.updateSolverStatus(ctx, &solver); err != nil {
					klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
			}
			// Check if Disco
		} else if candidatePhase.Phase == nodecorev1alpha1.PhaseTimeout {

		} else {
			time := time.Now().String()
			candidatePhase.Phase = nodecorev1alpha1.PhaseIdle
			candidatePhase.StartTime = time
			candidatePhase.LastChangeTime = time

			// Update the Solver status
			if err := r.updateSolverStatus(ctx, &solver); err != nil {
				klog.Errorf("Error when updating Solver %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
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

func (r *SolverReconciler) createOrUpdateDiscovery(ctx context.Context, solver *nodecorev1alpha1.Solver) error {
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
