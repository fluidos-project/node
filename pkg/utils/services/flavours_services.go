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

package services

import (
	"context"
	"sync"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/flags"
)

// FlavorService is the interface that wraps the basic Flavor methods and allows to manage the concurrent access to the Flavor CRs.
type FlavorService interface {
	sync.Mutex
	GetAllFlavors() ([]nodecorev1alpha1.Flavor, error)
	GetFlavorByID(flavorID string) (*nodecorev1alpha1.Flavor, error)
}

// GetAllFlavors returns all the Flavors in the cluster.
func GetAllFlavors(cl client.Client) ([]nodecorev1alpha1.Flavor, error) {
	var flavorList nodecorev1alpha1.FlavorList

	// List all Flavor CRs
	err := cl.List(context.Background(), &flavorList)
	if err != nil {
		klog.Errorf("Error when listing Flavors: %s", err)
		return nil, err
	}

	return flavorList.Items, nil
}

// GetFlavorByID returns the entire Flavor CR (not only spec) in the cluster that matches the flavorID.
func GetFlavorByID(flavorID string, cl client.Client) (*nodecorev1alpha1.Flavor, error) {
	// Get the flavor with the given ID (that is the name of the CR)
	flavor := &nodecorev1alpha1.Flavor{}
	err := cl.Get(context.Background(), client.ObjectKey{
		Namespace: flags.FluidosNamespace,
		Name:      flavorID,
	}, flavor)
	if err != nil {
		klog.Errorf("Error when getting Flavor %s: %s", flavorID, err)
		return nil, err
	}

	return flavor, nil
}
