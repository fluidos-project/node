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

import (
	"encoding/json"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
)

// GpuCharacteristics represents the characteristics of a Gpu.
type GpuCharacteristics struct {
	Model  string            `json:"model"`
	Cores  resource.Quantity `json:"cores"`
	Memory resource.Quantity `json:"memory"`
}

// Cmp compares models.GpuCharacteristics with nodecorev1alpha1.GPU.
func (gpu *GpuCharacteristics) Cmp(other *nodecorev1alpha1.GPU) int {
	if gpu.Model != other.Model {
		return strings.Compare(gpu.Model, other.Model)
	}
	if cmp := gpu.Cores.Cmp(other.Cores); cmp != 0 {
		return cmp
	}
	return gpu.Memory.Cmp(other.Memory)
}

// K8SliceCharacteristics represents the characteristics of a Kubernetes slice.
type K8SliceCharacteristics struct {
	Architecture string              `json:"architecture"`
	CPU          resource.Quantity   `json:"cpu"`
	Memory       resource.Quantity   `json:"memory"`
	Pods         resource.Quantity   `json:"pods"`
	Gpu          *GpuCharacteristics `json:"gpu,omitempty"`
	Storage      *resource.Quantity  `json:"storage,omitempty"`
}

// K8SliceProperties represents the properties of a Kubernetes slice.
type K8SliceProperties struct {
	Latency               int                        `json:"latency,omitempty"`
	SecurityStandards     []string                   `json:"securityStandards,omitempty"`
	CarbonFootprint       *CarbonFootprint           `json:"carbonFootprint,omitempty"`
	NetworkAuthorizations *NetworkAuthorizations     `json:"networkAuthorizations,omitempty"`
	AdditionalProperties  map[string]json.RawMessage `json:"additionalProperties,omitempty"`
}

// K8SlicePartitionability represents the partitionability of a Kubernetes slice.
type K8SlicePartitionability struct {
	CPUMin     resource.Quantity `json:"cpuMin"`
	MemoryMin  resource.Quantity `json:"memoryMin"`
	PodsMin    resource.Quantity `json:"podsMin"`
	CPUStep    resource.Quantity `json:"cpuStep"`
	MemoryStep resource.Quantity `json:"memoryStep"`
	PodsStep   resource.Quantity `json:"podsStep"`
}

// K8SlicePolicies represents the policies of a Kubernetes slice.
type K8SlicePolicies struct {
	Partitionability K8SlicePartitionability `json:"partitionability"`
}

// K8Slice represents a Kubernetes slice.
type K8Slice struct {
	Characteristics K8SliceCharacteristics `json:"charateristics"`
	Properties      K8SliceProperties      `json:"properties"`
	Policies        K8SlicePolicies        `json:"policies"`
}

// GetFlavorTypeName returns the type of the Flavor.
func (K8Slice) GetFlavorTypeName() FlavorTypeName {
	return K8SliceNameDefault
}

// K8SliceSelector represents the criteria for selecting a K8Slice Flavor.
type K8SliceSelector struct {
	Architecture *StringFilter           `json:"architecture,omitempty"`
	CPU          *ResourceQuantityFilter `scheme:"cpu,omitempty"`
	Memory       *ResourceQuantityFilter `scheme:"memory,omitempty"`
	Pods         *ResourceQuantityFilter `scheme:"pods,omitempty"`
	Storage      *ResourceQuantityFilter `scheme:"storage,omitempty"`
}

// GetSelectorType returns the type of the Selector.
func (ks K8SliceSelector) GetSelectorType() FlavorTypeName {
	return K8SliceNameDefault
}
