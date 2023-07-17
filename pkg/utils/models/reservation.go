package models

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
