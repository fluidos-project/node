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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//nolint:revive // Do not need to repeat the same comment
type Status string

// Status is the status of the allocation.
const (
	Active   Status = "Active"
	Reserved Status = "Reserved"
	Released Status = "Released"
	Inactive Status = "Inactive"
	Error    Status = "Error"
)

// AllocationSpec defines the desired state of Allocation.
type AllocationSpec struct {
	// This is the ID of the intent for which the allocation was created.
	// It is used by the Node Orchestrator to identify the correct allocation for a given intent
	IntentID string `json:"intentID"`

	// This flag indicates if the allocation is a forwarding allocation
	// if true it represents only a placeholder to undertand that the cluster is just a proxy to another cluster
	Forwarding bool `json:"forwarding,omitempty"`

	// This is the reference to the contract related to the allocation
	Contract GenericRef `json:"contract,omitempty"`
}

// AllocationStatus defines the observed state of Allocation.
type AllocationStatus struct {

	// This allow to know the current status of the allocation
	Status Status `json:"status,omitempty"`

	// The last time the allocation was updated
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`

	// Message contains the last message of the allocation
	Message string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=alloc;allocs

// Allocation is the Schema for the allocations API.
type Allocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AllocationSpec   `json:"spec,omitempty"`
	Status AllocationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AllocationList contains a list of Allocation.
type AllocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Allocation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Allocation{}, &AllocationList{})
}
