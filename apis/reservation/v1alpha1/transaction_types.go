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

// TransactionSpec defines the desired state of Transaction.
type TransactionSpec struct {
	// FlavorID is the ID of the flavor that is being reserved
	FlavorID string `json:"flavorID"`

	// Buyer is the buyer Identity of the Fluidos Node that is reserving the Flavor
	Buyer nodecorev1alpha1.NodeIdentity `json:"buyer"`

	// ClusterID is the Liqo ClusterID of the Fluidos Node that is reserving the Flavor
	ClusterID string `json:"clusterID"`

	// Configuration is the configuration of the flavor that is being reserved
	Configuration *nodecorev1alpha1.Configuration `json:"configuration,omitempty"`

	// ExpirationTime is the time when the reservation will expire
	ExpirationTime string `json:"expirationTime,omitempty"`
}

// TransactionStatus defines the observed state of Transaction.
type TransactionStatus struct {
	// This is the current phase of the reservation
	Phase nodecorev1alpha1.PhaseStatus `json:"phase"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Transaction is the Schema for the transactions API.
// +kubebuilder:printcolumn:name="Flavor ID",type="string",JSONPath=".spec.flavorID"
// +kubebuilder:printcolumn:name="Buyer Name",type="string",JSONPath=".spec.buyer.nodeID"
// +kubebuilder:printcolumn:name="Buyer IP",type="string",priority=1,JSONPath=".spec.buyer.ip"
// +kubebuilder:printcolumn:name="Buyer Domain",type="string",priority=1,JSONPath=".spec.buyer.domain"
// +kubebuilder:printcolumn:name="Cluster ID",type="string",JSONPath=".spec.clusterID"
// +kubebuilder:printcolumn:name="Start Time",type="string",JSONPath=".spec.startTime"
// +kubebuilder:resource:shortName=tr
type Transaction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransactionSpec   `json:"spec,omitempty"`
	Status TransactionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TransactionList contains a list of Transaction.
type TransactionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Transaction `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Transaction{}, &TransactionList{})
}
