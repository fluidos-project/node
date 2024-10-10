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

	var contractsBuyer reservationv1alpha1.ContractList
	var contractsSeller reservationv1alpha1.ContractList

	// Retrieve contracts where the cluster which is requesting the peering is the buyer
	if err := cl.List(context.Background(), &contractsBuyer, client.MatchingFields{"spec.buyer.additionalInformation.liqoID": clusterID}); err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Errorf("Error when listing Contracts: %s", err)
			return nil, err
		}
	}

	// Retrieve contracts where the cluster which is requesting the peering is the seller
	if err := cl.List(context.Background(), &contractsSeller, client.MatchingFields{"spec.seller.additionalInformation.liqoID": clusterID}); err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Errorf("Error when listing Contracts: %s", err)
			return nil, err
		}
	}

	contracts.Items = append(contracts.Items, contractsBuyer.Items...)
	contracts.Items = append(contracts.Items, contractsSeller.Items...)

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
		return addResources(make(map[string]*resource.Quantity), contract.Spec.Configuration, &contract.Spec.Flavor), nil
	}

	return addResourceByFlavor(make(map[string]*resource.Quantity), &contract.Spec.Flavor), nil
}

func multipleContractLogic(contracts []reservationv1alpha1.Contract) map[string]*resource.Quantity {
	resources := make(map[string]*resource.Quantity)
	for i := range contracts {
		if contracts[i].Spec.Configuration != nil {
			resources = addResources(resources, contracts[i].Spec.Configuration, &contracts[i].Spec.Flavor)
		} else {
			resources = addResourceByFlavor(resources, &contracts[i].Spec.Flavor)
		}
	}
	return resources
}

func addResourceByFlavor(resources map[string]*resource.Quantity, flavor *nodecorev1alpha1.Flavor) map[string]*resource.Quantity {
	var resourcesToAdd = make(map[string]*resource.Quantity)
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
		resourcesToAdd = mapK8SliceToResources(&k8sliceFlavor)
	case nodecorev1alpha1.TypeVM:
		// TODO (VM): Implement Liqo resource management for VMs
		klog.Errorf("VM flavor not supported yet")
	case nodecorev1alpha1.TypeService:
		// Force casting
		serviceFlavor := flavorData.(nodecorev1alpha1.ServiceFlavor)
		// For each characteristic, add the resource to the map
		resourcesToAdd = mapServiceWithConfigurationToResources(&serviceFlavor, nil)
	case nodecorev1alpha1.TypeSensor:
		// TODO (Sensor): Implement Liqo resource management for Sensors
	default:
		klog.Errorf("Flavor type %s not supported", flavorType)
		return nil
	}

	for key, value := range resourcesToAdd {
		if prevRes, ok := resources[key]; !ok {
			resources[key] = value
		} else {
			prevRes.Add(*value)
			resources[key] = prevRes
		}
	}

	return resources
}

// This function adds the resources of a contract to the existing resourceList.
func addResources(resources map[string]*resource.Quantity,
	configuration *nodecorev1alpha1.Configuration,
	flavor *nodecorev1alpha1.Flavor) map[string]*resource.Quantity {
	var resourcesToAdd = make(map[string]*resource.Quantity)
	// Parse configuration
	configurationType, configurationData, err := nodecorev1alpha1.ParseConfiguration(configuration, flavor)
	if err != nil {
		klog.Errorf("Error when parsing configuration: %s", err)
		return nil
	}

	// Parse flavor
	flavorType, flavorData, err := nodecorev1alpha1.ParseFlavorType(flavor)
	if err != nil {
		klog.Errorf("Error when parsing flavor: %s", err)
		return nil
	}

	switch configurationType {
	case nodecorev1alpha1.TypeK8Slice:
		// Force casting
		k8sliceConfiguration := configurationData.(nodecorev1alpha1.K8SliceConfiguration)
		// Obtain the resources from the configuration
		resourcesToAdd = mapK8SliceConfigurationToResources(&k8sliceConfiguration)
	case nodecorev1alpha1.TypeVM:
		// TODO (VM): Implement Liqo resource management for VMs
		klog.Errorf("VM configuration not supported yet")
	case nodecorev1alpha1.TypeService:
		// Force casting of the configuration
		serviceConfiguration, ok := configurationData.(nodecorev1alpha1.ServiceConfiguration)
		if !ok {
			klog.Errorf("Error when casting ServiceConfiguration")
			return nil
		}
		// Check if the flavor type matches the configuration type
		if flavorType != nodecorev1alpha1.TypeService {
			klog.Errorf("Flavor type %s does not match configuration type %s", flavorType, configurationType)
			return nil
		}
		// Force casting of the flavor
		serviceFlavor, ok := flavorData.(nodecorev1alpha1.ServiceFlavor)
		if !ok {
			klog.Errorf("Error when casting ServiceFlavor")
			return nil
		}
		// Obtain the resources from the configuration
		resourcesToAdd = mapServiceWithConfigurationToResources(&serviceFlavor, &serviceConfiguration)
	case nodecorev1alpha1.TypeSensor:
		// TODO (Sensor): Implement Liqo resource management for Sensors
		klog.Errorf("Sensor configuration not supported yet")
	default:
		klog.Errorf("Configuration type %s not supported", configurationType)
		return nil
	}

	// Add the resources of the flavor to the existing resources
	for key, value := range resourcesToAdd {
		if prevRes, ok := resources[key]; !ok {
			resources[key] = value
		} else {
			prevRes.Add(*value)
			resources[key] = prevRes
		}
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

func mapServiceWithConfigurationToResources(service *nodecorev1alpha1.ServiceFlavor,
	serviceConfiguration *nodecorev1alpha1.ServiceConfiguration) map[string]*resource.Quantity {
	var hostingPolicy nodecorev1alpha1.HostingPolicy

	// Check the hosting policy of the service
	if serviceConfiguration.HostingPolicy != nil {
		hostingPolicy = *serviceConfiguration.HostingPolicy
	} else {
		// No configuration is passed therefore we need to follow default hosting policies
		// Check the first default hosting policy of the service
		if len(service.HostingPolicies) == 0 {
			klog.Infof("No default hosting policies found for service %s", service.Name)
			klog.Infof("Proceeding with the default hosting policy: %s", nodecorev1alpha1.HostingPolicyProvider)
			hostingPolicy = nodecorev1alpha1.HostingPolicyProvider
		} else {
			klog.Infof("Default hosting policy found for service %s", service.Name)
			hostingPolicy = service.HostingPolicies[0]
		}
	}

	return mapServiceToResourcesWithHostingPolicy(service, hostingPolicy)
}

func mapServiceToResourcesWithHostingPolicy(
	service *nodecorev1alpha1.ServiceFlavor,
	hostingPolicy nodecorev1alpha1.HostingPolicy) map[string]*resource.Quantity {
	resources := make(map[string]*resource.Quantity)

	// Set default resources to minimum values: 500m CPU, 500MB memory, 10 pod
	resources[corev1.ResourceCPU.String()] = resource.NewMilliQuantity(800, resource.DecimalSI)
	resources[corev1.ResourceMemory.String()] = resource.NewScaledQuantity(800, resource.Mega)
	resources[corev1.ResourcePods.String()] = resource.NewQuantity(10, resource.DecimalSI)

	// Print default resources
	klog.Infof("Default resources for service %s:", service.Name)
	for key, value := range resources {
		klog.Infof("%s: %s", key, value.String())
	}

	switch hostingPolicy {
	case nodecorev1alpha1.HostingPolicyProvider:
		klog.Infof("Proceeding with the default hosting policy: %s", nodecorev1alpha1.HostingPolicyProvider)
		klog.Info("No resources will be allocated for the peering...")
		return resources
	case nodecorev1alpha1.HostingPolicyConsumer:
		klog.Infof("Proceeding with the default hosting policy: %s", nodecorev1alpha1.HostingPolicyConsumer)
		// TODO(Service): implement Liqo resource management for Services Consumer Hosted: resource should be read by the manifests resource limits
		klog.Errorf("Service Consumer Hosted not supported yet, returning no resources")
		return resources
	case nodecorev1alpha1.HostingPolicyShared:
		klog.Infof("Proceeding with the default hosting policy: %s", nodecorev1alpha1.HostingPolicyShared)
		// TODO(Service): implement Liqo resource management for Services Shared
		klog.Errorf("Service Shared not supported yet, returning no resources")
		return resources
	default:
		klog.Errorf("Default hosting policy %s not supported, returning no resources", service.HostingPolicies[0])
		return resources
	}
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
