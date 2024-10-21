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

// KnownClusterSpec defines the desired state of KnownCluster.
type KnownClusterSpec struct {

	// Address of the KnownCluster.
	Address string `json:"address"`
}

// KnownClusterStatus defines the observed state of KnownCluster.
type KnownClusterStatus struct {

	// This field represents the expiration time of the KnownCluster. It is used to determine when the KnownCluster is no longer valid.
	ExpirationTime string `json:"expirationTime"`

	// This field represents the last update time of the KnownCluster.
	LastUpdateTime string `json:"lastUpdateTime"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=kclust;kclusts

// KnownCluster is the Schema for the clusters API.
type KnownCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KnownClusterSpec   `json:"spec,omitempty"`
	Status KnownClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KnownClusterList contains a list of KnownCluster.
type KnownClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KnownCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KnownCluster{}, &KnownClusterList{})
}
