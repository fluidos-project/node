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
	FLUIDOS_NAMESPACE string = "fluidos"
)

// EXPIRATION flags
var (
	EXPIRATION_PHASE_RUNNING = 2 * time.Minute
	EXPIRATION_SOLVER        = 5 * time.Minute
	EXPIRATION_TRANSACTION   = 20 * time.Second
	EXPIRATION_CONTRACT      = 365 * 24 * time.Hour
	REFRESH_CACHE_INTERVAL   = 20 * time.Second
	LIQO_CHECK_INTERVAL      = 20 * time.Second
)

var (
	HTTP_PORT           string
	GRPC_PORT           string
	RESOURCE_NODE_LABEL string
)

var (
	RESOURCE_TYPE string
	AMOUNT        string
	CURRENCY      string
	PERIOD        string
	CPU_MIN       string
	MEMORY_MIN    string
	CPU_STEP      string
	MEMORY_STEP   string
	MIN_COUNT     int64
	MAX_COUNT     int64
)
