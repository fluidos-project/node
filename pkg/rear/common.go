package rearmanager

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	advertisementv1alpha1 "fluidos.eu/node/api/advertisement/v1alpha1"
	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
)

const (
	solverNamespace    = "default"
	discoveryNamespace = "default"
)

func forgeDiscovery(selector nodecorev1alpha1.FlavourSelector, solverID string) *advertisementv1alpha1.Discovery {
	return &advertisementv1alpha1.Discovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forgeDiscoveryName(solverID),
			Namespace: discoveryNamespace,
		},
		Spec: advertisementv1alpha1.DiscoverySpec{
			Selector:  selector,
			SolverID:  solverID,
			Subscribe: false,
		},
	}
}

func forgeDiscoveryName(solverID string) string {
	return fmt.Sprintf("%s-discovery", solverID)
}

func checkPhaseExpiration(phase nodecorev1alpha1.PhaseStatus) bool {
	// Check if the phase has expired
	t, err := time.Parse(time.RFC3339, phase.LastChangeTime)
	if err != nil {
		klog.Errorf("Error when parsing time %s: %s", phase.LastChangeTime, err)
		return true
	}
	if t.Add(EXPIRATION_PHASE_RUNNING).Before(time.Now()) {
		return true
	}
	return false
}

func getTimeNow() string {
	return time.Now().Format(time.RFC3339)
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

func setSolverPhase(solver *nodecorev1alpha1.Solver, phase nodecorev1alpha1.Phase, msg string) {
	t := getTimeNow()
	solver.Status.SolverPhase.Phase = phase
	solver.Status.SolverPhase.LastChangeTime = t
	solver.Status.SolverPhase.Message = msg
	solver.Status.SolverPhase.EndTime = t
}

func setPurchasingPhase(solver *nodecorev1alpha1.Solver, phase nodecorev1alpha1.Phase) {

	solver.Status.PurchasingPhase = phase
	solver.Status.SolverPhase.LastChangeTime = getTimeNow()
}

func setCandidatePhase(solver *nodecorev1alpha1.Solver, phase nodecorev1alpha1.Phase) {
	solver.Status.CandidatesPhase = phase
	solver.Status.SolverPhase.LastChangeTime = getTimeNow()
}

func setDiscoveryPhase(solver *nodecorev1alpha1.Solver, phase nodecorev1alpha1.Phase) {
	solver.Status.DiscoveryPhase = phase
	solver.Status.SolverPhase.LastChangeTime = getTimeNow()
}

func discoveryStatusCheck(solver *nodecorev1alpha1.Solver, discovery *advertisementv1alpha1.Discovery) {
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseSolved {
		klog.Infof("Discovery %s has found a candidate: %s", discovery.Name, discovery.Spec.PeeringCandidate)
		solver.Status.CandidatesPhase = nodecorev1alpha1.PhaseSolved
		solver.Status.PeeringCandidate = discovery.Spec.PeeringCandidate
		setDiscoveryPhase(solver, nodecorev1alpha1.PhaseSolved)
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseFailed {
		klog.Infof("Discovery %s has failed. Reason: %s", discovery.Name, discovery.Status.Phase.Message)
		klog.Infof("Peering candidate not found, Solver %s failed", solver.Name)
		solver.Status.CandidatesPhase = nodecorev1alpha1.PhaseFailed
		setDiscoveryPhase(solver, nodecorev1alpha1.PhaseFailed)
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseTimeout {
		klog.Infof("Discovery %s has timed out", discovery.Name)
		solver.Status.CandidatesPhase = nodecorev1alpha1.PhaseTimeout
		setDiscoveryPhase(solver, nodecorev1alpha1.PhaseTimeout)
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseRunning {
		klog.Infof("Discovery %s is running", discovery.Name)
		setDiscoveryPhase(solver, nodecorev1alpha1.PhaseRunning)
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseIdle {
		klog.Infof("Discovery %s is idle", discovery.Name)
		setDiscoveryPhase(solver, nodecorev1alpha1.PhaseIdle)
	}
}
