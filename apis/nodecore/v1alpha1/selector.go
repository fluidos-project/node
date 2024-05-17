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

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

const (
	// TypeMatchFilter is the name of the filter that matches a specific value.
	TypeMatchFilter FilterType = "Match"
	// TypeRangeFilter is the name of the filter that selects resources within a range.
	TypeRangeFilter FilterType = "Range"
)

// FilterType is the type of filter that can be applied to a resource quantity.
type FilterType string

// ResourceQuantityFilter is a filter that can be applied to a resource quantity.
type ResourceQuantityFilter struct {
	// Name indicates the type of the filter
	Name FilterType `json:"name"`
	// Filter data
	Data runtime.RawExtension `json:"data"`
}

// ResourceMatchSelector is a filter that selects resources that match a specific value.
type ResourceMatchSelector struct {
	// Value is the value to match
	Value resource.Quantity `json:"value"`
}

// ResourceRangeSelector is a filter that selects resources within a range.
type ResourceRangeSelector struct {
	// Min is the minimum value of the range
	Min *resource.Quantity `json:"min,omitempty"`
	// Max is the maximum value of the range
	Max *resource.Quantity `json:"max,omitempty"`
}

// ParseResourceQuantityFilter parses a ResourceQuantityFilter into a FilterType and the corresponding filter data.
// It also provides a set of validation rules for the filter data.
// Particularly for the ResourceRangeSelector, it checks that at least one of min or max is set and that min is less than max if both are set.
func ParseResourceQuantityFilter(rqf *ResourceQuantityFilter) (FilterType, interface{}, error) {
	var validationErr error

	klog.Infof("Parsing ResourceQuantityFilter %v - Name: %s", rqf, rqf.Name)

	switch rqf.Name {
	case TypeMatchFilter:
		// Unmarshal the data into a ResourceMatchSelector
		var rms ResourceMatchSelector
		validationErr = json.Unmarshal(rqf.Data.Raw, &rms)
		return TypeMatchFilter, rms, validationErr
	case TypeRangeFilter:
		// Unmarshal the data into a ResourceRangeSelector
		var rrs ResourceRangeSelector
		validationErr = json.Unmarshal(rqf.Data.Raw, &rrs)

		klog.Infof("ResourceRangeSelector: %v", rrs)
		// Check that at least one of min or max is set
		if rrs.Min == nil && rrs.Max == nil {
			klog.Error("at least one of min or max must be set")
			validationErr = fmt.Errorf("at least one of min or max must be set")
		} else
		// If both min and max are set, check that min is less than max
		if rrs.Min != nil && rrs.Max != nil {
			// Check that the min is less than the max
			if rrs.Min != nil && rrs.Max != nil && rrs.Min.Cmp(*rrs.Max) > 0 {
				klog.Errorf("min value %s is greater than max value %s", rrs.Min.String(), rrs.Max.String())
				validationErr = fmt.Errorf("min value %s is greater than max value %s", rrs.Min.String(), rrs.Max.String())
			}
		}
		return TypeRangeFilter, rrs, validationErr
	default:
		return "", nil, fmt.Errorf("unknown filter type %s", rqf.Name)
	}
}
