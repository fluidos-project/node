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

// CarbonFootprint represents the carbon footprint of a Flavor.
type CarbonFootprint struct {
	Embodied    int   `json:"embodied"`
	Operational []int `json:"operational"`
}

// Properties represents the properties of a Flavor.
type Properties struct {
	// Latency to reach the K8Slice Flavor
	Latency int `json:"latency,omitempty"`
	// Security standards complied by the K8Slice Flavor
	SecurityStandards []string `json:"securityStandards,omitempty"`
	// Carbon footprint of the K8Slice Flavor
	CarbonFootprint *CarbonFootprint `json:"carbon-footprint,omitempty"`
}
