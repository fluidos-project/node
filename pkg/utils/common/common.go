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

package common

import (
	"encoding/json"
	"fmt"
	"regexp"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"

	advertisementv1alpha1 "github.com/fluidos-project/node/apis/advertisement/v1alpha1"
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/namings"
	"github.com/fluidos-project/node/pkg/utils/parseutil"
)

// FilterFlavorsBySelector returns the Flavor CRs in the cluster that match the selector.
func FilterFlavorsBySelector(flavors []nodecorev1alpha1.Flavor, selector models.Selector) ([]nodecorev1alpha1.Flavor, error) {
	var flavorsSelected []nodecorev1alpha1.Flavor

	klog.Infof("Filtering flavors by selector: %v", selector)
	klog.Infof("Selector type: %s", selector.GetSelectorType())

	// Map the selector flavor type to the FlavorTypeIdentifier
	selectorType := models.MapFromFlavorTypeName(selector.GetSelectorType())

	// Get the Flavors that match the selector
	for i := range flavors {
		f := flavors[i]
		if f.Spec.FlavorType.TypeIdentifier == selectorType {
			klog.Infof("Flavor type matches selector type, which is %s", selector.GetSelectorType())
			if FilterFlavor(selector, &f) {
				flavorsSelected = append(flavorsSelected, f)
			}
		}
	}

	return flavorsSelected, nil
}

// FilterFlavor returns true if the Flavor CR fits the selector.
func FilterFlavor(selector models.Selector, flavorCR *nodecorev1alpha1.Flavor) bool {
	flavorTypeIdentifier, flavorTypeData, err := nodecorev1alpha1.ParseFlavorType(flavorCR)

	if err != nil {
		klog.Errorf("error parsing flavor type: %v", err)
		return false
	}

	switch flavorTypeIdentifier {
	case nodecorev1alpha1.TypeK8Slice:
		// Check if selector type matches flavor type
		if selector.GetSelectorType() != models.K8SliceNameDefault {
			klog.Errorf("selector type %s does not match flavor type %s", selector.GetSelectorType(), models.K8SliceNameDefault)
			return false
		}
		// Cast the selector to a K8Slice selector
		k8sliceFilters := selector.(models.K8SliceSelector)
		// Cast the flavor type data to a K8Slice CR
		flavorTypeCR := flavorTypeData.(nodecorev1alpha1.K8Slice)
		return filterFlavorK8Slice(&k8sliceFilters, &flavorTypeCR)
	case nodecorev1alpha1.TypeService:
		// Check if selector type matches flavor type
		if selector.GetSelectorType() != models.ServiceNameDefault {
			klog.Errorf("selector type %s does not match flavor type %s", selector.GetSelectorType(), models.ServiceNameDefault)
			return false
		}
		// Cast the selector to a Service selector
		serviceFilters := selector.(models.ServiceSelector)
		// Cast the flavor type data to a Service CR
		flavorTypeCR := flavorTypeData.(nodecorev1alpha1.ServiceFlavor)
		// Check if the flavor matches the Service selector
		return filterFlavorService(&serviceFilters, &flavorTypeCR)
	case nodecorev1alpha1.TypeVM:
		// TODO (VM): Implement VM filtering
		klog.Errorf("VM filtering not implemented")
		return false
	case nodecorev1alpha1.TypeSensor:
		// TODO (Sensor): Implement Sensor filtering
		klog.Errorf("Sensor filtering not implemented")
		return false
	default:
		// Flavor type not supported
		klog.Errorf("flavor type %s not supported", flavorCR.Spec.FlavorType.TypeIdentifier)
		return false
	}
}

func filterResourceQuantityFilter(selectorValue resource.Quantity, filter models.ResourceQuantityFilter) bool {
	switch filter.Name {
	case models.MatchFilter:
		// Parse the filter to a match filter
		var matchFilter models.ResourceQuantityMatchFilter
		err := json.Unmarshal(filter.Data, &matchFilter)
		if err != nil {
			klog.Errorf("Error unmarshalling match filter: %v", err)
			return false
		}
		// Check if the selector value matches the filter value
		if selectorValue.Cmp(matchFilter.Value) != 0 {
			klog.Infof("Match Filter: %d - Selector Value: %d", matchFilter.Value, selectorValue)
			return false
		}
	case models.RangeFilter:
		// Parse the filter to a range filter
		var rangeFilter models.ResourceQuantityRangeFilter
		err := json.Unmarshal(filter.Data, &rangeFilter)
		if err != nil {
			klog.Errorf("Error unmarshalling range filter: %v", err)
			return false
		}
		// Check if the selector value is within the range
		// If the rangeFilter.Min exists check if the selector value is greater or equal to it
		if rangeFilter.Min != nil {
			if selectorValue.Cmp(*rangeFilter.Min) < 0 {
				klog.Infof("Range Filter: %d-%d - Selector Value: %d", rangeFilter.Min, rangeFilter.Max, selectorValue)
				return false
			}
		}
		// If the rangeFilter.Max exists check if the selector value is less or equal to it
		if rangeFilter.Max != nil {
			if selectorValue.Cmp(*rangeFilter.Max) > 0 {
				klog.Infof("Range Filter: %d-%d - Selector Value: %d", rangeFilter.Min, rangeFilter.Max, selectorValue)
				return false
			}
		}
	}
	return true
}

func filterStringFilter(selectorValue string, filter models.StringFilter) bool {
	switch filter.Name {
	case models.MatchFilter:
		// Parse the filter to a match filter
		var matchFilter models.StringMatchFilter
		err := json.Unmarshal(filter.Data, &matchFilter)
		if err != nil {
			klog.Errorf("Error unmarshalling match filter: %v", err)
			return false
		}
		// Check if the selector value matches the filter value
		if selectorValue != matchFilter.Value {
			klog.Infof("Match Filter: %s - Selector Value: %s", matchFilter.Value, selectorValue)
			return false
		}
	case models.RangeFilter:
		// Parse the filter to a range filter
		var rangeFilter models.StringRangeFilter
		err := json.Unmarshal(filter.Data, &rangeFilter)
		if err != nil {
			klog.Errorf("Error unmarshalling range filter: %v", err)
			return false
		}
		// Check if the selector value matches the regex
		match, err := regexp.MatchString(rangeFilter.Regex, selectorValue)
		if err != nil {
			klog.Errorf("Error matching regex: %v", err)
			return false
		}
		if !match {
			klog.Infof("Range Filter: %s - Selector Value: %s", rangeFilter.Regex, selectorValue)
			return false
		}
	default:
		klog.Errorf("Filter name %s not supported", filter.Name)
		return false
	}
	return true
}

// filterFlavorK8Slice return true if the K8Slice Flavor CR fits the K8Slice selector.
func filterFlavorK8Slice(k8SliceSelector *models.K8SliceSelector, flavorTypeK8SliceCR *nodecorev1alpha1.K8Slice) bool {
	// Architecture Filter
	if k8SliceSelector.Architecture != nil {
		// Check if the flavor matches the Architecture filter
		architectureFilterModel := *k8SliceSelector.Architecture
		if !filterStringFilter(flavorTypeK8SliceCR.Characteristics.Architecture, architectureFilterModel) {
			return false
		}
	}
	// CPU Filter
	if k8SliceSelector.CPU != nil {
		// Check if the flavor matches the CPU filter
		cpuFilterModel := *k8SliceSelector.CPU
		if !filterResourceQuantityFilter(flavorTypeK8SliceCR.Characteristics.CPU, cpuFilterModel) {
			return false
		}
	}

	// Memory Filter
	if k8SliceSelector.Memory != nil {
		// Check if the flavor matches the Memory filter
		memoryFilterModel := *k8SliceSelector.Memory
		if !filterResourceQuantityFilter(flavorTypeK8SliceCR.Characteristics.Memory, memoryFilterModel) {
			return false
		}
	}

	// Pods Filter
	if k8SliceSelector.Pods != nil {
		// Check if the flavor matches the Pods filter
		podsFilterModel := *k8SliceSelector.Pods
		if !filterResourceQuantityFilter(flavorTypeK8SliceCR.Characteristics.Pods, podsFilterModel) {
			return false
		}
	}

	// Storage Filter
	if k8SliceSelector.Storage != nil {
		// Check if the flavor matches the Storage filter
		storageFilterModel := *k8SliceSelector.Storage
		if !filterResourceQuantityFilter(*flavorTypeK8SliceCR.Characteristics.Storage, storageFilterModel) {
			return false
		}
	}

	return true
}

// filterFlavorService return true if the Service Flavor CR fits the Service selector.
func filterFlavorService(serviceSelector *models.ServiceSelector, flavorTypeServiceCR *nodecorev1alpha1.ServiceFlavor) bool {
	// Category Filter
	if serviceSelector.Category != nil {
		// Check if the flavor matches the Category filter
		categoryFilterModel := *serviceSelector.Category
		if !filterStringFilter(flavorTypeServiceCR.Category, categoryFilterModel) {
			return false
		}
	}

	// Tags Filter
	if serviceSelector.Tags != nil {
		// Check if the flavor matches the Tags filter
		tagsFilterModel := *serviceSelector.Tags
		tagsCheck := false
		// For each tag in the flavor, check if any of them matches the filter
		for _, tag := range flavorTypeServiceCR.Tags {
			if filterStringFilter(tag, tagsFilterModel) {
				tagsCheck = true
				break
			}
		}
		if !tagsCheck {
			return false
		}
	}

	return true
}

/* filterFlavorVM return true if the VM Flavor CR fits the VM selector.
func filterFlavorVM(vmSelector *models.VMSelector, flavorTypeVMCR *nodecorev1alpha1.VMFlavor) bool {
	// TODO (VM): Implement VM filtering
	klog.Errorf("VM filtering not implemented")
	return false
}

// filterFlavorSensor return true if the Sensor Flavor CR fits the Sensor selector.
func filterFlavorSensor(sensorSelector *models.SensorSelector, flavorTypeSensorCR *nodecorev1alpha1.SensorFlavor) bool {
	// TODO (Sensor): Implement Sensor filtering
	klog.Errorf("Sensor filtering not implemented")
	return false
}*/

// FilterPeeringCandidate filters the peering candidate based on the solver's flavor selector.
func FilterPeeringCandidate(selector *nodecorev1alpha1.Selector, pc *advertisementv1alpha1.PeeringCandidate) bool {
	// Parsing the selector
	if selector == nil {
		klog.Infof("No selector provided")
		return true
	}
	s, err := parseutil.ParseFlavorSelector(selector)
	if err != nil {
		klog.Errorf("Error parsing selector: %v", err)
		return false
	}
	// Filter the peering candidate based on its flavor
	return FilterFlavor(s, &pc.Spec.Flavor)
}

// CheckSelector ia a func to check if the syntax of the Selector is right.
// Strict and range syntax cannot be used together.
func CheckSelector(selector models.Selector) error {
	// Parse the selector to check the syntax
	switch selector.GetSelectorType() {
	case models.K8SliceNameDefault:
		k8sliceSelector := selector.(models.K8SliceSelector)
		klog.Infof("Checking K8Slice selector: %v", k8sliceSelector)
		// Nothing is compulsory in the K8Slice selector
		return nil
	case models.VMNameDefault:
		// TODO (VM): Implement VM selector syntax check
		klog.Errorf("VM selector syntax check not implemented")
		return nil
	case models.ServiceNameDefault:
		serviceSelector := selector.(models.ServiceSelector)
		klog.Infof("Checking Service selector: %v", serviceSelector)
		// Nothing is compulsory in the Service selector
		// TODO (Service): Implement Service selector syntax check if needed
		return nil
	case models.SensorNameDefault:
		// TODO (Sensor): Implement Sensor selector syntax check
		klog.Errorf("Sensor selector syntax check not implemented")
		return nil
	default:
		return fmt.Errorf("selector type %s not supported", selector.GetSelectorType())
	}
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
	flavorName := namings.RetrieveFlavorNameFromPC(reservation.Spec.PeeringCandidate.Name)
	if reservation.Status.Phase.Phase == nodecorev1alpha1.PhaseSolved {
		klog.Infof("Reservation %s has reserved and purchase the flavor %s", reservation.Name, flavorName)
		solver.Status.ReservationPhase = nodecorev1alpha1.PhaseSolved
		solver.Status.ReserveAndBuy = nodecorev1alpha1.PhaseSolved
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Reservation: Flavor reserved and purchased")
	}
	if reservation.Status.Phase.Phase == nodecorev1alpha1.PhaseFailed {
		klog.Infof("Reservation %s has failed. Reason: %s", reservation.Name, reservation.Status.Phase.Message)
		solver.Status.ReservationPhase = nodecorev1alpha1.PhaseFailed
		solver.Status.ReserveAndBuy = nodecorev1alpha1.PhaseFailed
		solver.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation: Flavor reservation and purchase failed")
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

// AllocationStatusCheck checks the status of the allocation.
func AllocationStatusCheck(solver *nodecorev1alpha1.Solver, allocation *nodecorev1alpha1.Allocation) {
	klog.Infof("Allocation %s is in phase %s", allocation.Name, allocation.Status.Status)
	if allocation.Status.Status == nodecorev1alpha1.Active {
		klog.Infof("Allocation %s is active", allocation.Name)
		solver.Status.Peering = nodecorev1alpha1.PhaseSolved
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Allocation: active")
	}
	if allocation.Status.Status == nodecorev1alpha1.Provisioning {
		klog.Infof("Allocation %s is provisioning", allocation.Name)
		solver.Status.Peering = nodecorev1alpha1.PhaseRunning
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Allocation: provisioning")
	}
	if allocation.Status.Status == nodecorev1alpha1.ResourceCreation {
		klog.Infof("Allocation %s is creating resources", allocation.Name)
		solver.Status.Peering = nodecorev1alpha1.PhaseRunning
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Allocation: creating resources")
	}
	if allocation.Status.Status == nodecorev1alpha1.Peering {
		klog.Infof("Allocation %s is peering", allocation.Name)
		solver.Status.Peering = nodecorev1alpha1.PhaseRunning
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Allocation: peering")
	}
	if allocation.Status.Status == nodecorev1alpha1.Released {
		klog.Infof("Allocation %s is released", allocation.Name)
		solver.Status.Peering = nodecorev1alpha1.PhaseSolved
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Allocation: released")
	}
	if allocation.Status.Status == nodecorev1alpha1.Inactive {
		klog.Infof("Allocation %s is inactive", allocation.Name)
		solver.Status.Peering = nodecorev1alpha1.PhaseRunning
		solver.SetPhase(nodecorev1alpha1.PhaseRunning, "Allocation: inactive")
	}
	if allocation.Status.Status == nodecorev1alpha1.Error {
		klog.Infof("Allocation %s is in error", allocation.Name)
		solver.Status.Peering = nodecorev1alpha1.PhaseFailed
		solver.SetPhase(nodecorev1alpha1.PhaseFailed, "Allocation: error")
	}
}
