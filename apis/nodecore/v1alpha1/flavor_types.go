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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

const (
	// TypeK8Slice is the type of a K8Slice Flavor.
	TypeK8Slice FlavorTypeIdentifier = "K8Slice"
	// TypeVM is the type of a VM Flavor.
	TypeVM FlavorTypeIdentifier = "VM"
	// TypeService is the type of a Service Flavor.
	TypeService FlavorTypeIdentifier = "Service"
	// TypeSensor is the type of a Sensor Flavor.
	TypeSensor FlavorTypeIdentifier = "Sensor"
)

// FlavorTypeIdentifier is the identifier of a Flavor type.
type FlavorTypeIdentifier string

// FlavorType represents the type of a Flavor.
type FlavorType struct {
	// Type of the Flavor.
	TypeIdentifier FlavorTypeIdentifier `json:"typeIdentifier"`
	// Raw is the raw value of the Flavor.
	TypeData runtime.RawExtension `json:"typeData"`
}

// Price represents the price of a Flavor.
type Price struct {
	// Amount is the amount of the price.
	Amount string `json:"amount"`

	// Currency is the currency of the price.
	Currency string `json:"currency"`

	// Period is the period of the price.
	Period string `json:"period"`
}

// Location represents the location of a Flavor.
type Location struct {
	// Latitude is the latitude of the location.
	Latitude string `json:"latitude,omitempty"`

	// Longitude is the longitude of the location.
	Longitude string `json:"longitude,omitempty"`

	// Country is the country of the location.
	Country string `json:"country,omitempty"`

	// City is the city of the location.
	City string `json:"city,omitempty"`

	// AdditionalNotes are additional notes of the location.
	AdditionalNotes string `json:"additionalNotes,omitempty"`
}

// FlavorSpec defines the desired state of Flavor.
type FlavorSpec struct {
	// This specs are based on the REAR Protocol specifications.

	// ProviderID is the ID of the FLUIDOS Node ID that provides this Flavor.
	// It can correspond to ID of the owner FLUIDOS Node or to the ID of a FLUIDOS SuperNode that represents the entry point to a FLUIDOS Domain
	ProviderID string `json:"providerID"`

	// FlavorType is the type of the Flavor.
	FlavorType FlavorType `json:"flavorType"`

	// Owner contains the identity info of the owner of the Flavor. It can be unknown if the Flavor is provided by a reseller or a third party.
	Owner NodeIdentity `json:"owner"`

	// Price contains the price model of the Flavor.
	Price Price `json:"price"`

	// Availability is the availability flag of the Flavor.
	Availability bool `json:"availability"`

	// NetworkPropertyType is the network property type of the Flavor.
	NetworkPropertyType string `json:"networkPropertyType,omitempty"`

	// Location is the location of the Flavor.
	Location *Location `json:"location,omitempty"`
}

// FlavorStatus defines the observed state of Flavor.
type FlavorStatus struct {

	// This field represents the expiration time of the Flavor. It is used to determine when the Flavor is no longer valid.
	ExpirationTime string `json:"expirationTime"`

	// This field represents the creation time of the Flavor.
	CreationTime string `json:"creationTime"`

	// This field represents the last update time of the Flavor.
	LastUpdateTime string `json:"lastUpdateTime"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Flavor is the Schema for the flavors API.
// +kubebuilder:printcolumn:name="Provider ID",type=string,JSONPath=`.spec.providerID`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.flavorType.typeIdentifier`
// +kubebuilder:printcolumn:name="Owner Name",type=string,priority=1,JSONPath=`.spec.owner.nodeID`
// +kubebuilder:printcolumn:name="Available",type=boolean,JSONPath=`.spec.availability`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Kubernetes Node Owner",type=string,JSONPath=`.metadata.ownerReferences[0].name`
// +kubebuilder:resource:shortName=fl
type Flavor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlavorSpec   `json:"spec,omitempty"`
	Status FlavorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FlavorList contains a list of Flavor.
type FlavorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Flavor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Flavor{}, &FlavorList{})
}

// ParseFlavorType parses a Flavor into a the type and the unmarshalled raw value.
func ParseFlavorType(flavor *Flavor) (FlavorTypeIdentifier, interface{}, error) {
	var validationErr error

	klog.Infof("Parsing Flavor type %s", flavor.Spec.FlavorType.TypeIdentifier)

	switch flavor.Spec.FlavorType.TypeIdentifier {
	case TypeK8Slice:

		var k8slice *K8Slice
		// Parse K8Slice flavor
		k8slice, err := ParseK8SliceFlavor(flavor.Spec.FlavorType)
		if err != nil {
			return "", nil, err
		}

		return TypeK8Slice, *k8slice, validationErr

	case TypeVM:
		// TODO (VM): implement the VM flavor parsing
		return "", nil, fmt.Errorf("flavor type %s not supported", flavor.Spec.FlavorType.TypeIdentifier)

	case TypeService:
		var service *ServiceFlavor
		// Parse Service flavor
		service, _, err := ParseServiceFlavor(flavor.Spec.FlavorType)
		if err != nil {
			return "", nil, err
		}

		return TypeService, *service, validationErr

	case TypeSensor:
		// TODO (Sensor): implement the sensor flavor parsing
		return "", nil, fmt.Errorf("flavor type %s not supported", flavor.Spec.FlavorType.TypeIdentifier)

	default:
		return "", nil, fmt.Errorf("flavor type %s not supported", flavor.Spec.FlavorType.TypeIdentifier)
	}
}
