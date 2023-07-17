package parseutil

import (
	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	reservationv1alpha1 "fluidos.eu/node/api/reservation/v1alpha1"
	"fluidos.eu/node/pkg/utils/models"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ParseFlavourSelector parses FlavourSelector into a Selector
func ParseFlavourSelector(selector nodecorev1alpha1.FlavourSelector) (s models.Selector) {

	s.Architecture = selector.Architecture
	s.FlavourType = string(selector.FlavourType)

	if selector.MatchSelector != nil {
		cpu, _ := selector.MatchSelector.Cpu.AsInt64()
		memory, _ := selector.MatchSelector.Memory.AsInt64()
		ephStorage, _ := selector.MatchSelector.EphemeralStorage.AsInt64()
		storage, _ := selector.MatchSelector.Storage.AsInt64()
		gpu, _ := selector.MatchSelector.Gpu.AsInt64()

		s.MatchSelector = &models.MatchSelector{
			Cpu:              int(cpu),
			Memory:           int(memory),
			EphemeralStorage: int(ephStorage),
			Storage:          int(storage),
			Gpu:              int(gpu),
		}
	}

	if selector.RangeSelector != nil {
		moreThanCpu, _ := selector.RangeSelector.MoreThanCPU.AsInt64()
		moreThanMemory, _ := selector.RangeSelector.MoreThanMemory.AsInt64()
		moreThanEph, _ := selector.RangeSelector.MoreThanEph.AsInt64()
		moreThanStorage, _ := selector.RangeSelector.MoreThanStorage.AsInt64()
		moreThanGpu, _ := selector.RangeSelector.MoreThanGpu.AsInt64()
		lessThanCpu, _ := selector.RangeSelector.LessThanCPU.AsInt64()
		lessThanMemory, _ := selector.RangeSelector.LessThanMemory.AsInt64()
		lessThanEph, _ := selector.RangeSelector.LessThanEph.AsInt64()
		lessThanStorage, _ := selector.RangeSelector.LessThanStorage.AsInt64()
		lessThanGpu, _ := selector.RangeSelector.LessThanGpu.AsInt64()

		s.RangeSelector = &models.RangeSelector{
			MoreThanCPU:     int(moreThanCpu),
			MoreThanMemory:  int(moreThanMemory),
			MoreThanEph:     int(moreThanEph),
			MoreThanStorage: int(moreThanStorage),
			MoreThanGpu:     int(moreThanGpu),
			LessThanCPU:     int(lessThanCpu),
			LessThanMemory:  int(lessThanMemory),
			LessThanEph:     int(lessThanEph),
			LessThanStorage: int(lessThanStorage),
			LessThanGpu:     int(lessThanGpu),
		}
	}

	return
}

func ParsePartition(partition reservationv1alpha1.Partition) models.Partition {
	cpu, _ := partition.Cpu.AsInt64()
	memory, _ := partition.Memory.AsInt64()
	ephStorage, _ := partition.EphemeralStorage.AsInt64()
	storage, _ := partition.PersistentStorage.AsInt64()
	gpu, _ := partition.Gpu.AsInt64()

	return models.Partition{
		Cpu:              int(cpu),
		Memory:           int(memory),
		EphemeralStorage: int(ephStorage),
		Storage:          int(storage),
		Gpu:              int(gpu),
	}
}

func ParsePartitionFromObj(partition models.Partition) reservationv1alpha1.Partition {
	p := reservationv1alpha1.Partition{
		Architecture: partition.Architecture,
		Cpu:          *resource.NewQuantity(int64(partition.Cpu), resource.DecimalSI),
		Memory:       *resource.NewQuantity(int64(partition.Memory), resource.BinarySI),
	}
	if partition.EphemeralStorage != 0 {
		p.EphemeralStorage = *resource.NewQuantity(int64(partition.EphemeralStorage), resource.BinarySI)
	}
	if partition.Storage != 0 {
		p.PersistentStorage = *resource.NewQuantity(int64(partition.Storage), resource.BinarySI)
	}
	if partition.Gpu != 0 {
		p.Gpu = *resource.NewQuantity(int64(partition.Gpu), resource.DecimalSI)
	}
	return p
}

func ParseNodeIdentity(node nodecorev1alpha1.NodeIdentity) models.Owner {
	return models.Owner{
		NodeID: node.NodeID,
		IP:     node.IP,
		Domain: node.Domain,
	}
}

// ParseFlavourObject creates a Flavour Object from a Flavour CR
func ParseFlavour(flavour nodecorev1alpha1.Flavour) models.Flavour {
	cpu, _ := flavour.Spec.Characteristics.Cpu.AsInt64()
	ram, _ := flavour.Spec.Characteristics.Memory.AsInt64()
	obj := models.Flavour{
		FlavourID:  flavour.Name,
		Type:       string(flavour.Spec.Type),
		ProviderID: flavour.Spec.ProviderID,
		Characteristics: models.Characteristics{
			CPU:    int(cpu),
			Memory: int(ram),
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
	if ephStorage, ok := flavour.Spec.Characteristics.EphemeralStorage.AsInt64(); ok == true && ephStorage != 0 {
		obj.Characteristics.EphemeralStorage = int(ephStorage)
	}
	if storage, ok := flavour.Spec.Characteristics.PersistentStorage.AsInt64(); ok == true && storage != 0 {
		obj.Characteristics.PersistentStorage = int(storage)
	}
	if gpu, ok := flavour.Spec.Characteristics.Gpu.AsInt64(); ok == true && gpu != 0 {
		obj.Characteristics.GPU = int(gpu)
	}
	if flavour.Spec.Characteristics.Architecture != "" {
		obj.Characteristics.Architecture = flavour.Spec.Characteristics.Architecture
	}
	return obj
}

// ForgeContractObject creates a Contract Object
func ParseContract(contract *reservationv1alpha1.Contract) models.Contract {
	return models.Contract{
		ContractID:    contract.Name,
		Flavour:       ParseFlavour(contract.Spec.Flavour),
		Buyer:         ParseNodeIdentity(contract.Spec.Buyer),
		TransactionID: contract.Spec.TransactionID,
		Partition:     ParsePartition(contract.Spec.Partition),
	}
}

// ParseFlavourSpecToFlavour converts a FlavourSpec to a Flavour struct
/* func ParseCRToFlavour(flavourCR nodecorev1alpha1.Flavour) *models.Flavour {

	// It is converted in int since REAR for now only supports int
	cpuInt, _ := flavourCR.Spec.Characteristics.Cpu.AsInt64()
	memoryInt, _ := flavourCR.Spec.Characteristics.Memory.AsInt64()
	ephInt, _ := flavourCR.Spec.Characteristics.EphemeralStorage.AsInt64()
	storageInt, _ := flavourCR.Spec.Characteristics.PersistentStorage.AsInt64()
	gpuInt, _ := flavourCR.Spec.Characteristics.Gpu.AsInt64()

	return &models.Flavour{
		FlavourID:  flavourCR.Name,
		ProviderID: flavourCR.Spec.ProviderID,
		Type:       string(flavourCR.Spec.Type),
		Characteristics: models.Characteristics{
			CPU:               int(cpuInt),
			Memory:            int(memoryInt),
			EphemeralStorage:  int(ephInt),
			PersistentStorage: int(storageInt),
			GPU:               int(gpuInt),
			Architecture:      flavourCR.Spec.Characteristics.Architecture,
		},
		Owner: models.Owner{
			NodeID: flavourCR.Spec.Owner.NodeID,
			IP:     flavourCR.Spec.Owner.IP,
			Domain: flavourCR.Spec.Owner.Domain,
		},
		Policy: models.Policy{
			Partitionable: func() *models.Partitionable {
				if flavourCR.Spec.Policy.Partitionable != nil {
					return &models.Partitionable{
						CPUMinimum:    flavourCR.Spec.Policy.Partitionable.CpuMin,
						MemoryMinimum: flavourCR.Spec.Policy.Partitionable.MemoryMin,
						CPUStep:       flavourCR.Spec.Policy.Partitionable.CpuStep,
						MemoryStep:    flavourCR.Spec.Policy.Partitionable.MemoryStep,
					}
				}
				return nil
			}(),
			Aggregatable: func() *models.Aggregatable {
				if flavourCR.Spec.Policy.Aggregatable != nil {
					return &models.Aggregatable{
						MinCount: flavourCR.Spec.Policy.Aggregatable.MinCount,
						MaxCount: flavourCR.Spec.Policy.Aggregatable.MaxCount,
					}
				}
				return nil
			}(),
		},
		Price: models.Price{
			Amount:   flavourCR.Spec.Price.Amount,
			Currency: flavourCR.Spec.Price.Currency,
			Period:   flavourCR.Spec.Price.Period,
		},
		OptionalFields: models.OptionalFields{
			Availability: flavourCR.Spec.OptionalFields.Availability,
		},
	}
}
*/
