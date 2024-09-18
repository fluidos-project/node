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

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DiscoverySpec defines the desired state of Discovery.
type DiscoverySpec struct {

	// This is the Solver ID of the solver that creates and so asks for the discovery.
	// This is a reference to the Solver CRD
	SolverID string `json:"solverID"`

	// This is the FlavourSelector that describes the characteristics of the intent that the solver is looking to satisfy
	// This pattern corresponds to what has been defined in the REAR Protocol to do a discovery with a selector
	Selector *nodecorev1alpha1.Selector `json:"selector,omitempty"`

	// This flag indicates that needs to be established a subscription to the provider in case a match is found.
	// In order to have periodic updates of the status of the matching Flavor
	Subscribe bool `json:"subscribe"`
}

// DiscoveryStatus defines the observed state of Discovery.
type DiscoveryStatus struct {

	// This is the current phase of the discovery
	Phase nodecorev1alpha1.PhaseStatus `json:"phase"`

	// This is a list of the PeeringCandidates that have been found as a result of the discovery matching the solver
	PeeringCandidateList PeeringCandidateList `json:"peeringCandidateList,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Discovery is the Schema for the discoveries API.
// +kubebuilder:printcolumn:name="Solver ID",type=string,JSONPath=`.spec.solverID`
// +kubebuilder:printcolumn:name="Subscribe",type=boolean,JSONPath=`.spec.subscribe`
// +kubebuilder:printcolumn:name="PC Namespace",type=string,JSONPath=`.status.peeringCandidate.namespace`
// +kubebuilder:printcolumn:name="PC Name",type=string,JSONPath=`.status.peeringCandidate.name`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase.phase`
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.phase.message`
// +kubebuilder:resource:shortName=dis
type Discovery struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiscoverySpec   `json:"spec,omitempty"`
	Status DiscoveryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DiscoveryList contains a list of Discovery.
type DiscoveryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Discovery `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Discovery{}, &DiscoveryList{})
}
