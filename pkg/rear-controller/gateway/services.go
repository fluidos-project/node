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

package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"k8s.io/klog/v2"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
)

func searchFlavour(selector models.Selector, addr string) (*nodecorev1alpha1.Flavour, error) {
	var flavour models.Flavour
	//var flavoursCR []*nodecorev1alpha1.Flavour

	// Marshal the selector into JSON bytes
	selectorBytes, err := json.Marshal(selector)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(addr+"/listflavours/selector", "application/json", bytes.NewBuffer(selectorBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 (OK)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&flavour); err != nil {
		klog.Errorf("Error decoding the response body: %s", err)
		return nil, err
	}

	flavourCR := resourceforge.ForgeFlavourFromObj(flavour)

	return flavourCR, nil
}
