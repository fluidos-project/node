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

// ForgeVirtualNodeName creates a name for the VirtualNode starting from the cluster name of the remote cluster.
func ForgeVirtualNodeName(clusterName string) string {
	return fmt.Sprintf("liqo-%s", clusterName)
}

// ForgeContractName creates a name for the Contract.
func ForgeContractName(flavourID string) string {
	hash := ForgeHashString(flavourID, 4)
	return fmt.Sprintf("contract-%s-%s", flavourID, hash)
}

// ForgeAllocationName generates a name for the Allocation.
func ForgeAllocationName(flavourID string) string {
	hash := ForgeHashString(flavourID, 4)
	return fmt.Sprintf("allocation-%s-%s", flavourID, hash)
}

// ForgePeeringCandidateName generates a name for the PeeringCandidate.
func ForgePeeringCandidateName(flavourID string) string {
	return fmt.Sprintf("peeringcandidate-%s", flavourID)
}

// ForgeReservationName generates a name for the Reservation.
func ForgeReservationName(solverID string) string {
	return fmt.Sprintf("reservation-%s", solverID)
}

// ForgeFlavourName returns the name of the flavour following the pattern Domain-resourceType-rand(4).
func ForgeFlavourName(workerID, resourceType, domain string) string {
	var resType string
	if resourceType == "" {
		resType = flags.ResourceType
	} else {
		resType = resourceType
	}
	r, err := ForgeRandomString()
	if err != nil {
		klog.Errorf("Error when generating random string: %s", err)
	}

	return domain + "-" + resType + "-" + ForgeHashString(workerID+r, 8)
}

// ForgeDiscoveryName returns the name of the discovery following the pattern solverID-discovery.
func ForgeDiscoveryName(solverID string) string {
	return fmt.Sprintf("discovery-%s", solverID)
}

// RetrieveSolverNameFromDiscovery retrieves the solver name from the discovery name.
func RetrieveSolverNameFromDiscovery(discoveryName string) string {
	return strings.TrimPrefix(discoveryName, "discovery-")
}

// RetrieveSolverNameFromReservation retrieves the solver name from the reservation name.
func RetrieveSolverNameFromReservation(reservationName string) string {
	return strings.TrimPrefix(reservationName, "reservation-")
}

// ForgeTransactionID Generates a unique transaction ID using the current timestamp.
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

// RetrieveFlavourNameFromPC generates a name for the Flavour from the PeeringCandidate.
func RetrieveFlavourNameFromPC(pcName string) string {
	return strings.TrimPrefix(pcName, "peeringcandidate-")
}

// ForgeRandomString generates a random string of 16 bytes.
func ForgeRandomString() (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	randomString := hex.EncodeToString(randomBytes)

	return randomString, nil
}

// ForgeHashString computes SHA-256 Hash of the NodeUID.
func ForgeHashString(input string, length int) string {
	hash := sha256.Sum256([]byte(input))
	hashString := hex.EncodeToString(hash[:])
	uniqueString := hashString[:length]

	return uniqueString
}
