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

// TelemetryServer defines the telemetry server configuration.
type TelemetryServer struct {
	// Endpoint is the endpoint where the telemetry is sent by the provider
	Endpoint string `json:"endpoint"`
	// Intents is the list of intents
	Intents []string `json:"intents,omitempty"`
}

// ReservationSpec defines the desired state of Reservation.
type ReservationSpec struct {

	// SolverID is the ID of the solver that asks for the reservation
	SolverID string `json:"solverID"`

	// This is the Node identity of the buyer FLUIDOS Node.
	Buyer nodecorev1alpha1.NodeIdentity `json:"buyer"`

	// This is the Node identity of the seller FLUIDOS Node.
	Seller nodecorev1alpha1.NodeIdentity `json:"seller"`

	// Configuration is the configuration of the flavour that is being reserved
	Configuration *nodecorev1alpha1.Configuration `json:"configuration,omitempty"`

	// Reserve indicates if the reservation is a reserve or not
	Reserve bool `json:"reserve,omitempty"`

	// Purchase indicates if the reservation is an purchase or not
	Purchase bool `json:"purchase,omitempty"`

	// PeeringCandidate is the reference to the PeeringCandidate of the Reservation
	PeeringCandidate nodecorev1alpha1.GenericRef `json:"peeringCandidate,omitempty"`

	// IngressTelemetryEndpoint is the endpoint where the ingress telemetry is sent by the provider
	IngressTelemetryEndpoint *TelemetryServer `json:"ingressTelemetryEndpoint,omitempty"`
}

// ReservationStatus defines the observed state of Reservation.
type ReservationStatus struct {
	// This is the current phase of the reservation
	Phase nodecorev1alpha1.PhaseStatus `json:"phase"`

	// ReservePhase is the current phase of the reservation
	ReservePhase nodecorev1alpha1.Phase `json:"reservePhase,omitempty"`

	// PurchasePhase is the current phase of the reservation
	PurchasePhase nodecorev1alpha1.Phase `json:"purchasePhase,omitempty"`

	// TransactionID is the ID of the transaction that this reservation is part of
	TransactionID string `json:"transactionID,omitempty"`

	// Contract is the reference to the Contract of the Reservation
	Contract nodecorev1alpha1.GenericRef `json:"contract,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Reservation is the Schema for the reservations API.
// +kubebuilder:printcolumn:name="Solver ID",type=string,JSONPath=`.spec.solverID`
// +kubebuilder:printcolumn:name="Reserve",type=boolean,JSONPath=`.spec.reserve`
// +kubebuilder:printcolumn:name="Purchase",type=boolean,JSONPath=`.spec.purchase`
// +kubebuilder:printcolumn:name="Seller Name",type=string,JSONPath=`.spec.seller.nodeID`
// +kubebuilder:printcolumn:name="Seller Domain",type=string,priority=1,JSONPath=`.spec.buyer.domain`
// +kubebuilder:printcolumn:name="Peering Candidate",type=string,priority=1,JSONPath=`.spec.peeringCandidate.name`
// +kubebuilder:printcolumn:name="Transaction ID",type=string,JSONPath=`.status.transactionID`
// +kubebuilder:printcolumn:name="Reserve Phase",type=string,priority=1,JSONPath=`.status.reservePhase`
// +kubebuilder:printcolumn:name="Purchase Phase",type=string,priority=1,JSONPath=`.status.purchasePhase`
// +kubebuilder:printcolumn:name="Contract Name",type=string,JSONPath=`.status.contract.name`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase.phase`
// +kubebuilder:printcolumn:name="Message",type=string,priority=1,JSONPath=`.status.phase.message`
// +kubebuilder:resource:shortName=res
type Reservation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReservationSpec   `json:"spec,omitempty"`
	Status ReservationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReservationList contains a list of Reservation.
type ReservationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Reservation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Reservation{}, &ReservationList{})
}
