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

package models

import "encoding/json"

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
	ConfigurationTemplate json.RawMessage `json:"configurationTemplate"`
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

// ServiceSelector represents the criteria for selecting a Service Flavor.
type ServiceSelector struct {
	Category *StringFilter `json:"category,omitempty"`
	Tags     *StringFilter `json:"tags,omitempty"`
	// TODO (Service): Add more filters
}

// GetSelectorType returns the type of the Selector.
func (ss ServiceSelector) GetSelectorType() FlavorTypeName {
	return ServiceNameDefault
}
