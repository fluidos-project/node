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

package models

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

// Flavour represents a Flavour object with its characteristics and policies.
type Flavour struct {
	FlavourID         string          `json:"flavourID"`
	ProviderID        string          `json:"providerID"`
	Type              string          `json:"type"`
	Characteristics   Characteristics `json:"characteristics"`
	Policy            Policy          `json:"policy"`
	Owner             NodeIdentity    `json:"owner"`
	Price             Price           `json:"price"`
	ExpirationTime    time.Time       `json:"expirationTime"`
	QuantityAvailable int             `json:"quantityAvailable,omitempty"`
	OptionalFields    OptionalFields  `json:"optionalFields"`
}

// Characteristics represents the characteristics of a Flavour, such as CPU and RAM.
type Characteristics struct {
	CPU               resource.Quantity `json:"cpu,omitempty"`
	Memory            resource.Quantity `json:"memory,omitempty"`
	Pods              resource.Quantity `json:"pods,omitempty"`
	PersistentStorage resource.Quantity `json:"storage,omitempty"`
	EphemeralStorage  resource.Quantity `json:"ephemeralStorage,omitempty"`
	Gpu               resource.Quantity `json:"gpu,omitempty"`
	Architecture      string            `json:"architecture,omitempty"`
}

// Policy represents the policy associated with a Flavour, which can be either Partitionable or Aggregatable.
type Policy struct {
	Partitionable *Partitionable `json:"partitionable,omitempty"`
	Aggregatable  *Aggregatable  `json:"aggregatable,omitempty"`
}

// Partitionable represents the partitioning properties of a Flavour, such as the minimum and incremental values of CPU and RAM.
type Partitionable struct {
	CPUMinimum    resource.Quantity `json:"cpuMinimum"`
	MemoryMinimum resource.Quantity `json:"memoryMinimum"`
	PodsMinimum   resource.Quantity `json:"podsMinimum"`
	CPUStep       resource.Quantity `json:"cpuStep"`
	MemoryStep    resource.Quantity `json:"memoryStep"`
	PodsStep      resource.Quantity `json:"podsStep"`
}

// Aggregatable represents the aggregation properties of a Flavour, such as the minimum instance count.
type Aggregatable struct {
	MinCount int `json:"minCount"`
	MaxCount int `json:"maxCount"`
}

// NodeIdentity represents the owner of a Flavour, with associated ID, IP, and domain name.
type NodeIdentity struct {
	NodeID string `json:"ID"`
	IP     string `json:"IP"`
	Domain string `json:"domain"`
}

// Price represents the price of a Flavour, with the amount, currency, and period associated.
type Price struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Period   string `json:"period"`
}

// OptionalFields represents the optional fields of a Flavour, such as availability.
type OptionalFields struct {
	Availability bool   `json:"availability,omitempty"`
	WorkerID     string `json:"workerID"`
}

// Selector represents the criteria for selecting Flavours.
type Selector struct {
	FlavourType   string         `json:"type,omitempty"`
	Architecture  string         `json:"architecture,omitempty"`
	RangeSelector *RangeSelector `json:"rangeSelector,omitempty"`
	MatchSelector *MatchSelector `json:"matchSelector,omitempty"`
}

// MatchSelector represents the criteria for selecting Flavours through a strict match.
type MatchSelector struct {
	CPU              resource.Quantity `json:"cpu,omitempty"`
	Memory           resource.Quantity `json:"memory,omitempty"`
	Pods             resource.Quantity `json:"pods,omitempty"`
	Storage          resource.Quantity `json:"storage,omitempty"`
	EphemeralStorage resource.Quantity `json:"ephemeralStorage,omitempty"`
	Gpu              resource.Quantity `json:"gpu,omitempty"`
}

// RangeSelector represents the criteria for selecting Flavours through a range.
type RangeSelector struct {
	MinCPU     resource.Quantity `json:"minCpu,omitempty"`
	MinMemory  resource.Quantity `json:"minMemory,omitempty"`
	MinPods    resource.Quantity `json:"minPods,omitempty"`
	MinStorage resource.Quantity `json:"minStorage,omitempty"`
	MinEph     resource.Quantity `json:"minEph,omitempty"`
	MinGpu     resource.Quantity `json:"minGpu,omitempty"`
	MaxCPU     resource.Quantity `json:"maxCpu,omitempty"`
	MaxMemory  resource.Quantity `json:"maxMemory,omitempty"`
	MaxPods    resource.Quantity `json:"maxPods,omitempty"`
	MaxStorage resource.Quantity `json:"maxStorage,omitempty"`
	MaxEph     resource.Quantity `json:"maxEph,omitempty"`
	MaxGpu     resource.Quantity `json:"maxGpu,omitempty"`
}
