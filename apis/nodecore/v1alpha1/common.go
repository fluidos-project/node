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

import "k8s.io/apimachinery/pkg/api/resource"

// Set of constants for the phases of the FLUIDOS Node modules.
const (
	PhaseSolved     Phase = "Solved"
	PhaseFailed     Phase = "Failed"
	PhaseRunning    Phase = "Running"
	PhaseAllocating Phase = "Allocating"
	PhaseIdle       Phase = "Idle"
	PhaseTimeout    Phase = "Timed Out"
	PhaseActive     Phase = "Active"
	PhasePending    Phase = "Pending"
	PhaseInactive   Phase = "Inactive"
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

// NodeIdentity is the identity of a FLUIDOS Node.
type NodeIdentity struct {
	Domain string `json:"domain"`
	NodeID string `json:"nodeID"`
	IP     string `json:"ip"`
}

// Partition is the partition of the flavour.
type Partition struct {
	Architecture     string            `json:"architecture"`
	CPU              resource.Quantity `json:"cpu"`
	Memory           resource.Quantity `json:"memory"`
	Pods             resource.Quantity `json:"pods"`
	Gpu              resource.Quantity `json:"gpu,omitempty"`
	EphemeralStorage resource.Quantity `json:"ephemeral-storage,omitempty"`
	Storage          resource.Quantity `json:"storage,omitempty"`
}

// LiqoCredentials contains the credentials of a Liqo cluster to enstablish a peering.
type LiqoCredentials struct {
	ClusterID   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
	Token       string `json:"token"`
	Endpoint    string `json:"endpoint"`
}
