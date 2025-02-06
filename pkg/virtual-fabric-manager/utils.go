// Copyright 2022-2025 FLUIDOS Project
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

package virtualfabricmanager

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
)

func getContractResourcesByClusterID(contract *reservationv1alpha1.Contract) (map[corev1.ResourceName]string, error) {
	if contract.Spec.Configuration != nil {
		return addResources(make(map[corev1.ResourceName]string), contract.Spec.Configuration, &contract.Spec.Flavor)
	}

	return addResourceByFlavor(make(map[corev1.ResourceName]string), &contract.Spec.Flavor)
}

func addResourceByFlavor(resources map[corev1.ResourceName]string, flavor *nodecorev1alpha1.Flavor) (map[corev1.ResourceName]string, error) {
	klog.InfofDepth(1, "Current resources: %v", len(resources))
	var resourcesToAdd = make(map[corev1.ResourceName]string)
	// Parse flavor
	flavorType, flavorData, err := nodecorev1alpha1.ParseFlavorType(flavor)
	if err != nil {
		klog.Errorf("Error when parsing flavor: %s", err)
		return nil, err
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
		return nil, err
	}

	resources = resourcesToAdd

	return resources, nil
}

// This function adds the resources of a contract to the existing resourceList.
func addResources(resources map[corev1.ResourceName]string,
	configuration *nodecorev1alpha1.Configuration,
	flavor *nodecorev1alpha1.Flavor) (map[corev1.ResourceName]string, error) {
	var resourcesToAdd = make(map[corev1.ResourceName]string)
	klog.InfofDepth(1, "Current resources: %v", len(resources))
	// Parse configuration
	configurationType, configurationData, err := nodecorev1alpha1.ParseConfiguration(configuration, flavor)
	if err != nil {
		klog.Errorf("Error when parsing configuration: %s", err)
		return nil, err
	}

	// Parse flavor
	flavorType, flavorData, err := nodecorev1alpha1.ParseFlavorType(flavor)
	if err != nil {
		klog.Errorf("Error when parsing flavor: %s", err)
		return nil, err
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
			return nil, err
		}
		// Check if the flavor type matches the configuration type
		if flavorType != nodecorev1alpha1.TypeService {
			klog.Errorf("Flavor type %s does not match configuration type %s", flavorType, configurationType)
			return nil, err
		}
		// Force casting of the flavor
		serviceFlavor, ok := flavorData.(nodecorev1alpha1.ServiceFlavor)
		if !ok {
			klog.Errorf("Error when casting ServiceFlavor")
			return nil, err
		}
		// Obtain the resources from the configuration
		resourcesToAdd = mapServiceWithConfigurationToResources(&serviceFlavor, &serviceConfiguration)
	case nodecorev1alpha1.TypeSensor:
		// TODO (Sensor): Implement Liqo resource management for Sensors
		klog.Errorf("Sensor configuration not supported yet")
	default:
		klog.Errorf("Configuration type %s not supported", configurationType)
		return nil, err
	}

	resources = resourcesToAdd

	return resources, nil
}

func mapK8SliceConfigurationToResources(k8SliceConfiguration *nodecorev1alpha1.K8SliceConfiguration) map[corev1.ResourceName]string {
	resources := make(map[corev1.ResourceName]string)
	resources[corev1.ResourceCPU] = k8SliceConfiguration.CPU.String()
	resources[corev1.ResourceMemory] = k8SliceConfiguration.Memory.String()
	resources[corev1.ResourcePods] = k8SliceConfiguration.Pods.String()
	if k8SliceConfiguration.Storage != nil {
		resources[corev1.ResourceStorage] = k8SliceConfiguration.Storage.String()
		resources[corev1.ResourceEphemeralStorage] = k8SliceConfiguration.Storage.String()
	}
	return resources
}

func mapServiceWithConfigurationToResources(service *nodecorev1alpha1.ServiceFlavor,
	serviceConfiguration *nodecorev1alpha1.ServiceConfiguration) map[corev1.ResourceName]string {
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
	hostingPolicy nodecorev1alpha1.HostingPolicy) map[corev1.ResourceName]string {
	resources := make(map[corev1.ResourceName]string)

	// Set default resources to minimum values: 500m CPU, 500MB memory, 10 pod
	resources[corev1.ResourceCPU] = resource.NewMilliQuantity(800, resource.DecimalSI).String()
	resources[corev1.ResourceMemory] = resource.NewScaledQuantity(800, resource.Mega).String()
	resources[corev1.ResourcePods] = resource.NewQuantity(10, resource.DecimalSI).String()

	// Print default resources
	klog.Infof("Default resources for service %s:", service.Name)
	for key, value := range resources {
		klog.Infof("%s: %s", key, value)
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

func mapK8SliceToResources(k8Slice *nodecorev1alpha1.K8Slice) map[corev1.ResourceName]string {
	resources := make(map[corev1.ResourceName]string)
	resources[corev1.ResourceCPU] = k8Slice.Characteristics.CPU.String()
	resources[corev1.ResourceMemory] = k8Slice.Characteristics.Memory.String()
	resources[corev1.ResourcePods] = k8Slice.Characteristics.Pods.String()
	if k8Slice.Characteristics.Storage != nil {
		resources[corev1.ResourceStorage] = k8Slice.Characteristics.Storage.String()
		resources[corev1.ResourceEphemeralStorage] = k8Slice.Characteristics.Storage.String()
	}
	return resources
}
