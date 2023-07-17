package contractmanager

import "time"

// Selector represents the criteria for selecting Flavours.
type Selector struct {
	FlavourType      string `json:"type,omitempty"`
	CPU              int    `json:"cpu,omitempty"`
	RAM              int    `json:"ram,omitempty"`
	EphemeralStorage int    `json:"ephemeral-storage,omitempty"`
}

// Transaction contains information regarding the transaction for a flavour
type Transaction struct {
	TransactionID string    `json:"transactionID"`
	FlavourID     string    `json:"flavourID"`
	StartTime     time.Time `json:"startTime,omitempty"`
}

// Purchase contains information regarding the purchase for a flavour
type Purchase struct {
	TransactionID string `json:"transactionID"`
	FlavourID     string `json:"flavourID"`
	BuyerID       string `json:"buyerID"`
}

// ResponsePurchase contain information after purchase a Flavour
type ResponsePurchase struct {
	FlavourID string `json:"flavourID"`
	BuyerID   string `json:"buyerID"`
	Status    string `json:"status"`
}

// transactions is a map of Transaction
var transactions map[string]Transaction
