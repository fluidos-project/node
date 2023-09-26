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

package discoverymanager

const (
	SERVER_ADDR       = "http://localhost:14144/api"
	K8S_TYPE          = "k8s-fluidos"
	DEFAULT_NAMESPACE = "default"
	CLIENT_ID         = "topix.fluidos.eu"
)

// We define different server addresses, much more dynamic and less hardcoded
var (
	SERVER_ADDRESSES = []string{"http://localhost:14144/api"}
)
