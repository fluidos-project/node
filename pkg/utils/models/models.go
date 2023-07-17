package models

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
	CPU               int    `json:"cpu,omitempty"`
	Memory            int    `json:"memory,omitempty"`
	PersistentStorage int    `json:"storage,omitempty"`
	EphemeralStorage  int    `json:"ephemeralStorage,omitempty"`
	GPU               int    `json:"gpu,omitempty"`
	Architecture      string `json:"architecture,omitempty"`
}

// Policy represents the policy associated with a Flavour, which can be either Partitionable or Aggregatable.
type Policy struct {
	Partitionable *Partitionable `json:"partitionable,omitempty"`
	Aggregatable  *Aggregatable  `json:"aggregatable,omitempty"`
}

// Partitionable represents the partitioning properties of a Flavour, such as the minimum and incremental values of CPU and RAM.
type Partitionable struct {
	CPUMinimum    int `json:"cpuMinimum"`
	MemoryMinimum int `json:"memoryMinimum"`
	CPUStep       int `json:"cpuStep"`
	MemoryStep    int `json:"memoryStep"`
}

// Aggregatable represents the aggregation properties of a Flavour, such as the minimum instance count.
type Aggregatable struct {
	MinCount int `json:"minCount"`
	MaxCount int `json:"maxCount"`
}

// Owner represents the owner of a Flavour, with associated ID, IP, and domain name.
type Owner struct {
	NodeID string `json:"ID"`
	IP     string `json:"IP"`
	Domain string `json:"domain"`
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
/* type Selector struct {
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
} */

type Selector struct {
	FlavourType   string         `json:"type,omitempty"`
	Architecture  string         `json:"architecture,omitempty"`
	RangeSelector *RangeSelector `json:"rangeSelector,omitempty"`
	MatchSelector *MatchSelector `json:"matchSelector,omitempty"`
}

// MatchSelector represents the criteria for selecting Flavours through a strict match.
type MatchSelector struct {
	Cpu              int `json:"cpu,omitempty"`
	Memory           int `json:"memory,omitempty"`
	Storage          int `json:"storage,omitempty"`
	EphemeralStorage int `json:"ephemeralStorage,omitempty"`
	Gpu              int `json:"gpu,omitempty"`
}

// RangeSelector represents the criteria for selecting Flavours through a range.
type RangeSelector struct {
	MoreThanCPU     int `json:"moreThanCPU,omitempty"`
	MoreThanMemory  int `json:"moreThanMemory,omitempty"`
	MoreThanEph     int `json:"moreThanEph,omitempty"`
	MoreThanStorage int `json:"moreThanStorage,omitempty"`
	MoreThanGpu     int `json:"moreThanGpu,omitempty"`
	LessThanCPU     int `json:"lessThanCPU,omitempty"`
	LessThanMemory  int `json:"lessThanMemory,omitempty"`
	LessThanEph     int `json:"lessThanEph,omitempty"`
	LessThanStorage int `json:"lessThanStorage,omitempty"`
	LessThanGpu     int `json:"lessThanGpu,omitempty"`
}

// Partition represents the partitioning properties of a Flavour
type Partition struct {
	Architecture     string `json:"architecture"`
	Cpu              int    `json:"cpu"`
	Memory           int    `json:"memory"`
	EphemeralStorage int    `json:"ephemeral-storage,omitempty"`
	Gpu              int    `json:"gpu,omitempty"`
	Storage          int    `json:"storage,omitempty"`
}

// Transaction contains information regarding the transaction for a flavour
type Transaction struct {
	TransactionID string    `json:"transactionID"`
	FlavourID     string    `json:"flavourID"`
	Partition     Partition `json:"partition"`
	Buyer         Owner     `json:"buyer"`
	StartTime     string    `json:"startTime,omitempty"`
}

// Contract represents a Contract object with its characteristics
type Contract struct {
	ContractID       string            `json:"contractID"`
	TransactionID    string            `json:"transactionID"`
	Flavour          Flavour           `json:"flavour"`
	Buyer            Owner             `json:"buyerID"`
	Seller           Owner             `json:"seller"`
	ExpirationTime   string            `json:"expirationTime,omitempty"`
	ExtraInformation map[string]string `json:"extraInformation,omitempty"`
	Partition        Partition         `json:"partition,omitempty"`
}

// Purchase contains information regarding the purchase for a flavour
type Purchase struct {
	Partition     Selector `json:"partition"`
	TransactionID string   `json:"transactionID"`
	FlavourID     string   `json:"flavourID"`
	BuyerID       string   `json:"buyerID"`
}

type PurchaseRequest struct {
	TransactionID string `json:"transactionID"`
	FlavourID     string `json:"flavourID"`
	BuyerID       string `json:"buyerID"`
}

// ResponsePurchase contain information after purchase a Flavour
type ResponsePurchase struct {
	Contract Contract `json:"contract"`
	Status   string   `json:"status"`
}

type ReserveRequest struct {
	FlavourID string    `json:"flavourID"`
	Buyer     Owner     `json:"buyerID"`
	Partition Partition `json:"partition"`
}

// AddFlavourToPcCache adds a new entry to the FlavourPeeringCandidateMap map if it does not exist
/* func AddFlavourToPcCache(flavourID string) {
	peeringCandidateID := namings.ForgePeeringCandidateName(flavourID)
	if _, ok := FlavourPeeringCandidateMap[peeringCandidateID]; !ok {
		FlavourPeeringCandidateMap[peeringCandidateID] = flavourID
	}
} */