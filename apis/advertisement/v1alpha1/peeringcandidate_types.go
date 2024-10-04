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

// PeeringCandidateSpec defines the desired state of PeeringCandidate.
type PeeringCandidateSpec struct {
	InterestedSolverIDs []string                `json:"interestedSolverIDs"`
	Flavor              nodecorev1alpha1.Flavor `json:"flavor"`
	Available           bool                    `json:"available"`
}

// PeeringCandidateStatus defines the observed state of PeeringCandidate.
type PeeringCandidateStatus struct {

	// This field represents the creation time of the PeeringCandidate.
	CreationTime string `json:"creationTime"`

	// This field represents the last update time of the PeeringCandidate.
	LastUpdateTime string `json:"lastUpdateTime"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PeeringCandidate is the Schema for the peeringcandidates API.
type PeeringCandidate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PeeringCandidateSpec   `json:"spec,omitempty"`
	Status PeeringCandidateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PeeringCandidateList contains a list of PeeringCandidate.
type PeeringCandidateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PeeringCandidate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PeeringCandidate{}, &PeeringCandidateList{})
}
