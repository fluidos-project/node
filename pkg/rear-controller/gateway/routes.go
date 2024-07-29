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

package gateway

// Routes defines the routes for the rear controller.
var Routes = struct {
	// Flavors is the route to get all the flavors.
	Flavors string
	// K8SliceFlavors is the route to get all the K8Slice flavors.
	K8SliceFlavors string
	// VMFlavors is the route to get all the VM flavors.
	VMFlavors string
	// ServiceFlavors is the route to get all the service flavors.
	ServiceFlavors string
	// SensorsFlavors is the route to get all the sensors flavors.
	SensorsFlavors string
	// Reserve is the route to reserve a flavor.
	Reserve string
	// Purchase is the route to purchase a flavor.
	Purchase string
}{
	Flavors:        "/api/v2/flavors",
	K8SliceFlavors: "/api/v2/flavors/k8slice",
	VMFlavors:      "/api/v2/flavors/vm",
	ServiceFlavors: "/api/v2/flavors/service",
	SensorsFlavors: "/api/v2/flavors/sensors",
	Reserve:        "/api/v2/reservations",
	Purchase:       "/api/v2/transactions/{transactionID}/purchase",
}
