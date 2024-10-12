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

// ServiceSelector is the selector for a Service.
type ServiceSelector struct {
	// CategoryFilter is the Category filter of the ServiceSelector.
	CategoryFilter *StringFilter `json:"categoryFilter,omitempty"`
	// TagsFilter is the Tags filter of the ServiceSelector.
	TagsFilter *StringFilter `json:"tagsFilter,omitempty"`
	// TODO(Service): Add more filters here
}

// GetFlavorTypeSelector returns the type of the Flavor.
func (*ServiceSelector) GetFlavorTypeSelector() FlavorTypeIdentifier {
	return TypeService
}

// ParseServiceSelector parses the K8SliceSelector into a map of filters.
func ParseServiceSelector(serviceSelector *ServiceSelector) (map[FilterType]interface{}, error) {
	filters := make(map[FilterType]interface{})
	if serviceSelector.CategoryFilter != nil {
		klog.Info("Parsing Category filter")
		// Parse Category filter
		categoryFilterType, categoryFilterData, err := ParseStringFilter(serviceSelector.CategoryFilter)
		if err != nil {
			return nil, err
		}
		filters[categoryFilterType] = categoryFilterData
	}

	if serviceSelector.TagsFilter != nil {
		klog.Info("Parsing Tags filter")
		// Parse Tags filter
		tagsFilterType, tagsFilterData, err := ParseStringFilter(serviceSelector.TagsFilter)
		if err != nil {
			return nil, err
		}
		filters[tagsFilterType] = tagsFilterData
	}

	// TODO(Service): Add more filters parsing functions here

	return filters, nil
}
