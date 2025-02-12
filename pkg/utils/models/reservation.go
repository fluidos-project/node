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
)

// Configuration represents the configuration properties of a Flavor.
type Configuration struct {
	Type FlavorTypeName  `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ConfigurationData represents the data of a Configuration.
type ConfigurationData interface {
	GetConfigurationType() FlavorTypeName
}

// Transaction contains information regarding the transaction for a flavor.
type Transaction struct {
	TransactionID  string         `json:"transactionID"`
	FlavorID       string         `json:"flavorID"`
	Configuration  *Configuration `json:"configuration,omitempty"`
	Buyer          NodeIdentity   `json:"buyer"`
	ClusterID      string         `json:"clusterID"`
	ExpirationTime string         `json:"expirationTime"`
}

// TelemetryServer represents a TelemetryServer object with its characteristics.
type TelemetryServer struct {
	Endpoint string   `json:"endpoint"`
	Intents  []string `json:"intents"`
}

// Contract represents a Contract object with its characteristics.
type Contract struct {
	ContractID               string            `json:"contractID"`
	TransactionID            string            `json:"transactionID"`
	Flavor                   Flavor            `json:"flavor"`
	Buyer                    NodeIdentity      `json:"buyerID"`
	BuyerClusterID           string            `json:"buyerClusterID"`
	Seller                   NodeIdentity      `json:"seller"`
	PeeringTargetCredentials LiqoCredentials   `json:"peeringTargetCredentials"`
	ExpirationTime           string            `json:"expirationTime,omitempty"`
	ExtraInformation         map[string]string `json:"extraInformation,omitempty"`
	Configuration            *Configuration    `json:"configuration,omitempty"`
	NetworkRequests          string            `json:"networkAuthorizations,omitempty"`
	IngressTelemetryEndpoint *TelemetryServer  `json:"ingressTelemetryEndpoint,omitempty"`
}

// LiqoCredentials contains the credentials of a Liqo cluster to establish a peering.
type LiqoCredentials struct {
	ClusterID   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
	Token       string `json:"token"`
	Endpoint    string `json:"endpoint"`
}
