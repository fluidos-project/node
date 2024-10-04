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

import "k8s.io/apimachinery/pkg/api/resource"

// K8SliceConfiguration is the partition of the flavor K8Slice.
type K8SliceConfiguration struct {
	// CPU is the CPU of the K8Slice partition.
	CPU resource.Quantity `json:"cpu"`
	// Memory is the Memory of the K8Slice partition.
	Memory resource.Quantity `json:"memory"`
	// Pods is the Pods of the K8Slice partition.
	Pods resource.Quantity `json:"pods"`
	// Gpu is the GPU of the K8Slice partition.
	Gpu *GPU `json:"gpu,omitempty"`
	// Storage is the Storage of the K8Slice partition.
	Storage *resource.Quantity `json:"storage,omitempty"`
}
