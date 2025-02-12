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

package getters

import (
	"context"
	"fmt"

	"github.com/liqotech/liqo/pkg/auth"
	"github.com/liqotech/liqo/pkg/utils"
	foreigncluster "github.com/liqotech/liqo/pkg/utils/foreignCluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	networkv1alpha1 "github.com/fluidos-project/node/apis/network/v1alpha1"
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/consts"
	"github.com/fluidos-project/node/pkg/utils/flags"
)

// GetNodeIdentity retrieves the node identity from the local cluster.
func GetNodeIdentity(ctx context.Context, cl client.Client) *nodecorev1alpha1.NodeIdentity {
	cm := &corev1.ConfigMap{}

	// Get the node identity
	err := cl.Get(ctx, types.NamespacedName{
		Name:      consts.NodeIdentityConfigMapName,
		Namespace: flags.FluidosNamespace,
	}, cm)
	if err != nil {
		klog.Errorf("Error getting the configmap: %s", err)
		return nil
	}

	// Get the control plane IP address if not available in the ConfigMap
	if cm.Data["ip"] == "" {
		ep := &corev1.Endpoints{}

		// Get the "kubernetes" endpoint
		err := cl.Get(ctx, types.NamespacedName{
			Name:      "kubernetes",
			Namespace: metav1.NamespaceDefault,
		}, ep)
		if err != nil {
			klog.Errorf("Error getting the endpoint: %s", err)
			return nil
		}

		cm.Data["ip"] = ep.Subsets[0].Addresses[0].IP
	}

	return &nodecorev1alpha1.NodeIdentity{
		NodeID: cm.Data["nodeID"],
		Domain: cm.Data["domain"],
		IP:     cm.Data["ip"] + ":" + cm.Data["port"],
	}
}

// GetLocalProviders retrieves the list of local providers ip addresses from the KnownCluster CRs.
func GetLocalProviders(ctx context.Context, cl client.Client) []string {
	knownclusters := networkv1alpha1.KnownClusterList{}
	result := []string{}

	// Get the list of KnownClusters
	if err := cl.List(ctx, &knownclusters); err != nil {
		klog.Errorf("Error when listing KnownClusters: %s", err)
		return nil
	}

	if len(knownclusters.Items) == 0 {
		klog.Infof("No KnownCluster found")
		return nil
	}

	for i := range knownclusters.Items {
		result = append(result, knownclusters.Items[i].Spec.Address)
	}

	return result
}

// GetLiqoCredentials retrieves the Liqo credentials from the local cluster.
func GetLiqoCredentials(ctx context.Context, cl client.Client) (*nodecorev1alpha1.LiqoCredentials, error) {
	localToken, err := auth.GetToken(ctx, cl, consts.LiqoNamespace)
	if err != nil {
		return nil, err
	}

	clusterIdentity, err := utils.GetClusterIdentityWithControllerClient(ctx, cl, consts.LiqoNamespace)
	if err != nil {
		return nil, err
	}

	authEP, err := foreigncluster.GetHomeAuthURL(ctx, cl, consts.LiqoNamespace)
	if err != nil {
		return nil, err
	}

	// If the local cluster has not a cluster name, we print the use the local clusterID to not leave this field empty.
	// This can be changed by the user when pasting this value in a remote cluster.
	if clusterIdentity.ClusterName == "" {
		clusterIdentity.ClusterName = clusterIdentity.ClusterID
	}

	return &nodecorev1alpha1.LiqoCredentials{
		ClusterName: clusterIdentity.ClusterName,
		ClusterID:   clusterIdentity.ClusterID,
		Endpoint:    authEP,
		Token:       localToken,
	}, nil
}

// GetReservationByTransactionID retrieves the Reservation with the given transaction ID.
func GetReservationByTransactionID(ctx context.Context, c client.Client, transactionID string) (*reservationv1alpha1.Reservation, error) {
	// Get all the Reservations
	reservations := &reservationv1alpha1.ReservationList{}
	if err := c.List(ctx, reservations, client.InNamespace(flags.FluidosNamespace)); err != nil {
		klog.Errorf("Error getting the Reservations: %s", err)
		return nil, err
	}

	for i := range reservations.Items {
		reservation := reservations.Items[i]
		if reservation.Status.TransactionID == transactionID {
			return &reservation, nil
		}
	}

	return nil, fmt.Errorf("reservation with transaction ID %s not found", transactionID)
}

// GetAllocationByClusterIDSpec retrieves the name of the allocation with the given in its Specs.
func GetAllocationByClusterIDSpec(ctx context.Context, c client.Client, clusterID string) (nodecorev1alpha1.AllocationList, error) {
	filteredAllocations := nodecorev1alpha1.AllocationList{}
	allocationList := nodecorev1alpha1.AllocationList{}
	// Get all the Allocations
	if err := c.List(ctx, &allocationList, client.InNamespace(flags.FluidosNamespace)); err != nil {
		klog.Errorf("Error getting the Allocations: %s", err)
		return filteredAllocations, err
	}
	// Filter the Allocation based on the contract spec:
	// the clusterID field should match spec.seller.additionalInformation.liqoID or spec.buyer.additionalInformation.liqoID
	for i := range allocationList.Items {
		allocation := allocationList.Items[i]
		// Get the contract
		contract := &reservationv1alpha1.Contract{}
		if err := c.Get(
			ctx,
			types.NamespacedName{
				Name:      allocation.Spec.Contract.Name,
				Namespace: allocation.Spec.Contract.Namespace,
			},
			contract,
		); err != nil {
			klog.Errorf("Error getting the Contract: %s", err)
			return filteredAllocations, err
		}
		// Check if the clusterID is in the contract
		if contract.Spec.Seller.AdditionalInformation.LiqoID == clusterID || contract.Spec.Buyer.AdditionalInformation.LiqoID == clusterID {
			// Add the Allocation to the filtered list
			filteredAllocations.Items = append(filteredAllocations.Items, allocation)
		}
	}

	return filteredAllocations, nil
}

// GetBlueprint retrieves a ServiceBlueprint from the Flavor name.
func GetBlueprint(ctx context.Context, c client.Client, flavorName string) (*nodecorev1alpha1.ServiceBlueprint, error) {
	// ServiceBlueprint is the owner of the Flavor
	flavor := &nodecorev1alpha1.Flavor{}
	if err := c.Get(ctx, types.NamespacedName{Name: flavorName, Namespace: flags.FluidosNamespace}, flavor); err != nil {
		klog.Errorf("Error getting the Flavor: %s", err)
		return nil, err
	}

	// Get the owners of the Flavor
	owners := flavor.GetOwnerReferences()
	if len(owners) == 0 {
		klog.Errorf("No owner found for the Flavor %s", flavorName)
		return nil, fmt.Errorf("no owner found for the Flavor %s", flavorName)
	}

	// Get the ServiceBlueprint
	blueprint := &nodecorev1alpha1.ServiceBlueprint{}
	if err := c.Get(ctx, types.NamespacedName{Name: owners[0].Name, Namespace: flags.FluidosNamespace}, blueprint); err != nil {
		klog.Errorf("Error getting the ServiceBlueprint: %s", err)
		return nil, err
	}

	return blueprint, nil
}

// GetAllocationByContractName retrieves the Allocation with the given contract name.
func GetAllocationByContractName(ctx context.Context, c client.Client, contractName string) (*nodecorev1alpha1.Allocation, error) {
	// Get all the Allocations with spec.contract.name == contractName
	allocations := &nodecorev1alpha1.AllocationList{}
	if err := c.List(ctx, allocations, client.InNamespace(flags.FluidosNamespace)); err != nil {
		klog.Errorf("Error getting the Allocations: %s", err)
		return nil, err
	}

	for i := range allocations.Items {
		allocation := allocations.Items[i]
		if allocation.Spec.Contract.Name == contractName {
			return &allocation, nil
		}
	}

	return nil, fmt.Errorf("allocation with contract name %s not found", contractName)
}
