package contractmanager

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// generateTransactionID Generates a unique transaction ID using the current timestamp
func generateTransactionID() (string, error) {
	// Generate a random byte slice
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Convert the random bytes to a hexadecimal string
	transactionID := hex.EncodeToString(randomBytes)

	// Add a timestamp to ensure uniqueness
	transactionID = fmt.Sprintf("%s-%d", transactionID, time.Now().UnixNano())

	return transactionID, nil
}
