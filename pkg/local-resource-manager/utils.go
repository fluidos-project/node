package localResourceManager

import (
	"crypto/sha256"
	"encoding/hex"
)

// generateUniqueString computes SHA-256 Hash of the NodeUID
func generateUniqueString(input string) string {
	hash := sha256.Sum256([]byte(input))
	hashString := hex.EncodeToString(hash[:])
	uniqueString := hashString[:6]

	return uniqueString
}
