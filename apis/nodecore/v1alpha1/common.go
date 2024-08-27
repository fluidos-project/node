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
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

// Set of constants for the phases of the FLUIDOS Node modules.
const (
	PhaseSolved     Phase = "Solved"
	PhaseFailed     Phase = "Failed"
	PhaseRunning    Phase = "Running"
	PhaseAllocating Phase = "Allocating"
	PhaseIdle       Phase = "Idle"
	PhaseTimeout    Phase = "Timed Out"
	PhaseActive     Phase = "Active"
	PhasePending    Phase = "Pending"
	PhaseInactive   Phase = "Inactive"
)

// GenericRef represents a reference to a generic Kubernetes resource,
// and it is composed of the resource name and (optionally) its namespace.
type GenericRef struct {
	// The name of the resource to be referenced.
	Name string `json:"name,omitempty"`

	// The namespace containing the resource to be referenced. It should be left
	// empty in case of cluster-wide resources.
	Namespace string `json:"namespace,omitempty"`

	// The API version of the resource to be referenced.
	APIVersion string `json:"apiVersion,omitempty"`

	// The kind of the resource to be referenced.
	Kind string `json:"kind,omitempty"`
}

func (g *GenericRef) String() string {
	return fmt.Sprintf("%s:%s - %s/%s", g.APIVersion, g.Kind, g.Namespace, g.Name)
}

// NodeIdentity is the identity of a FLUIDOS Node.
type NodeIdentity struct {
	Domain                string                      `json:"domain"`
	NodeID                string                      `json:"nodeID"`
	IP                    string                      `json:"ip"`
	AdditionalInformation *NodeIdentityAdditionalInfo `json:"additionalInformation,omitempty"`
}

// NodeIdentityAdditionalInfo contains additional information about the node.
type NodeIdentityAdditionalInfo struct {
	LiqoID string `json:"liqoID,omitempty"`
}

// Configuration represents the configuration of a FLUIDOS Node.
type Configuration struct {
	// Identifier is the identifier of the configuration.
	ConfigurationTypeIdentifier FlavorTypeIdentifier `json:"type"`
	// ConfigurationData is the data of the configuration.
	ConfigurationData runtime.RawExtension `json:"data"`
}

// LiqoCredentials contains the credentials of a Liqo cluster to enstablish a peering.
type LiqoCredentials struct {
	ClusterID   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
	Token       string `json:"token"`
	Endpoint    string `json:"endpoint"`
}

// ParseConfiguration parses the configuration data into the correct type.
// Returns the FlavorTypeIdentifier, aka the ConfigurationTypeIdentifier and the configuration data.
func ParseConfiguration(configuration *Configuration, flavor *Flavor) (FlavorTypeIdentifier, interface{}, error) {
	var validationError error

	switch configuration.ConfigurationTypeIdentifier {
	case TypeK8Slice:
		var k8SliceConfiguration K8SliceConfiguration
		validationError = json.Unmarshal(configuration.ConfigurationData.Raw, &k8SliceConfiguration)
		return TypeK8Slice, k8SliceConfiguration, validationError
	case TypeService:
		var serviceConfiguration ServiceConfiguration
		validationError = json.Unmarshal(configuration.ConfigurationData.Raw, &serviceConfiguration)
		if validationError != nil {
			return "", nil, validationError
		}
		// Parse Flavor
		flavorType, flavorData, err := ParseFlavorType(flavor)
		if err != nil {
			return "", nil, err
		}
		// Forcing the type to be ServiceFlavor
		if flavorType != TypeService {
			return "", nil, fmt.Errorf("flavor type %s does not match the configuration type %s", flavorType, TypeService)
		}
		// Get ServiceFlavor
		serviceFlavor := flavorData.(ServiceFlavor)
		validationError = serviceConfiguration.Validate(&serviceFlavor)

		return TypeService, serviceConfiguration, validationError
	case TypeVM:
		// TODO (VM): implement the VM configuration parsing
		return TypeVM, nil, nil
	case TypeSensor:
		// TODO (Sensor): implement the sensor configuration parsing
		return TypeSensor, nil, nil
	default:
		return "", nil, fmt.Errorf("partition type %s not supported", configuration.ConfigurationTypeIdentifier)
	}
}
