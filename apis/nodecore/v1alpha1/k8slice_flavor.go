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
	"k8s.io/apimachinery/pkg/api/resource"
)

// K8Slice represents a K8Slice Flavor.
type K8Slice struct {
	// Characteristics of the K8Slice Flavor
	Characteristics K8SliceCharacteristics `json:"characteristics"`
	// Properties of the K8Slice Flavor
	Properties Properties `json:"properties"`
	// Policies of the K8Slice Flavor
	Policies Policies `json:"policies"`
}

// GetFlavorType returns the type of the Flavor.
func (k8s *K8Slice) GetFlavorType() FlavorTypeIdentifier {
	return TypeK8Slice
}

// K8SliceCharacteristics represents the characteristics of a K8Slice Flavor, such as the CPU, RAM, and storage.
type K8SliceCharacteristics struct {
	// CPU is the number of CPU cores of the K8Slice Flavor.
	CPU resource.Quantity `json:"cpu"`
	// Memory is the amount of RAM of the K8Slice Flavor.
	Memory resource.Quantity `json:"memory"`
	// Pods is the maximum number of pods schedulable on this K8Slice Flavor.
	Pods resource.Quantity `json:"pods"`
	// GPU is the number of GPU cores of the K8Slice Flavor.
	Gpu *GPU `json:"gpu,omitempty"`
	// Storage is the amount of storage offered by this K8Slice Flavor.
	Storage *resource.Quantity `json:"storage,omitempty"`
}

// GPU represents the GPU characteristics of a K8Slice Flavor.
type GPU struct {
	// Model of the GPU
	Model string `json:"model"`
	// Number of GPU cores
	Cores resource.Quantity `json:"cores"`
	// Memory of the GPU
	Memory resource.Quantity `json:"memory"`
}

// Policies represents the policies of a K8Slice Flavor, such as the partitionability of the K8Slice Flavor.
type Policies struct {
	// Partitionability of the K8Slice Flavor
	Partitionability Partitionability `json:"partitionability,omitempty"`
}

// Partitionability represents the partitioning properties of a K8Slice Flavor, such as the minimum and incremental values of CPU and RAM.
type Partitionability struct {
	// CPUMin is the minimum number of CPU cores in which the K8Slice Flavor can be partitioned.
	CPUMin resource.Quantity `json:"cpuMin"`
	// MemoryMin is the minimum amount of RAM in which the K8Slice Flavor can be partitioned.
	MemoryMin resource.Quantity `json:"memoryMin"`
	// PodsMin is the minimum number of pods in which the K8Slice Flavor can be partitioned.
	PodsMin resource.Quantity `json:"podsMin"`
	// GpuMin is the minimum number of GPU cores in which the K8Slice Flavor can be partitioned.
	GpuMin resource.Quantity `json:"gpuMin,omitempty"`
	// CPUStep is the incremental value of CPU cores in which the K8Slice Flavor can be partitioned.
	CPUStep resource.Quantity `json:"cpuStep"`
	// MemoryStep is the incremental value of RAM in which the K8Slice Flavor can be partitioned.
	MemoryStep resource.Quantity `json:"memoryStep"`
	// PodsStep is the incremental value of pods in which the K8Slice Flavor can be partitioned.
	PodsStep resource.Quantity `json:"podsStep"`
	// GpuStep is the incremental value of GPU cores in which the K8Slice Flavor can be partitioned.
	GpuStep resource.Quantity `json:"gpuStep,omitempty"`
}
