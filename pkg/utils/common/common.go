// Copyright 2022-2023 FLUIDOS Project
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

package common

import (
	"fmt"

	"k8s.io/klog/v2"

	advertisementv1alpha1 "github.com/fluidos-project/node/apis/advertisement/v1alpha1"
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/namings"
	"github.com/fluidos-project/node/pkg/utils/parseutil"
)

// FilterFlavoursBySelector returns the Flavour CRs in the cluster that match the selector.
func FilterFlavoursBySelector(flavours []nodecorev1alpha1.Flavour, selector *models.Selector) ([]nodecorev1alpha1.Flavour, error) {
	var flavoursSelected []nodecorev1alpha1.Flavour

	// Get the Flavours that match the selector
	for i := range flavours {
		f := flavours[i]
		if string(f.Spec.Type) == selector.FlavourType {
			// filter function
			if FilterFlavour(selector, &f) {
				flavoursSelected = append(flavoursSelected, f)
			}
		}
	}

	return flavoursSelected, nil
}

// FilterFlavour filters the Flavour CRs in the cluster that match the selector.
func FilterFlavour(selector *models.Selector, f *nodecorev1alpha1.Flavour) bool {
	if f.Spec.Characteristics.Architecture != selector.Architecture {
		klog.Infof("Flavour %s has different architecture: %s - Selector: %s", f.Name, f.Spec.Characteristics.Architecture, selector.Architecture)
		return false
	}

	if selector.MatchSelector != nil {
		if !filterByMatchSelector(selector, f) {
			return false
		}
	}

	if selector.RangeSelector != nil && selector.MatchSelector == nil {
		if !filterByRangeSelector(selector, f) {
			return false
		}
	}

	return true
}

func filterByMatchSelector(selector *models.Selector, f *nodecorev1alpha1.Flavour) bool {
	if selector.MatchSelector.CPU.CmpInt64(0) == 0 && f.Spec.Characteristics.Cpu.Cmp(selector.MatchSelector.CPU) != 0 {
		klog.Infof("MatchSelector Cpu: %d - Flavour Cpu: %d", selector.MatchSelector.CPU, f.Spec.Characteristics.Cpu)
		return false
	}

	if selector.MatchSelector.Memory.CmpInt64(0) == 0 && f.Spec.Characteristics.Memory.Cmp(selector.MatchSelector.Memory) != 0 {
		klog.Infof("MatchSelector Memory: %d - Flavour Memory: %d", selector.MatchSelector.Memory, f.Spec.Characteristics.Memory)
		return false
	}

	if selector.MatchSelector.Pods.CmpInt64(0) == 0 && f.Spec.Characteristics.Pods.Cmp(selector.MatchSelector.Pods) != 0 {
		klog.Infof("MatchSelector Pods: %d - Flavour Pods: %d", selector.MatchSelector.Pods, f.Spec.Characteristics.Pods)
		return false
	}

	if selector.MatchSelector.EphemeralStorage.CmpInt64(0) == 0 &&
		f.Spec.Characteristics.EphemeralStorage.Cmp(selector.MatchSelector.EphemeralStorage) != 0 {
		klog.Infof("MatchSelector EphemeralStorage: %d - Flavour EphemeralStorage: %d",
			selector.MatchSelector.EphemeralStorage, f.Spec.Characteristics.EphemeralStorage)
		return false
	}

	if selector.MatchSelector.Storage.CmpInt64(0) == 0 && f.Spec.Characteristics.PersistentStorage.Cmp(selector.MatchSelector.Storage) != 0 {
		klog.Infof("MatchSelector Storage: %d - Flavour Storage: %d", selector.MatchSelector.Storage, f.Spec.Characteristics.PersistentStorage)
		return false
	}

	if selector.MatchSelector.Gpu.CmpInt64(0) == 0 && f.Spec.Characteristics.Gpu.Cmp(selector.MatchSelector.Gpu) != 0 {
		klog.Infof("MatchSelector GPU: %d - Flavour GPU: %d", selector.MatchSelector.Gpu, f.Spec.Characteristics.Gpu)
		return false
	}
	return true
}

func filterByRangeSelector(selector *models.Selector, f *nodecorev1alpha1.Flavour) bool {
	if selector.RangeSelector.MinCPU.CmpInt64(0) != 0 && f.Spec.Characteristics.Cpu.Cmp(selector.RangeSelector.MinCPU) < 0 {
		klog.Infof("RangeSelector MinCpu: %d - Flavour Cpu: %d", selector.RangeSelector.MinCPU, f.Spec.Characteristics.Cpu)
		return false
	}

	if selector.RangeSelector.MinMemory.CmpInt64(0) != 0 && f.Spec.Characteristics.Memory.Cmp(selector.RangeSelector.MinMemory) < 0 {
		klog.Infof("RangeSelector MinMemory: %d - Flavour Memory: %d", selector.RangeSelector.MinMemory, f.Spec.Characteristics.Memory)
		return false
	}

	if selector.RangeSelector.MinPods.CmpInt64(0) != 0 && f.Spec.Characteristics.Pods.Cmp(selector.RangeSelector.MinPods) < 0 {
		klog.Infof("RangeSelector MinPods: %d - Flavour Pods: %d", selector.RangeSelector.MinPods, f.Spec.Characteristics.Pods)
		return false
	}

	if selector.RangeSelector.MinEph.CmpInt64(0) != 0 && f.Spec.Characteristics.EphemeralStorage.Cmp(selector.RangeSelector.MinEph) < 0 {
		klog.Infof("RangeSelector MinEph: %d - Flavour EphemeralStorage: %d", selector.RangeSelector.MinEph, f.Spec.Characteristics.EphemeralStorage)
		return false
	}

	if selector.RangeSelector.MinStorage.CmpInt64(0) != 0 && f.Spec.Characteristics.PersistentStorage.Cmp(selector.RangeSelector.MinStorage) < 0 {
		klog.Infof("RangeSelector MinStorage: %d - Flavour Storage: %d", selector.RangeSelector.MinStorage, f.Spec.Characteristics.PersistentStorage)
		return false
	}

	if selector.RangeSelector.MinGpu.CmpInt64(0) != 0 && f.Spec.Characteristics.Gpu.Cmp(selector.RangeSelector.MinGpu) < 0 {
		return false
	}

	if selector.RangeSelector.MaxCPU.CmpInt64(0) != 0 && f.Spec.Characteristics.Cpu.Cmp(selector.RangeSelector.MaxCPU) > 0 {
		return false
	}

	if selector.RangeSelector.MaxMemory.CmpInt64(0) != 0 && f.Spec.Characteristics.Memory.Cmp(selector.RangeSelector.MaxMemory) > 0 {
		return false
	}

	if selector.RangeSelector.MaxPods.CmpInt64(0) != 0 && f.Spec.Characteristics.Pods.Cmp(selector.RangeSelector.MaxPods) > 0 {
		return false
	}

	if selector.RangeSelector.MaxEph.CmpInt64(0) != 0 && f.Spec.Characteristics.EphemeralStorage.Cmp(selector.RangeSelector.MaxEph) > 0 {
		return false
	}

	if selector.RangeSelector.MaxStorage.CmpInt64(0) != 0 && f.Spec.Characteristics.PersistentStorage.Cmp(selector.RangeSelector.MaxStorage) > 0 {
		return false
	}

	if selector.RangeSelector.MaxGpu.CmpInt64(0) != 0 && f.Spec.Characteristics.Gpu.Cmp(selector.RangeSelector.MaxGpu) > 0 {
		return false
	}
	return true
}

// FilterPeeringCandidate filters the peering candidate based on the solver's flavour selector.
func FilterPeeringCandidate(selector *nodecorev1alpha1.FlavourSelector, pc *advertisementv1alpha1.PeeringCandidate) bool {
	s := parseutil.ParseFlavourSelector(selector)
	return FilterFlavour(s, &pc.Spec.Flavour)
}

// CheckSelector ia a func to check if the syntax of the Selector is right.
// Strict and range syntax cannot be used together.
func CheckSelector(selector *models.Selector) error {
	if selector.MatchSelector != nil && selector.RangeSelector != nil {
		return fmt.Errorf("selector syntax error: strict and range syntax cannot be used together")
	}
	return nil
}

// SOLVER PHASE SETTERS

// DiscoveryStatusCheck checks the status of the discovery.
func DiscoveryStatusCheck(solver *nodecorev1alpha1.Solver, discovery *advertisementv1alpha1.Discovery) {
	if discovery.Status.Phase.Phase == nodecorev1alpha1.PhaseSolved {
		klog.Infof("Discovery %s has found candidates: %s", discovery.Name, discovery.Status.PeeringCandidateList)
		solver.Status.FindCandidate = nodecorev1alpha1.PhaseSolved
		solver.Status.DiscoveryPhase = nodecorev1alpha1.PhaseSolved
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Solver has completed the Discovery phase")
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

// ReservationStatusCheck checks the status of the reservation.
func ReservationStatusCheck(solver *nodecorev1alpha1.Solver, reservation *reservationv1alpha1.Reservation) {
	klog.Infof("Reservation %s is in phase %s", reservation.Name, reservation.Status.Phase.Phase)
	flavourName := namings.RetrieveFlavourNameFromPC(reservation.Spec.PeeringCandidate.Name)
	if reservation.Status.Phase.Phase == nodecorev1alpha1.PhaseSolved {
		klog.Infof("Reservation %s has reserved and purchase the flavour %s", reservation.Name, flavourName)
		solver.Status.ReservationPhase = nodecorev1alpha1.PhaseSolved
		solver.Status.ReserveAndBuy = nodecorev1alpha1.PhaseSolved
		solver.Status.Contract = reservation.Status.Contract
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Reservation: Flavour reserved and purchased")
	}
	if reservation.Status.Phase.Phase == nodecorev1alpha1.PhaseFailed {
		klog.Infof("Reservation %s has failed. Reason: %s", reservation.Name, reservation.Status.Phase.Message)
		solver.Status.ReservationPhase = nodecorev1alpha1.PhaseFailed
		solver.Status.ReserveAndBuy = nodecorev1alpha1.PhaseFailed
		solver.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation: Flavour reservation and purchase failed")
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
