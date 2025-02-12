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

// PurchaseRequest is the request model for purchasing a Flavor.
type PurchaseRequest struct {
	// LiqoCredentials contains the Liqo credentials of the buyer.
	// This field is optional and should be used only if the buyer is the Liqo peering target cluster, based on the Flavor you are going to purchase.
	// This field may be dismissed in the future version of Liqo.
	LiqoCredentials *LiqoCredentials `json:"liqoCredentials,omitempty"`
	// IngressTelemetryEndpoint is the endpoint where the buyer wants to receive the telemetry data.
	// This field is optional and should be used only if continuous telemetry is needed.
	IngressTelemetryEndpoint *TelemetryServer `json:"ingressTelemetryEndpoint,omitempty"`
}

// ReserveRequest is the request model for reserving a Flavor.
type ReserveRequest struct {
	FlavorID      string         `json:"flavorID"`
	Buyer         NodeIdentity   `json:"buyerID"`
	Configuration *Configuration `json:"configuration,omitempty"`
}
