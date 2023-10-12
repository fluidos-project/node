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

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
)

// LiqoCredentials contains the credentials of a Liqo cluster to enstablish a peering.
type LiqoCredentials struct {
	ClusterID   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
	Token       string `json:"token"`
	Endpoint    string `json:"endpoint"`
}

// ContractSpec defines the desired state of Contract
type ContractSpec struct {

	// TODO: ClusterID is the ID used by Liqo to search a contract during the peering phase.

	// This is the flavour on which the contract is based. It is used to lifetime maintain the critical characteristics of the contract.
	Flavour nodecorev1alpha1.Flavour `json:"flavour"`

	// TransactionID is the ID of the transaction that this contract is part of
	TransactionID string `json:"transactionID"`

	// The partition represents the dimension of the resources sold/bought.
	// So it will reflect the dimension of the resources allocated on the remote cluster and reflected on the local virtual node.
	Partition Partition `json:"partition,omitempty"`

	// This is the Node identity of the buyer FLUIDOS Node.
	Buyer nodecorev1alpha1.NodeIdentity `json:"buyer"`

	// BuyerClusterID is the Liqo ClusterID used by the seller to search a contract and the related resources during the peering phase.
	BuyerClusterID string `json:"buyerClusterID"`

	// This is the Node identity of the seller FLUIDOS Node.
	Seller nodecorev1alpha1.NodeIdentity `json:"seller"`

	// This credentials will be used by the customer to connect and enstablish a peering with the seller FLUIDOS Node through Liqo.
	SellerCredentials LiqoCredentials `json:"sellerCredentials"`

	// This is the expiration time of the contract. It can be empty if the contract is not time limited.
	ExpirationTime string `json:"expirationTime,omitempty"`

	// This contains additional information about the contract if needed.
	ExtraInformation map[string]string `json:"extraInformation,omitempty"`
}

// ContractStatus defines the observed state of Contract
type ContractStatus struct {

	// This is the status of the contract.
	Phase nodecorev1alpha1.PhaseStatus `json:"phase"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Contract is the Schema for the contracts API
type Contract struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContractSpec   `json:"spec,omitempty"`
	Status ContractStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ContractList contains a list of Contract
type ContractList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Contract `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Contract{}, &ContractList{})
}
