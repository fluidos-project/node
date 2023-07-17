package common

import (
	"fmt"
	"time"

	"k8s.io/klog/v2"

	advertisementv1alpha1 "fluidos.eu/node/api/advertisement/v1alpha1"
	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	reservationv1alpha1 "fluidos.eu/node/api/reservation/v1alpha1"
	"fluidos.eu/node/pkg/utils/models"
	"fluidos.eu/node/pkg/utils/namings"
	"fluidos.eu/node/pkg/utils/parseutil"
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

// FilterFlavoursBySelector returns the Flavour CRs in the cluster that match the selector
func FilterFlavoursBySelector(flavours []nodecorev1alpha1.Flavour, selector models.Selector) ([]nodecorev1alpha1.Flavour, error) {
	var flavoursSelected []nodecorev1alpha1.Flavour

	// Get the Flavours that match the selector
	for _, f := range flavours {
		if string(f.Spec.Type) == selector.FlavourType {
			// filter function
			if FilterFlavour(selector, f) {
				flavoursSelected = append(flavoursSelected, f)
			}

		}
	}

	return flavoursSelected, nil
}

// filterFlavour filters the Flavour CRs in the cluster that match the selector
func FilterFlavour(selector models.Selector, f nodecorev1alpha1.Flavour) bool {

	if f.Spec.Characteristics.Architecture != selector.Architecture {
		return false
	}

	if selector.MatchSelector != nil {
		if selector.MatchSelector.Cpu == 0 && f.Spec.Characteristics.Cpu.CmpInt64(int64(selector.MatchSelector.Cpu)) != 0 {
			return false
		}

		if selector.MatchSelector.Memory == 0 && f.Spec.Characteristics.Memory.CmpInt64(int64(selector.MatchSelector.Memory)) != 0 {
			return false
		}

		if selector.MatchSelector.EphemeralStorage == 0 && f.Spec.Characteristics.EphemeralStorage.CmpInt64(int64(selector.MatchSelector.EphemeralStorage)) < 0 {
			return false
		}

		if selector.MatchSelector.Storage == 0 && f.Spec.Characteristics.PersistentStorage.CmpInt64(int64(selector.MatchSelector.Storage)) < 0 {
			return false
		}

		if selector.MatchSelector.Gpu == 0 && f.Spec.Characteristics.Gpu.CmpInt64(int64(selector.MatchSelector.Gpu)) < 0 {
			return false
		}
	}

	if selector.RangeSelector != nil && selector.MatchSelector == nil {

		if selector.RangeSelector.MoreThanCPU != 0 && f.Spec.Characteristics.Cpu.CmpInt64(int64(selector.RangeSelector.MoreThanCPU)) < 0 {
			return false
		}

		if selector.RangeSelector.MoreThanMemory != 0 && f.Spec.Characteristics.Memory.CmpInt64(int64(selector.RangeSelector.MoreThanMemory)) < 0 {
			return false
		}

		if selector.RangeSelector.MoreThanEph != 0 && f.Spec.Characteristics.EphemeralStorage.CmpInt64(int64(selector.RangeSelector.MoreThanEph)) < 0 {
			return false
		}

		if selector.RangeSelector.MoreThanStorage != 0 && f.Spec.Characteristics.PersistentStorage.CmpInt64(int64(selector.RangeSelector.MoreThanStorage)) < 0 {
			return false
		}

		if selector.RangeSelector.MoreThanGpu != 0 && f.Spec.Characteristics.Gpu.CmpInt64(int64(selector.RangeSelector.MoreThanGpu)) < 0 {
			return false
		}

		if selector.RangeSelector.LessThanCPU != 0 && f.Spec.Characteristics.Cpu.CmpInt64(int64(selector.RangeSelector.LessThanCPU)) > 0 {
			return false
		}

		if selector.RangeSelector.LessThanMemory != 0 && f.Spec.Characteristics.Memory.CmpInt64(int64(selector.RangeSelector.LessThanMemory)) > 0 {
			return false
		}

		if selector.RangeSelector.LessThanEph != 0 && f.Spec.Characteristics.EphemeralStorage.CmpInt64(int64(selector.RangeSelector.LessThanEph)) > 0 {
			return false
		}

		if selector.RangeSelector.LessThanStorage != 0 && f.Spec.Characteristics.PersistentStorage.CmpInt64(int64(selector.RangeSelector.LessThanStorage)) > 0 {
			return false
		}

		if selector.RangeSelector.LessThanGpu != 0 && f.Spec.Characteristics.Gpu.CmpInt64(int64(selector.RangeSelector.LessThanGpu)) > 0 {
			return false
		}
	}

	return true
}

// FilterPeeringCandidate filters the peering candidate based on the solver's flavour selector
func FilterPeeringCandidate(selector *nodecorev1alpha1.FlavourSelector, pc *advertisementv1alpha1.PeeringCandidate) bool {
	s := parseutil.ParseFlavourSelector(*selector)
	return FilterFlavour(s, pc.Spec.Flavour)
}

// CheckSelector ia a func to check if the syntax of the Selector is right.
// Strict and range syntax cannot be used together
func CheckSelector(selector models.Selector) error {

	if selector.MatchSelector != nil && selector.RangeSelector != nil {
		return fmt.Errorf("selector syntax error: strict and range syntax cannot be used together")
	}
	return nil
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

// SOLVER PHASE SETTERS

// SetSolverPhase sets the solver phase
func SetSolverPhase(solver *nodecorev1alpha1.Solver, phase nodecorev1alpha1.Phase, msg string) {
	t := GetTimeNow()
	solver.Status.SolverPhase.Phase = phase
	solver.Status.SolverPhase.LastChangeTime = t
	solver.Status.SolverPhase.Message = msg
	solver.Status.SolverPhase.EndTime = t
}

// SetPurchasePhase sets the purchase phase of the solver
func SetPurchasingPhaseSolver(solver *nodecorev1alpha1.Solver, phase nodecorev1alpha1.Phase) {
	solver.Status.ReserveAndBuy = phase
	solver.Status.SolverPhase.LastChangeTime = GetTimeNow()
}

// SetCandidatePhaseSolver sets the candidate phase of the solver
func SetCandidatePhaseSolver(solver *nodecorev1alpha1.Solver, phase nodecorev1alpha1.Phase) {
	solver.Status.FindCandidate = phase
	solver.Status.SolverPhase.LastChangeTime = GetTimeNow()
}

// SetDiscoveryPhaseSolver sets the discovery phase of the solver
func SetDiscoveryPhaseSolver(solver *nodecorev1alpha1.Solver, phase nodecorev1alpha1.Phase) {
	solver.Status.DiscoveryPhase = phase
	solver.Status.SolverPhase.LastChangeTime = GetTimeNow()
}

// SetReservationPhaseSolver sets the reservation phase of the solver
func SetReservationPhaseSolver(solver *nodecorev1alpha1.Solver, phase nodecorev1alpha1.Phase) {
	solver.Status.ReservationPhase = phase
	solver.Status.SolverPhase.LastChangeTime = GetTimeNow()
}

// DiscoveryStatusCheck checks the status of the discovery
func DiscoveryStatusCheck(solver *nodecorev1alpha1.Solver, discovery *advertisementv1alpha1.Discovery) {
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseSolved {
		klog.Infof("Discovery %s has found a candidate: %s", discovery.Name, discovery.Spec.PeeringCandidate)
		solver.Status.FindCandidate = nodecorev1alpha1.PhaseSolved
		solver.Status.PeeringCandidate = discovery.Spec.PeeringCandidate
		SetDiscoveryPhaseSolver(solver, nodecorev1alpha1.PhaseSolved)
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseFailed {
		klog.Infof("Discovery %s has failed. Reason: %s", discovery.Name, discovery.Status.Phase.Message)
		klog.Infof("Peering candidate not found, Solver %s failed", solver.Name)
		solver.Status.FindCandidate = nodecorev1alpha1.PhaseFailed
		SetDiscoveryPhaseSolver(solver, nodecorev1alpha1.PhaseFailed)
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseTimeout {
		klog.Infof("Discovery %s has timed out", discovery.Name)
		solver.Status.FindCandidate = nodecorev1alpha1.PhaseTimeout
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

func ReservationStatusCheck(solver *nodecorev1alpha1.Solver, reservation *reservationv1alpha1.Reservation) {
	flavourName := namings.RetrieveFlavourNameFromPC(reservation.Spec.PeeringCandidate.Name)
	if reservation.Status.Phase.Phase == nodecorev1alpha1.PhaseSolved {
		klog.Infof("Reservation %s has reserved and purchase the flavour %s", reservation.Name, flavourName)
		solver.Status.ReservationPhase = nodecorev1alpha1.PhaseSolved
		solver.Status.ReserveAndBuy = nodecorev1alpha1.PhaseSolved
	}
	if reservation.Status.Phase.Phase == nodecorev1alpha1.PhaseFailed {
		klog.Infof("Reservation %s has failed. Reason: %s", reservation.Name, reservation.Status.Phase.Message)
		solver.Status.ReservationPhase = nodecorev1alpha1.PhaseFailed
		solver.Status.ReserveAndBuy = nodecorev1alpha1.PhaseFailed
	}
	if reservation.Status.Phase.Phase == nodecorev1alpha1.PhaseRunning {
		if reservation.Status.ReservePhase == nodecorev1alpha1.PhaseRunning {
			klog.Infof("Reservation %s is running", reservation.Name)
			solver.Status.SolverPhase.Phase = nodecorev1alpha1.PhaseRunning
			solver.Status.SolverPhase.LastChangeTime = GetTimeNow()
			solver.Status.SolverPhase.Message = "Reservation is running"
		}
		if reservation.Status.PurchasePhase == nodecorev1alpha1.PhaseRunning {
			klog.Infof("Purchasing %s is running", reservation.Name)
			solver.Status.SolverPhase.Phase = nodecorev1alpha1.PhaseRunning
			solver.Status.SolverPhase.LastChangeTime = GetTimeNow()
			solver.Status.SolverPhase.Message = "Purchasing is running"
		}
	}
	if reservation.Status.Phase.Phase == nodecorev1alpha1.PhaseIdle {
		klog.Infof("Reservation %s is idle", reservation.Name)
		SetReservationPhaseSolver(solver, nodecorev1alpha1.PhaseIdle)
	}
}
