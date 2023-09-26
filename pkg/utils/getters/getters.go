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

package getters

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/consts"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"k8s.io/apimachinery/pkg/types"
)

func GetNodeIdentity(ctx context.Context, cl client.Client) *nodecorev1alpha1.NodeIdentity {

	cm := &corev1.ConfigMap{}

	// Get the node identity
	err := cl.Get(ctx, types.NamespacedName{
		Name:      consts.NODE_IDENTITY_CONFIG_MAP_NAME,
		Namespace: flags.FLUIDOS_NAMESPACE,
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

// This function retrieves the list of local providers ip addresses from the Network Manager configMap
func GetLocalProviders(ctx context.Context, cl client.Client) []string {
	cm := &corev1.ConfigMap{}

	// Get the configmap
	err := cl.Get(ctx, types.NamespacedName{
		Name:      consts.NETWORK_CONFIG_MAP_NAME,
		Namespace: flags.FLUIDOS_NAMESPACE,
	}, cm)
	if err != nil {
		klog.Errorf("Error getting the configmap: %s", err)
		return nil
	}
	return strings.Split(cm.Data["local"], ",")
}
