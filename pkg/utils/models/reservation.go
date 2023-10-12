// Copyright 2022-2023 FLUIDOS Project
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

// Partition represents the partitioning properties of a Flavour
type Partition struct {
	Architecture     string `json:"architecture"`
	Cpu              int    `json:"cpu"`
	Memory           int    `json:"memory"`
	EphemeralStorage int    `json:"ephemeral-storage,omitempty"`
	Gpu              int    `json:"gpu,omitempty"`
	Storage          int    `json:"storage,omitempty"`
}

// Transaction contains information regarding the transaction for a flavour
type Transaction struct {
	TransactionID string       `json:"transactionID"`
	FlavourID     string       `json:"flavourID"`
	Partition     Partition    `json:"partition"`
	Buyer         NodeIdentity `json:"buyer"`
	ClusterID     string       `json:"clusterID"`
	StartTime     string       `json:"startTime"`
}

// Contract represents a Contract object with its characteristics
type Contract struct {
	ContractID        string            `json:"contractID"`
	TransactionID     string            `json:"transactionID"`
	Flavour           Flavour           `json:"flavour"`
	Buyer             NodeIdentity      `json:"buyerID"`
	BuyerClusterID    string            `json:"buyerClusterID"`
	Seller            NodeIdentity      `json:"seller"`
	SellerCredentials LiqoCredentials   `json:"sellerCredentials"`
	ExpirationTime    string            `json:"expirationTime,omitempty"`
	ExtraInformation  map[string]string `json:"extraInformation,omitempty"`
	Partition         Partition         `json:"partition,omitempty"`
}

// LiqoCredentials contains the credentials of a Liqo cluster to enstablish a peering.
type LiqoCredentials struct {
	ClusterID   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
	Token       string `json:"token"`
	Endpoint    string `json:"endpoint"`
}
