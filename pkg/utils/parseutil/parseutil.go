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
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/models"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ParseFlavourSelector parses FlavourSelector into a Selector
func ParseFlavourSelector(selector *nodecorev1alpha1.FlavourSelector) (s *models.Selector) {

	s.Architecture = selector.Architecture
	s.FlavourType = string(selector.FlavourType)

	if selector.MatchSelector != nil {
		s.MatchSelector = &models.MatchSelector{
			Cpu:              selector.MatchSelector.Cpu,
			Memory:           selector.MatchSelector.Memory,
			EphemeralStorage: selector.MatchSelector.EphemeralStorage,
			Storage:          selector.MatchSelector.Storage,
			Gpu:              selector.MatchSelector.Gpu,
		}
	}

	if selector.RangeSelector != nil {
		s.RangeSelector = &models.RangeSelector{
			MinCpu:     selector.RangeSelector.MinCpu,
			MinMemory:  selector.RangeSelector.MinMemory,
			MinEph:     selector.RangeSelector.MinEph,
			MinStorage: selector.RangeSelector.MinStorage,
			MinGpu:     selector.RangeSelector.MinGpu,
			MaxCpu:     selector.RangeSelector.MaxCpu,
			MaxMemory:  selector.RangeSelector.MaxMemory,
			MaxEph:     selector.RangeSelector.MaxEph,
			MaxStorage: selector.RangeSelector.MaxStorage,
			MaxGpu:     selector.RangeSelector.MaxGpu,
		}
	}

	return
}

func ParsePartition(partition *reservationv1alpha1.Partition) *models.Partition {
	return &models.Partition{
		Cpu:              partition.Cpu,
		Memory:           partition.Memory,
		EphemeralStorage: partition.EphemeralStorage,
		Storage:          partition.Storage,
		Gpu:              partition.Gpu,
	}
}

func ParsePartitionFromObj(partition *models.Partition) *reservationv1alpha1.Partition {
	return &reservationv1alpha1.Partition{
		Architecture:     partition.Architecture,
		Cpu:              partition.Cpu,
		Memory:           partition.Memory,
		Gpu:              partition.Gpu,
		Storage:          partition.Storage,
		EphemeralStorage: partition.EphemeralStorage,
	}
}

func ParseNodeIdentity(node nodecorev1alpha1.NodeIdentity) models.NodeIdentity {
	return models.NodeIdentity{
		NodeID: node.NodeID,
		IP:     node.IP,
		Domain: node.Domain,
	}
}

// ParseFlavourObject creates a Flavour Object from a Flavour CR
func ParseFlavour(flavour nodecorev1alpha1.Flavour) models.Flavour {
	return models.Flavour{
		FlavourID:  flavour.Name,
		Type:       string(flavour.Spec.Type),
		ProviderID: flavour.Spec.ProviderID,
		Characteristics: models.Characteristics{
			CPU:               flavour.Spec.Characteristics.Cpu,
			Memory:            flavour.Spec.Characteristics.Memory,
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
						CPUStep:       flavour.Spec.Policy.Partitionable.CpuStep,
						MemoryStep:    flavour.Spec.Policy.Partitionable.MemoryStep,
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

// ForgeContractObject creates a Contract Object
func ParseContract(contract *reservationv1alpha1.Contract) models.Contract {
	return models.Contract{
		ContractID:     contract.Name,
		Flavour:        ParseFlavour(contract.Spec.Flavour),
		Buyer:          ParseNodeIdentity(contract.Spec.Buyer),
		BuyerClusterID: contract.Spec.BuyerClusterID,
		TransactionID:  contract.Spec.TransactionID,
		Partition:      ParsePartition(contract.Spec.Partition),
		Seller:         ParseNodeIdentity(contract.Spec.Seller),
		SellerCredentials: models.LiqoCredentials{
			ClusterID:   contract.Spec.SellerCredentials.ClusterID,
			ClusterName: contract.Spec.SellerCredentials.ClusterName,
			Token:       contract.Spec.SellerCredentials.Token,
			Endpoint:    contract.Spec.SellerCredentials.Endpoint,
		},
	}
}

func ParseQuantityFromString(s string) resource.Quantity {
	i, err := resource.ParseQuantity(s)
	if err != nil {
		return *resource.NewQuantity(0, resource.DecimalSI)
	}
	return i
}
