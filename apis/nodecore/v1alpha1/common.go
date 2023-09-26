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

package v1alpha1

const (
	//PhaseReady   Phase = "Ready"
	PhaseSolved   Phase = "Solved"
	PhaseFailed   Phase = "Failed"
	PhaseRunning  Phase = "Running"
	PhaseIdle     Phase = "Idle"
	PhaseTimeout  Phase = "Timed Out"
	PhaseBackoff  Phase = "Backoff"
	PhaseActive   Phase = "Active"
	PhasePending  Phase = "Pending"
	PhaseInactive Phase = "Inactive"
)

// GenericRef represents a reference to a generic Kubernetes resource,
// and it is composed of the resource name and (optionally) its namespace.
type GenericRef struct {
	// The name of the resource to be referenced.
	Name string `json:"name,omitempty"`

	// The namespace containing the resource to be referenced. It should be left
	// empty in case of cluster-wide resources.
	Namespace string `json:"namespace,omitempty"`
}

type NodeIdentity struct {
	Domain string `json:"domain"`
	NodeID string `json:"nodeID"`
	IP     string `json:"ip"`
}

// toString() returns a string representation of the GenericRef.
func (r GenericRef) toString() string {
	if r.Namespace != "" {
		return r.Namespace + "/" + r.Name
	}
	return r.Name
}
