package models

import (
	"time"

	"fluidos.eu/node/pkg/utils/namings"
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
	CPU              int    `json:"cpu,omitempty"`
	RAM              int    `json:"ram,omitempty"`
	EphemeralStorage int    `json:"ephemeralStorage,omitempty"`
	Architecture     string `json:"architecture,omitempty"`
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
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Period   string `json:"period"`
}

// OptionalFields represents the optional fields of a Flavour, such as availability.
type OptionalFields struct {
	Availability bool   `json:"availability,omitempty"`
	WorkerID     string `json:"workerID"`
}

// Selector represents the criteria for selecting Flavours.
type Selector struct {
	FlavourType      string `json:"type,omitempty"`
	Architecture     string `json:"architecture,omitempty"`
	Cpu              int    `json:"cpu,omitempty"`
	Memory           int    `json:"memory,omitempty"`
	EphemeralStorage int    `json:"ephemeral-storage,omitempty"`
	MoreThanCpu      int    `json:"moreThanCpu,omitempty"`
	MoreThanMemory   int    `json:"moreThanMemory,omitempty"`
	MoreThanEph      int    `json:"moreThanEph,omitempty"`
	LessThanCpu      int    `json:"lessThanCpu,omitempty"`
	LessThanMemory   int    `json:"lessThanMemory,omitempty"`
	LessThanEph      int    `json:"lessThanEph,omitempty"`
}

// Partition represents the partitioning properties of a Flavour
type Partition struct {
	Architecture     string `json:"architecture,omitempty"`
	Cpu              int    `json:"cpu,omitempty"`
	Memory           int    `json:"memory,omitempty"`
	EphemeralStorage int    `json:"ephemeral-storage,omitempty"`
}

// Transaction contains information regarding the transaction for a flavour
type Transaction struct {
	TransactionID string `json:"transactionID"`
	FlavourID     string `json:"flavourID"`
	Buyer         Owner  `json:"buyer"`
	StartTime     string `json:"startTime,omitempty"`
}

// Contract represents a Contract object with its characteristics
type Contract struct {
	ContractID       string            `json:"contractID"`
	TransactionID    string            `json:"transactionID"`
	Flavour          Flavour           `json:"flavour"`
	BuyerID          string            `json:"buyerID"`
	Seller           Owner             `json:"seller"`
	ExpirationTime   string            `json:"expirationTime,omitempty"`
	ExtraInformation map[string]string `json:"extraInformation,omitempty"`
	Partition        Partition         `json:"partition"`
}

// Purchase contains information regarding the purchase for a flavour
type Purchase struct {
	Partition     Partition `json:"partition"`
	TransactionID string    `json:"transactionID"`
	FlavourID     string    `json:"flavourID"`
	BuyerID       string    `json:"buyerID"`
}

// ResponsePurchase contain information after purchase a Flavour
type ResponsePurchase struct {
	Contract Contract `json:"contract"`
	Status   string   `json:"status"`
}

// TransactionCache is the current transaction used in the Reservation controller
var TransactionCache Transaction

// Transactions is a map of Transaction
var Transactions map[string]Transaction

// FlavourPeeringCandidateMap is a map of Flavour and PeeringCandidate
var FlavourPeeringCandidateMap = make(map[string]string)

// AddFlavourToPcCache adds a new entry to the FlavourPeeringCandidateMap map if it does not exist
func AddFlavourToPcCache(flavourID string) {
	peeringCandidateID := namings.ForgePeeringCandidateName(flavourID)
	if _, ok := FlavourPeeringCandidateMap[peeringCandidateID]; !ok {
		FlavourPeeringCandidateMap[peeringCandidateID] = flavourID
	}
}

// fromNodeInfo creates from params a new NodeInfo Struct
func FromNodeInfo(uid, name, arch, os string, metrics ResourceMetrics) *NodeInfo {
	return &NodeInfo{
		UID:             uid,
		Name:            name,
		Architecture:    arch,
		OperatingSystem: os,
		ResourceMetrics: metrics,
	}
}

// fromResourceMetrics creates from params a new ResourceMetrics Struct
func FromResourceMetrics(cpuTotal, cpuUsed, memoryTotal, memoryUsed, ephStorage resource.Quantity) *ResourceMetrics {
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
func CompareByCPUAvailable(node1, node2 NodeInfo) bool {
	cmpResult := node1.ResourceMetrics.CPUAvailable.Cmp(node2.ResourceMetrics.CPUAvailable)
	if cmpResult == 1 {
		return true
	} else {
		return false
	}
}

// maxNode find the node with the maximum value based on the provided comparison function
func MaxNode(nodes []NodeInfo, compareFunc func(NodeInfo, NodeInfo) bool) NodeInfo {
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
