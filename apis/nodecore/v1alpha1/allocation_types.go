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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//nolint:revive // Do not need to repeat the same comment
type NodeType string

//nolint:revive // Do not need to repeat the same comment
type Status string

//nolint:revive // Do not need to repeat the same comment
type Destination string

// NodeType is the type of the node: Node (Physical node of the cluster) or VirtualNode (Remote node owned by a different cluster).
const (
	Node        NodeType = "Node"
	VirtualNode NodeType = "VirtualNode"
)

// Status is the status of the allocation.
const (
	Active   Status = "Active"
	Reserved Status = "Reserved"
	Released Status = "Released"
	Inactive Status = "Inactive"
	Error    Status = "Error"
)

// Destination is the destination of the allocation: Local (the allocation will be used locally)
// or Remote (the allocation will be used from a remote cluster).
const (
	Remote Destination = "Remote"
	Local  Destination = "Local"
)

// AllocationSpec defines the desired state of Allocation
type AllocationSpec struct {
	// This is the ID of the cluster that owns the allocation.
	RemoteClusterID string `json:"remoteClusterID,omitempty"`

	// This is the ID of the intent for which the allocation was created.
	// It is used by the Node Orchestrator to identify the correct allocation for a given intent
	IntentID string `json:"intentID"`

	// This is the corresponding Node or VirtualNode local name
	NodeName string `json:"nodeName"`

	// This specifies the type of the node: Node (Physical node of the cluster) or VirtualNode (Remote node owned by a different cluster)
	Type NodeType `json:"type"`

	// This specifies if the destination of the allocation is local or remote so if the allocation will be used locally or from a remote cluster
	Destination Destination `json:"destination"`

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

// Allocation is the Schema for the allocations API.
type Allocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AllocationSpec   `json:"spec,omitempty"`
	Status AllocationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AllocationList contains a list of Allocation
type AllocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Allocation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Allocation{}, &AllocationList{})
}
