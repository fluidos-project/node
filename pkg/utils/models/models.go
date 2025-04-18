// Copyright 2022-2025 FLUIDOS Project
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
	"time"

	"k8s.io/klog/v2"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/consts"
)

// Flavor represents a Flavor object with its characteristics and policies.
type Flavor struct {
	FlavorID            string       `json:"flavorID"`
	ProviderID          string       `json:"providerID"`
	Type                FlavorType   `json:"type"`
	NetworkPropertyType string       `json:"networkPropertyType,omitempty"`
	Timestamp           time.Time    `json:"timestamp"`
	Location            *Location    `json:"location,omitempty"`
	Price               Price        `json:"price"`
	Owner               NodeIdentity `json:"owner"`
	Availability        bool         `json:"availability"`
}

const (
	// RangeMinKey is the key for the identifier of the minimum value in a range.
	RangeMinKey = "min"
	// RangeMaxKey is the key for the identifier of the maximum value in a range.
	RangeMaxKey = "max"
)

// FlavorType represents the type of a Flavor.
type FlavorType struct {
	Name FlavorTypeName  `json:"name"`
	Data json.RawMessage `json:"data"`
}

// FlavorTypeData represents the data of a FlavorType.
type FlavorTypeData interface {
	GetFlavorTypeName() FlavorTypeName
}

// FlavorTypeName represents the name of a FlavorType.
type FlavorTypeName string

// CarbonFootprint represents the carbon footprint of a Flavor, with embodied and operational values.
type CarbonFootprint struct {
	Embodied    int   `json:"embodied"`
	Operational []int `json:"operational"`
}

const (
	// K8SliceNameDefault is the default name for a K8Slice Flavor.
	K8SliceNameDefault FlavorTypeName = "k8slice"
	// VMNameDefault is the default name for a VM Flavor.
	VMNameDefault FlavorTypeName = "vm"
	// ServiceNameDefault is the default name for a Service Flavor.
	ServiceNameDefault FlavorTypeName = "service"
	// SensorNameDefault is the default name for a Sensor Flavor.
	SensorNameDefault FlavorTypeName = "sensor"
)

// Location represents the location of a Flavor, with latitude, longitude, altitude, and additional notes.
type Location struct {
	Latitude        string `json:"latitude,omitempty"`
	Longitude       string `json:"longitude,omitempty"`
	Country         string `json:"altitude,omitempty"`
	City            string `json:"city,omitempty"`
	AdditionalNotes string `json:"additionalNotes,omitempty"`
}

// NodeIdentityAdditionalInfo represents additional information about a NodeIdentity.
type NodeIdentityAdditionalInfo struct {
	LiqoID string `json:"liqoID,omitempty"`
}

// NodeIdentity represents the owner of a Flavor, with associated ID, IP, and domain name.
type NodeIdentity struct {
	NodeID                string                      `json:"ID"`
	IP                    string                      `json:"IP"`
	Domain                string                      `json:"domain"`
	AdditionalInformation *NodeIdentityAdditionalInfo `json:"additionalInformation,omitempty"`
}

// Price represents the price of a Flavor, with the amount, currency, and period associated.
type Price struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Period   string `json:"period"`
}

// Selector represents the criteria for selecting Flavors.
type Selector interface {
	GetSelectorType() FlavorTypeName
}

// MapToFlavorTypeName maps a nodecorev1alpha1.FlavorTypeIdentifier to a models.FlavorTypeName.
func MapToFlavorTypeName(flavorType nodecorev1alpha1.FlavorTypeIdentifier) FlavorTypeName {
	switch flavorType {
	case nodecorev1alpha1.TypeK8Slice:
		return K8SliceNameDefault
	case nodecorev1alpha1.TypeVM:
		return VMNameDefault
	case nodecorev1alpha1.TypeService:
		return ServiceNameDefault
	default:
		return ""
	}
}

// MapFromFlavorTypeName maps a models.FlavorTypeName to a nodecorev1alpha1.FlavorTypeIdentifier.
func MapFromFlavorTypeName(flavorType FlavorTypeName) nodecorev1alpha1.FlavorTypeIdentifier {
	switch flavorType {
	case K8SliceNameDefault:
		return nodecorev1alpha1.TypeK8Slice
	case VMNameDefault:
		return nodecorev1alpha1.TypeVM
	case ServiceNameDefault:
		return nodecorev1alpha1.TypeService
	case SensorNameDefault:
		return nodecorev1alpha1.TypeSensor
	default:
		return ""
	}
}

// MapToFilterType maps a nodecorev1alpha1.FilterType to a models.FilterType.
func MapToFilterType(filterType nodecorev1alpha1.FilterType) FilterType {
	switch filterType {
	case nodecorev1alpha1.TypeMatchFilter:
		return MatchFilter
	case nodecorev1alpha1.TypeRangeFilter:
		return RangeFilter
	default:
		return ""
	}
}

// MapFromFilterType maps a models.FilterType to a nodecorev1alpha1.FilterType.
func MapFromFilterType(filterType FilterType) nodecorev1alpha1.FilterType {
	switch filterType {
	case MatchFilter:
		return nodecorev1alpha1.TypeMatchFilter
	case RangeFilter:
		return nodecorev1alpha1.TypeRangeFilter
	default:
		return ""
	}
}

// MapFromModelHostingPolicy maps a models.HostingPolicy to a nodecorev1alpha1.HostingPolicy.
func MapFromModelHostingPolicy(hostingPolicy HostingPolicy) nodecorev1alpha1.HostingPolicy {
	switch hostingPolicy {
	case HostingPolicyProvider:
		return nodecorev1alpha1.HostingPolicyProvider
	case HostingPolicyConsumer:
		return nodecorev1alpha1.HostingPolicyConsumer
	case HostingPolicyShared:
		return nodecorev1alpha1.HostingPolicyShared
	default:
		return ""
	}
}

// MapToModelHostingPolicy maps a nodecorev1alpha1.HostingPolicy to a models.HostingPolicy.
func MapToModelHostingPolicy(hostingPolicy nodecorev1alpha1.HostingPolicy) HostingPolicy {
	switch hostingPolicy {
	case nodecorev1alpha1.HostingPolicyProvider:
		return HostingPolicyProvider
	case nodecorev1alpha1.HostingPolicyConsumer:
		return HostingPolicyConsumer
	case nodecorev1alpha1.HostingPolicyShared:
		return HostingPolicyShared
	default:
		return ""
	}
}

// MapToServiceCategory maps a string to a consts.ServiceCategory.
func MapToServiceCategory(serviceCategory string) consts.ServiceCategory {
	klog.Info("Service category: ", serviceCategory)
	switch serviceCategory {
	case string(consts.Database):
		return consts.Database
	case string(consts.MessageQueue):
		return consts.MessageQueue
		// TODO(Service): Implement the mapping for the other service categories according to the ontology.
	default:
		return ""
	}
}
