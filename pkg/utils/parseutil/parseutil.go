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

package parseutil

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/models"
)

// ParseFlavorSelector parses FlavorSelector into a Selector.
func ParseFlavorSelector(selector *nodecorev1alpha1.Selector) (models.Selector, error) {
	// Parse the Selector
	klog.Infof("Parsing the selector %s", selector.FlavorType)
	selectorIdentifier, selectorStruct, err := nodecorev1alpha1.ParseSolverSelector(selector)
	if err != nil {
		return nil, err
	}

	klog.Infof("Selector type: %s", selectorIdentifier)

	switch selectorIdentifier {
	case nodecorev1alpha1.TypeK8Slice:
		// Check if the selectorStruct is nil
		if selectorStruct == nil {
			return models.K8SliceSelector{
				CPU:     nil,
				Memory:  nil,
				Pods:    nil,
				Storage: nil,
			}, nil
		}
		// Force casting of selectorStruct to K8Slice
		selectorStruct := selectorStruct.(nodecorev1alpha1.K8SliceSelector)
		klog.Info("Forced casting of selectorStruct to K8Slice")
		// Print the selectorStruct
		klog.Infof("SelectorStruct: %v", selectorStruct)
		// Generate the model for the K8Slice selector
		k8SliceSelector, err := parseK8SliceFilters(&selectorStruct)
		if err != nil {
			return nil, err
		}

		klog.Infof("K8SliceSelector: %v", k8SliceSelector)

		return *k8SliceSelector, nil

	case nodecorev1alpha1.TypeVM:
		// Force casting of selectorStruct to VM
		// TODO: Implement the parsing of the VM selector
		return nil, fmt.Errorf("VM selector not implemented")

	case nodecorev1alpha1.TypeService:
		// Force casting of selectorStruct to Service
		// TODO: Implement the parsing of the Service selector
		return nil, fmt.Errorf("service selector not implemented")

	default:
		return nil, fmt.Errorf("unknown selector type")
	}
}

// ParseResourceQuantityFilter parses a filter of type ResourceQuantityFilter into a ResourceQuantityFilter model.
func ParseResourceQuantityFilter(filter *nodecorev1alpha1.ResourceQuantityFilter) (models.ResourceQuantityFilter, error) {
	var filterModel models.ResourceQuantityFilter

	if filter == nil {
		return filterModel, nil
	}

	klog.Infof("Parsing the filter %s", filter.Name)
	switch filter.Name {
	case nodecorev1alpha1.TypeMatchFilter:
		klog.Info("Parsing the filter as a MatchFilter")
		// Unmarshal the data into a ResourceMatchSelector
		var matchFilter nodecorev1alpha1.ResourceMatchSelector
		err := json.Unmarshal(filter.Data.Raw, &matchFilter)
		if err != nil {
			return filterModel, err
		}

		matchFilterData := models.ResourceQuantityMatchFilter{
			Value: matchFilter.Value.DeepCopy(),
		}
		// Marshal the filter data into JSON
		matchFilterDataJSON, err := json.Marshal(matchFilterData)
		if err != nil {
			return filterModel, err
		}

		// Generate the model for the filter
		filterModel = models.ResourceQuantityFilter{
			Name: models.MatchFilter,
			Data: matchFilterDataJSON,
		}
		klog.Infof("Filter model: %v", filterModel)
	case nodecorev1alpha1.TypeRangeFilter:
		klog.Info("Parsing the filter as a RangeFilter")
		// Unmarshal the data into a ResourceRangeSelector
		var rangeFilter nodecorev1alpha1.ResourceRangeSelector
		err := json.Unmarshal(filter.Data.Raw, &rangeFilter)
		if err != nil {
			return filterModel, err
		}

		rangeFilterData := models.ResourceQuantityRangeFilter{
			Min: rangeFilter.Min,
			Max: rangeFilter.Max,
		}

		// Marshal the filter data into JSON
		rangeFilterDataJSON, err := json.Marshal(rangeFilterData)
		if err != nil {
			return filterModel, err
		}

		// Generate the model for the filter
		filterModel = models.ResourceQuantityFilter{
			Name: models.RangeFilter,
			Data: rangeFilterDataJSON,
		}
		klog.Infof("Filter model: %v", filterModel)
	default:
		return filterModel, fmt.Errorf("unknown filter type")
	}

	return filterModel, nil
}

func parseK8SliceFilters(k8sSelector *nodecorev1alpha1.K8SliceSelector) (*models.K8SliceSelector, error) {
	var cpuFilterModel, memoryFilterModel, podsFilterModel, storageFilterModel models.ResourceQuantityFilter

	// Parse the CPU filter
	klog.Info("Parsing the CPU filter")
	cpuFilterModel, err := ParseResourceQuantityFilter(k8sSelector.CPUFilter)
	if err != nil {
		return nil, err
	}
	if cpuFilterModel.Data == nil {
		klog.Info("CPU filter is nil")
	}

	if k8sSelector.MemoryFilter != nil {
		klog.Info("Parsing the Memory filter")
		// Parse the Memory filter
		switch k8sSelector.MemoryFilter.Name {
		case nodecorev1alpha1.TypeMatchFilter:
			klog.Info("Parsing the Memory filter as a MatchFilter")
			// Unmarshal the data into a ResourceMatchSelector
			var memoryFilter nodecorev1alpha1.ResourceMatchSelector
			err := json.Unmarshal(k8sSelector.MemoryFilter.Data.Raw, &memoryFilter)
			if err != nil {
				return nil, err
			}

			memoryFilterData := models.ResourceQuantityMatchFilter{
				Value: memoryFilter.Value.DeepCopy(),
			}
			// Marshal the Memory filter data into JSON
			memoryFilterDataJSON, err := json.Marshal(memoryFilterData)
			if err != nil {
				return nil, err
			}

			// Generate the model for the Memory filter
			memoryFilterModel = models.ResourceQuantityFilter{
				Name: models.MatchFilter,
				Data: memoryFilterDataJSON,
			}
			klog.Infof("Memory filter model: %v", memoryFilterModel)
		case nodecorev1alpha1.TypeRangeFilter:
			klog.Info("Parsing the Memory filter as a RangeFilter")
			// Unmarshal the data into a ResourceRangeSelector
			var memoryFilter nodecorev1alpha1.ResourceRangeSelector
			err := json.Unmarshal(k8sSelector.MemoryFilter.Data.Raw, &memoryFilter)
			if err != nil {
				return nil, err
			}

			memoryFilterData := models.ResourceQuantityRangeFilter{
				Min: memoryFilter.Min,
				Max: memoryFilter.Max,
			}
			// Marshal the Memory filter data into JSON
			memoryFilterDataJSON, err := json.Marshal(memoryFilterData)
			if err != nil {
				return nil, err
			}

			// Generate the model for the Memory filter
			memoryFilterModel = models.ResourceQuantityFilter{
				Name: models.RangeFilter,
				Data: memoryFilterDataJSON,
			}
			klog.Infof("Memory filter model: %v", memoryFilterModel)
		default:
			return nil, fmt.Errorf("unknown filter type")
		}
	}

	if k8sSelector.PodsFilter != nil {
		klog.Info("Parsing the Pods filter")
		// Parse the Pods filter
		switch k8sSelector.PodsFilter.Name {
		case nodecorev1alpha1.TypeMatchFilter:
			klog.Info("Parsing the Pods filter as a MatchFilter")
			// Unmarshal the data into a ResourceMatchSelector
			var podsFilter nodecorev1alpha1.ResourceMatchSelector
			err := json.Unmarshal(k8sSelector.PodsFilter.Data.Raw, &podsFilter)
			if err != nil {
				return nil, err
			}

			podsFilterData := models.ResourceQuantityMatchFilter{
				Value: podsFilter.Value.DeepCopy(),
			}
			// Marshal the Pods filter data into JSON
			podsFilterDataJSON, err := json.Marshal(podsFilterData)
			if err != nil {
				return nil, err
			}

			// Generate the model for the Pods filter
			podsFilterModel = models.ResourceQuantityFilter{
				Name: models.MatchFilter,
				Data: podsFilterDataJSON,
			}
			klog.Infof("Pods filter model: %v", podsFilterModel)
		case nodecorev1alpha1.TypeRangeFilter:
			klog.Info("Parsing the Pods filter as a RangeFilter")
			// Unmarshal the data into a ResourceRangeSelector
			var podsFilter nodecorev1alpha1.ResourceRangeSelector
			err := json.Unmarshal(k8sSelector.PodsFilter.Data.Raw, &podsFilter)
			if err != nil {
				return nil, err
			}

			podsFilterData := models.ResourceQuantityRangeFilter{
				Min: podsFilter.Min,
				Max: podsFilter.Max,
			}
			// Marshal the Pods filter data into JSON
			podsFilterDataJSON, err := json.Marshal(podsFilterData)
			if err != nil {
				return nil, err
			}

			// Generate the model for the Pods filter
			podsFilterModel = models.ResourceQuantityFilter{
				Name: models.RangeFilter,
				Data: podsFilterDataJSON,
			}
			klog.Infof("Pods filter model: %v", podsFilterModel)
		default:
			return nil, fmt.Errorf("unknown filter type")
		}
	}

	if k8sSelector.StorageFilter != nil {
		klog.Info("Parsing the Storage filter")
		// Parse the Storage filter
		switch k8sSelector.StorageFilter.Name {
		case nodecorev1alpha1.TypeMatchFilter:
			klog.Info("Parsing the Storage filter as a MatchFilter")
			// Unmarshal the data into a ResourceMatchSelector
			var storageFilter nodecorev1alpha1.ResourceMatchSelector
			err := json.Unmarshal(k8sSelector.StorageFilter.Data.Raw, &storageFilter)
			if err != nil {
				return nil, err
			}

			storageFilterData := models.ResourceQuantityMatchFilter{
				Value: storageFilter.Value.DeepCopy(),
			}
			// Marshal the Storage filter data into JSON
			storageFilterDataJSON, err := json.Marshal(storageFilterData)
			if err != nil {
				return nil, err
			}

			// Generate the model for the Storage filter
			storageFilterModel = models.ResourceQuantityFilter{
				Name: models.MatchFilter,
				Data: storageFilterDataJSON,
			}
			klog.Infof("Storage filter model: %v", storageFilterModel)
		case nodecorev1alpha1.TypeRangeFilter:
			klog.Info("Parsing the Storage filter as a RangeFilter")
			// Unmarshal the data into a ResourceRangeSelector
			var storageFilter nodecorev1alpha1.ResourceRangeSelector
			err := json.Unmarshal(k8sSelector.StorageFilter.Data.Raw, &storageFilter)
			if err != nil {
				return nil, err
			}

			storageFilterData := models.ResourceQuantityRangeFilter{
				Min: storageFilter.Min,
				Max: storageFilter.Max,
			}
			// Marshal the Storage filter data into JSON
			storageFilterDataJSON, err := json.Marshal(storageFilterData)
			if err != nil {
				return nil, err
			}

			// Generate the model for the Storage filter
			storageFilterModel = models.ResourceQuantityFilter{
				Name: models.RangeFilter,
				Data: storageFilterDataJSON,
			}
			klog.Infof("Storage filter model: %v", storageFilterModel)
		default:
			return nil, fmt.Errorf("unknown filter type")
		}
	}

	// Generate the model for the K8Slice selector
	k8SliceSelector := models.K8SliceSelector{
		CPU: func() *models.ResourceQuantityFilter {
			if k8sSelector.CPUFilter != nil {
				return &cpuFilterModel
			}
			return nil
		}(),
		Memory: func() *models.ResourceQuantityFilter {
			if k8sSelector.MemoryFilter != nil {
				return &memoryFilterModel
			}
			return nil
		}(),
		Pods: func() *models.ResourceQuantityFilter {
			if k8sSelector.PodsFilter != nil {
				return &podsFilterModel
			}
			return nil
		}(),
		Storage: func() *models.ResourceQuantityFilter {
			if k8sSelector.StorageFilter != nil {
				return &storageFilterModel
			}
			return nil
		}(),
	}

	return &k8SliceSelector, nil
}

// ParseConfiguration creates a Configuration Object from a Configuration CR.
func ParseConfiguration(configuration *nodecorev1alpha1.Configuration) *models.Configuration {
	// Parse the Configuration
	configurationType, configurationStruct, err := nodecorev1alpha1.ParseConfiguration(configuration)
	if err != nil {
		return nil
	}

	switch configurationType {
	case nodecorev1alpha1.TypeK8Slice:
		// Force casting of configuration to K8Slice
		configuration := configurationStruct.(nodecorev1alpha1.K8SliceConfiguration)
		k8sliceConfigurationJSON := models.K8SliceConfiguration{
			CPU:    configuration.CPU,
			Memory: configuration.Memory,
			Pods:   configuration.Pods,
			Gpu: func() *models.GpuCharacteristics {
				if configuration.Gpu != nil {
					return &models.GpuCharacteristics{
						Model:  configuration.Gpu.Model,
						Cores:  configuration.Gpu.Cores,
						Memory: configuration.Gpu.Memory,
					}
				}
				return nil
			}(),
			Storage: configuration.Storage,
		}
		// Marshal the K8Slice configuration to JSON
		configurationData, err := json.Marshal(k8sliceConfigurationJSON)
		if err != nil {
			klog.Errorf("Error marshaling the K8Slice configuration: %s", err)
			return nil
		}
		return &models.Configuration{
			Type: models.K8SliceNameDefault,
			Data: configurationData,
		}
		// TODO: Implement the other configuration types, if any
	default:
		return nil
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

// ParseFlavor creates a Flavor Object from a Flavor CR.
func ParseFlavor(flavor *nodecorev1alpha1.Flavor) *models.Flavor {
	var modelFlavor models.Flavor

	flavorType, flavorTypeStruct, errParse := nodecorev1alpha1.ParseFlavorType(flavor)
	if errParse != nil {
		return nil
	}

	var modelFlavorType models.FlavorType

	switch flavorType {
	case nodecorev1alpha1.TypeK8Slice:
		// Force casting of flavorTypeStruct to K8Slice
		flavorTypeStruct := flavorTypeStruct.(nodecorev1alpha1.K8Slice)
		modelFlavorTypeData := models.K8Slice{
			Characteristics: models.K8SliceCharacteristics{
				CPU:    flavorTypeStruct.Characteristics.CPU,
				Memory: flavorTypeStruct.Characteristics.Memory,
				Pods:   flavorTypeStruct.Characteristics.Pods,
				Gpu: func() *models.GpuCharacteristics {
					if flavorTypeStruct.Characteristics.Gpu != nil {
						return &models.GpuCharacteristics{
							Model:  flavorTypeStruct.Characteristics.Gpu.Model,
							Cores:  flavorTypeStruct.Characteristics.Gpu.Cores,
							Memory: flavorTypeStruct.Characteristics.Gpu.Memory,
						}
					}
					return nil
				}(),
				Storage: flavorTypeStruct.Characteristics.Storage,
			},
			Properties: models.K8SliceProperties{
				Latency:           flavorTypeStruct.Properties.Latency,
				SecurityStandards: flavorTypeStruct.Properties.SecurityStandards,
				CarbonFootprint: func() *models.CarbonFootprint {
					if flavorTypeStruct.Properties.CarbonFootprint != nil {
						return &models.CarbonFootprint{
							Embodied:    flavorTypeStruct.Properties.CarbonFootprint.Embodied,
							Operational: flavorTypeStruct.Properties.CarbonFootprint.Operational,
						}
					}
					return nil
				}(),
			},
			Policies: models.K8SlicePolicies{
				Partitionability: models.K8SlicePartitionability{
					CPUMin:     flavorTypeStruct.Policies.Partitionability.CPUMin,
					MemoryMin:  flavorTypeStruct.Policies.Partitionability.MemoryMin,
					PodsMin:    flavorTypeStruct.Policies.Partitionability.PodsMin,
					CPUStep:    flavorTypeStruct.Policies.Partitionability.CPUStep,
					MemoryStep: flavorTypeStruct.Policies.Partitionability.MemoryStep,
					PodsStep:   flavorTypeStruct.Policies.Partitionability.PodsStep,
				},
			},
		}

		// Encode the K8Slice data into JSON
		encodedFlavorTypeData, err := json.Marshal(modelFlavorTypeData)
		if err != nil {
			klog.Errorf("Error encoding the K8Slice data: %s", err)
			return nil
		}

		modelFlavorType = models.FlavorType{
			Name: models.K8SliceNameDefault,
			Data: encodedFlavorTypeData,
		}
	// TODO: Implement the other flavor types
	default:
		klog.Errorf("Unknown flavor type: %s", flavorType)
		return nil
	}

	modelFlavor = models.Flavor{
		FlavorID:            flavor.Name,
		Type:                modelFlavorType,
		ProviderID:          flavor.Spec.ProviderID,
		NetworkPropertyType: flavor.Spec.NetworkPropertyType,
		Timestamp:           flavor.CreationTimestamp.Time,
		Location: func() *models.Location {
			if flavor.Spec.Location != nil {
				location := models.Location{
					Latitude:        flavor.Spec.Location.Latitude,
					Longitude:       flavor.Spec.Location.Longitude,
					Country:         flavor.Spec.Location.Country,
					City:            flavor.Spec.Location.City,
					AdditionalNotes: flavor.Spec.Location.AdditionalNotes,
				}
				return &location
			}
			return nil
		}(),
		Price: models.Price{
			Amount:   flavor.Spec.Price.Amount,
			Currency: flavor.Spec.Price.Currency,
			Period:   flavor.Spec.Price.Period,
		},
		Owner:        ParseNodeIdentity(flavor.Spec.Owner),
		Availability: flavor.Spec.Availability,
	}

	return &modelFlavor
}

// ParseContract creates a Contract Object.
func ParseContract(contract *reservationv1alpha1.Contract) *models.Contract {
	return &models.Contract{
		ContractID:     contract.Name,
		Flavor:         *ParseFlavor(&contract.Spec.Flavor),
		Buyer:          ParseNodeIdentity(contract.Spec.Buyer),
		BuyerClusterID: contract.Spec.BuyerClusterID,
		TransactionID:  contract.Spec.TransactionID,
		Configuration: func() *models.Configuration {
			if contract.Spec.Configuration != nil {
				configuration := ParseConfiguration(contract.Spec.Configuration)
				return configuration
			}
			return nil
		}(),
		Seller: ParseNodeIdentity(contract.Spec.Seller),
		PeeringTargetCredentials: models.LiqoCredentials{
			ClusterID:   contract.Spec.PeeringTargetCredentials.ClusterID,
			ClusterName: contract.Spec.PeeringTargetCredentials.ClusterName,
			Token:       contract.Spec.PeeringTargetCredentials.Token,
			Endpoint:    contract.Spec.PeeringTargetCredentials.Endpoint,
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
