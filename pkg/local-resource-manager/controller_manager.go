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

package localResourceManager

import (
	"context"
	"log"

	"github.com/fluidos-project/node/pkg/utils/resourceforge"
	"github.com/fluidos-project/node/pkg/utils/services"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// clusterRole
//+kubebuilder:rbac:groups=nodecore.github.com/fluidos-project/,resources=flavours,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups=metrics.k8s.io,resources=pods,verbs=get;list;watch

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
