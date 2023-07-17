package resourceforge

import (
	advertisementv1alpha1 "fluidos.eu/node/api/advertisement/v1alpha1"
	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	reservationv1alpha1 "fluidos.eu/node/api/reservation/v1alpha1"
	"fluidos.eu/node/pkg/utils/common"
	"fluidos.eu/node/pkg/utils/flags"
	"fluidos.eu/node/pkg/utils/models"
	"fluidos.eu/node/pkg/utils/namings"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// ForgeContractFromModel creates a Contract from a reservation
func ForgeContractFromModel(reservation *reservationv1alpha1.Reservation, contract models.Contract) *reservationv1alpha1.Contract {
	return &reservationv1alpha1.Contract{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "contract-" + reservation.Name,
			Namespace: flags.CONTRACT_DEFAULT_NAMESPACE,
		},
		Spec: reservationv1alpha1.ContractSpec{
			Flavour: *ForgeFlavourCustomResource(contract.Flavour),
			Buyer: nodecorev1alpha1.NodeIdentity{
				NodeID: reservation.Spec.Buyer.NodeID,
			},
			Seller: nodecorev1alpha1.NodeIdentity{
				NodeID: reservation.Spec.Seller.NodeID,
			},
		},
		Status: reservationv1alpha1.ContractStatus{
			Phase: nodecorev1alpha1.PhaseStatus{
				Phase:     nodecorev1alpha1.PhaseActive,
				StartTime: common.GetTimeNow(),
			},
		},
	}
}

// ForgeFlavourCustomResource creates a Flavour CR from a Flavour Object (REAR)
func ForgeFlavourCustomResource(flavour models.Flavour) *nodecorev1alpha1.Flavour {
	return &nodecorev1alpha1.Flavour{
		ObjectMeta: metav1.ObjectMeta{
			Name:      flavour.FlavourID,
			Namespace: flags.DEFAULT_NAMESPACE,
		},
		Spec: nodecorev1alpha1.FlavourSpec{
			ProviderID: flavour.Owner.ID,
			Type:       nodecorev1alpha1.K8S,
			Characteristics: nodecorev1alpha1.Characteristics{
				Cpu:    *resource.NewQuantity(int64(flavour.Characteristics.CPU), resource.DecimalSI),
				Memory: *resource.NewQuantity(int64(flavour.Characteristics.RAM), resource.BinarySI),
			},
			Policy: nodecorev1alpha1.Policy{
				// Check if flavour.Partitionable is not nil before setting Partitionable
				Partitionable: func() *nodecorev1alpha1.Partitionable {
					if flavour.Policy.Partitionable != nil {
						return &nodecorev1alpha1.Partitionable{
							CpuMin:     flavour.Policy.Partitionable.CPUMinimum,
							MemoryMin:  flavour.Policy.Partitionable.RAMMinimum,
							CpuStep:    flavour.Policy.Partitionable.CPUStep,
							MemoryStep: flavour.Policy.Partitionable.RAMStep,
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
				Domain: flavour.Owner.DomainName,
				IP:     flavour.Owner.IP,
				NodeID: flavour.Owner.ID,
			},
			Price: nodecorev1alpha1.Price{
				Amount:   flavour.Price.Amount,
				Currency: flavour.Price.Currency,
				Period:   flavour.Price.Period,
			},
		},
	}
}

// ForgePeeringCandidateCustomResources creates a PeeringCandidate CR from a Flavour and a Discovery
func ForgePeeringCandidateCustomResources(flavourPeeringCandidate *nodecorev1alpha1.Flavour, discovery *advertisementv1alpha1.Discovery, reserved bool) (pc *advertisementv1alpha1.PeeringCandidate) {
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
		pc.Spec.SolverID = discovery.Spec.SolverID
		pc.Spec.Reserved = true
	}

	return
}

// ForgeReservationCustomResource creates a Reservation CR from a PeeringCandidate
func ForgeReservationCustomResource(peeringCandidate advertisementv1alpha1.PeeringCandidate) *reservationv1alpha1.Reservation {
	return &reservationv1alpha1.Reservation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeReservationName(peeringCandidate.Name),
			Namespace: flags.RESERVATION_DEFAULT_NAMESPACE,
		},
		Spec: reservationv1alpha1.ReservationSpec{
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

// ForgeTransaction creates a transaction from a Transaction object
func ForgeTransaction(reservation *models.Transaction) *reservationv1alpha1.Transaction {
	return &reservationv1alpha1.Transaction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      reservation.TransactionID,
			Namespace: flags.TRANSACTION_DEFAULT_NAMESPACE,
		},
		Spec: reservationv1alpha1.TransactionSpec{
			FlavourID: reservation.FlavourID,
			StartTime: reservation.StartTime,
		},
	}
}

// ForgeContractCustomResource creates a Contract CR
func ForgeContractCustomResource(flavour nodecorev1alpha1.Flavour, buyerID string) *reservationv1alpha1.Contract {
	return &reservationv1alpha1.Contract{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeContractName(flavour.Name),
			Namespace: flags.CONTRACT_DEFAULT_NAMESPACE,
		},
		Spec: reservationv1alpha1.ContractSpec{
			Flavour: flavour,
			Buyer: nodecorev1alpha1.NodeIdentity{
				NodeID: buyerID,
			},
			Seller: flavour.Spec.Owner,
		},
		Status: reservationv1alpha1.ContractStatus{
			Phase: nodecorev1alpha1.PhaseStatus{
				Phase:     nodecorev1alpha1.PhaseActive,
				StartTime: common.GetTimeNow(),
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

// ForgeTransaction creates a new transaction
func ForgeTransactionObject(flavourID, transactionID string, buyer models.Owner) models.Transaction {
	return models.Transaction{
		TransactionID: transactionID,
		Buyer:         buyer,
		FlavourID:     flavourID,
		StartTime:     common.GetTimeNow(),
	}
}

// ForgeFlavourObject creates a Flavour Object from a Flavour CR
func ForgeFlavourObject(flavour *nodecorev1alpha1.Flavour) models.Flavour {
	cpu, _ := flavour.Spec.Characteristics.Cpu.AsInt64()
	ram, _ := flavour.Spec.Characteristics.Memory.AsInt64()
	return models.Flavour{
		FlavourID: flavour.Name,
		Owner: models.Owner{
			ID:         flavour.Spec.Owner.NodeID,
			IP:         flavour.Spec.Owner.IP,
			DomainName: flavour.Spec.Owner.Domain,
		},
		Characteristics: models.Characteristics{
			CPU: int(cpu),
			RAM: int(ram),
		},
		Policy: models.Policy{
			Partitionable: &models.Partitionable{
				CPUMinimum: flavour.Spec.Policy.Partitionable.CpuMin,
				RAMMinimum: flavour.Spec.Policy.Partitionable.MemoryMin,
				CPUStep:    flavour.Spec.Policy.Partitionable.CpuStep,
				RAMStep:    flavour.Spec.Policy.Partitionable.MemoryStep,
			},
			Aggregatable: &models.Aggregatable{
				MinCount: flavour.Spec.Policy.Aggregatable.MinCount,
				MaxCount: flavour.Spec.Policy.Aggregatable.MaxCount,
			},
		},
		Price: models.Price{
			Amount:   flavour.Spec.Price.Amount,
			Currency: flavour.Spec.Price.Currency,
			Period:   flavour.Spec.Price.Period,
		},
		OptionalFields: models.OptionalFields{
			Availability: flavour.Spec.OptionalFields.Availability,
			WorkerID:     flavour.Spec.OptionalFields.WorkerID,
		},
	}
}

// ForgeContractObject creates a Contract Object
func ForgeContractObject(contract *reservationv1alpha1.Contract, buyerID, transactionID string, partition models.Partition) models.Contract {
	return models.Contract{
		ContractID:    contract.Name,
		Flavour:       ForgeFlavourObject(&contract.Spec.Flavour),
		BuyerID:       buyerID,
		TransactionID: transactionID,
		Partition:     partition,
	}
}

// ForgeResponsePurchaseObject creates a new response purchase
func ForgeResponsePurchaseObject(contract models.Contract) models.ResponsePurchase {
	return models.ResponsePurchase{
		Contract: contract,
		Status:   "Completed",
	}
}
