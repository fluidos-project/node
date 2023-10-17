// Copyright 2022-2023 FLUIDOS Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package namings

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"github.com/fluidos-project/node/pkg/utils/flags"
)

// ForgeContractName creates a name for the Contract CR
func ForgeContractName(flavourID string) string {
	hash := ForgeHashString(flavourID, 4)
	return fmt.Sprintf("contract-%s-%s", flavourID, hash)
}

// ForgePeeringCandidateName generates a name for the PeeringCandidate
func ForgePeeringCandidateName(flavourID string) string {
	return fmt.Sprintf("peeringcandidate-%s", flavourID)
}

// ForgeReservationName generates a name for the Reservation
func ForgeReservationName(solverID string) string {
	return fmt.Sprintf("reservation-%s", solverID)
}

// ForgeFlavourName returns the name of the flavour following the pattern Domain-Type-rand(4)
func ForgeFlavourName(WorkerID, domain string) string {
	rand, err := ForgeRandomString()
	if err != nil {
		klog.Errorf("Error when generating random string: %s", err)
	}

	return domain + "-" + flags.RESOURCE_TYPE + "-" + ForgeHashString(WorkerID+rand, 8)
}

// ForgeDiscoveryName returns the name of the discovery following the pattern solverID-discovery
func ForgeDiscoveryName(solverID string) string {
	return fmt.Sprintf("discovery-%s", solverID)
}

func RetrieveSolverNameFromDiscovery(discoveryName string) string {
	return strings.TrimPrefix(discoveryName, "discovery-")
}

func RetrieveSolverNameFromReservation(reservationName string) string {
	return strings.TrimPrefix(reservationName, "reservation-")
}

// ForgeTransactionID Generates a unique transaction ID using the current timestamp
func ForgeTransactionID() (string, error) {
	// Convert the random bytes to a hexadecimal string
	transactionID, err := ForgeRandomString()
	if err != nil {
		return "", err
	}

	// Add a timestamp to ensure uniqueness
	transactionID = fmt.Sprintf("%s-%d", transactionID, time.Now().UnixNano())

	return transactionID, nil
}

// RetrieveFlavourNameFromPC generates a name for the Flavour from the PeeringCandidate
func RetrieveFlavourNameFromPC(pcName string) string {
	return strings.TrimPrefix(pcName, "peeringcandidate-")
}

// ForgePrefixClientID generates a prefix for the client ID
func ForgeRandomString() (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	randomString := hex.EncodeToString(randomBytes)

	return randomString, nil
}

// ForgeHashString computes SHA-256 Hash of the NodeUID
func ForgeHashString(input string, lenght int) string {
	hash := sha256.Sum256([]byte(input))
	hashString := hex.EncodeToString(hash[:])
	uniqueString := hashString[:lenght]

	return uniqueString
}
