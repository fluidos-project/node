// Copyright 2022-2024 FLUIDOS Project
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

	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/flags"
)

// ForgeVirtualNodeName creates a name for the VirtualNode starting from the cluster name of the remote cluster.
func ForgeVirtualNodeName(clusterName string) string {
	return fmt.Sprintf("liqo-%s", clusterName)
}

// ForgeContractName creates a name for the Contract.
func ForgeContractName(flavorID string) string {
	hash := ForgeHashString(flavorID, 4)
	contractName := fmt.Sprintf("contract-%s-%s", flavorID, hash)
	// Replace all dots with dashes
	return strings.ReplaceAll(contractName, ".", "-")
}

// ForgeAllocationName generates a name for the Allocation.
func ForgeAllocationName(flavorID string) string {
	hash := ForgeHashString(flavorID, 4)
	return fmt.Sprintf("allocation-%s-%s", flavorID, hash)
}

// ForgePeeringCandidateName generates a name for the PeeringCandidate.
func ForgePeeringCandidateName(flavorID string) string {
	return fmt.Sprintf("peeringcandidate-%s", flavorID)
}

// ForgeReservationName generates a name for the Reservation.
func ForgeReservationName(solverID string) string {
	return fmt.Sprintf("reservation-%s", solverID)
}

// ForgeFlavorName returns the name of the flavor following the pattern Domain-resourceType-rand(4).
func ForgeFlavorName(resourceType, domain string) string {
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
	// Trim the random string to 4 characters
	r = r[:4]

	name := domain + "-" + resType + "-" + r

	// Force lowercase
	return strings.ToLower(name)
}

// ForgePartitionName generates a name for the Partition.
func ForgePartitionName(partitionType string) string {
	// Generate Random String
	r, err := ForgeRandomString()
	if err != nil {
		klog.Errorf("Error when generating random string: %s", err)
	}

	// Append the random string to the partition type
	return fmt.Sprintf("%s-%s", partitionType, r)
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

// RetrieveFlavorNameFromPC generates a name for the Flavor from the PeeringCandidate.
func RetrieveFlavorNameFromPC(pcName string) string {
	return strings.TrimPrefix(pcName, "peeringcandidate-")
}

// ForgeKnownClusterName generates a name for the Cluster.
func ForgeKnownClusterName(nodeID string) string {
	return fmt.Sprintf("knowncluster-%s", nodeID)
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

// ForgeNamespaceName creates a namespace name from a contract.
func ForgeNamespaceName(contract *reservationv1alpha1.Contract) string {
	var namespaceName string
	if contract == nil {
		// Create random string
		rnd, err := ForgeRandomString()
		if err != nil {
			klog.Errorf("Error when generating random string: %s", err)
		}
		rnd = rnd[:4]
		namespaceName = fmt.Sprintf("fluidos-namespace-%s", rnd)
	} else {
		namespaceName = contract.Name
	}

	return namespaceName
}

// ForgeSecretName creates a secret name from a contract.
func ForgeSecretName(contract *reservationv1alpha1.Contract) string {
	var secretName string
	if contract == nil {
		// Create random string
		rnd, err := ForgeRandomString()
		if err != nil {
			klog.Errorf("Error when generating random string: %s", err)
		}
		rnd = rnd[:4]
		secretName = fmt.Sprintf("fluidos-secret-%s", rnd)
	} else {
		secretName = fmt.Sprintf("fluidos-secret-%s", contract.Name)
	}

	return secretName
}
