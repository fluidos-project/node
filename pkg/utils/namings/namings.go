package namings

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"fluidos.eu/node/pkg/utils/flags"
)

// ForgeContractName creates a name for the Contract CR
func ForgeContractName(flavourID string) string {
	hash := ForgeUniqueString(flavourID)
	return fmt.Sprintf("contract-%s-%s", flavourID, hash)
}

// ForgePeeringCandidateName generates a name for the PeeringCandidate
func ForgePeeringCandidateName(flavorName string) string {
	return fmt.Sprintf("peeringcandidate-%s", flavorName)
}

// ForgeReservationName generates a name for the Reservation
func ForgeReservationName(flavorName string) string {
	hash := ForgeUniqueString(flavorName)
	return fmt.Sprintf("reservation-%s-%s", flavorName, hash)
}

// ForgeFlavourName returns the name of the flavour following the pattern nodeID-Type-rand(4)
func ForgeFlavourName(id string) string {
	return id + "-" + flags.RESOURCE_TYPE + "-" + ForgeUniqueString(id)
}

// ForgeDiscoveryName returns the name of the discovery following the pattern solverID-discovery
func ForgeDiscoveryName(solverID string) string {
	return fmt.Sprintf("%s-discovery", solverID)
}

// ForgeTransactionID Generates a unique transaction ID using the current timestamp
func ForgeTransactionID() (string, error) {
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

// ForgeFlavourNameFromPC generates a name for the Flavour from the PeeringCandidate
func ForgeFlavourNameFromPC(pcName string) string {
	return strings.TrimPrefix(pcName, "peeringcandidate-")
}

// ForgePrefixClientID generates a prefix for the client ID
func ForgePrefixClientID() (string, error) {
	randomBytes := make([]byte, 10)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	prefixClientID := hex.EncodeToString(randomBytes)

	return prefixClientID + "fluidos.eu", nil
}

// ForgeUniqueString computes SHA-256 Hash of the NodeUID
func ForgeUniqueString(input string) string {
	hash := sha256.Sum256([]byte(input))
	hashString := hex.EncodeToString(hash[:])
	uniqueString := hashString[:4]

	return uniqueString
}
