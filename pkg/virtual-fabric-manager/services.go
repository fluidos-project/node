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

package virtualfabricmanager

import (
	"context"
	"fmt"

	discoveryv1alpha1 "github.com/liqotech/liqo/apis/discovery/v1alpha1"
	offloadingv1alpha1 "github.com/liqotech/liqo/apis/offloading/v1alpha1"
	"github.com/liqotech/liqo/pkg/discovery"
	"github.com/liqotech/liqo/pkg/utils"
	foreigncluster "github.com/liqotech/liqo/pkg/utils/foreignCluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	klog "k8s.io/klog/v2"
	pointer "k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/fluidos-project/node/pkg/utils/consts"
)

// clusterRole
//+kubebuilder:rbac:groups=discovery.liqo.io,resources=foreignclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=*,verbs=get;list;watch

// PeerWithCluster creates a ForeignCluster resource to peer with a remote cluster.
func PeerWithCluster(ctx context.Context, cl client.Client, clusterID,
	clusterName, clusterAuthURL, clusterToken string) (*discoveryv1alpha1.ForeignCluster, error) {
	// Retrieve the cluster identity associated with the current cluster.
	clusterIdentity, err := utils.GetClusterIdentityWithControllerClient(ctx, cl, consts.LiqoNamespace)
	if err != nil {
		return nil, err
	}

	// Check whether cluster IDs are the same, as we cannot peer with ourselves.
	if clusterIdentity.ClusterID == clusterID {
		return nil, fmt.Errorf("the Cluster ID of the remote cluster is the same of that of the local cluster")
	}

	// Create the secret containing the authentication token.
	err = storeInSecret(ctx, cl, clusterID, clusterToken, consts.LiqoNamespace)
	if err != nil {
		return nil, err
	}
	return enforceForeignCluster(ctx, cl, clusterID, clusterName, clusterAuthURL)
}

func enforceForeignCluster(ctx context.Context, cl client.Client,
	clusterID, clusterName, clusterAuthURL string) (*discoveryv1alpha1.ForeignCluster, error) {
	fc, err := foreigncluster.GetForeignClusterByID(ctx, cl, clusterID)
	if client.IgnoreNotFound(err) == nil {
		fc = &discoveryv1alpha1.ForeignCluster{ObjectMeta: metav1.ObjectMeta{Name: clusterName,
			Labels: map[string]string{discovery.ClusterIDLabel: clusterID}}}
	} else if err != nil {
		return nil, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, cl, fc, func() error {
		if fc.Spec.PeeringType != discoveryv1alpha1.PeeringTypeUnknown && fc.Spec.PeeringType != discoveryv1alpha1.PeeringTypeOutOfBand {
			return fmt.Errorf("a peering of type %s already exists towards remote cluster %q, cannot be changed to %s",
				fc.Spec.PeeringType, clusterName, discoveryv1alpha1.PeeringTypeOutOfBand)
		}

		fc.Spec.PeeringType = discoveryv1alpha1.PeeringTypeOutOfBand
		fc.Spec.ClusterIdentity.ClusterID = clusterID
		if fc.Spec.ClusterIdentity.ClusterName == "" {
			fc.Spec.ClusterIdentity.ClusterName = clusterName
		}

		fc.Spec.ForeignAuthURL = clusterAuthURL
		fc.Spec.ForeignProxyURL = ""
		fc.Spec.OutgoingPeeringEnabled = discoveryv1alpha1.PeeringEnabledYes
		if fc.Spec.IncomingPeeringEnabled == "" {
			fc.Spec.IncomingPeeringEnabled = discoveryv1alpha1.PeeringEnabledAuto
		}
		if fc.Spec.InsecureSkipTLSVerify == nil {
			//nolint:staticcheck // referring to the Liqo implementation
			fc.Spec.InsecureSkipTLSVerify = pointer.BoolPtr(true)
		}
		return nil
	})

	return fc, err
}

func storeInSecret(ctx context.Context, cl client.Client,
	clusterID, authToken, liqoNamespace string) error {
	secretName := fmt.Sprintf("%v%v", consts.LiqoAuthTokenSecretNamePrefix, clusterID)
	secret := &corev1.Secret{}

	err := cl.Get(ctx, types.NamespacedName{Name: secretName}, secret)
	if client.IgnoreNotFound(err) == nil {
		return createAuthTokenSecret(ctx, cl, secretName, liqoNamespace, clusterID, authToken)
	}
	if err != nil {
		klog.Error(err)
		return err
	}

	// the secret already exists, update it
	return updateAuthTokenSecret(ctx, cl, secret, clusterID, authToken)
}

func updateAuthTokenSecret(ctx context.Context, cl client.Client,
	secret *corev1.Secret, clusterID, authToken string) error {
	labels := secret.GetLabels()
	labels[discovery.ClusterIDLabel] = clusterID
	labels[discovery.AuthTokenLabel] = ""
	secret.SetLabels(labels)

	if secret.StringData == nil {
		secret.StringData = map[string]string{}
	}
	secret.StringData[consts.LiqoTokenKey] = authToken

	err := cl.Update(ctx, secret)
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

func createAuthTokenSecret(ctx context.Context, cl client.Client,
	secretName, liqoNamespace, clusterID, authToken string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: liqoNamespace,
			Labels: map[string]string{
				discovery.ClusterIDLabel: clusterID,
				discovery.AuthTokenLabel: "",
			},
		},
		StringData: map[string]string{
			"token": authToken,
		},
	}

	err := cl.Create(ctx, secret)
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

// OffloadNamespace creates a NamespaceOffloading inside the specified namespace with given pod offloading strategy and cluster selector.
func OffloadNamespace(ctx context.Context, cl client.Client, namespaceName string, strategy offloadingv1alpha1.PodOffloadingStrategyType,
	clusterTargetID string) (*offloadingv1alpha1.NamespaceOffloading, error) {
	nodeValues := make([]string, 0)
	nodeValues = append(nodeValues, clusterTargetID)

	// Create a NamespaceOffloading
	namespaceOffloading := &offloadingv1alpha1.NamespaceOffloading{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "offloading",
			Namespace: namespaceName,
		},
		Spec: offloadingv1alpha1.NamespaceOffloadingSpec{
			NamespaceMappingStrategy: offloadingv1alpha1.EnforceSameNameMappingStrategyType,
			PodOffloadingStrategy:    strategy,
			ClusterSelector: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      consts.LiqoRemoteClusterIDLabel,
								Operator: corev1.NodeSelectorOpIn,
								Values:   nodeValues,
							},
						},
					},
				},
			},
		},
	}

	klog.Infof("Creating NamespaceOffloading %s in namespace %s", namespaceOffloading.Name, namespaceOffloading.Namespace)

	err := cl.Create(ctx, namespaceOffloading)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return namespaceOffloading, nil
}
