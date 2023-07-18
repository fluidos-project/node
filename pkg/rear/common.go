package rearmanager

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	advertisementv1alpha1 "fluidos.eu/node/api/advertisement/v1alpha1"
	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
)

const (
	solverNamespace    = "default"
	discoveryNamespace = "default"
)

func forgeDiscovery(selector nodecorev1alpha1.FlavourSelector, solverID string) *advertisementv1alpha1.Discovery {
	time := metav1.Now().String()
	return &advertisementv1alpha1.Discovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forgeDiscoveryName(solverID),
			Namespace: discoveryNamespace,
		},
		Spec: advertisementv1alpha1.DiscoverySpec{
			Selector:  selector,
			SolverID:  solverID,
			Subscribe: false,
		},
		Status: advertisementv1alpha1.DiscoveryStatus{
			Phase: nodecorev1alpha1.PhaseStatus{
				Phase:          nodecorev1alpha1.PhaseIdle,
				StartTime:      time,
				LastChangeTime: time,
			},
		},
	}
}

func forgeDiscoveryName(solverID string) string {
	return fmt.Sprintf("%s-discovery", solverID)
}
