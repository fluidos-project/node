package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"

	"fluidos.eu/node/pkg/utils/models"
)

// buildSelector builds a selector from a request body
func buildSelector(body []byte) (models.Selector, error) {
	// Parse the request body into the APIRequest struct
	var selector models.Selector
	err := json.Unmarshal(body, &selector)
	if err != nil {
		return models.Selector{}, err
	}
	return selector, nil
}

// getTransaction returns a transaction from the transactions map
func (g *Gateway) GetTransaction(transactionID string) (models.Transaction, error) {
	transaction, exists := g.Transactions[transactionID]
	if !exists {
		return models.Transaction{}, fmt.Errorf("Transaction not found")
	}
	return transaction, nil
}

// SearchTransaction returns a transaction from the transactions map
func (g *Gateway) SearchTransaction(buyerID string, flavourID string) (models.Transaction, bool) {
	for _, t := range g.Transactions {
		if t.Buyer.NodeID == buyerID && t.FlavourID == flavourID {
			return t, true
		}
	}
	return models.Transaction{}, false
}

// addNewTransacion add a new transaction to the transactions map
func (g *Gateway) addNewTransacion(transaction models.Transaction) {
	g.Transactions[transaction.TransactionID] = transaction
}

// removeTransaction removes a transaction from the transactions map
func (g *Gateway) removeTransaction(transactionID string) {
	delete(g.Transactions, transactionID)
}

// handleError handles errors by sending an error response
func handleError(w http.ResponseWriter, err error, statusCode int) {
	http.Error(w, err.Error(), statusCode)
}

// encodeResponse encodes the response as JSON and writes it to the response writer
func encodeResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
	}
}
