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

package parseutil

import (
	"k8s.io/apimachinery/pkg/api/resource"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/models"
)

// ParseFlavourSelector parses FlavourSelector into a Selector.
func ParseFlavourSelector(selector *nodecorev1alpha1.FlavourSelector) *models.Selector {
	s := &models.Selector{
		Architecture: selector.Architecture,
		FlavourType:  selector.FlavourType,
	}

	if selector.MatchSelector != nil {
		s.MatchSelector = &models.MatchSelector{
			CPU:              selector.MatchSelector.CPU,
			Memory:           selector.MatchSelector.Memory,
			Pods:             selector.MatchSelector.Pods,
			EphemeralStorage: selector.MatchSelector.EphemeralStorage,
			Storage:          selector.MatchSelector.Storage,
			Gpu:              selector.MatchSelector.Gpu,
		}
	}

	if selector.RangeSelector != nil {
		s.RangeSelector = &models.RangeSelector{
			MinCPU:     selector.RangeSelector.MinCpu,
			MinMemory:  selector.RangeSelector.MinMemory,
			MinPods:    selector.RangeSelector.MinPods,
			MinEph:     selector.RangeSelector.MinEph,
			MinStorage: selector.RangeSelector.MinStorage,
			MinGpu:     selector.RangeSelector.MinGpu,
			MaxCPU:     selector.RangeSelector.MaxCpu,
			MaxMemory:  selector.RangeSelector.MaxMemory,
			MaxPods:    selector.RangeSelector.MaxPods,
			MaxEph:     selector.RangeSelector.MaxEph,
			MaxStorage: selector.RangeSelector.MaxStorage,
			MaxGpu:     selector.RangeSelector.MaxGpu,
		}
	}

	return s
}

// ParsePartition creates a Partition Object from a Partition CR.
func ParsePartition(partition *nodecorev1alpha1.Partition) *models.Partition {
	return &models.Partition{
		CPU:              partition.CPU,
		Memory:           partition.Memory,
		Pods:             partition.Pods,
		EphemeralStorage: partition.EphemeralStorage,
		Storage:          partition.Storage,
		Gpu:              partition.Gpu,
	}
}

// ParsePartitionFromObj creates a Partition CR from a Partition Object.
func ParsePartitionFromObj(partition *models.Partition) *nodecorev1alpha1.Partition {
	return &nodecorev1alpha1.Partition{
		Architecture:     partition.Architecture,
		CPU:              partition.CPU,
		Memory:           partition.Memory,
		Pods:             partition.Pods,
		Gpu:              partition.Gpu,
		Storage:          partition.Storage,
		EphemeralStorage: partition.EphemeralStorage,
	}
}

// ParseNodeIdentity creates a NodeIdentity Object from a NodeIdentity CR.
func ParseNodeIdentity(node nodecorev1alpha1.NodeIdentity) models.NodeIdentity {
	return models.NodeIdentity{
		NodeID: node.NodeID,
		IP:     node.IP,
		Domain: node.Domain,
	}
}

// ParseFlavour creates a Flavour Object from a Flavour CR.
func ParseFlavour(flavour *nodecorev1alpha1.Flavour) *models.Flavour {
	return &models.Flavour{
		FlavourID:  flavour.Name,
		Type:       string(flavour.Spec.Type),
		ProviderID: flavour.Spec.ProviderID,
		Characteristics: models.Characteristics{
			Architecture:      flavour.Spec.Characteristics.Architecture,
			CPU:               flavour.Spec.Characteristics.Cpu,
			Memory:            flavour.Spec.Characteristics.Memory,
			Pods:              flavour.Spec.Characteristics.Pods,
			PersistentStorage: flavour.Spec.Characteristics.PersistentStorage,
			EphemeralStorage:  flavour.Spec.Characteristics.EphemeralStorage,
			Gpu:               flavour.Spec.Characteristics.Gpu,
		},
		Owner: ParseNodeIdentity(flavour.Spec.Owner),
		Policy: models.Policy{
			Partitionable: func() *models.Partitionable {
				if flavour.Spec.Policy.Partitionable != nil {
					return &models.Partitionable{
						CPUMinimum:    flavour.Spec.Policy.Partitionable.CpuMin,
						MemoryMinimum: flavour.Spec.Policy.Partitionable.MemoryMin,
						PodsMinimum:   flavour.Spec.Policy.Partitionable.PodsMin,
						CPUStep:       flavour.Spec.Policy.Partitionable.CpuStep,
						MemoryStep:    flavour.Spec.Policy.Partitionable.MemoryStep,
						PodsStep:      flavour.Spec.Policy.Partitionable.PodsStep,
					}
				}
				return nil
			}(),
			Aggregatable: func() *models.Aggregatable {
				if flavour.Spec.Policy.Aggregatable != nil {
					return &models.Aggregatable{
						MinCount: flavour.Spec.Policy.Aggregatable.MinCount,
						MaxCount: flavour.Spec.Policy.Aggregatable.MaxCount,
					}
				}
				return nil
			}(),
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

// ParseContract creates a Contract Object.
func ParseContract(contract *reservationv1alpha1.Contract) *models.Contract {
	return &models.Contract{
		ContractID:     contract.Name,
		Flavour:        *ParseFlavour(&contract.Spec.Flavour),
		Buyer:          ParseNodeIdentity(contract.Spec.Buyer),
		BuyerClusterID: contract.Spec.BuyerClusterID,
		TransactionID:  contract.Spec.TransactionID,
		Partition: func() *models.Partition {
			if contract.Spec.Partition != nil {
				return ParsePartition(contract.Spec.Partition)
			}
			return nil
		}(),
		Seller: ParseNodeIdentity(contract.Spec.Seller),
		SellerCredentials: models.LiqoCredentials{
			ClusterID:   contract.Spec.SellerCredentials.ClusterID,
			ClusterName: contract.Spec.SellerCredentials.ClusterName,
			Token:       contract.Spec.SellerCredentials.Token,
			Endpoint:    contract.Spec.SellerCredentials.Endpoint,
		},
		ExpirationTime:   contract.Spec.ExpirationTime,
		ExtraInformation: contract.Spec.ExtraInformation,
	}
}

// ParseQuantityFromString parses a string into a resource.Quantity.
func ParseQuantityFromString(s string) resource.Quantity {
	i, err := resource.ParseQuantity(s)
	if err != nil {
		return *resource.NewQuantity(0, resource.DecimalSI)
	}
	return i
}
