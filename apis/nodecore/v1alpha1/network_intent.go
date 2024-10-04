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

	"k8s.io/apimachinery/pkg/runtime"
)

// PodNamespaceSelector represents the pod namespace selector of a SourceDestination.
type PodNamespaceSelector struct {
	// Pod is the pod selector of the SourceDestination.
	Pod map[string]string `json:"pod"`
	// Namespace is the namespace selector of the SourceDestination.
	Namespace map[string]string `json:"namespace"`
}

// CIDRSelector represents the CIDR selector of a SourceDestination.
type CIDRSelector string

// ResourceSelectorIdentifier represents the type of a ResourceSelector.
type ResourceSelectorIdentifier string

const (
	// PodNamespaceSelectorType is the type of a PodNamespaceSelector.
	PodNamespaceSelectorType ResourceSelectorIdentifier = "PodNamespaceSelector"
	// CIDRSelectorType is the type of a CIDRSelector.
	CIDRSelectorType ResourceSelectorIdentifier = "CIDRSelector"
)

// ResourceSelector represents the resource selector of a SourceDestination.
type ResourceSelector struct {
	// TypeIdentifier is the type of the resource selector.
	TypeIdentifier ResourceSelectorIdentifier `json:"typeIdentifier"`
	// Selector is the selector of the resource selector.
	Selector runtime.RawExtension `json:"selector"`
}

// SourceDestination can represent either the source or destination of a network intent.
type SourceDestination struct {
	// IsHotCluster is true if the source/destination is a hot cluster.
	IsHotCluster bool `json:"isHotCluster"`
	// ResourceSelector is the resource selector of the source/destination.
	ResourceSelector ResourceSelector `json:"resourceSelector"`
}

// NetworkIntent represents the network intent of a Flavor.
type NetworkIntent struct {
	// Name of the network intent
	Name string `json:"name"`
	// Source of the network intent
	Source SourceDestination `json:"source"`
	// Destination of the network intent
	Destination SourceDestination `json:"destination"`
	// DestinationPort of the network intent
	DestinationPort string `json:"destinationPort"`
	// ProtocolType of the network intent
	ProtocolType string `json:"protocolType"`
}

// ParseResourceSelector parses a ResourceSelector into a ResourceSelectorIdentifier and a selector.
func ParseResourceSelector(selector ResourceSelector) (ResourceSelectorIdentifier, interface{}, error) {
	switch selector.TypeIdentifier {
	case PodNamespaceSelectorType:
		podNamespaceSelector := PodNamespaceSelector{}
		err := json.Unmarshal(selector.Selector.Raw, &podNamespaceSelector)
		if err != nil {
			return "", nil, err
		}
		return PodNamespaceSelectorType, podNamespaceSelector, nil
	case CIDRSelectorType:
		cidrSelector := CIDRSelector("")
		err := json.Unmarshal(selector.Selector.Raw, &cidrSelector)
		if err != nil {
			return "", nil, err
		}
		return CIDRSelectorType, cidrSelector, nil
	default:
		return "", nil, fmt.Errorf("unknown resource selector type: %s", selector.TypeIdentifier)
	}
}
