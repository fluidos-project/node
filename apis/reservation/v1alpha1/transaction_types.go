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
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TransactionSpec defines the desired state of Transaction
type TransactionSpec struct {
	// FlavourID is the ID of the flavour that is being reserved
	FlavourID string `json:"flavourID"`

	// Buyer is the buyer Identity of the Fluidos Node that is reserving the Flavour
	Buyer nodecorev1alpha1.NodeIdentity `json:"buyer"`

	// Partition is the partition of the flavour that is being reserved
	Partition Partition `json:"partition,omitempty"`

	// StartTime is the time at which the reservation should start
	StartTime string `json:"startTime,omitempty"`
}

// TransactionStatus defines the observed state of Transaction
type TransactionStatus struct {
	// This is the current phase of the reservation
	Phase nodecorev1alpha1.PhaseStatus `json:"phase"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Transaction is the Schema for the transactions API
type Transaction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransactionSpec   `json:"spec,omitempty"`
	Status TransactionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TransactionList contains a list of Transaction
type TransactionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Transaction `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Transaction{}, &TransactionList{})
}
