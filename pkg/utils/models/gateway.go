package models

// PurchaseRequest is the request model for purchasing a Flavour
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

// ReserveRequest is the request model for reserving a Flavour
type ReserveRequest struct {
	FlavourID string    `json:"flavourID"`
	Buyer     Owner     `json:"buyerID"`
	Partition Partition `json:"partition"`
}
