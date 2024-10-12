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

	"github.com/xeipuuv/gojsonschema"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

// ServiceConfiguration represents the configuration of a Service.
type ServiceConfiguration struct {
	// HostingPolicy is the hosting policy chosen for the service, where the service will be hosted.
	HostingPolicy *HostingPolicy `json:"hostingPolicy,omitempty"`
	// ConfigurationData is the data of the specific service configuration, compliant with the ConfigurationTemplate defined in the Service Flavor.
	ConfigurationData runtime.RawExtension `json:"configurationData"`
}

// Validate validates the ServiceConfiguration over the given ServiceFlavor,
// in particular checking that the ConfigurationData is compliant with the ConfigurationTemplate defined in the Service Flavor.
func (sc *ServiceConfiguration) Validate(serviceFlavor *ServiceFlavor) error {
	configTemplate := serviceFlavor.ConfigurationTemplate
	// Parse the ConfigurationTemplate into a string
	configTemplateString := string(configTemplate.Raw)
	klog.Infof("ConfigurationTemplate: %v", configTemplateString)
	// Create JSON Schema from the ConfigurationTemplate
	loader := gojsonschema.NewStringLoader(configTemplateString)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		klog.Errorf("Error creating JSON Schema from ConfigurationTemplate: %v", err)
		return err
	}

	// Parse the ConfigurationData into a string
	configDataString := string(sc.ConfigurationData.Raw)
	klog.Infof("ConfigurationData: %v", configDataString)
	// Compare the ConfigurationData with the ConfigurationTemplate
	dataLoader := gojsonschema.NewStringLoader(configDataString)
	result, err := schema.Validate(dataLoader)
	if err != nil {
		klog.Errorf("Error validating ConfigurationData with ConfigurationTemplate: %v", err)
		return err
	}
	klog.Infof("ConfigurationData is valid: %v", result.Valid())
	if !result.Valid() {
		for _, desc := range result.Errors() {
			klog.Errorf("- %s\n", desc)
		}
		return fmt.Errorf("ConfigurationData is not valid")
	}

	return nil
}
