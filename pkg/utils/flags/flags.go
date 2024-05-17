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

package flags

import "time"

// Namespace flags.
var (
	FluidosNamespace = "fluidos"
)

// Expiration/Time flags.
var (
	ExpirationPhaseRunning = 2 * time.Minute
	ExpirationSolver       = 5 * time.Minute
	ExpirationTransaction  = 20 * time.Second
	ExpirationContract     = 365 * 24 * time.Hour
	RefreshCacheInterval   = 20 * time.Second
	LiqoCheckInterval      = 20 * time.Second
)

// Configs flags.
var (
	HTTPPort          string
	GRPCPort          string
	ResourceNodeLabel string
)

// Customization flags.
var (
	ResourceType string
	AMOUNT       string
	CURRENCY     string
	PERIOD       string
	CPUMin       string
	MemoryMin    string
	PodsMin      string
	CPUStep      string
	MemoryStep   string
	PodsStep     string
	MinCount     int64
	MaxCount     int64
)
