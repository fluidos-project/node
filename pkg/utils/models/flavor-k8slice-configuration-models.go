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

import "k8s.io/apimachinery/pkg/api/resource"

// K8SliceConfiguration represents the configuration properties of a K8Slice Flavor.
type K8SliceConfiguration struct {
	CPU     resource.Quantity   `json:"cpu"`
	Memory  resource.Quantity   `json:"memory"`
	Pods    resource.Quantity   `json:"pods"`
	Gpu     *GpuCharacteristics `json:"gpu,omitempty"`
	Storage *resource.Quantity  `json:"storage,omitempty"`
}

// GetConfigurationType returns the type of the Configuration.
func (p *K8SliceConfiguration) GetConfigurationType() FlavorTypeName {
	return K8SliceNameDefault
}
