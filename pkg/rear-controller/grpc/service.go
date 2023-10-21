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

package grpc

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
)

func getContractResourcesByClusterID(cl client.Client, clusterID string) (*corev1.ResourceList, error) {
	var contracts reservationv1alpha1.ContractList

	if err := cl.List(context.Background(), &contracts, client.MatchingFields{"spec.clusterID": clusterID}); err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Errorf("Error when listing Contracts: %s", err)
			return nil, err
		}
	}

	if len(contracts.Items) == 0 {
		klog.Errorf("No contracts found for cluster %s", clusterID)
		return nil, fmt.Errorf("No contracts found for cluster %s", clusterID)
	}

	if len(contracts.Items) > 1 {
		resources := multipleContractLogic(contracts.Items)
		return resources, nil
	}

	contract := contracts.Items[0]

	return mapQuantityToResourceList(contract.Spec.Partition), nil
}

func multipleContractLogic(contracts []reservationv1alpha1.Contract) *corev1.ResourceList {
	resources := &corev1.ResourceList{}
	for _, contract := range contracts {
		resources = addResources(*resources, contract.Spec.Partition)
	}
	return resources
}

// This function adds the resources of a contract to the existing resourceList
func addResources(resources corev1.ResourceList, partition *nodecorev1alpha1.Partition) *corev1.ResourceList {
	for key, value := range *mapQuantityToResourceList(partition) {
		if prevRes, ok := resources[key]; !ok {
			resources[key] = value
		} else {
			prevRes.Add(value)
			resources[key] = prevRes
		}
	}
	return &resources
}

func mapQuantityToResourceList(partition *nodecorev1alpha1.Partition) *corev1.ResourceList {
	return &corev1.ResourceList{
		corev1.ResourceCPU:              partition.CPU,
		corev1.ResourceMemory:           partition.Memory,
		corev1.ResourceEphemeralStorage: partition.EphemeralStorage,
		corev1.ResourceStorage:          partition.Storage,
	}
}
