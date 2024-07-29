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
				Architecture: nil,
				CPU:          nil,
				Memory:       nil,
				Pods:         nil,
				Storage:      nil,
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

// ParseStringFilter parses a filter of type StringFilter into a StringFilter model.
func ParseStringFilter(filter *nodecorev1alpha1.StringFilter) (models.StringFilter, error) {
	var filterModel models.StringFilter

	if filter == nil {
		return filterModel, nil
	}

	klog.Infof("Parsing the filter %s", filter.Name)
	switch filter.Name {
	case nodecorev1alpha1.TypeMatchFilter:
		klog.Info("Parsing the filter as a MatchFilter")
		// Unmarshal the data into a StringMatchSelector
		var matchFilter nodecorev1alpha1.StringMatchSelector
		err := json.Unmarshal(filter.Data.Raw, &matchFilter)
		if err != nil {
			return filterModel, err
		}

		matchFilterData := models.StringMatchFilter{
			Value: matchFilter.Value,
		}
		// Marshal the filter data into JSON
		matchFilterDataJSON, err := json.Marshal(matchFilterData)
		if err != nil {
			return filterModel, err
		}

		// Generate the model for the filter
		filterModel = models.StringFilter{
			Name: models.MatchFilter,
			Data: matchFilterDataJSON,
		}
		klog.Infof("Filter model: %v", filterModel)
	case nodecorev1alpha1.TypeRangeFilter:
		klog.Info("Parsing the filter as a RangeFilter")
		// Unmarshal the data into a StringRangeSelector
		var rangeFilter nodecorev1alpha1.StringRangeSelector
		err := json.Unmarshal(filter.Data.Raw, &rangeFilter)
		if err != nil {
			return filterModel, err
		}

		rangeFilterData := models.StringRangeFilter{
			Regex: rangeFilter.Regex,
		}

		// Marshal the filter data into JSON
		rangeFilterDataJSON, err := json.Marshal(rangeFilterData)
		if err != nil {
			return filterModel, err
		}

		// Generate the model for the filter
		filterModel = models.StringFilter{
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
	var architectureFilterModel models.StringFilter
	var cpuFilterModel, memoryFilterModel, podsFilterModel, storageFilterModel models.ResourceQuantityFilter

	// Parse the Architecture filter
	klog.Info("Parsing the Architecture filter")
	architectureFilterModel, err := ParseStringFilter(k8sSelector.ArchitectureFilter)
	if err != nil {
		return nil, err
	}
	if architectureFilterModel.Data == nil {
		klog.Info("Architecture filter is nil")
	}

	// Parse the CPU filter
	klog.Info("Parsing the CPU filter")
	cpuFilterModel, err = ParseResourceQuantityFilter(k8sSelector.CPUFilter)
	if err != nil {
		return nil, err
	}
	if cpuFilterModel.Data == nil {
		klog.Info("CPU filter is nil")
	}

	// Parse the Memory filter
	klog.Info("Parsing the Memory filter")
	memoryFilterModel, err = ParseResourceQuantityFilter(k8sSelector.MemoryFilter)
	if err != nil {
		return nil, err
	}
	if memoryFilterModel.Data == nil {
		klog.Info("Memory filter is nil")
	}

	// Parse the Pods filter
	klog.Info("Parsing the Pods filter")
	podsFilterModel, err = ParseResourceQuantityFilter(k8sSelector.PodsFilter)
	if err != nil {
		return nil, err
	}
	if podsFilterModel.Data == nil {
		klog.Info("Pods filter is nil")
	}

	// Parse the Storage filter
	klog.Info("Parsing the Storage filter")
	storageFilterModel, err = ParseResourceQuantityFilter(k8sSelector.StorageFilter)
	if err != nil {
		return nil, err
	}
	if storageFilterModel.Data == nil {
		klog.Info("Storage filter is nil")
	}

	// Generate the model for the K8Slice selector
	k8SliceSelector := models.K8SliceSelector{
		Architecture: func() *models.StringFilter {
			if k8sSelector.ArchitectureFilter != nil {
				return &architectureFilterModel
			}
			return nil
		}(),
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

// ParseSourceDestination creates a SourceDestination Object from a SourceDestination CR.
func ParseSourceDestination(sourceDestination nodecorev1alpha1.SourceDestination) (*models.SourceDestination, error) {
	var resourceSelector models.ResourceSelector
	// Parse the ResourceSelector
	typeIdentifier, selector, err := nodecorev1alpha1.ParseResourceSelector(sourceDestination.ResourceSelector)
	if err != nil {
		return nil, err
	}
	switch typeIdentifier {
	case nodecorev1alpha1.CIDRSelectorType:
		// Force casting of selector to CIDRSelector
		cidrSelector := selector.(nodecorev1alpha1.CIDRSelector)
		// Marshal the CIDRSelector to JSON
		cidrSelectorData, err := json.Marshal(cidrSelector)
		if err != nil {
			return nil, err
		}
		resourceSelector = models.ResourceSelector{
			TypeIdentifier: models.CIDRSelectorType,
			Selector:       cidrSelectorData,
		}
	case nodecorev1alpha1.PodNamespaceSelectorType:
		// Force casting of selector to PodNamespaceSelector
		podNamespaceSelector := selector.(nodecorev1alpha1.PodNamespaceSelector)
		// Marshal the PodNamespaceSelector to JSON
		podNamespaceSelectorData, err := json.Marshal(podNamespaceSelector)
		if err != nil {
			return nil, err
		}
		resourceSelector = models.ResourceSelector{
			TypeIdentifier: models.CIDRSelectorType,
			Selector:       podNamespaceSelectorData,
		}
	default:
		return nil, fmt.Errorf("unknown resource selector type")
	}
	sourceDestinationModel := models.SourceDestination{
		IsHotCluster:     sourceDestination.IsHotCluster,
		ResourceSelector: resourceSelector,
	}

	return &sourceDestinationModel, nil
}

// ParseNetworkAuthorizations creates a NetworkAuthorizations Object from a NetworkAuthorizations CR.
func ParseNetworkAuthorizations(networkAuthorizations *nodecorev1alpha1.NetworkAuthorizations) (*models.NetworkAuthorizations, error) {
	var deniedCommunications []models.NetworkIntent
	var mandatoryCommunications []models.NetworkIntent

	// DeniedCommuncations
	for i := range networkAuthorizations.DeniedCommunications {
		// Parse the NetworkIntent
		networkIntent := networkAuthorizations.DeniedCommunications[i]
		source, err := ParseSourceDestination(networkIntent.Source)
		if err != nil {
			return nil, err
		}
		destination, err := ParseSourceDestination(networkIntent.Destination)
		if err != nil {
			return nil, err
		}
		deniedCommunications = append(deniedCommunications, models.NetworkIntent{
			Name:            networkIntent.Name,
			Source:          *source,
			Destination:     *destination,
			DestinationPort: networkIntent.DestinationPort,
			ProtocolType:    networkIntent.ProtocolType,
		})
	}

	// MandatoryCommunications
	for i := range networkAuthorizations.MandatoryCommunications {
		// Parse the NetworkIntent
		networkIntent := networkAuthorizations.MandatoryCommunications[i]
		source, err := ParseSourceDestination(networkIntent.Source)
		if err != nil {
			return nil, err
		}
		destination, err := ParseSourceDestination(networkIntent.Destination)
		if err != nil {
			return nil, err
		}
		mandatoryCommunications = append(mandatoryCommunications, models.NetworkIntent{
			Name:            networkIntent.Name,
			Source:          *source,
			Destination:     *destination,
			DestinationPort: networkIntent.DestinationPort,
			ProtocolType:    networkIntent.ProtocolType,
		})
	}

	return &models.NetworkAuthorizations{
		DeniedCommunications:    deniedCommunications,
		MandatoryCommunications: mandatoryCommunications,
	}, nil
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
				Architecture: flavorTypeStruct.Characteristics.Architecture,
				CPU:          flavorTypeStruct.Characteristics.CPU,
				Memory:       flavorTypeStruct.Characteristics.Memory,
				Pods:         flavorTypeStruct.Characteristics.Pods,
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
				NetworkAuthorizations: func() *models.NetworkAuthorizations {
					if flavorTypeStruct.Properties.NetworkAuthorizations != nil {
						na, err := ParseNetworkAuthorizations(flavorTypeStruct.Properties.NetworkAuthorizations)
						if err != nil {
							return nil
						}
						return na
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
		NetworkRequests:  contract.Spec.NetworkRequests,
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
