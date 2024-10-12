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
	"bytes"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// ServiceBlueprintSpec defines the desired state of ServiceBlueprint.
type ServiceBlueprintSpec struct {
	// Name of the Service Blueprint.
	Name string `json:"name"`

	// Description of the Service Blueprint.
	Description string `json:"description"`

	// Category of the Service Blueprint.
	Category string `json:"category"`

	// Tags of the Service Blueprint.
	Tags []string `json:"tags"`

	// HostingPolicies of the Service Blueprint.
	// If empty, the default behavior is to host on the provider cluster.
	// If multiple policies are specified, the first one is the default.
	HostingPolicies []HostingPolicy `json:"hostingPolicies"`

	// Templates of the Service Blueprint.
	Templates []ServiceTemplate `json:"templates"`
}

// ServiceTemplate defines the template of a Service.
type ServiceTemplate struct {
	// Name of the Service Template.
	Name string `json:"name"`
	// Description of the Service Template.
	Description string `json:"description,omitempty"`
	// YAML template of the Service.
	ServiceTemplateData runtime.RawExtension `json:"serviceData"`
}

// ServiceBlueprintStatus defines the observed state of ServiceBlueprint.
type ServiceBlueprintStatus struct {
	// ServiceFlavor linked to the Service Blueprint.
	ServiceFlavors []ServiceFlavor `json:"serviceFlavors,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ServiceBlueprint is the Schema for the serviceblueprints API
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.spec.description`
// +kubebuilder:printcolumn:name="Category",type=string,JSONPath=`.spec.category`
// +kubebuilder:printcolumn:name="Tags",type=string,JSONPath=`.spec.tags`
// +kubebuilder:printcolumn:name="ServiceFlavors",type=string,JSONPath=`.status.serviceFlavors`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=sbp
type ServiceBlueprint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBlueprintSpec   `json:"spec,omitempty"`
	Status ServiceBlueprintStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ServiceBlueprintList contains a list of ServiceBlueprint.
type ServiceBlueprintList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceBlueprint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceBlueprint{}, &ServiceBlueprintList{})
}

// ValidateAndExtractManifests extracts manifests from ServiceTemplateData and validates them.
func ValidateAndExtractManifests(templates []ServiceTemplate) ([]*unstructured.Unstructured, error) {
	var manifests []*unstructured.Unstructured

	for _, template := range templates {
		// Step 1: Parse the raw data from ServiceTemplateData (YAML or JSON)
		parsedManifests, err := parseRawServiceTemplateData(template.ServiceTemplateData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse service template data for template %s: %w", template.Name, err)
		}

		// Step 2: Validate the manifests
		for _, manifest := range parsedManifests {
			if err := validateManifest(manifest); err != nil {
				return nil, fmt.Errorf("validation failed for template %s: %w", template.Name, err)
			}
			manifests = append(manifests, manifest)
		}
	}

	return manifests, nil
}

// parseRawServiceTemplateData parses the raw YAML/JSON data from ServiceTemplateData.
func parseRawServiceTemplateData(data runtime.RawExtension) ([]*unstructured.Unstructured, error) {
	var manifests []*unstructured.Unstructured

	// Decoding YAML to JSON
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	// Split multiple documents if necessary (YAML can have multiple docs separated by `---`)
	docs := bytes.Split(data.Raw, []byte("\n---\n"))

	for _, doc := range docs {
		if len(bytes.TrimSpace(doc)) == 0 {
			continue // Skip empty docs
		}

		// Parse into Unstructured to handle dynamic manifest types
		obj := &unstructured.Unstructured{}
		_, _, err := decoder.Decode(doc, nil, obj)
		if err != nil {
			return nil, fmt.Errorf("failed to decode manifest: %w", err)
		}

		manifests = append(manifests, obj)
	}

	return manifests, nil
}

// validateManifest performs basic validation on the Kubernetes manifest.
func validateManifest(manifest *unstructured.Unstructured) error {
	// Basic validation for required fields in Kubernetes objects
	if manifest.GetKind() == "" || manifest.GetAPIVersion() == "" {
		return fmt.Errorf("manifest missing kind or apiVersion")
	}

	// Add more specific validation logic if necessary

	return nil
}
