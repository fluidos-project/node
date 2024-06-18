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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	K8S FlavourType = "k8s-fluidos"
)

type FlavourType string

type Characteristics struct {

	// Architecture is the architecture of the Flavour.
	Architecture string `json:"architecture"`

	// CPU is the number of CPU cores of the Flavour.
	Cpu resource.Quantity `json:"cpu"`

	// Memory is the amount of RAM of the Flavour.
	Memory resource.Quantity `json:"memory"`

	// Pods is the maximum number of pods of the Flavour.
	Pods resource.Quantity `json:"pods"`

	// GPU is the number of GPU cores of the Flavour.
	Gpu resource.Quantity `json:"gpu,omitempty"`

	// EphemeralStorage is the amount of ephemeral storage of the Flavour.
	EphemeralStorage resource.Quantity `json:"ephemeral-storage,omitempty"`

	// PersistentStorage is the amount of persistent storage of the Flavour.
	PersistentStorage resource.Quantity `json:"persistent-storage,omitempty"`
}

type Policy struct {

	// Partitionable contains the partitioning properties of the Flavour.
	Partitionable *Partitionable `json:"partitionable,omitempty"`

	// Aggregatable contains the aggregation properties of the Flavour.
	Aggregatable *Aggregatable `json:"aggregatable,omitempty"`
}

// Partitionable represents the partitioning properties of a Flavour, such as the minimum and incremental values of CPU and RAM.
type Partitionable struct {
	// CpuMin is the minimum requirable number of CPU cores of the Flavour.
	CpuMin resource.Quantity `json:"cpuMin"`

	// MemoryMin is the minimum requirable amount of RAM of the Flavour.
	MemoryMin resource.Quantity `json:"memoryMin"`

	// PodsMin is the minimum requirable number of pods of the Flavour.
	PodsMin resource.Quantity `json:"podsMin"`

	// CpuStep is the incremental value of CPU cores of the Flavour.
	CpuStep resource.Quantity `json:"cpuStep"`

	// MemoryStep is the incremental value of RAM of the Flavour.
	MemoryStep resource.Quantity `json:"memoryStep"`

	// PodsStep is the incremental value of pods of the Flavour.
	PodsStep resource.Quantity `json:"podsStep"`
}

// Aggregatable represents the aggregation properties of a Flavour, such as the minimum instance count.
type Aggregatable struct {
	// MinCount is the minimum requirable number of instances of the Flavour.
	MinCount int `json:"minCount"`

	// MaxCount is the maximum requirable number of instances of the Flavour.
	MaxCount int `json:"maxCount"`
}

type Price struct {

	// Amount is the amount of the price.
	Amount string `json:"amount"`

	// Currency is the currency of the price.
	Currency string `json:"currency"`

	// Period is the period of the price.
	Period string `json:"period"`
}

type OptionalFields struct {

	// Availability is the availability flag of the Flavour.
	// It is a field inherited from the REAR Protocol specifications.
	Availability bool `json:"availability,omitempty"`

	// WorkerID is the ID of the worker that provides the Flavour.
	WorkerID string `json:"workerID,omitempty"`
}

// FlavourSpec defines the desired state of Flavour
type FlavourSpec struct {
	// This specs are based on the REAR Protocol specifications.

	// ProviderID is the ID of the FLUIDOS Node ID that provides this Flavour.
	// It can correspond to ID of the owner FLUIDOS Node or to the ID of a FLUIDOS SuperNode that represents the entry point to a FLUIDOS Domain
	ProviderID string `json:"providerID"`

	// Type is the type of the Flavour. Currently, only K8S is supported.
	Type FlavourType `json:"type"`

	// Characteristics contains the characteristics of the Flavour.
	// They are based on the type of the Flavour and can change depending on it. In this case, the type is K8S so the characteristics are CPU, Memory, GPU and EphemeralStorage.
	Characteristics Characteristics `json:"characteristics"`

	// Policy contains the policy of the Flavour. The policy describes the partitioning and aggregation properties of the Flavour.
	Policy Policy `json:"policy"`

	// Owner contains the identity info of the owner of the Flavour. It can be unknown if the Flavour is provided by a reseller or a third party.
	Owner NodeIdentity `json:"owner"`

	// Price contains the price model of the Flavour.
	Price Price `json:"price"`

	// This field is used to specify the optional fields that can be retrieved from the Flavour.
	// In the future it will be expanded to include more optional fields defined in the REAR Protocol or custom ones.
	OptionalFields OptionalFields `json:"optionalFields"`

	// This field is used to represent the available quantity of transactions of the Flavour.
	QuantityAvailable int `json:"quantityAvailable,omitempty"`
}

// FlavourStatus defines the observed state of Flavour.
type FlavourStatus struct {

	// This field represents the expiration time of the Flavour. It is used to determine when the Flavour is no longer valid.
	ExpirationTime string `json:"expirationTime"`

	// This field represents the creation time of the Flavour.
	CreationTime string `json:"creationTime"`

	// This field represents the last update time of the Flavour.
	LastUpdateTime string `json:"lastUpdateTime"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// +kubebuilder:printcolumn:name="Provider ID",type=string,JSONPath=`.spec.providerID`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="CPU",type=string,priority=1,JSONPath=`.spec.characteristics.cpu`
// +kubebuilder:printcolumn:name="Memory",type=string,priority=1,JSONPath=`.spec.characteristics.memory`
// +kubebuilder:printcolumn:name="Owner Name",type=string,priority=1,JSONPath=`.spec.owner.nodeID`
// +kubebuilder:printcolumn:name="Owner Domain",type=string,priority=1,JSONPath=`.spec.owner.domain`
// +kubebuilder:printcolumn:name="Available",type=boolean,JSONPath=`.spec.optionalFields.availability`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// Flavour is the Schema for the flavours API.
type Flavour struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlavourSpec   `json:"spec,omitempty"`
	Status FlavourStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FlavourList contains a list of Flavour
type FlavourList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Flavour `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Flavour{}, &FlavourList{})
}
