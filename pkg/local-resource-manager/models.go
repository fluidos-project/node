package localResourceManager

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// fromNodeInfo creates from params a new NodeInfo Struct
func fromNodeInfo(uid, name, arch, os string, metrics ResourceMetrics) *NodeInfo {
	return &NodeInfo{
		UID:             uid,
		Name:            name,
		Architecture:    arch,
		OperatingSystem: os,
		ResourceMetrics: metrics,
	}
}

// fromResourceMetrics creates from params a new ResourceMetrics Struct
func fromResourceMetrics(cpuTotal, cpuUsed, memoryTotal, memoryUsed, ephStorage resource.Quantity) *ResourceMetrics {
	cpuAvail := cpuTotal.DeepCopy()
	memAvail := memoryTotal.DeepCopy()

	cpuAvail.Sub(cpuUsed)
	memAvail.Sub(memoryUsed)

	return &ResourceMetrics{
		CPUTotal:         cpuTotal,
		CPUAvailable:     cpuAvail,
		MemoryTotal:      memoryTotal,
		MemoryAvailable:  memAvail,
		EphemeralStorage: ephStorage,
	}
}

// compareByCPUAvailable is a custom comparison function based on CPUAvailable
func compareByCPUAvailable(node1, node2 NodeInfo) bool {
	cmpResult := node1.ResourceMetrics.CPUAvailable.Cmp(node2.ResourceMetrics.CPUAvailable)
	if cmpResult == 1 {
		return true
	} else {
		return false
	}
}

// maxNode find the node with the maximum value based on the provided comparison function
func maxNode(nodes []NodeInfo, compareFunc func(NodeInfo, NodeInfo) bool) NodeInfo {
	if len(nodes) == 0 {
		panic("Empty node list")
	}

	maxNode := nodes[0]
	for _, node := range nodes {
		if compareFunc(node, maxNode) {
			maxNode = node
		}
	}
	return maxNode
}

// Flavour represents a Flavour object with its characteristics and policies.
type Flavour struct {
	FlavourID       string          `json:"flavourID"`
	ProviderID      string          `json:"providerID"`
	Type            string          `json:"type"`
	Characteristics Characteristics `json:"characteristics"`
	Policy          Policy          `json:"policy"`
	Owner           Owner           `json:"owner"`
	Price           Price           `json:"price"`
	ExpirationTime  time.Time       `json:"expirationTime"`
	OptionalFields  OptionalFields  `json:"optionalFields"`
}

// Characteristics represents the characteristics of a Flavour, such as CPU and RAM.
type Characteristics struct {
	CPU int `json:"cpu"`
	RAM int `json:"ram"`
}

// Policy represents the policy associated with a Flavour, which can be either Partitionable or Aggregatable.
type Policy struct {
	Partitionable *Partitionable `json:"partitionable,omitempty"`
	Aggregatable  *Aggregatable  `json:"aggregatable,omitempty"`
}

// Partitionable represents the partitioning properties of a Flavour, such as the minimum and incremental values of CPU and RAM.
type Partitionable struct {
	CPUMinimum int `json:"cpuMinimum"`
	RAMMinimum int `json:"ramMinimum"`
	CPUStep    int `json:"cpuStep"`
	RAMStep    int `json:"ramStep"`
}

// Aggregatable represents the aggregation properties of a Flavour, such as the minimum instance count.
type Aggregatable struct {
	MinCount int `json:"minCount"`
	MaxCount int `json:"maxCount"`
}

// Owner represents the owner of a Flavour, with associated ID, IP, and domain name.
type Owner struct {
	ID         string `json:"ID"`
	IP         string `json:"IP"`
	DomainName string `json:"domainName"`
}

// Price represents the price of a Flavour, with the amount, currency, and period associated.
type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Period   string  `json:"period"`
}

// OptionalFields represents the optional fields of a Flavour, such as availability.
type OptionalFields struct {
	Availability bool `json:"availability,omitempty"`
}
