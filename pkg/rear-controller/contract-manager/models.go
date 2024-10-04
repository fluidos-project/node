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

package contractmanager

import "time"

// Selector represents the criteria for selecting Flavours.
type Selector struct {
	FlavourType      string `json:"type,omitempty"`
	CPU              int    `json:"cpu,omitempty"`
	RAM              int    `json:"ram,omitempty"`
	EphemeralStorage int    `json:"ephemeral-storage,omitempty"`
}

// Transaction contains information regarding the transaction for a flavor.
type Transaction struct {
	TransactionID  string    `json:"transactionID"`
	FlavorID       string    `json:"flavorID"`
	ExpirationTime time.Time `json:"startTime,omitempty"`
}

// Purchase contains information regarding the purchase for a flavor.
type Purchase struct {
	TransactionID string `json:"transactionID"`
	FlavorID      string `json:"flavorID"`
	BuyerID       string `json:"buyerID"`
}
