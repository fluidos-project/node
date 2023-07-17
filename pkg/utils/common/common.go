package common

import (
	"time"

	"k8s.io/klog/v2"

	advertisementv1alpha1 "fluidos.eu/node/api/advertisement/v1alpha1"
	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	reservationv1alpha1 "fluidos.eu/node/api/reservation/v1alpha1"
)

// GetTimeNow returns the current time in RFC3339 format
func GetTimeNow() string {
	return time.Now().Format(time.RFC3339)
}

// CheckExpiration checks if the timestamp has expired
func CheckExpiration(timestamp string, expTime time.Duration) bool {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		klog.Errorf("Error parsing the transaction start time: %s", err)
		return false
	}
	return time.Since(t) > expTime
}

// FilterPeeringCandidate filters the peering candidate based on the solver's flavour selector
func FilterPeeringCandidate(selector *nodecorev1alpha1.FlavourSelector, pc *advertisementv1alpha1.PeeringCandidate) bool {

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

// SetSolverPhase sets the solver phase
func SetSolverPhase(solver *nodecorev1alpha1.Solver, phase nodecorev1alpha1.Phase, msg string) {
	t := GetTimeNow()
	solver.Status.SolverPhase.Phase = phase
	solver.Status.SolverPhase.LastChangeTime = t
	solver.Status.SolverPhase.Message = msg
	solver.Status.SolverPhase.EndTime = t
}

// SetPurchasePhase sets the purchase phase
func SetPurchasingPhase(solver *nodecorev1alpha1.Solver, phase nodecorev1alpha1.Phase) {
	solver.Status.PurchasingPhase = phase
	solver.Status.SolverPhase.LastChangeTime = GetTimeNow()
}

// SetCandidatePhase sets the candidate phase
func SetCandidatePhase(solver *nodecorev1alpha1.Solver, phase nodecorev1alpha1.Phase) {
	solver.Status.CandidatesPhase = phase
	solver.Status.SolverPhase.LastChangeTime = GetTimeNow()
}

// SetDiscoveryPhaseSolver sets the discovery phase
func SetDiscoveryPhaseSolver(solver *nodecorev1alpha1.Solver, phase nodecorev1alpha1.Phase) {
	solver.Status.DiscoveryPhase = phase
	solver.Status.SolverPhase.LastChangeTime = GetTimeNow()
}

// SetDiscoveryPhase sets the phase of the discovery
func SetDiscoveryPhase(d *advertisementv1alpha1.Discovery, phase nodecorev1alpha1.Phase, msg string) {
	d.Status.Phase.Phase = phase
	d.Status.Phase.LastChangeTime = GetTimeNow()
	d.Status.Phase.Message = msg
}

// SetReservationPhase sets the phase of the discovery
func SetReservationPhase(r *reservationv1alpha1.Reservation, phase nodecorev1alpha1.Phase, msg string) {
	r.Status.Phase.Phase = phase
	r.Status.Phase.LastChangeTime = GetTimeNow()
	r.Status.Phase.Message = msg
}

// SetReserveStatus sets the status of the reserve (if it is a reserve)
func SetReserveStatus(r *reservationv1alpha1.Reservation, status nodecorev1alpha1.Phase) {
	r.Status.ReservePhase = status
}

// SetPurchaseStatus sets the status of the purchase (if it is a purchase)
func SetPurchaseStatus(r *reservationv1alpha1.Reservation, status nodecorev1alpha1.Phase) {
	r.Status.PurchasePhase = status
}

// DiscoveryStatusCheck checks the status of the discovery
func DiscoveryStatusCheck(solver *nodecorev1alpha1.Solver, discovery *advertisementv1alpha1.Discovery) {
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseSolved {
		klog.Infof("Discovery %s has found a candidate: %s", discovery.Name, discovery.Spec.PeeringCandidate)
		solver.Status.CandidatesPhase = nodecorev1alpha1.PhaseSolved
		solver.Status.PeeringCandidate = discovery.Spec.PeeringCandidate
		SetDiscoveryPhaseSolver(solver, nodecorev1alpha1.PhaseSolved)
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseFailed {
		klog.Infof("Discovery %s has failed. Reason: %s", discovery.Name, discovery.Status.Phase.Message)
		klog.Infof("Peering candidate not found, Solver %s failed", solver.Name)
		solver.Status.CandidatesPhase = nodecorev1alpha1.PhaseFailed
		SetDiscoveryPhaseSolver(solver, nodecorev1alpha1.PhaseFailed)
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseTimeout {
		klog.Infof("Discovery %s has timed out", discovery.Name)
		solver.Status.CandidatesPhase = nodecorev1alpha1.PhaseTimeout
		SetDiscoveryPhaseSolver(solver, nodecorev1alpha1.PhaseTimeout)
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseRunning {
		klog.Infof("Discovery %s is running", discovery.Name)
		SetDiscoveryPhaseSolver(solver, nodecorev1alpha1.PhaseRunning)
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseIdle {
		klog.Infof("Discovery %s is idle", discovery.Name)
		SetDiscoveryPhaseSolver(solver, nodecorev1alpha1.PhaseIdle)
	}
}
