package parseutil

import (
	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	"fluidos.eu/node/pkg/utils/flags"
	"fluidos.eu/node/pkg/utils/models"
)

// ParseSelectorValues parses the values of the selector in int64 (REAR only supports int64 for now)
func ParseSelectorValues(selector nodecorev1alpha1.FlavourSelector) map[string]interface{} {
	cpuInt, _ := selector.Cpu.AsInt64()
	ramInt, _ := selector.Memory.AsInt64()
	ephStorage, _ := selector.EphemeralStorage.AsInt64()

	return map[string]interface{}{
		"cpu":    int(cpuInt),
		"ram":    int(ramInt),
		"memory": int(ephStorage),
		"type":   flags.RESOURCE_TYPE,
	}
}

// ParseFlavourSpecToFlavour converts a FlavourSpec to a Flavour struct
func ParseCRToFlavour(flavourCR nodecorev1alpha1.Flavour) *models.Flavour {

	// It is converted in int since REAR for now only supports int
	cpuInt, _ := flavourCR.Spec.Characteristics.Cpu.AsInt64()
	ramInt, _ := flavourCR.Spec.Characteristics.Memory.AsInt64()
	ephInt, _ := flavourCR.Spec.Characteristics.EphemeralStorage.AsInt64()

	return &models.Flavour{
		FlavourID:  flavourCR.Name,
		ProviderID: flavourCR.Spec.ProviderID,
		Type:       string(flavourCR.Spec.Type),
		Characteristics: models.Characteristics{
			CPU:              int(cpuInt),
			RAM:              int(ramInt),
			EphemeralStorage: int(ephInt),
			Architecture:     flavourCR.Spec.Characteristics.Architecture,
		},
		Owner: models.Owner{
			ID:         flavourCR.Spec.Owner.NodeID,
			IP:         flavourCR.Spec.Owner.IP,
			DomainName: flavourCR.Spec.Owner.Domain,
		},
		Policy: models.Policy{

			Partitionable: func() *models.Partitionable {
				if flavourCR.Spec.Policy.Partitionable != nil {
					return &models.Partitionable{
						CPUMinimum: flavourCR.Spec.Policy.Partitionable.CpuMin,
						RAMMinimum: flavourCR.Spec.Policy.Partitionable.MemoryMin,
						CPUStep:    flavourCR.Spec.Policy.Partitionable.CpuStep,
						RAMStep:    flavourCR.Spec.Policy.Partitionable.MemoryStep,
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
