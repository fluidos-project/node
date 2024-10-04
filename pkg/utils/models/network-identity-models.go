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

import "encoding/json"

// ResourceSelectorIdentifier represents the identifier of a ResourceSelector.
type ResourceSelectorIdentifier string

const (
	// PodNamespaceSelectorType is the type of a PodNamespaceSelector.
	PodNamespaceSelectorType ResourceSelectorIdentifier = "PodNamespaceSelector"
	// CIDRSelectorType is the type of a CIDRSelector.
	CIDRSelectorType ResourceSelectorIdentifier = "CIDRSelector"
)

// NetworkIntent represents the network intent of a Flavor, with source and destination.
type NetworkIntent struct {
	Name            string            `json:"name"`
	Source          SourceDestination `json:"source"`
	Destination     SourceDestination `json:"destination"`
	DestinationPort string            `json:"destinationPort"`
	ProtocolType    string            `json:"protocolType"`
}

// NetworkAuthorizations represents the network authorizations of a Flavor, with denied and mandatory communications.
type NetworkAuthorizations struct {
	DeniedCommunications    []NetworkIntent `json:"deniedCommunications"`
	MandatoryCommunications []NetworkIntent `json:"mandatoryCommunications"`
}

// SourceDestination represents the source or destination of a network intent.
type SourceDestination struct {
	IsHotCluster     bool             `json:"isHotCluster"`
	ResourceSelector ResourceSelector `json:"resourceSelectors"`
}

// ResourceSelector represents the resource selector of a SourceDestination.
type ResourceSelector struct {
	TypeIdentifier ResourceSelectorIdentifier `json:"typeIdentifier"`
	Selector       json.RawMessage            `json:"selector"`
}

// CIDRSelector represents the CIDR selector of a SourceDestination.
type CIDRSelector string

// PodNamespaceSelector represents the pod namespace selector of a SourceDestination.
type PodNamespaceSelector struct {
	Pod       []keyValuePair `json:"pod"`
	Namespace []keyValuePair `json:"namespace"`
}

type keyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
