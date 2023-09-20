package localResourceManager

import (
	"context"
	"log"

	"fluidos.eu/node/pkg/utils/resourceforge"
	"fluidos.eu/node/pkg/utils/services"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO: If the local resource manager restarts,
// ensure to check and subtract the already allocated resources from the node
// resources calculation.

// StartController starts the controller
func StartController(cl client.Client) {

	klog.Info("Getting nodes resources...")
	nodes, err := services.GetNodesResources(context.Background(), cl)
	if err != nil {
		log.Printf("Error getting nodes resources: %v", err)
		return
	}

	klog.Infof("Creating Flavour CRs: found %d nodes", len(*nodes))

	// For each node create a Flavour
	for _, node := range *nodes {
		flavour := resourceforge.ForgeFlavourFromMetrics(node)
		err := cl.Create(context.Background(), flavour)
		if err != nil {
			log.Printf("Error creating Flavour CR: %v", err)
			return
		} else {
			log.Printf("Flavour created successfully")
			return
		}
	}

}
