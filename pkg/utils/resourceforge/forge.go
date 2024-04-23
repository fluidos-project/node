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

package resourceforge

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	advertisementv1alpha1 "github.com/fluidos-project/node/apis/advertisement/v1alpha1"
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/namings"
	"github.com/fluidos-project/node/pkg/utils/parseutil"
	"github.com/fluidos-project/node/pkg/utils/tools"
)

// ForgeDiscovery creates a Discovery CR from a FlavourSelector and a solverID.
func ForgeDiscovery(selector *nodecorev1alpha1.FlavourSelector, solverID string) *advertisementv1alpha1.Discovery {
	return &advertisementv1alpha1.Discovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeDiscoveryName(solverID),
			Namespace: flags.FluidoNamespace,
		},
		Spec: advertisementv1alpha1.DiscoverySpec{
			Selector: func() *nodecorev1alpha1.FlavourSelector {
				if selector != nil {
					return selector
				}
				return nil
			}(),
			SolverID:  solverID,
			Subscribe: false,
		},
	}
}

// ForgePeeringCandidate creates a PeeringCandidate CR from a Flavour and a Discovery.
func ForgePeeringCandidate(flavourPeeringCandidate *nodecorev1alpha1.Flavour,
	solverID string, available bool) (pc *advertisementv1alpha1.PeeringCandidate) {
	pc = &advertisementv1alpha1.PeeringCandidate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgePeeringCandidateName(flavourPeeringCandidate.Name),
			Namespace: flags.FluidoNamespace,
		},
		Spec: advertisementv1alpha1.PeeringCandidateSpec{
			Flavour: nodecorev1alpha1.Flavour{
				ObjectMeta: metav1.ObjectMeta{
					Name:      flavourPeeringCandidate.Name,
					Namespace: flavourPeeringCandidate.Namespace,
				},
				Spec: flavourPeeringCandidate.Spec,
			},
			Available: available,
		},
	}
	pc.Spec.SolverID = solverID
	return
}

// ForgeReservation creates a Reservation CR from a PeeringCandidate.
func ForgeReservation(pc *advertisementv1alpha1.PeeringCandidate,
	partition *nodecorev1alpha1.Partition,
	ni nodecorev1alpha1.NodeIdentity) *reservationv1alpha1.Reservation {
	solverID := pc.Spec.SolverID
	reservation := &reservationv1alpha1.Reservation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeReservationName(solverID),
			Namespace: flags.FluidoNamespace,
		},
		Spec: reservationv1alpha1.ReservationSpec{
			SolverID: solverID,
			Buyer:    ni,
			Seller: nodecorev1alpha1.NodeIdentity{
				Domain: pc.Spec.Flavour.Spec.Owner.Domain,
				NodeID: pc.Spec.Flavour.Spec.Owner.NodeID,
				IP:     pc.Spec.Flavour.Spec.Owner.IP,
			},
			PeeringCandidate: nodecorev1alpha1.GenericRef{
				Name:      pc.Name,
				Namespace: pc.Namespace,
			},
			Reserve:  true,
			Purchase: true,
			Partition: func() *nodecorev1alpha1.Partition {
				if partition != nil {
					return partition
				}
				return nil
			}(),
		},
	}
	if partition != nil {
		reservation.Spec.Partition = partition
	}
	return reservation
}

// ForgeContract creates a Contract CR.
func ForgeContract(flavour *nodecorev1alpha1.Flavour, transaction *models.Transaction,
	lc *nodecorev1alpha1.LiqoCredentials) *reservationv1alpha1.Contract {
	return &reservationv1alpha1.Contract{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeContractName(flavour.Name),
			Namespace: flags.FluidoNamespace,
		},
		Spec: reservationv1alpha1.ContractSpec{
			Flavour: *flavour,
			Buyer: nodecorev1alpha1.NodeIdentity{
				Domain: transaction.Buyer.Domain,
				IP:     transaction.Buyer.IP,
				NodeID: transaction.Buyer.NodeID,
			},
			BuyerClusterID:    transaction.ClusterID,
			Seller:            flavour.Spec.Owner,
			SellerCredentials: *lc,
			TransactionID:     transaction.TransactionID,
			Partition: func() *nodecorev1alpha1.Partition {
				if transaction.Partition != nil {
					return parseutil.ParsePartitionFromObj(transaction.Partition)
				}
				return nil
			}(),
			ExpirationTime:   time.Now().Add(flags.ExpirationContract).Format(time.RFC3339),
			ExtraInformation: nil,
		},
		Status: reservationv1alpha1.ContractStatus{
			Phase: nodecorev1alpha1.PhaseStatus{
				Phase:     nodecorev1alpha1.PhaseActive,
				StartTime: tools.GetTimeNow(),
			},
		},
	}
}

// ForgeFlavourFromMetrics creates a new flavour custom resource from the metrics of the node.
func ForgeFlavourFromMetrics(node *models.NodeInfo, ni nodecorev1alpha1.NodeIdentity) (flavour *nodecorev1alpha1.Flavour) {
	return &nodecorev1alpha1.Flavour{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeFlavourName(node.UID, "", ni.Domain),
			Namespace: flags.FluidoNamespace,
		},
		Spec: nodecorev1alpha1.FlavourSpec{
			ProviderID: ni.NodeID,
			Type:       nodecorev1alpha1.K8S,
			Characteristics: nodecorev1alpha1.Characteristics{
				Architecture:      node.Architecture,
				Cpu:               node.ResourceMetrics.CPUAvailable,
				Memory:            node.ResourceMetrics.MemoryAvailable,
				Pods:              node.ResourceMetrics.PodsAvailable,
				EphemeralStorage:  node.ResourceMetrics.EphemeralStorage,
				PersistentStorage: parseutil.ParseQuantityFromString("0"),
				Gpu:               parseutil.ParseQuantityFromString("0"),
			},
			Policy: nodecorev1alpha1.Policy{
				Partitionable: &nodecorev1alpha1.Partitionable{
					CpuMin:     parseutil.ParseQuantityFromString(flags.CPUMin),
					MemoryMin:  parseutil.ParseQuantityFromString(flags.MemoryMin),
					PodsMin:    parseutil.ParseQuantityFromString(flags.PodsMin),
					CpuStep:    parseutil.ParseQuantityFromString(flags.CPUStep),
					MemoryStep: parseutil.ParseQuantityFromString(flags.MemoryStep),
					PodsStep:   parseutil.ParseQuantityFromString(flags.PodsStep),
				},
				Aggregatable: &nodecorev1alpha1.Aggregatable{
					MinCount: int(flags.MinCount),
					MaxCount: int(flags.MaxCount),
				},
			},
			Owner: ni,
			Price: nodecorev1alpha1.Price{
				Amount:   flags.AMOUNT,
				Currency: flags.CURRENCY,
				Period:   flags.PERIOD,
			},
			OptionalFields: nodecorev1alpha1.OptionalFields{
				Availability: true,
				// This previously was the node UID that maybe is not the best choice to manage the scheduling
				WorkerID: node.Name,
			},
		},
	}
}

// ForgeFlavourFromRef creates a new flavour starting from a Reference Flavour and the new Characteristics.
func ForgeFlavourFromRef(f *nodecorev1alpha1.Flavour, char *nodecorev1alpha1.Characteristics) (flavour *nodecorev1alpha1.Flavour) {
	return &nodecorev1alpha1.Flavour{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeFlavourName(f.Spec.OptionalFields.WorkerID, string(f.Spec.Type), f.Spec.Owner.Domain),
			Namespace: flags.FluidoNamespace,
		},
		Spec: nodecorev1alpha1.FlavourSpec{
			ProviderID:      f.Spec.ProviderID,
			Type:            f.Spec.Type,
			Characteristics: *char,
			Policy:          f.Spec.Policy,
			Owner:           f.Spec.Owner,
			Price:           f.Spec.Price,
			OptionalFields: nodecorev1alpha1.OptionalFields{
				Availability: true,
				WorkerID:     f.Spec.OptionalFields.WorkerID,
			},
		},
	}
}

// FORGER FUNCTIONS FROM OBJECTS

// ForgeTransactionObj creates a new Transaction object.
func ForgeTransactionObj(id string, req *models.ReserveRequest) *models.Transaction {
	return &models.Transaction{
		TransactionID: id,
		Buyer:         req.Buyer,
		ClusterID:     req.ClusterID,
		FlavourID:     req.FlavourID,
		Partition: func() *models.Partition {
			if req.Partition != nil {
				return req.Partition
			}
			return nil
		}(),
		StartTime: tools.GetTimeNow(),
	}
}

// ForgeContractObj creates a new Contract object.
func ForgeContractObj(contract *reservationv1alpha1.Contract) models.Contract {
	return models.Contract{
		ContractID:     contract.Name,
		Flavour:        *parseutil.ParseFlavour(&contract.Spec.Flavour),
		Buyer:          parseutil.ParseNodeIdentity(contract.Spec.Buyer),
		BuyerClusterID: contract.Spec.BuyerClusterID,
		Seller:         parseutil.ParseNodeIdentity(contract.Spec.Seller),
		SellerCredentials: models.LiqoCredentials{
			ClusterID:   contract.Spec.SellerCredentials.ClusterID,
			ClusterName: contract.Spec.SellerCredentials.ClusterName,
			Token:       contract.Spec.SellerCredentials.Token,
			Endpoint:    contract.Spec.SellerCredentials.Endpoint,
		},
		Partition: func() *models.Partition {
			if contract.Spec.Partition != nil {
				return parseutil.ParsePartition(contract.Spec.Partition)
			}
			return nil
		}(),
		TransactionID:  contract.Spec.TransactionID,
		ExpirationTime: contract.Spec.ExpirationTime,
		ExtraInformation: func() map[string]string {
			if contract.Spec.ExtraInformation != nil {
				return contract.Spec.ExtraInformation
			}
			return nil
		}(),
	}
}

// ForgeResponsePurchaseObj creates a new response purchase.
func ForgeResponsePurchaseObj(contract *models.Contract) *models.ResponsePurchase {
	return &models.ResponsePurchase{
		Contract: *contract,
		Status:   "Completed",
	}
}

// ForgeContractFromObj creates a Contract from a reservation.
func ForgeContractFromObj(contract *models.Contract) *reservationv1alpha1.Contract {
	return &reservationv1alpha1.Contract{
		ObjectMeta: metav1.ObjectMeta{
			Name:      contract.ContractID,
			Namespace: flags.FluidoNamespace,
		},
		Spec: reservationv1alpha1.ContractSpec{
			Flavour: *ForgeFlavourFromObj(&contract.Flavour),
			Buyer: nodecorev1alpha1.NodeIdentity{
				Domain: contract.Buyer.Domain,
				IP:     contract.Buyer.IP,
				NodeID: contract.Buyer.NodeID,
			},
			BuyerClusterID: contract.BuyerClusterID,
			Seller: nodecorev1alpha1.NodeIdentity{
				NodeID: contract.Seller.NodeID,
				IP:     contract.Seller.IP,
				Domain: contract.Seller.Domain,
			},
			SellerCredentials: nodecorev1alpha1.LiqoCredentials{
				ClusterID:   contract.SellerCredentials.ClusterID,
				ClusterName: contract.SellerCredentials.ClusterName,
				Token:       contract.SellerCredentials.Token,
				Endpoint:    contract.SellerCredentials.Endpoint,
			},
			TransactionID: contract.TransactionID,
			Partition: func() *nodecorev1alpha1.Partition {
				if contract.Partition != nil {
					return parseutil.ParsePartitionFromObj(contract.Partition)
				}
				return nil
			}(),
			ExpirationTime: contract.ExpirationTime,
			ExtraInformation: func() map[string]string {
				if contract.ExtraInformation != nil {
					return contract.ExtraInformation
				}
				return nil
			}(),
		},
		Status: reservationv1alpha1.ContractStatus{
			Phase: nodecorev1alpha1.PhaseStatus{
				Phase:     nodecorev1alpha1.PhaseActive,
				StartTime: tools.GetTimeNow(),
			},
		},
	}
}

// ForgeTransactionFromObj creates a transaction from a Transaction object.
func ForgeTransactionFromObj(transaction *models.Transaction) *reservationv1alpha1.Transaction {
	return &reservationv1alpha1.Transaction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      transaction.TransactionID,
			Namespace: flags.FluidoNamespace,
		},
		Spec: reservationv1alpha1.TransactionSpec{
			FlavourID: transaction.FlavourID,
			StartTime: transaction.StartTime,
			Buyer: nodecorev1alpha1.NodeIdentity{
				Domain: transaction.Buyer.Domain,
				IP:     transaction.Buyer.IP,
				NodeID: transaction.Buyer.NodeID,
			},
			ClusterID: transaction.ClusterID,
			Partition: func() *nodecorev1alpha1.Partition {
				if transaction.Partition != nil {
					return parseutil.ParsePartitionFromObj(transaction.Partition)
				}
				return nil
			}(),
		},
	}
}

// ForgeFlavourFromObj creates a Flavour CR from a Flavour Object (REAR).
func ForgeFlavourFromObj(flavour *models.Flavour) *nodecorev1alpha1.Flavour {
	f := &nodecorev1alpha1.Flavour{
		ObjectMeta: metav1.ObjectMeta{
			Name:      flavour.FlavourID,
			Namespace: flags.FluidoNamespace,
		},
		Spec: nodecorev1alpha1.FlavourSpec{
			ProviderID: flavour.Owner.NodeID,
			Type:       nodecorev1alpha1.K8S,
			Characteristics: nodecorev1alpha1.Characteristics{
				Cpu:               flavour.Characteristics.CPU,
				Memory:            flavour.Characteristics.Memory,
				Pods:              flavour.Characteristics.Pods,
				Architecture:      flavour.Characteristics.Architecture,
				EphemeralStorage:  flavour.Characteristics.EphemeralStorage,
				PersistentStorage: flavour.Characteristics.PersistentStorage,
				Gpu:               flavour.Characteristics.Gpu,
			},
			Policy: nodecorev1alpha1.Policy{
				// Check if flavour.Partitionable is not nil before setting Partitionable
				Partitionable: func() *nodecorev1alpha1.Partitionable {
					if flavour.Policy.Partitionable != nil {
						return &nodecorev1alpha1.Partitionable{
							CpuMin:     flavour.Policy.Partitionable.CPUMinimum,
							MemoryMin:  flavour.Policy.Partitionable.MemoryMinimum,
							CpuStep:    flavour.Policy.Partitionable.CPUStep,
							MemoryStep: flavour.Policy.Partitionable.MemoryStep,
						}
					}
					return nil
				}(),
				Aggregatable: func() *nodecorev1alpha1.Aggregatable {
					if flavour.Policy.Aggregatable != nil {
						return &nodecorev1alpha1.Aggregatable{
							MinCount: flavour.Policy.Aggregatable.MinCount,
							MaxCount: flavour.Policy.Aggregatable.MaxCount,
						}
					}
					return nil
				}(),
			},
			Owner: nodecorev1alpha1.NodeIdentity{
				Domain: flavour.Owner.Domain,
				IP:     flavour.Owner.IP,
				NodeID: flavour.Owner.NodeID,
			},
			Price: nodecorev1alpha1.Price{
				Amount:   flavour.Price.Amount,
				Currency: flavour.Price.Currency,
				Period:   flavour.Price.Period,
			},
			OptionalFields: nodecorev1alpha1.OptionalFields{
				Availability: flavour.OptionalFields.Availability,
				WorkerID:     flavour.OptionalFields.WorkerID,
			},
		},
	}
	return f
}

// ForgePartition creates a Partition from a FlavourSelector.
func ForgePartition(selector *nodecorev1alpha1.FlavourSelector) *nodecorev1alpha1.Partition {
	return &nodecorev1alpha1.Partition{
		Architecture:     selector.Architecture,
		CPU:              selector.RangeSelector.MinCpu,
		Memory:           selector.RangeSelector.MinMemory,
		Pods:             selector.RangeSelector.MinPods,
		EphemeralStorage: selector.RangeSelector.MinEph,
		Storage:          selector.RangeSelector.MinStorage,
		Gpu:              selector.RangeSelector.MinGpu,
	}
}

// ForgeAllocation creates an Allocation from a Contract.
func ForgeAllocation(contract *reservationv1alpha1.Contract, intentID, nodeName string,
	destination nodecorev1alpha1.Destination, nodeType nodecorev1alpha1.NodeType) *nodecorev1alpha1.Allocation {
	return &nodecorev1alpha1.Allocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeAllocationName(contract.Spec.Flavour.Name),
			Namespace: flags.FluidoNamespace,
		},
		Spec: nodecorev1alpha1.AllocationSpec{
			RemoteClusterID: func() string {
				if nodeType == nodecorev1alpha1.Node {
					return contract.Spec.BuyerClusterID
				}
				return contract.Spec.SellerCredentials.ClusterID
			}(),
			IntentID:    intentID,
			NodeName:    nodeName,
			Type:        nodeType,
			Destination: destination,
			Forwarding:  false,
			Contract: nodecorev1alpha1.GenericRef{
				Name:      contract.Name,
				Namespace: contract.Namespace,
			},
		},
	}
}
