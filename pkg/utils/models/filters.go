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

package models

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/api/resource"
)

// ResourceQuantityFilter represents a filter for a resource quantity.
type ResourceQuantityFilter struct {
	Name FilterType      `scheme:"name"`
	Data json.RawMessage `scheme:"data"`
}

// ResourceQuantityFilterData represents the data of a ResourceQuantityFilter.
type ResourceQuantityFilterData interface {
	GetFilterType() FilterType
}

// StringFilter represents a filter for a string.
type StringFilter struct {
	Name FilterType      `scheme:"name"`
	Data json.RawMessage `scheme:"data"`
}

// StringFilterData represents the data of a StringFilter.
type StringFilterData interface {
	GetFilterType() FilterType
}

// FilterType represents the type of a Filter.
type FilterType string

const (
	// MatchFilter is the identifier for a match filter.
	MatchFilter FilterType = "Match"
	// RangeFilter is the identifier for a range filter.
	RangeFilter FilterType = "Range"
)

// ResourceQuantityMatchFilter represents a match filter for a resource quantity.
type ResourceQuantityMatchFilter struct {
	Value resource.Quantity `scheme:"value"`
}

// GetFilterType returns the type of the Filter.
func (fq ResourceQuantityMatchFilter) GetFilterType() FilterType {
	return MatchFilter
}

// ResourceQuantityRangeFilter represents a range filter for a resource quantity.
type ResourceQuantityRangeFilter struct {
	Min *resource.Quantity `scheme:"min,omitempty"`
	Max *resource.Quantity `scheme:"max,omitempty"`
}

// GetFilterType returns the type of the Filter.
func (fq ResourceQuantityRangeFilter) GetFilterType() FilterType {
	return RangeFilter
}

// StringMatchFilter represents a match filter for a string.
type StringMatchFilter struct {
	Value string `scheme:"value"`
}

// GetFilterType returns the type of the Filter.
func (fq StringMatchFilter) GetFilterType() FilterType {
	return MatchFilter
}

// StringRangeFilter represents a range filter for a string.
type StringRangeFilter struct {
	Regex string `scheme:"regex"`
}
