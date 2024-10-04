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

package v1alpha1

import "k8s.io/klog/v2"

// K8SliceSelector is the selector for a K8Slice.
type K8SliceSelector struct {
	// ArchitectureFilter is the Architecture filter of the K8SliceSelector.
	ArchitectureFilter *StringFilter `json:"architectureFilter,omitempty"`

	// CPUFilter is the CPU filter of the K8SliceSelector.
	CPUFilter *ResourceQuantityFilter `json:"cpuFilter,omitempty"`

	// MemoryFilter is the Memory filter of the K8SliceSelector.
	MemoryFilter *ResourceQuantityFilter `json:"memoryFilter,omitempty"`

	// PodsFilter is the Pods filter of the K8SliceSelector.
	PodsFilter *ResourceQuantityFilter `json:"podsFilter,omitempty"`

	// StorageFilter is the Storage filter of the K8SliceSelector.
	StorageFilter *ResourceQuantityFilter `json:"storageFilter,omitempty"`
}

// GetFlavorTypeSelector returns the type of the Flavor.
func (*K8SliceSelector) GetFlavorTypeSelector() FlavorTypeIdentifier {
	return TypeK8Slice
}

// ParseK8SliceSelector parses the K8SliceSelector into a map of filters.
func ParseK8SliceSelector(k8SliceSelector *K8SliceSelector) (map[FilterType]interface{}, error) {
	filters := make(map[FilterType]interface{})
	if k8SliceSelector.ArchitectureFilter != nil {
		klog.Info("Parsing Architecture filter")
		// Parse Architecture filter
		architectureFilterType, architectureFilterData, err := ParseStringFilter(k8SliceSelector.ArchitectureFilter)
		if err != nil {
			return nil, err
		}
		filters[architectureFilterType] = architectureFilterData
	}
	if k8SliceSelector.CPUFilter != nil {
		klog.Info("Parsing CPU filter")
		// Parse CPU filter
		cpuFilterType, cpuFilterData, err := ParseResourceQuantityFilter(k8SliceSelector.CPUFilter)
		if err != nil {
			return nil, err
		}
		filters[cpuFilterType] = cpuFilterData
	}
	if k8SliceSelector.MemoryFilter != nil {
		klog.Info("Parsing Memory filter")
		// Parse Memory filter
		memoryFilterType, memoryFilterData, err := ParseResourceQuantityFilter(k8SliceSelector.MemoryFilter)
		if err != nil {
			return nil, err
		}
		filters[memoryFilterType] = memoryFilterData
	}
	if k8SliceSelector.PodsFilter != nil {
		klog.Info("Parsing Pods filter")
		// Parse Pods filter
		podsFilterType, podsFilterData, err := ParseResourceQuantityFilter(k8SliceSelector.PodsFilter)
		if err != nil {
			return nil, err
		}
		filters[podsFilterType] = podsFilterData
	}
	if k8SliceSelector.StorageFilter != nil {
		klog.Info("Parsing Storage filter")
		// Parse Storage filter
		storageFilterType, storageFilterData, err := ParseResourceQuantityFilter(k8SliceSelector.StorageFilter)
		if err != nil {
			return nil, err
		}
		filters[storageFilterType] = storageFilterData
	}

	return filters, nil
}
