package models

import (
	"time"
)

// Flavour represents a Flavour object with its characteristics and policies.
type Flavour struct {
	FlavourID       string          `json:"flavourID"`
	ProviderID      string          `json:"providerID"`
	Type            string          `json:"type"`
	Characteristics Characteristics `json:"characteristics"`
	Policy          Policy          `json:"policy"`
	Owner           NodeIdentity    `json:"owner"`
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

// NodeIdentity represents the owner of a Flavour, with associated ID, IP, and domain name.
type NodeIdentity struct {
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
	MinCpu     int `json:"minCpu,omitempty"`
	MinMemory  int `json:"minMemory,omitempty"`
	MinStorage int `json:"minStorage,omitempty"`
	MinEph     int `json:"minEph,omitempty"`
	MinGpu     int `json:"minGpu,omitempty"`
	MaxCpu     int `json:"maxCpu,omitempty"`
	MaxMemory  int `json:"maxMemory,omitempty"`
	MaxStorage int `json:"maxStorage,omitempty"`
	MaxEph     int `json:"maxEph,omitempty"`
	MaxGpu     int `json:"maxGpu,omitempty"`
}
