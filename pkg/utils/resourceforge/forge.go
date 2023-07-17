package resourceforge

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	advertisementv1alpha1 "fluidos.eu/node/api/advertisement/v1alpha1"
	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	reservationv1alpha1 "fluidos.eu/node/api/reservation/v1alpha1"
	"fluidos.eu/node/pkg/utils/tools"
	"fluidos.eu/node/pkg/utils/flags"
	"fluidos.eu/node/pkg/utils/models"
	"fluidos.eu/node/pkg/utils/namings"
	"fluidos.eu/node/pkg/utils/parseutil"
)

// ForgeDiscovery creates a Discovery CR from a FlavourSelector and a solverID
func ForgeDiscovery(selector nodecorev1alpha1.FlavourSelector, solverID string) *advertisementv1alpha1.Discovery {
	return &advertisementv1alpha1.Discovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeDiscoveryName(solverID),
			Namespace: flags.DISCOVERY_DEFAULT_NAMESPACE,
		},
		Spec: advertisementv1alpha1.DiscoverySpec{
			Selector:  selector,
			SolverID:  solverID,
			Subscribe: false,
		},
	}
}

// ForgePeeringCandidate creates a PeeringCandidate CR from a Flavour and a Discovery
func ForgePeeringCandidate(flavourPeeringCandidate *nodecorev1alpha1.Flavour, solverID string, reserved bool) (pc *advertisementv1alpha1.PeeringCandidate) {
	pc = &advertisementv1alpha1.PeeringCandidate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgePeeringCandidateName(flavourPeeringCandidate.Name),
			Namespace: flags.PC_DEFAULT_NAMESPACE,
		},
		Spec: advertisementv1alpha1.PeeringCandidateSpec{
			Flavour: nodecorev1alpha1.Flavour{
				ObjectMeta: metav1.ObjectMeta{
					Name:      flavourPeeringCandidate.Name,
					Namespace: flavourPeeringCandidate.Namespace,
				},
				Spec: flavourPeeringCandidate.Spec,
			},
		},
	}

	if reserved {
		pc.Spec.SolverID = solverID
		pc.Spec.Reserved = true
	}

	return
}

// ForgeReservation creates a Reservation CR from a PeeringCandidate
func ForgeReservation(peeringCandidate advertisementv1alpha1.PeeringCandidate) *reservationv1alpha1.Reservation {
	solverID := peeringCandidate.Spec.SolverID
	return &reservationv1alpha1.Reservation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeReservationName(solverID),
			Namespace: flags.RESERVATION_DEFAULT_NAMESPACE,
		},
		Spec: reservationv1alpha1.ReservationSpec{
			SolverID: solverID,
			Buyer: nodecorev1alpha1.NodeIdentity{
				Domain: flags.DOMAIN,
				NodeID: flags.CLIENT_ID,
				IP:     flags.IP_ADDR,
			},
			Seller: nodecorev1alpha1.NodeIdentity{
				Domain: peeringCandidate.Spec.Flavour.Spec.Owner.Domain,
				NodeID: peeringCandidate.Spec.Flavour.Spec.Owner.NodeID,
				IP:     peeringCandidate.Spec.Flavour.Spec.Owner.IP,
			},
			PeeringCandidate: nodecorev1alpha1.GenericRef{
				Name:      peeringCandidate.Name,
				Namespace: peeringCandidate.Namespace,
			},
			Reserve:  true,
			Purchase: true,
		},
	}
}

// ForgeContract creates a Contract CR
func ForgeContract(flavour nodecorev1alpha1.Flavour, transaction models.Transaction) *reservationv1alpha1.Contract {
	return &reservationv1alpha1.Contract{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeContractName(flavour.Name),
			Namespace: flags.CONTRACT_DEFAULT_NAMESPACE,
		},
		Spec: reservationv1alpha1.ContractSpec{
			Flavour: flavour,
			Buyer: nodecorev1alpha1.NodeIdentity{
				Domain: transaction.Buyer.Domain,
				IP:     transaction.Buyer.IP,
				NodeID: transaction.Buyer.NodeID,
			},
			Seller:         flavour.Spec.Owner,
			TransactionID:  transaction.TransactionID,
			Partition:      parseutil.ParsePartitionFromObj(transaction.Partition),
			ExpirationTime: time.Now().Add(flags.EXPIRATION_CONTRACT).Format(time.RFC3339),
		},
		Status: reservationv1alpha1.ContractStatus{
			Phase: nodecorev1alpha1.PhaseStatus{
				Phase:     nodecorev1alpha1.PhaseActive,
				StartTime: tools.GetTimeNow(),
			},
		},
	}
}

// ForgeFlavourFromMetrics creates a new flavour custom resource from the metrics of the node
func ForgeFlavourFromMetrics(node models.NodeInfo) (flavour *nodecorev1alpha1.Flavour) {
	return &nodecorev1alpha1.Flavour{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeFlavourName(node.UID),
			Namespace: flags.FLAVOUR_DEFAULT_NAMESPACE,
		},
		Spec: nodecorev1alpha1.FlavourSpec{
			ProviderID: node.UID,
			Type:       nodecorev1alpha1.K8S,
			Characteristics: nodecorev1alpha1.Characteristics{
				Architecture:     node.Architecture,
				Cpu:              node.ResourceMetrics.CPUAvailable,
				Memory:           node.ResourceMetrics.MemoryAvailable,
				EphemeralStorage: node.ResourceMetrics.EphemeralStorage,
			},
			Policy: nodecorev1alpha1.Policy{
				Partitionable: &nodecorev1alpha1.Partitionable{
					CpuMin:     int(flags.CPU_MIN),
					MemoryMin:  int(flags.MEMORY_MIN),
					CpuStep:    int(flags.CPU_STEP),
					MemoryStep: int(flags.MEMORY_STEP),
				},
				Aggregatable: &nodecorev1alpha1.Aggregatable{
					MinCount: int(flags.MIN_COUNT),
					MaxCount: int(flags.MAX_COUNT),
				},
			},
			Owner: nodecorev1alpha1.NodeIdentity{
				Domain: flags.DOMAIN,
				IP:     flags.IP_ADDR,
				NodeID: flags.CLIENT_ID,
			},
			Price: nodecorev1alpha1.Price{
				Amount:   flags.AMOUNT,
				Currency: flags.CURRENCY,
				Period:   flags.PERIOD,
			},
			OptionalFields: nodecorev1alpha1.OptionalFields{
				Availability: true,
				WorkerID:     node.UID,
			},
		},
	}
}

// FORGER FUNCTIONS FROM OBJECTS

// ForgeTransaction creates a new transaction
func ForgeTransactionObj(ID string, req models.ReserveRequest) models.Transaction {
	return models.Transaction{
		TransactionID: ID,
		Buyer:         req.Buyer,
		FlavourID:     req.FlavourID,
		Partition:     req.Partition,
		StartTime:     tools.GetTimeNow(),
	}
}

func ForgeContractObj(contract *reservationv1alpha1.Contract) models.Contract {
	return models.Contract{
		ContractID:     contract.Name,
		Flavour:        parseutil.ParseFlavour(contract.Spec.Flavour),
		Buyer:          parseutil.ParseNodeIdentity(contract.Spec.Buyer),
		Seller:         parseutil.ParseNodeIdentity(contract.Spec.Seller),
		Partition:      parseutil.ParsePartition(contract.Spec.Partition),
		TransactionID:  contract.Spec.TransactionID,
		ExpirationTime: contract.Spec.ExpirationTime,
	}
}

// ForgeResponsePurchaseObj creates a new response purchase
func ForgeResponsePurchaseObj(contract models.Contract) models.ResponsePurchase {
	return models.ResponsePurchase{
		Contract: contract,
		Status:   "Completed",
	}
}

// ForgeContractFromObj creates a Contract from a reservation
func ForgeContractFromObj(contract models.Contract) *reservationv1alpha1.Contract {
	return &reservationv1alpha1.Contract{
		ObjectMeta: metav1.ObjectMeta{
			Name:      contract.ContractID,
			Namespace: flags.CONTRACT_DEFAULT_NAMESPACE,
		},
		Spec: reservationv1alpha1.ContractSpec{
			Flavour: *ForgeFlavourFromObj(contract.Flavour),
			Buyer: nodecorev1alpha1.NodeIdentity{
				Domain: contract.Buyer.Domain,
				IP:     contract.Buyer.IP,
				NodeID: contract.Buyer.NodeID,
			},
			Seller: nodecorev1alpha1.NodeIdentity{
				NodeID: contract.Seller.NodeID,
				IP:     contract.Seller.IP,
				Domain: contract.Seller.Domain,
			},
			TransactionID:  contract.TransactionID,
			Partition:      parseutil.ParsePartitionFromObj(contract.Partition),
			ExpirationTime: contract.ExpirationTime,
		},
		Status: reservationv1alpha1.ContractStatus{
			Phase: nodecorev1alpha1.PhaseStatus{
				Phase:     nodecorev1alpha1.PhaseActive,
				StartTime: tools.GetTimeNow(),
			},
		},
	}
}

// ForgeTransactionFromObj creates a transaction from a Transaction object
func ForgeTransactionFromObj(reservation *models.Transaction) *reservationv1alpha1.Transaction {
	return &reservationv1alpha1.Transaction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      reservation.TransactionID,
			Namespace: flags.TRANSACTION_DEFAULT_NAMESPACE,
		},
		Spec: reservationv1alpha1.TransactionSpec{
			FlavourID: reservation.FlavourID,
			StartTime: reservation.StartTime,
			Buyer: nodecorev1alpha1.NodeIdentity{
				Domain: reservation.Buyer.Domain,
				IP:     reservation.Buyer.IP,
				NodeID: reservation.Buyer.NodeID,
			},
			Partition: parseutil.ParsePartitionFromObj(reservation.Partition),
		},
	}
}

// ForgeFlavourFromObj creates a Flavour CR from a Flavour Object (REAR)
func ForgeFlavourFromObj(flavour models.Flavour) *nodecorev1alpha1.Flavour {
	f := &nodecorev1alpha1.Flavour{
		ObjectMeta: metav1.ObjectMeta{
			Name:      flavour.FlavourID,
			Namespace: flags.DEFAULT_NAMESPACE,
		},
		Spec: nodecorev1alpha1.FlavourSpec{
			ProviderID: flavour.Owner.NodeID,
			Type:       nodecorev1alpha1.K8S,
			Characteristics: nodecorev1alpha1.Characteristics{
				Cpu:    *resource.NewQuantity(int64(flavour.Characteristics.CPU), resource.DecimalSI),
				Memory: *resource.NewQuantity(int64(flavour.Characteristics.Memory), resource.BinarySI),
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
		},
	}

	if flavour.Characteristics.EphemeralStorage != 0 {
		f.Spec.Characteristics.EphemeralStorage = *resource.NewQuantity(int64(flavour.Characteristics.EphemeralStorage), resource.BinarySI)
	}
	if flavour.Characteristics.PersistentStorage != 0 {
		f.Spec.Characteristics.PersistentStorage = *resource.NewQuantity(int64(flavour.Characteristics.PersistentStorage), resource.BinarySI)
	}
	if flavour.Characteristics.GPU != 0 {
		f.Spec.Characteristics.Gpu = *resource.NewQuantity(int64(flavour.Characteristics.GPU), resource.DecimalSI)
	}
	if flavour.Characteristics.Architecture != "" {
		f.Spec.Characteristics.Architecture = flavour.Characteristics.Architecture
	}
	return f
}
