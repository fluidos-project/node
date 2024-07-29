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
	"strings"

	"github.com/liqotech/liqo/pkg/auth"
	"github.com/liqotech/liqo/pkg/utils"
	foreigncluster "github.com/liqotech/liqo/pkg/utils/foreignCluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
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

	return &nodecorev1alpha1.NodeIdentity{
		NodeID: cm.Data["nodeID"],
		Domain: cm.Data["domain"],
		IP:     cm.Data["ip"],
	}
}

// GetLocalProviders retrieves the list of local providers ip addresses from the Network Manager configMap.
func GetLocalProviders(ctx context.Context, cl client.Client) []string {
	cm := &corev1.ConfigMap{}

	// Get the configmap
	err := cl.Get(ctx, types.NamespacedName{
		Name:      consts.NetworkConfigMapName,
		Namespace: flags.FluidosNamespace,
	}, cm)
	if err != nil {
		klog.Errorf("Error getting the configmap: %s", err)
		return nil
	}
	return strings.Split(cm.Data["local"], ",")
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

// GetAllocationNameByClusterIDSpec retrieves the name of the allocation with the given in its Specs.
func GetAllocationNameByClusterIDSpec(ctx context.Context, c client.Client, clusterID string) *string {
	allocationList := nodecorev1alpha1.AllocationList{}
	if err := c.List(ctx, &allocationList, client.MatchingFields{"spec.remoteClusterID": clusterID}); err != nil {
		klog.Infof("Error when getting the allocation list: %s", err)
		return nil
	}
	if len(allocationList.Items) == 0 {
		klog.Infof("No Allocations found with the ClusterID %s", clusterID)
		return nil
	}
	return &allocationList.Items[0].Name
}
