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

package services

import (
	"context"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetAllFlavours returns all the Flavours in the cluster
func GetAllFlavours(cl client.Client) ([]nodecorev1alpha1.Flavour, error) {

	var flavourList nodecorev1alpha1.FlavourList

	// List all Flavour CRs
	err := cl.List(context.Background(), &flavourList)
	if err != nil {
		klog.Errorf("Error when listing Flavours: %s", err)
		return nil, err
	}

	return flavourList.Items, nil
}

// GetFlavourByID returns the entire Flavour CR (not only spec) in the cluster that matches the flavourID
func GetFlavourByID(flavourID string, cl client.Client) (*nodecorev1alpha1.Flavour, error) {

	// Get the flavour with the given ID (that is the name of the CR)
	flavour := &nodecorev1alpha1.Flavour{}
	err := cl.Get(context.Background(), client.ObjectKey{
		Namespace: flags.FLUIDOS_NAMESPACE,
		Name:      flavourID,
	}, flavour)
	if err != nil {
		klog.Errorf("Error when getting Flavour %s: %s", flavourID, err)
		return nil, err
	}

	return flavour, nil
}
