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
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/liqotech/liqo/apis/authentication/v1beta1"
	corev1beta1 "github.com/liqotech/liqo/apis/core/v1beta1"
	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	offloadingv1beta1 "github.com/liqotech/liqo/apis/offloading/v1beta1"
	liqoConsts "github.com/liqotech/liqo/pkg/consts"
	authenticationForge "github.com/liqotech/liqo/pkg/liqo-controller-manager/authentication/forge"
	authenticationUtils "github.com/liqotech/liqo/pkg/liqo-controller-manager/authentication/utils"
	networkForgeLiqo "github.com/liqotech/liqo/pkg/liqo-controller-manager/networking/forge"
	"github.com/liqotech/liqo/pkg/utils"
	ipamLiqo "github.com/liqotech/liqo/pkg/utils/ipam"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	klog "k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	reservation "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/consts"
)

// clusterRole
//+kubebuilder:rbac:groups=discovery.liqo.io,resources=foreignclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=*,verbs=get;list;watch

// EncodeKubeconfig encodes a clientcmdapi.Config into a Base64 string.
func EncodeKubeconfig(kubeconfig *clientcmdapi.Config) (string, error) {
	if kubeconfig == nil {
		return "", fmt.Errorf("kubeconfig is nil")
	}

	// Convert the Kubeconfig struct to YAML
	kubeconfigYAML, err := clientcmd.Write(*kubeconfig)
	if err != nil {
		return "", fmt.Errorf("failed to serialize kubeconfig: %w", err)
	}

	// Encode the YAML to Base64
	encodedKubeconfig := base64.StdEncoding.EncodeToString(kubeconfigYAML)
	return encodedKubeconfig, nil
}

// DecodeKubeconfig decodes a Base64 string into a clientcmdapi.Config.
func DecodeKubeconfig(encodedKubeconfig string) (*clientcmdapi.Config, error) {
	if encodedKubeconfig == "" {
		return nil, fmt.Errorf("encoded Kubeconfig string is empty")
	}

	// Decode Base64 string
	kubeconfigYAML, err := base64.StdEncoding.DecodeString(encodedKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Base64 string: %w", err)
	}

	// Deserialize the YAML into a clientcmdapi.Config
	kubeconfig, err := clientcmd.Load(kubeconfigYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig YAML: %w", err)
	}

	return kubeconfig, nil
}

// CreateKubeconfigForPeering creates a kubeconfig for peering with a remote cluster.
func CreateKubeconfigForPeering(ctx context.Context, cl client.Client, consumerClusterID string) (*clientcmdapi.Config, error) {
	// Create a Service Account
	sa, err := createOrGetPeeringServiceAccount(ctx, consumerClusterID, cl)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	// Create a ClusterRole
	cr, err := createOrGetPeeringClusterRole(ctx, consumerClusterID, cl)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	// Create a ClusterRoleBinding
	_, err = createOrGetPeeringClusterRoleBinding(ctx, consumerClusterID, cr, sa, cl)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	// Create a Secret
	secret, err := createOrGetPeeringSecret(ctx, consumerClusterID, sa, cl)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// Sleep for a while to allow the Secret to be populated.
	time.Sleep(5 * time.Second)

	// Retrieve the token from the secret just created.
	if err := cl.Get(ctx, client.ObjectKey{Name: secret.Name, Namespace: secret.Namespace}, secret); err != nil {
		klog.Error(err)
		return nil, err
	}

	token := string(secret.Data[corev1.ServiceAccountTokenKey])
	caCert := string(secret.Data[corev1.ServiceAccountRootCAKey])
	// TODO: Retrieve the server URL from the remote cluster.
	serverURL, err := getControlPlaneURL(ctx, cl)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// Create a kubeconfig for peering with the remote cluster.
	kubeConfig := &clientcmdapi.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: map[string]*clientcmdapi.Cluster{
			consumerClusterID: {
				Server:                   serverURL,
				CertificateAuthorityData: []byte(caCert),
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			consumerClusterID: {
				Token: token,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			consumerClusterID: {
				Cluster:  consumerClusterID,
				AuthInfo: consumerClusterID,
			},
		},
		CurrentContext: consumerClusterID,
	}

	return kubeConfig, nil
}

// getControlPlaneURL retrieves the public control plane URL of the cluster.
func getControlPlaneURL(ctx context.Context, cl client.Client) (string, error) {
	port := "6443"

	// Get control plane node list
	nodeList := &corev1.NodeList{}
	if err := cl.List(ctx, nodeList); err != nil {
		klog.Error(err)
		return "", err
	}

	// Iterate over nodes to find the control plane node
	for i := range nodeList.Items {
		node := &nodeList.Items[i]
		klog.InfofDepth(1, "Node: %s - Found %d labels", node.Name, len(node.Labels))

		_, existsControl := node.Labels["node-role.kubernetes.io/control-plane"]
		_, existsMaster := node.Labels["node-role.kubernetes.io/master"]

		if existsControl || existsMaster {
			// Get the control plane node IP
			for _, address := range node.Status.Addresses {
				if address.Type == corev1.NodeInternalIP {
					return "https://" + address.Address + ":" + port, nil
				}
			}
		}
	}

	return "", fmt.Errorf("unable to retrieve the control plane URL")
}

// createPeeringServiceAccount creates a ServiceAccount to be used for peering with a remote cluster.
func createOrGetPeeringServiceAccount(ctx context.Context, consumerClusterID string, cl client.Client) (*corev1.ServiceAccount, error) {
	// Get the ServiceAccount if it already exists
	// If the ServiceAccount does not exist, create it
	sa := &corev1.ServiceAccount{}
	err := cl.Get(ctx, client.ObjectKey{Name: "liqo-cluster-" + consumerClusterID, Namespace: consts.LiqoNamespace}, sa)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Error(err)
			return nil, err
		}
		klog.InfofDepth(1, "ServiceAccount does not exist, creating a new one")

		sa = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "liqo-cluster-" + consumerClusterID,
				Namespace: consts.LiqoNamespace,
			},
		}

		err = cl.Create(ctx, sa)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
	}
	return sa, nil
}

// createPeeringClusterRole creates a ClusterRole to be used for peering with a remote cluster.
func createOrGetPeeringClusterRole(ctx context.Context, consumerClusterID string, cl client.Client) (*rbacv1.ClusterRole, error) {
	// Get the ClusterRole if it already exists
	// If the ClusterRole does not exist, create it
	cr := &rbacv1.ClusterRole{}
	err := cl.Get(ctx, client.ObjectKey{Name: "liqo-cluster-" + consumerClusterID}, cr)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Error(err)
			return nil, err
		}
		klog.InfofDepth(1, "ClusterRole does not exist, creating a new one")

		cr = &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: "liqo-cluster-" + consumerClusterID,
			},
			// TODO: Define the exact rules for Liqo peering
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{"networking.liqo.io"},
					Resources: []string{"*"},
					Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
				},
				{
					APIGroups: []string{"offloading.liqo.io"},
					Resources: []string{"*"},
					Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
				},
				{
					APIGroups: []string{"authentication.liqo.io"},
					Resources: []string{"*"},
					Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
				},
				{
					APIGroups: []string{"core.liqo.io"},
					Resources: []string{"*"},
					Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
				},
				{
					APIGroups: []string{"ipam.liqo.io"},
					Resources: []string{"*"},
					Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
				},
				{
					APIGroups: []string{""},
					Resources: []string{"*"},
					Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
				},
			},
		}

		err := cl.Create(ctx, cr)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
	}
	return cr, nil
}

// createPeeringClusterRoleBinding creates a ClusterRoleBinding to be used for peering with a remote cluster.
func createOrGetPeeringClusterRoleBinding(
	ctx context.Context,
	consumerClusterID string,
	cr *rbacv1.ClusterRole,
	sa *corev1.ServiceAccount,
	cl client.Client) (*rbacv1.ClusterRoleBinding, error) {
	// Get the ClusterRoleBinding if it already exists
	// If the ClusterRoleBinding does not exist, create it
	crb := &rbacv1.ClusterRoleBinding{}
	err := cl.Get(ctx, client.ObjectKey{Name: "liqo-cluster-" + consumerClusterID, Namespace: consts.LiqoNamespace}, crb)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Error(err)
			return nil, err
		}
		klog.InfofDepth(1, "ClusterRoleBinding does not exist, creating a new one")
		crb = &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "liqo-cluster-" + consumerClusterID,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      sa.Name,
					Namespace: sa.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind:     "ClusterRole",
				Name:     cr.Name,
				APIGroup: "rbac.authorization.k8s.io",
			},
		}

		err := cl.Create(ctx, crb)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
	}
	return crb, nil
}

// createOrGetPeeringSecret creates a Secret to be used for peering with a remote cluster.
func createOrGetPeeringSecret(ctx context.Context, consumerClusterID string, sa *corev1.ServiceAccount, cl client.Client) (*corev1.Secret, error) {
	// Get the Secret if it already exists
	// If the Secret does not exist, create it
	secret := &corev1.Secret{}
	err := cl.Get(ctx, client.ObjectKey{Name: "liqo-cluster-" + consumerClusterID, Namespace: consts.LiqoNamespace}, secret)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Error(err)
			return nil, err
		}
		klog.InfofDepth(1, "Secret not found, creating a new one")
		// Create the Secret
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "liqo-cluster-" + consumerClusterID,
				Namespace: consts.LiqoNamespace,
				Annotations: map[string]string{
					"kubernetes.io/service-account.name": sa.Name,
				},
			},
			Type: corev1.SecretTypeServiceAccountToken,
		}

		err := cl.Create(ctx, secret)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
	}

	return secret, nil
}

func createTenantNamespace(ctx context.Context, cl client.Client, clusterID corev1beta1.ClusterID) (string, error) {
	// Create tenant namespace
	name := "liqo-tenant-" + string(clusterID)
	tenantNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				liqoConsts.RemoteClusterID:      string(clusterID),
				liqoConsts.TenantNamespaceLabel: "true",
			},
		},
	}
	err := cl.Create(ctx, tenantNamespace)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	klog.InfofDepth(1, "Tenant namespace %s created in %s cluster", name, clusterID)
	return name, nil
}

// CreateKubeClientFromConfig creates a Kubernetes client from a clientcmdapi.Config.
func CreateKubeClientFromConfig(kubeconfig *clientcmdapi.Config, localScheme *runtime.Scheme) (client.Client, *rest.Config, error) {
	if kubeconfig == nil {
		return nil, nil, fmt.Errorf("kubeconfig is nil")
	}

	// Convert clientcmdapi.Config to a rest.Config
	restConfig, err := clientcmd.NewDefaultClientConfig(*kubeconfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create REST config: %w", err)
	}

	// Create the Kubernetes client using the REST config
	k8sClient, err := client.New(restConfig, client.Options{
		Scheme: localScheme,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return k8sClient, restConfig, nil
}

// EstablishNetwork enables the networking module of Liqo between two clusters.
func EstablishNetwork(
	ctx context.Context,
	localClient client.Client,
	localRestConfig *rest.Config,
	remoteClient client.Client,
	remoteRestConfig *rest.Config,
) (localConn, remoteConn *networkingv1beta1.Connection, localNsName, remoteNsName string, er error) {
	// Retrieve remote liqo cluster id

	klog.InfofDepth(1, "Establishing network...")

	// Transform the client to a clientSet
	remoteKubeClient, err := kubernetes.NewForConfig(remoteRestConfig)
	if err != nil {
		klog.Errorf("Error creating the clientSet: %s", err)
		return nil, nil, "", "", err
	}

	remoteClusterIdentity, err := utils.GetClusterID(ctx, remoteKubeClient, consts.LiqoNamespace)
	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	localKubeClient, err := kubernetes.NewForConfig(localRestConfig)
	if err != nil {
		klog.Errorf("Error creating the clientSet: %s", err)
		return nil, nil, "", "", err
	}

	// Retrieve local liqo cluster id
	localClusterIdentity, err := utils.GetClusterID(ctx, localKubeClient, consts.LiqoNamespace)
	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	klog.InfoDepth(1, "Creating tenant namespaces...")

	// Create local tenant namespaces
	localNamespaceName, err := createTenantNamespace(ctx, localClient, remoteClusterIdentity)
	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	// Create remote tenant namespaces
	remoteNamespaceName, err := createTenantNamespace(ctx, remoteClient, localClusterIdentity)
	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	klog.InfoDepth(1, "Creating configurations...")

	// Create local configuration
	localConfiguration, err := createConfiguration(
		ctx,
		localClient,
		remoteNamespaceName,
		localClusterIdentity,
	)

	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	// Remote cluster applies Local configuration
	err = remoteClient.Create(ctx, localConfiguration)
	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	klog.InfofDepth(1, "Local configuration %s created", localConfiguration.Name)

	// Create remote configuration
	remoteConfiguration, err := createConfiguration(
		ctx,
		remoteClient,
		localNamespaceName,
		remoteClusterIdentity,
	)

	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	// Local cluster applies Remote configuration
	err = localClient.Create(ctx, remoteConfiguration)
	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	klog.InfofDepth(1, "Remote configuration %s created", remoteConfiguration.Name)

	klog.InfoDepth(1, "Creating Gateway Server and Client...")

	gwServer, gatewayServerIP, gatewayServerPort, remoteSecretRef, err := createGatewayServer(
		ctx,
		localClusterIdentity,
		remoteClient,
		remoteKubeClient,
		remoteNamespaceName,
	)
	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	klog.InfofDepth(
		1,
		"Gateway Server IP %s and Port %d created. Remote secret ref %s detected.",
		gatewayServerIP,
		gatewayServerPort,
		remoteSecretRef.Name,
	)

	gwClient, localSecretRef, err := createGatewayClient(
		ctx,
		localClient,
		localKubeClient,
		localNamespaceName,
		remoteClusterIdentity,
		gatewayServerIP,
		gatewayServerPort,
	)
	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	klog.InfofDepth(1, "Gatewat Client created. Local secret reference %s found", localSecretRef.Name)

	// Generate public key on Local cluster
	localPublicKey, err := generatePublicKey(
		ctx,
		localClient,
		localClusterIdentity,
		remoteNamespaceName,
		localSecretRef,
	)
	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	// Create public key on Local cluster
	err = remoteClient.Create(ctx, localPublicKey)
	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	klog.InfofDepth(1, "Local public key %s created", localPublicKey.Name)

	// Generate public key on Remote cluster
	remotePublicKey, err := generatePublicKey(
		ctx,
		remoteClient,
		remoteClusterIdentity,
		localNamespaceName,
		remoteSecretRef,
	)
	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	// Create public key on Remote cluster
	err = localClient.Create(ctx, remotePublicKey)
	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	klog.InfofDepth(1, "Remote public key %s created", remotePublicKey.Name)

	klog.InfoDepth(1, "Checking connections both on local and remote clusters...")

	localConnection, remoteConnection, err := checkConnections(
		ctx,
		localClient,
		localClusterIdentity,
		remoteClient,
		remoteClusterIdentity,
		gwClient.Name,
		gwServer.Name,
	)
	if err != nil {
		klog.Error(err)
		return nil, nil, "", "", err
	}

	return localConnection, remoteConnection, localNamespaceName, remoteNamespaceName, nil
}

func checkConnections(
	ctx context.Context,
	localClient client.Client,
	localClusterIdentity corev1beta1.ClusterID,
	remoteClient client.Client,
	remoteClusterIdentity corev1beta1.ClusterID,
	gwClientName,
	gwServerName string,
) (localConn, remoteConn *networkingv1beta1.Connection, er error) {
	// Check connection on local cluster
	localConnection := &networkingv1beta1.Connection{}
	remoteConnection := &networkingv1beta1.Connection{}

	timeout := time.After(2 * time.Minute)
	tick := time.Tick(5 * time.Second)

outerLoopLocalConnection:
	for {
		select {
		case <-timeout:
			return nil, nil, fmt.Errorf("timed out waiting for local connection to be ready")
		case <-tick:
			// Get the local connection, between all the connections in the cluster
			// Choose the one labeled liqo.io/remote-cluster-id to be the remote cluster id and owner reference to be the gateway client
			localConnections := &networkingv1beta1.ConnectionList{}
			err := localClient.List(ctx, localConnections, client.MatchingLabels{liqoConsts.RemoteClusterID: string(remoteClusterIdentity)})
			if err != nil {
				if client.IgnoreNotFound(err) != nil {
					klog.Error(err)
					return nil, nil, err
				}
			} else {
				break outerLoopLocalConnection
			}
			// Get the connection with the gateway client as owner reference
			for i := range localConnections.Items {
				conn := &localConnections.Items[i]
				if conn.OwnerReferences[0].Name == gwClientName {
					localConnection = conn
					// Check the connection status
					if localConnection.Status.Value == networkingv1beta1.Connected {
						break outerLoopLocalConnection
					}
					klog.InfofDepth(1, "Local connection %s not ready, it is in status %s", localConnection.Name, localConnection.Status.Value)
				}
			}
		}
	}

	klog.InfofDepth(1, "Local connection %s found", localConnection.Name)

	// Check connection on remote cluster
outerLoopRemoteConnection:
	for {
		select {
		case <-timeout:
			return nil, nil, fmt.Errorf("timed out waiting for remote connection to be ready")
		case <-tick:
			// Get the remote connection, between all the connections in the cluster
			// Choose the one labeled liqo.io/remote-cluster-id to be the remote cluster id and owner reference to be the gateway server
			remoteConnections := &networkingv1beta1.ConnectionList{}
			err := remoteClient.List(ctx, remoteConnections, client.MatchingLabels{liqoConsts.RemoteClusterID: string(localClusterIdentity)})
			if err != nil {
				if client.IgnoreNotFound(err) != nil {
					klog.Error(err)
					return nil, nil, err
				}
			} else {
				break outerLoopRemoteConnection
			}
			// Get the connection with the gateway server as owner reference
			for i := range remoteConnections.Items {
				conn := &remoteConnections.Items[i]
				if conn.OwnerReferences[0].Name == gwServerName {
					remoteConnection = conn
					// Check the connection status
					if remoteConnection.Status.Value == networkingv1beta1.Connected {
						break outerLoopRemoteConnection
					}
					klog.InfofDepth(1, "Remote connection %s not ready, it is in status %s", remoteConnection.Name, remoteConnection.Status.Value)
				}
			}
		}
	}

	klog.InfofDepth(1, "Remote connection %s found", remoteConnection.Name)

	return localConnection, remoteConnection, nil
}

func createGatewayClient(
	ctx context.Context,
	localClient client.Client,
	localKubeClient kubernetes.Interface,
	localNamespaceName string,
	remoteClusterIdentity corev1beta1.ClusterID,
	gatewayServerIP []string,
	gatewayServerPort int32,
) (gwCl *networkingv1beta1.GatewayClient, localSecRef *corev1.ObjectReference, er error) {
	// Create GatewayClient on Local cluster
	gwClientName := string(remoteClusterIdentity)
	gatewayClient, err := networkForgeLiqo.GatewayClient(
		localNamespaceName,
		&gwClientName,
		&networkForgeLiqo.GwClientOptions{
			KubeClient:        localKubeClient,
			RemoteClusterID:   remoteClusterIdentity,
			GatewayType:       networkForgeLiqo.DefaultGwClientType,
			TemplateName:      networkForgeLiqo.DefaultGwClientTemplateName,
			TemplateNamespace: consts.LiqoNamespace,
			MTU:               1340,
			Addresses:         gatewayServerIP,
			Port:              gatewayServerPort,
			Protocol:          networkForgeLiqo.DefaultProtocol,
		},
	)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	// Create GatewayClient on Local cluster
	err = localClient.Create(ctx, gatewayClient)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	klog.InfofDepth(1, "GatewayClient %s created", gwClientName)

	klog.InfoDepth(1, "Creating public keys...")

	// Retrieve local localSecret for the local cluster
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(5 * time.Second)

outerLoopGwClient:
	for {
		select {
		case <-timeout:
			return nil, nil, fmt.Errorf("timed out waiting for GatewayClient Secret Ref to be ready")
		case <-tick:
			// Retrieve the gateway client
			err = localClient.Get(ctx, client.ObjectKey{Name: gwClientName, Namespace: localNamespaceName}, gatewayClient)
			if err != nil {
				klog.Error(err)
				return nil, nil, err
			}
			if gatewayClient.Status.SecretRef != nil {
				break outerLoopGwClient
			}
		}
	}

	localSecretRef := gatewayClient.Status.SecretRef

	return gatewayClient, localSecretRef, nil
}

func createGatewayServer(
	ctx context.Context,
	localClusterIdentity corev1beta1.ClusterID,
	remoteClient client.Client,
	remoteKubeClient kubernetes.Interface,
	remoteNamespaceName string,
) (
	gatewayServer *networkingv1beta1.GatewayServer,
	gatewayServerIP []string,
	gatewayServerPort int32,
	remoteSecretRef *corev1.ObjectReference,
	err error,
) {
	// Create GatewayServer on Remote cluster
	gwServerName := string(localClusterIdentity)
	gatewayServer, err = networkForgeLiqo.GatewayServer(
		remoteNamespaceName,
		&gwServerName,
		&networkForgeLiqo.GwServerOptions{
			KubeClient:        remoteKubeClient,
			RemoteClusterID:   localClusterIdentity,
			GatewayType:       networkForgeLiqo.DefaultGwServerType,
			TemplateName:      networkForgeLiqo.DefaultGwServerTemplateName,
			TemplateNamespace: consts.LiqoNamespace,
			ServiceType:       corev1.ServiceTypeNodePort,
			MTU:               1340,
			Port:              networkForgeLiqo.DefaultGwServerPort,
		},
	)
	if err != nil {
		klog.Error(err)
		return nil, nil, 0, nil, err
	}

	err = remoteClient.Create(ctx, gatewayServer)
	if err != nil {
		klog.Error(err)
		return nil, nil, 0, nil, err
	}

	klog.InfofDepth(1, "GatewayServer %s created", gwServerName)

	// Wait for the GatewayServer to be ready with a timeout
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(5 * time.Second)

outerLoopGwServer:
	for {
		select {
		case <-timeout:
			return nil, nil, 0, nil, fmt.Errorf("timed out waiting for GatewayServer to be ready")
		case <-tick:
			// Retrieve the GatewayServer IP
			err = remoteClient.Get(ctx, client.ObjectKey{Name: gwServerName, Namespace: remoteNamespaceName}, gatewayServer)
			if err != nil {
				klog.Error(err)
				return nil, nil, 0, nil, err
			}
			if gatewayServer.Status.Endpoint != nil && gatewayServer.Status.Endpoint.Addresses != nil && gatewayServer.Status.SecretRef != nil {
				break outerLoopGwServer
			}
		}
	}

	// Retrieve the GatewayServer IP
	gatewayServerIP = gatewayServer.Status.Endpoint.Addresses
	gatewayServerPort = gatewayServer.Status.Endpoint.Port
	remoteSecretRef = gatewayServer.Status.SecretRef

	return gatewayServer, gatewayServerIP, gatewayServerPort, remoteSecretRef, nil
}

// Authentication enables the authentication module of Liqo between two clusters.
func Authentication(
	ctx context.Context,
	localClient client.Client,
	localRestConfig *rest.Config,
	remoteClient client.Client,
	remoteRestConfig *rest.Config,
	localNamespaceName,
	remoteNamespaceName string) error {
	klog.Infof("Authentication...")

	// Transform the client to a clientSet
	localKubeClient, err := kubernetes.NewForConfig(localRestConfig)
	if err != nil {
		klog.Errorf("Error creating the clientSet: %s", err)
		return err
	}

	remoteKubeClient, err := kubernetes.NewForConfig(remoteRestConfig)
	if err != nil {
		klog.Errorf("Error creating the clientSet: %s", err)
		return err
	}

	// Get local cluster id
	localClusterIdentity, err := utils.GetClusterID(ctx, localKubeClient, consts.LiqoNamespace)
	if err != nil {
		klog.Error(err)
		return err
	}

	// Get remote cluster id
	remoteClusterIdentity, err := utils.GetClusterID(ctx, remoteKubeClient, consts.LiqoNamespace)
	if err != nil {
		klog.Error(err)
		return err
	}

	// Forge Nonce secret
	nonceSecret := authenticationForge.Nonce(remoteNamespaceName)

	err = authenticationForge.MutateNonce(nonceSecret, localClusterIdentity)
	if err != nil {
		klog.Error(err)
		return err
	}

	// Create Nonce secret on Remote cluster
	err = remoteClient.Create(ctx, nonceSecret)
	if err != nil {
		klog.Error(err)
		return err
	}

	klog.InfofDepth(1, "Nonce secret %s created in remote cluster %s", nonceSecret.Name, remoteClusterIdentity)

	// Get signed nonce
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(5 * time.Second)
outerLoopNonce:
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for signed nonce")
		case <-tick:
			// Retrieve the nonce secret
			err = remoteClient.Get(ctx, client.ObjectKey{Name: nonceSecret.Name, Namespace: remoteNamespaceName}, nonceSecret)
			if err != nil {
				klog.Error(err)
				return err
			}
			if nonceSecret.Data["nonce"] != nil {
				break outerLoopNonce
			}
		}
	}

	nonceData := nonceSecret.Data["nonce"]

	// Ensure signed nonce
	klog.InfoDepth(1, "Ensuring signed nonce...")
	err = authenticationUtils.EnsureSignedNonceSecret(
		ctx,
		localClient,
		remoteClusterIdentity,
		localNamespaceName,
		ptr.To(string(nonceData)),
	)
	if err != nil {
		klog.Error(err)
		return err
	}

	// Retrieving signed nonce
	klog.InfoDepth(1, "Retrieving signed nonce...")
	signedNonce := []byte{}
	_ = signedNonce
outerLoopSignedNonce:
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for signed nonce")
		case <-tick:
			// Retrieve the signed nonce secret
			signedNonce, err = authenticationUtils.RetrieveSignedNonce(ctx, localClient, remoteClusterIdentity)
			if err != nil {
				if client.IgnoreNotFound(err) != nil {
					klog.Error(err)
					return err
				}
			}
			if signedNonce != nil {
				break outerLoopSignedNonce
			}
		}
	}

	tenant, err := authenticationUtils.GenerateTenant(ctx, localClient, localClusterIdentity, consts.LiqoNamespace, signedNonce, nil)
	if err != nil {
		klog.Error(err)
		return err
	}

	// Create Tenant on Remote cluster
	err = remoteClient.Create(ctx, tenant)
	if err != nil {
		klog.Error(err)
		return err
	}

	klog.InfofDepth(1, "Tenant %s created in remote cluster %s", tenant.Name, remoteClusterIdentity)

	// Wait for tenant status to be ready
outerLoopTenant:
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for tenant to be ready")
		case <-tick:
			// Retrieve the tenant
			err = remoteClient.Get(ctx, client.ObjectKey{Name: tenant.Name, Namespace: tenant.Namespace}, tenant)
			if err != nil {
				klog.Error(err)
				return err
			}
			if tenant.Status.AuthParams != nil && tenant.Status.TenantNamespace != "" {
				break outerLoopTenant
			}
		}
	}

	klog.InfofDepth(1, "Tenant %s ready", tenant.Name)

	klog.Infof("Creating Identity...")
	// From the provider cluster generate identity controlplane
	identity, err := authenticationUtils.GenerateIdentityControlPlane(
		ctx,
		remoteClient,
		localClusterIdentity,
		localNamespaceName,
		remoteClusterIdentity,
	)

	if err != nil {
		klog.Error(err)
		return err
	}

	// Create Identity on Local cluster
	err = localClient.Create(ctx, identity)
	if err != nil {
		klog.Error(err)
		return err
	}

	klog.InfofDepth(1, "Identity %s created in local cluster %s", identity.Name, localClusterIdentity)

	return nil
}

// Offloading enables the offloading module of Liqo between two clusters.
func Offloading(
	ctx context.Context,
	localClient client.Client,
	remoteRestConfig *rest.Config,
	localNamespaceName string,
	contract *reservation.Contract,
) error {
	// Forge resourceslice
	rs := authenticationForge.ResourceSlice(contract.Name, localNamespaceName)
	if rs == nil {
		return fmt.Errorf("unable to forge resourceslice")
	}

	remoteKubeClient, err := kubernetes.NewForConfig(remoteRestConfig)
	if err != nil {
		klog.Errorf("Error creating the clientSet: %s", err)
		return err
	}

	// Get remote cluster id
	remoteClusterIdentity, err := utils.GetClusterID(ctx, remoteKubeClient, consts.LiqoNamespace)
	if err != nil {
		klog.Error(err)
		return err
	}

	err = authenticationForge.MutateResourceSlice(
		rs,
		remoteClusterIdentity,
		&authenticationForge.ResourceSliceOptions{
			Class: v1beta1.ResourceSliceClassDefault,
			Resources: func() map[corev1.ResourceName]string {
				resources, err := getContractResourcesByClusterID(contract)
				if err != nil {
					klog.Error(err)
					return nil
				}
				return resources
			}(),
		},
		true,
	)
	if err != nil {
		klog.Error(err)
		return err
	}

	// Create resourceslice on Local cluster
	err = localClient.Create(ctx, rs)
	if err != nil {
		klog.Error(err)
		return err
	}

	// Wait for resource slice status authentication to be ready and resources accepted
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(5 * time.Second)

outerLoopQuotas:
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for quotas to be ready")
		case <-tick:
			// Retrieve the resourceslice
			err = localClient.Get(ctx, client.ObjectKey{Name: rs.Name, Namespace: rs.Namespace}, rs)
			if err != nil {
				klog.Error(err)
				return err
			}
			if rs.Status.Resources != nil && rs.Status.AuthParams != nil {
				break outerLoopQuotas
			}
		}
	}

	klog.InfofDepth(1, "ResourceSlice %s created in local cluster %s", rs.Name, localNamespaceName)

	klog.InfoDepth(1, "Waiting for VirtualNode...")
	// Wait for virtualnode created in local cluster
outerLoopVirtualNode:
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for virtual node to be ready")
		case <-tick:
			// Retrieve the virtual node
			virtualNode := &offloadingv1beta1.VirtualNode{}
			err = localClient.Get(ctx, client.ObjectKey{Name: rs.Name, Namespace: rs.Namespace}, virtualNode)
			if err != nil {
				if client.IgnoreNotFound(err) != nil {
					klog.Error(err)
					return err
				}
			}
			break outerLoopVirtualNode
		}
	}

	klog.InfofDepth(1, "VirtualNode %s created in namespace %s", rs.Name, localNamespaceName)

	return nil
}

func generatePublicKey(
	ctx context.Context,
	cl client.Client,
	localClusterID corev1beta1.ClusterID,
	namespaceName string,
	secretRef *corev1.ObjectReference) (*networkingv1beta1.PublicKey, error) {
	// Retrieve the secret
	secret := &corev1.Secret{}
	err := cl.Get(ctx, client.ObjectKey{Name: secretRef.Name, Namespace: secretRef.Namespace}, secret)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// Retrieve the public key from field in the secret
	stringPublicKey := string(secret.Data["publicKey"])
	if stringPublicKey == "" {
		return nil, fmt.Errorf("public key not found")
	}

	stringLocalClusterIdentity := string(localClusterID)

	// Generate public key on Local cluster
	publicKey, err := networkForgeLiqo.PublicKey(
		namespaceName,
		&stringLocalClusterIdentity,
		localClusterID,
		[]byte(stringPublicKey),
	)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return publicKey, nil
}

func createConfiguration(
	ctx context.Context,
	cl client.Client,
	destinationNamespace string,
	localClusterID corev1beta1.ClusterID) (*networkingv1beta1.Configuration, error) {
	klog.InfoDepth(1, "Creating configuration...")
	// Retrieve local pod CIDR
	localPodCIDR, err := ipamLiqo.GetPodCIDR(ctx, cl)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// Retrieve local external CIDR
	localExternalCIDR, err := ipamLiqo.GetExternalCIDR(ctx, cl)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// Create local configuration
	configuration := networkForgeLiqo.Configuration(
		string(localClusterID),
		destinationNamespace,
		localClusterID,
		localPodCIDR,
		localExternalCIDR,
	)

	if configuration == nil {
		return nil, fmt.Errorf("unable to create local configuration")
	}

	return configuration, nil
}

// PeerWithCluster creates a ForeignCluster resource to peer with a remote cluster.
func PeerWithCluster(
	ctx context.Context,
	localClient client.Client,
	localRestConfig *rest.Config,
	remoteclient client.Client,
	remoteRestConfig *rest.Config,
	contract *reservation.Contract) (*networkingv1beta1.Connection, error) {
	// Establish network with remote cluster
	localConnection, _, localNamespaceName, remoteNamespaceName, err := EstablishNetwork(
		ctx,
		localClient,
		localRestConfig,
		remoteclient,
		remoteRestConfig)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// Authenticate with remote cluster
	err = Authentication(ctx, localClient, localRestConfig, remoteclient, remoteRestConfig, localNamespaceName, remoteNamespaceName)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	err = Offloading(ctx, localClient, remoteRestConfig, localNamespaceName, contract)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return localConnection, nil
}

// OffloadNamespace creates a NamespaceOffloading inside the specified namespace with given pod offloading strategy and cluster selector.
func OffloadNamespace(ctx context.Context, cl client.Client, namespaceName string, strategy offloadingv1beta1.PodOffloadingStrategyType,
	clusterTargetID string) (*offloadingv1beta1.NamespaceOffloading, error) {
	nodeValues := make([]string, 0)
	nodeValues = append(nodeValues, clusterTargetID)

	// Create a NamespaceOffloading
	namespaceOffloading := &offloadingv1beta1.NamespaceOffloading{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "offloading",
			Namespace: namespaceName,
		},
		Spec: offloadingv1beta1.NamespaceOffloadingSpec{
			NamespaceMappingStrategy: offloadingv1beta1.EnforceSameNameMappingStrategyType,
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
