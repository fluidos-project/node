package common

import (
	"fmt"

	"k8s.io/klog/v2"

	advertisementv1alpha1 "fluidos.eu/node/api/advertisement/v1alpha1"
	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	reservationv1alpha1 "fluidos.eu/node/api/reservation/v1alpha1"
	"fluidos.eu/node/pkg/utils/models"
	"fluidos.eu/node/pkg/utils/namings"
	"fluidos.eu/node/pkg/utils/parseutil"
)

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
		klog.Infof("Flavour %s has different architecture: %s - Selector: %s", f.Name, f.Spec.Characteristics.Architecture, selector.Architecture)
		return false
	}

	if selector.MatchSelector != nil {
		if selector.MatchSelector.Cpu == 0 && f.Spec.Characteristics.Cpu.CmpInt64(int64(selector.MatchSelector.Cpu)) != 0 {
			klog.Infof("MatchSelector Cpu: %d - Flavour Cpu: %d", selector.MatchSelector.Cpu, f.Spec.Characteristics.Cpu)
			return false
		}

		if selector.MatchSelector.Memory == 0 && f.Spec.Characteristics.Memory.CmpInt64(int64(selector.MatchSelector.Memory)) != 0 {
			klog.Infof("MatchSelector Memory: %d - Flavour Memory: %d", selector.MatchSelector.Memory, f.Spec.Characteristics.Memory)
			return false
		}

		if selector.MatchSelector.EphemeralStorage == 0 && f.Spec.Characteristics.EphemeralStorage.CmpInt64(int64(selector.MatchSelector.EphemeralStorage)) < 0 {
			klog.Infof("MatchSelector EphemeralStorage: %d - Flavour EphemeralStorage: %d", selector.MatchSelector.EphemeralStorage, f.Spec.Characteristics.EphemeralStorage)
			return false
		}

		if selector.MatchSelector.Storage == 0 && f.Spec.Characteristics.PersistentStorage.CmpInt64(int64(selector.MatchSelector.Storage)) < 0 {
			klog.Infof("MatchSelector Storage: %d - Flavour Storage: %d", selector.MatchSelector.Storage, f.Spec.Characteristics.PersistentStorage)
			return false
		}

		if selector.MatchSelector.Gpu == 0 && f.Spec.Characteristics.Gpu.CmpInt64(int64(selector.MatchSelector.Gpu)) < 0 {
			klog.Infof("MatchSelector GPU: %d - Flavour GPU: %d", selector.MatchSelector.Gpu, f.Spec.Characteristics.Gpu)
			return false
		}
	}

	if selector.RangeSelector != nil && selector.MatchSelector == nil {

		if selector.RangeSelector.MinCpu != 0 && f.Spec.Characteristics.Cpu.CmpInt64(int64(selector.RangeSelector.MinCpu)) < 0 {
			klog.Infof("RangeSelector MinCpu: %d - Flavour Cpu: %d", selector.RangeSelector.MinCpu, f.Spec.Characteristics.Cpu)
			return false
		}

		if selector.RangeSelector.MinMemory != 0 && f.Spec.Characteristics.Memory.CmpInt64(int64(selector.RangeSelector.MinMemory)) < 0 {
			klog.Infof("RangeSelector MinMemory: %d - Flavour Memory: %d", selector.RangeSelector.MinMemory, f.Spec.Characteristics.Memory)
			return false
		}

		if selector.RangeSelector.MinEph != 0 && f.Spec.Characteristics.EphemeralStorage.CmpInt64(int64(selector.RangeSelector.MinEph)) < 0 {
			klog.Infof("RangeSelector MinEph: %d - Flavour EphemeralStorage: %d", selector.RangeSelector.MinEph, f.Spec.Characteristics.EphemeralStorage)
			return false
		}

		if selector.RangeSelector.MinStorage != 0 && f.Spec.Characteristics.PersistentStorage.CmpInt64(int64(selector.RangeSelector.MinStorage)) < 0 {
			klog.Infof("RangeSelector MinStorage: %d - Flavour Storage: %d", selector.RangeSelector.MinStorage, f.Spec.Characteristics.PersistentStorage)
			return false
		}

		if selector.RangeSelector.MinGpu != 0 && f.Spec.Characteristics.Gpu.CmpInt64(int64(selector.RangeSelector.MinGpu)) < 0 {
			return false
		}

		if selector.RangeSelector.MaxCpu != 0 && f.Spec.Characteristics.Cpu.CmpInt64(int64(selector.RangeSelector.MaxCpu)) > 0 {
			return false
		}

		if selector.RangeSelector.MaxMemory != 0 && f.Spec.Characteristics.Memory.CmpInt64(int64(selector.RangeSelector.MaxMemory)) > 0 {
			return false
		}

		if selector.RangeSelector.MaxEph != 0 && f.Spec.Characteristics.EphemeralStorage.CmpInt64(int64(selector.RangeSelector.MaxEph)) > 0 {
			return false
		}

		if selector.RangeSelector.MaxStorage != 0 && f.Spec.Characteristics.PersistentStorage.CmpInt64(int64(selector.RangeSelector.MaxStorage)) > 0 {
			return false
		}

		if selector.RangeSelector.MaxGpu != 0 && f.Spec.Characteristics.Gpu.CmpInt64(int64(selector.RangeSelector.MaxGpu)) > 0 {
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

// SOLVER PHASE SETTERS

// DiscoveryStatusCheck checks the status of the discovery
func DiscoveryStatusCheck(solver *nodecorev1alpha1.Solver, discovery *advertisementv1alpha1.Discovery) {
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseSolved {
		klog.Infof("Discovery %s has found a candidate: %s", discovery.Name, discovery.Status.PeeringCandidate)
		solver.Status.FindCandidate = nodecorev1alpha1.PhaseSolved
		solver.Status.PeeringCandidate = discovery.Status.PeeringCandidate
		solver.Status.DiscoveryPhase = nodecorev1alpha1.PhaseSolved
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Solver has found a candidate")
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseFailed {
		klog.Infof("Discovery %s has failed. Reason: %s", discovery.Name, discovery.Status.Phase.Message)
		klog.Infof("Peering candidate not found, Solver %s failed", solver.Name)
		solver.Status.FindCandidate = nodecorev1alpha1.PhaseFailed
		solver.Status.DiscoveryPhase = nodecorev1alpha1.PhaseFailed
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseTimeout {
		klog.Infof("Discovery %s has timed out", discovery.Name)
		solver.Status.FindCandidate = nodecorev1alpha1.PhaseTimeout
		solver.Status.DiscoveryPhase = nodecorev1alpha1.PhaseTimeout
		solver.SetPhase(nodecorev1alpha1.PhaseTimeout, "Discovery has expired before finding a candidate")
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseRunning {
		klog.Infof("Discovery %s is running", discovery.Name)
		solver.SetDiscoveryStatus(nodecorev1alpha1.PhaseRunning)
	}
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseIdle {
		klog.Infof("Discovery %s is idle", discovery.Name)
		solver.SetDiscoveryStatus(nodecorev1alpha1.PhaseIdle)
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
			solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Reservation: Reserve is running")
		}
		if reservation.Status.PurchasePhase == nodecorev1alpha1.PhaseRunning {
			klog.Infof("Purchasing %s is running", reservation.Name)
			solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Reservation: Purchase is running")
		}
	}
	if reservation.Status.Phase.Phase == nodecorev1alpha1.PhaseIdle {
		klog.Infof("Reservation %s is idle", reservation.Name)
		solver.SetReservationStatus(nodecorev1alpha1.PhaseIdle)
	}
}