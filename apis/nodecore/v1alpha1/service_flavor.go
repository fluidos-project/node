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

	"github.com/xeipuuv/gojsonschema"
	"k8s.io/apimachinery/pkg/runtime"
)

// ServiceFlavor represents a Service Flavor description.
type ServiceFlavor struct {
	// Name of the Service Flavor.
	Name string `json:"name"`
	// Description of the Service Flavor.
	Description string `json:"description"`
	// Category of the Service Flavor.
	Category string `json:"category"`
	// Tags of the Service Flavor.
	Tags []string `json:"tags"`
	// HostingPolicies of the Service Flavor.
	HostingPolicies []HostingPolicy `json:"hostingPolicies"`
	// ConfigurationTemplate of the Service Flavor. JSON Schema with the parameters that can be configured.
	ConfigurationTemplate runtime.RawExtension `json:"configurationTemplate"`
}

// ServiceIdentifier represents the identifier of a Service.
type ServiceIdentifier string

// GetFlavorType returns the type of the Flavor.
func (sf *ServiceFlavor) GetFlavorType() FlavorTypeIdentifier {
	return TypeService
}

// HostingPolicy represents the hosting policy chosen for the service.
type HostingPolicy string

const (
	// HostingPolicyProvider represents the hosting policy where the service will be hosted by a provider.
	HostingPolicyProvider HostingPolicy = "Provider"
	// HostingPolicyConsumer represents the hosting policy where the service will be hosted by a consumer.
	HostingPolicyConsumer HostingPolicy = "Consumer"
	// HostingPolicyShared represents the hosting policy where the service can be hosted by both provider and consumer
	// and the exact hosting policy is not defined.
	HostingPolicyShared HostingPolicy = "Shared"
)

// ParseServiceFlavor parses the ServiceFlavor from a string.
func ParseServiceFlavor(flavorType FlavorType) (*ServiceFlavor, *gojsonschema.Schema, error) {
	serviceFlavor := &ServiceFlavor{}
	// Check type of the Flavor
	if flavorType.TypeIdentifier != TypeService {
		return nil, nil, fmt.Errorf("flavor type is not a Service")
	}

	// Unmarshal the raw data into the K8Slice struct
	if err := json.Unmarshal(flavorType.TypeData.Raw, serviceFlavor); err != nil {
		return nil, nil, err
	}

	// parse the configuration template
	schema, err := parseConfigurationTemplate(serviceFlavor.ConfigurationTemplate)
	if err != nil {
		return nil, nil, err
	}
	return serviceFlavor, schema, nil
}

// parseConfigurationTemplate parses the ConfigurationTemplate from a runtime.RawExtension to a JSON Schema.
func parseConfigurationTemplate(configurationTemplate runtime.RawExtension) (*gojsonschema.Schema, error) {
	var schemaMap map[string]interface{}

	// Decode the raw extension into a map
	if err := json.Unmarshal(configurationTemplate.Raw, &schemaMap); err != nil {
		return nil, err
	}

	// Convert the map into a JSON Schema
	schemaLoader := gojsonschema.NewGoLoader(schemaMap)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to load JSON schema: %w", err)
	}

	return schema, nil
}
