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

package grpc

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
)

func getContractResourcesByClusterID(cl client.Client, clusterID string) (map[string]*resource.Quantity, error) {
	var contracts reservationv1alpha1.ContractList

	if err := cl.List(context.Background(), &contracts, client.MatchingFields{"spec.buyerClusterID": clusterID}); err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Errorf("Error when listing Contracts: %s", err)
			return nil, err
		}
	}

	if len(contracts.Items) == 0 {
		klog.Errorf("No contracts found for cluster %s", clusterID)
		return nil, fmt.Errorf("no contracts found for cluster %s", clusterID)
	}

	if len(contracts.Items) > 1 {
		resources := multipleContractLogic(contracts.Items)
		return resources, nil
	}

	contract := contracts.Items[0]

	if contract.Spec.Configuration != nil {
		return addResources(make(map[string]*resource.Quantity), contract.Spec.Configuration), nil
	}

	return addResourceByFlavor(make(map[string]*resource.Quantity), &contract.Spec.Flavor), nil
}

func multipleContractLogic(contracts []reservationv1alpha1.Contract) map[string]*resource.Quantity {
	resources := make(map[string]*resource.Quantity)
	for i := range contracts {
		if contracts[i].Spec.Configuration != nil {
			resources = addResources(resources, contracts[i].Spec.Configuration)
		} else {
			resources = addResourceByFlavor(resources, &contracts[i].Spec.Flavor)
		}
	}
	return resources
}

func addResourceByFlavor(resources map[string]*resource.Quantity, flavor *nodecorev1alpha1.Flavor) map[string]*resource.Quantity {
	// Parse flavor
	flavorType, flavorData, err := nodecorev1alpha1.ParseFlavorType(flavor)
	if err != nil {
		klog.Errorf("Error when parsing flavor: %s", err)
		return nil
	}
	switch flavorType {
	case nodecorev1alpha1.TypeK8Slice:
		// Force casting
		k8sliceFlavor := flavorData.(nodecorev1alpha1.K8Slice)
		// For each characteristic, add the resource to the map
		for key, value := range mapK8SliceToResources(&k8sliceFlavor) {
			if prevRes, ok := resources[key]; !ok {
				resources[key] = value
			} else {
				prevRes.Add(*value)
				resources[key] = prevRes
			}
		}
	default:
		klog.Errorf("Flavor type %s not supported", flavorType)
		return nil
	}

	return resources
}

// This function adds the resources of a contract to the existing resourceList.
func addResources(resources map[string]*resource.Quantity, configuration *nodecorev1alpha1.Configuration) map[string]*resource.Quantity {
	// Parse configuration
	configurationType, configurationData, err := nodecorev1alpha1.ParseConfiguration(configuration)
	if err != nil {
		klog.Errorf("Error when parsing configuration: %s", err)
		return nil
	}
	switch configurationType {
	case nodecorev1alpha1.TypeK8Slice:

		// Force casting
		k8sliceConfiguration := configurationData.(nodecorev1alpha1.K8SliceConfiguration)

		for key, value := range mapK8SliceConfigurationToResources(&k8sliceConfiguration) {
			if prevRes, ok := resources[key]; !ok {
				resources[key] = value
			} else {
				prevRes.Add(*value)
				resources[key] = prevRes
			}
		}
	default:
		klog.Errorf("Configuration type %s not supported", configurationType)
		return nil
	}
	return resources
}

func mapK8SliceConfigurationToResources(k8SliceConfiguration *nodecorev1alpha1.K8SliceConfiguration) map[string]*resource.Quantity {
	resources := make(map[string]*resource.Quantity)
	resources[corev1.ResourceCPU.String()] = &k8SliceConfiguration.CPU
	resources[corev1.ResourceMemory.String()] = &k8SliceConfiguration.Memory
	resources[corev1.ResourcePods.String()] = &k8SliceConfiguration.Pods
	if k8SliceConfiguration.Storage != nil {
		resources[corev1.ResourceStorage.String()] = k8SliceConfiguration.Storage
		resources[corev1.ResourceEphemeralStorage.String()] = k8SliceConfiguration.Storage
	}
	return resources
}

func mapK8SliceToResources(k8Slice *nodecorev1alpha1.K8Slice) map[string]*resource.Quantity {
	resources := make(map[string]*resource.Quantity)
	resources[corev1.ResourceCPU.String()] = &k8Slice.Characteristics.CPU
	resources[corev1.ResourceMemory.String()] = &k8Slice.Characteristics.Memory
	resources[corev1.ResourcePods.String()] = &k8Slice.Characteristics.Pods
	if k8Slice.Characteristics.Storage != nil {
		resources[corev1.ResourceStorage.String()] = k8Slice.Characteristics.Storage
		resources[corev1.ResourceEphemeralStorage.String()] = k8Slice.Characteristics.Storage
	}
	return resources
}
