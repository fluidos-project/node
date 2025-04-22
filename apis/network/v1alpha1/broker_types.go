// Copyright 2022-2025 FLUIDOS Project
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

// BrokerSpec defines the desired state of Broker.
type BrokerSpec struct {

	// Address of the Broker.
	Address string `json:"address"`
	Name    string `json:"name"`
	ClCert  string `json:"clcert"`
	CaCert  string `json:"cacert"`
	Role    string `json:"role"`
}

// BrokerStatus defines the observed state of Broker.
type BrokerStatus struct {

	// This field represents the expiration time of the Broker. It is used to determine when the Broker is no longer valid.
	ExpirationTime string `json:"expirationTime"`

	// This field represents the last update time of the Broker.
	LastUpdateTime string `json:"lastUpdateTime"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=broker;brokers

// Broker is the Schema for the clusters API.
type Broker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BrokerSpec   `json:"spec,omitempty"`
	Status BrokerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BrokerList contains a list of Broker.
type BrokerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Broker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Broker{}, &BrokerList{})
}
