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

package flags

import "time"

// NAMESPACES flags
var (
	DEFAULT_NAMESPACE             string = "default"
	PC_DEFAULT_NAMESPACE          string = "default"
	SOLVER_DEFAULT_NAMESPACE      string = "default"
	DISCOVERY_DEFAULT_NAMESPACE   string = "default"
	FLAVOUR_DEFAULT_NAMESPACE     string = "default"
	CONTRACT_DEFAULT_NAMESPACE    string = "default"
	RESERVATION_DEFAULT_NAMESPACE string = "default"
	TRANSACTION_DEFAULT_NAMESPACE string = "default"
)

// EXPIRATION flags
var (
	EXPIRATION_PHASE_RUNNING = 2 * time.Minute
	EXPIRATION_SOLVER        = 5 * time.Minute
	EXPIRATION_TRANSACTION   = 20 * time.Second
	EXPIRATION_CONTRACT      = 365 * 24 * time.Hour
	REFRESH_CACHE_INTERVAL   = 20 * time.Second
)

// TODO: TO BE REVIEWED
var (
	// THESE SHOULD BE THE NODE IDENTITY OF THE FLUIDOS NODE
	CLIENT_ID string
	DOMAIN    string
	IP_ADDR   string
	// THIS SHOULD BE PROVIDED BY THE NETWORK MANAGER
	SERVER_ADDR      string
	SERVER_ADDRESSES = []string{SERVER_ADDR}
	// REAR Gateway http port
	HTTP_PORT string
	// REAR Gateway grpc address
	GRPC_PORT string
)

var (
	RESOURCE_TYPE string
	AMOUNT        string
	CURRENCY      string
	PERIOD        string
	CPU_MIN       int64
	MEMORY_MIN    int64
	CPU_STEP      int64
	MEMORY_STEP   int64
	MIN_COUNT     int64
	MAX_COUNT     int64
)
