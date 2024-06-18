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

package localresourcemanager

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
)

// clusterRole
//+kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavours,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups=metrics.k8s.io,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=metrics.k8s.io,resources=nodes,verbs=get;list;watch

// TODO: If the local resource manager restarts,
// ensure to check and subtract the already allocated resources from the node
// resources calculation.

// Start starts the controller.

func canMergeFlavours(flavour1, flavour2 *nodecorev1alpha1.Flavour) bool {
	if flavour1.Spec.ProviderID != flavour2.Spec.ProviderID {
		return false
	}

	if flavour1.TypeMeta != flavour2.TypeMeta {
		return false
	}

	if flavour1.Namespace != flavour2.Namespace {
		return false
	}

	if flavour1.Spec.OptionalFields.Availability != flavour2.Spec.OptionalFields.Availability {
		return false
	}

	if flavour1.Spec.Owner != flavour2.Spec.Owner {
		return false
	}

	if flavour1.Spec.Price != flavour2.Spec.Price {
		return false
	}

	if flavour1.Spec.Type != flavour2.Spec.Type {
		return false
	}

	if !reflect.DeepEqual(flavour1.Spec.Policy, flavour2.Spec.Policy) {
		return false
	}

	if flavour1.Spec.Characteristics.Architecture != flavour2.Spec.Characteristics.Architecture {
		return false
	}

	if flavour1.Spec.Characteristics.Gpu != flavour2.Spec.Characteristics.Gpu {
		return false
	}

	if flavour1.Spec.Characteristics.Pods != flavour2.Spec.Characteristics.Pods {
		return false
	}

	if flavour1.Spec.Characteristics.Cpu.Value() != flavour2.Spec.Characteristics.Cpu.Value() {
		return false
	}

	if flavour1.Spec.Characteristics.Memory.Value()/
		flavour1.Spec.Policy.Partitionable.MemoryStep.Value() !=
		flavour2.Spec.Characteristics.Memory.Value()/
			flavour2.Spec.Policy.Partitionable.MemoryStep.Value() {
		return false
	}

	if flavour1.Spec.Characteristics.EphemeralStorage.Value() != flavour2.Spec.Characteristics.EphemeralStorage.Value() {
		return false
	}

	if flavour1.Spec.Characteristics.PersistentStorage.Value() != flavour2.Spec.Characteristics.PersistentStorage.Value() {
		return false
	}

	return true
}

func mergeFlavours(flavours []nodecorev1alpha1.Flavour) []nodecorev1alpha1.Flavour {
	mergedFlavours := []nodecorev1alpha1.Flavour{}

	for len(flavours) > 0 {
		mergedFlavour := flavours[0]
		mergedFlavour.Spec.QuantityAvailable = 1

		mergedFlavour.Spec.Characteristics.Memory.Set(
			(mergedFlavour.Spec.Characteristics.Memory.Value() /
				mergedFlavour.Spec.Policy.Partitionable.MemoryStep.Value()) *
				mergedFlavour.Spec.Policy.Partitionable.MemoryStep.Value(),
		)

		mergedFlavour.Spec.Characteristics.Cpu.Set(
			(mergedFlavour.Spec.Characteristics.Cpu.Value() /
				mergedFlavour.Spec.Policy.Partitionable.CpuStep.Value()) *
				mergedFlavour.Spec.Policy.Partitionable.CpuStep.Value(),
		)

		flavours = flavours[1:]

		for i := len(flavours) - 1; i >= 0; i-- {
			if canMergeFlavours(&mergedFlavour, &flavours[i]) {
				mergedFlavour.Spec.QuantityAvailable += flavours[i].Spec.QuantityAvailable
				flavours = append(flavours[:i], flavours[i+1:]...)
			}
		}

		mergedFlavours = append(mergedFlavours, mergedFlavour)
	}

	return mergedFlavours
}

func Start(ctx context.Context, cl client.Client) error {
	klog.Info("Getting FLUIDOS Node identity...")

	nodeIdentity := getters.GetNodeIdentity(ctx, cl)
	if nodeIdentity == nil {
		klog.Info("Error getting FLUIDOS Node identity")
		return fmt.Errorf("error getting FLUIDOS Node identity")
	}

	// Check current flavours
	flavours := &nodecorev1alpha1.FlavourList{}
	err := cl.List(ctx, flavours)
	if err != nil {
		log.Printf("Error getting flavours: %v", err)
		return err
	}

	if len(flavours.Items) == 0 {
		// Verify with current node status
		klog.Info("Getting nodes resources...")
		nodes, err := GetNodesResources(ctx, cl)
		if err != nil {
			log.Printf("Error getting nodes resources: %v", err)
			return err
		}

		klog.Infof("Creating Flavours: found %d nodes", len(nodes))

		forgedFlavours := []nodecorev1alpha1.Flavour{}

		// For each node create a Flavour
		// This creates a Flavour
		for i := range nodes {
			flavour := resourceforge.ForgeFlavourFromMetrics(&nodes[i], *nodeIdentity)
			forgedFlavours = append(forgedFlavours, *flavour)
		}

		// Perform Flavour merge
		mergedFlavours := mergeFlavours(forgedFlavours)

		for i := range mergedFlavours {
			if err := cl.Create(ctx, &mergedFlavours[i]); err != nil {
				log.Printf("Error creating Flavour: %v", err)
				return err
			}
			klog.Infof("Flavour created: %s", mergedFlavours[i].Name)
		}
	}

	return nil
}
