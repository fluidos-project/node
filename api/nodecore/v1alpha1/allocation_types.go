/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeType string
type Status string

const (
	Node        NodeType = "Node"
	VirtualNode NodeType = "VirtualNode"
)

const (
	Active   Status = "Active"
	Reserved Status = "Reserved"
	Released Status = "Released"
	Inactive Status = "Inactive"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AllocationSpec defines the desired state of Allocation
type AllocationSpec struct {

	// CustomerID

	// This is the ID of the intent for which the allocation was created. It is used by the Node Orchestrator to identify the correct allocation for a given intent
	IntentID string `json:"intentID"`

	// This is the corresponding Node or VirtualNode name
	LocalNode string `json:"localNode"`

	// This specifies the type of the node: Node (Physical node of the cluster) or VirtualNode (Remote node owned by a different cluster)
	Type NodeType `json:"type"`

	// This flag indicates if the allocation is a forwarding allocation, if true it represents only a placeholder to undertand that the cluster is just a proxy to another cluster
	Forwarding bool `json:"forwarding"`

	// This Flavour describes the characteristics of the allocation, it is based on the Flavour CRD from which it was created
	Flavour Flavour `json:"flavour"`

	// This is the dimension of the allocation, it is based on the Flavour CRD from which it was created
	Partition FlavourSelector `json:"partition,omitempty"`
}

// AllocationStatus defines the observed state of Allocation
type AllocationStatus struct {

	// This allow to know the current status of the allocation
	Status Status `json:"status"`

	// The creation time of the allocation object
	CreationTime metav1.Time `json:"creationTime"`

	// The last time the allocation was updated
	LastUpdateTime metav1.Time `json:"lastUpdateTime"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Allocation is the Schema for the allocations API
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
