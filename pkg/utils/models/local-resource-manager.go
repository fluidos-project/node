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
	"k8s.io/apimachinery/pkg/api/resource"
)

// NodeInfo represents a node and its resources.
type NodeInfo struct {
	UID             string          `json:"uid"`
	Name            string          `json:"name"`
	Architecture    string          `json:"architecture"`
	OperatingSystem string          `json:"os"`
	ResourceMetrics ResourceMetrics `json:"resources"`
}

// GPUMetrics represents GPU metrics.
type GPUMetrics struct {
	Model           string            `json:"model"`
	CoresTotal      resource.Quantity `json:"totalCores"`
	CoresAvailable  resource.Quantity `json:"availableCores"`
	MemoryTotal     resource.Quantity `json:"totalMemory"`
	MemoryAvailable resource.Quantity `json:"availableMemory"`
}

// ResourceMetrics represents resources of a certain node.
type ResourceMetrics struct {
	CPUTotal         resource.Quantity `json:"totalCPU"`
	CPUAvailable     resource.Quantity `json:"availableCPU"`
	MemoryTotal      resource.Quantity `json:"totalMemory"`
	MemoryAvailable  resource.Quantity `json:"availableMemory"`
	PodsTotal        resource.Quantity `json:"totalPods"`
	PodsAvailable    resource.Quantity `json:"availablePods"`
	EphemeralStorage resource.Quantity `json:"ephemeralStorage"`
	GPU              GPUMetrics        `json:"gpu"`
}
