package models

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// NodeInfo represents a node and its resources
type NodeInfo struct {
	UID             string          `json:"uid"`
	Name            string          `json:"name"`
	Architecture    string          `json:"architecture"`
	OperatingSystem string          `json:"os"`
	ResourceMetrics ResourceMetrics `json:"resources"`
}

// ResourceMetrics represents resources of a certain node
type ResourceMetrics struct {
	CPUTotal         resource.Quantity `json:"totalCPU"`
	CPUAvailable     resource.Quantity `json:"availableCPU"`
	MemoryTotal      resource.Quantity `json:"totalMemory"`
	MemoryAvailable  resource.Quantity `json:"availableMemory"`
	EphemeralStorage resource.Quantity `json:"ephemeralStorage"`
}
