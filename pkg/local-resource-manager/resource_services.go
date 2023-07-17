package localResourceManager

import (
	"context"
	"log"

	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	reservationv1alpha1 "fluidos.eu/node/api/reservation/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = runtime.NewScheme()
	uids   []string
)

func init() {
	_ = metricsv1beta1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = nodecorev1alpha1.AddToScheme(scheme)
	_ = reservationv1alpha1.AddToScheme(scheme)
}

// GetKClient creates a kubernetes API client and returns it.
func GetKClient(ctx context.Context) (client.Client, error) {
	config := ctrl.GetConfigOrDie()

	cl, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		klog.Fatalf("error creating manager: %", err)
	}

	return cl, nil
}

// GetNodesResources retrieves the metrics from all the worker nodes in the cluster
func GetNodesResources(ctx context.Context, cl client.Client) (*[]NodeInfo, error) {
	// Set a label selector to filter worker nodes
	labelSelector := labels.Set{workerLabelKey: ""}.AsSelector()

	// Get a list of nodes
	nodes := &corev1.NodeList{}
	err := cl.List(ctx, nodes, &client.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	// Get a list of nodes metrics
	nodesMetrics := &metricsv1beta1.NodeMetricsList{}
	err = cl.List(ctx, nodesMetrics, &client.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	var nodesInfo []NodeInfo
	// Print the name of each node
	for _, node := range nodes.Items {
		for _, metrics := range nodesMetrics.Items {
			if node.Name != metrics.Name {
				// So that we can select just the nodes that we want
				continue
			}

			metricsStruct := getNodeResourceMetrics(&metrics, &node)
			nodeInfo := getNodeInfo(&node, metricsStruct)
			nodesInfo = append(nodesInfo, *nodeInfo)
			uids = append(uids, string(node.UID))
		}
	}

	return &nodesInfo, nil
}

// getNodeResourceMetrics gets a ResourceMetrics struct
func getNodeResourceMetrics(nodeMetrics *metricsv1beta1.NodeMetrics, node *corev1.Node) *ResourceMetrics {
	cpuTotal := node.Status.Allocatable.Cpu()
	cpuUsed := nodeMetrics.Usage.Cpu()
	memoryTotal := node.Status.Allocatable.Memory()
	memoryUsed := nodeMetrics.Usage.Memory()
	ephemeralStorage := nodeMetrics.Usage.StorageEphemeral()
	return fromResourceMetrics(*cpuTotal, *cpuUsed, *memoryTotal, *memoryUsed, *ephemeralStorage)
}

// getNodeInfo gets a NodeInfo struct
func getNodeInfo(node *corev1.Node, metrics *ResourceMetrics) *NodeInfo {
	return fromNodeInfo(string(node.UID), node.Name, node.Status.NodeInfo.Architecture, node.Status.NodeInfo.OperatingSystem, *metrics)
}

// createFlavourCustomResource creates a new flavour custom resource
func createFlavourCustomResource(maxCPUAvailableNode NodeInfo, cl client.Client) error {
	flavour := (&nodecorev1alpha1.Flavour{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flavour-topix",
			Namespace: "default", // Specify the namespace where you want to create the CR
		},
		Spec: nodecorev1alpha1.FlavourSpec{
			FlavourID:  string(nodecorev1alpha1.K8S) + "-" + generateUniqueString(maxCPUAvailableNode.UID),
			ProviderID: maxCPUAvailableNode.UID,
			Type:       nodecorev1alpha1.K8S,
			Characteristics: nodecorev1alpha1.Characteristics{
				Cpu:              maxCPUAvailableNode.ResourceMetrics.CPUAvailable,
				Memory:           maxCPUAvailableNode.ResourceMetrics.MemoryAvailable,
				EphemeralStorage: maxCPUAvailableNode.ResourceMetrics.EphemeralStorage,
			},
			Policy: nodecorev1alpha1.Policy{
				Partitionable: &nodecorev1alpha1.Partitionable{
					CpuMin:     2,
					MemoryMin:  4,
					CpuStep:    1,
					MemoryStep: 2,
				},
				Aggregatable: &nodecorev1alpha1.Aggregatable{
					MinCount: 1,
					MaxCount: 5,
				},
			},
			Owner: nodecorev1alpha1.NodeIdentity{
				Domain: DOMAIN,
				IP:     IP_ADDR,
				NodeID: maxCPUAvailableNode.UID,
			},
			Price: nodecorev1alpha1.Price{
				Amount:   AMOUNT,
				Currency: CURRENCY,
				Period:   PERIOD,
			},
			OptionalFields: nodecorev1alpha1.OptionalFields{
				Availability: true,
			},
		},
	})

	err := cl.Create(context.Background(), flavour)
	if err != nil {
		return err
	} else {
		log.Printf("Flavour created successfully")
		return nil
	}
}
