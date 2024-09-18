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

package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"k8s.io/klog/v2"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
)

func searchFlavorWithSelector(ctx context.Context, selector models.Selector, addr string) ([]*nodecorev1alpha1.Flavor, error) {
	var flavors []models.Flavor

	var url string

	v, err := selectorToQueryParams(selector)
	if err != nil {
		return nil, err
	}

	// Differentiate the URL request based on the selector type
	switch selector.GetSelectorType() {
	case models.K8SliceNameDefault:
		url = fmt.Sprintf("http://%s%s", addr, Routes.K8SliceFlavors)
		// Convert the selector to query parameters
	// TODO: Implement the other selector types

	default:
		return nil, fmt.Errorf("unsupported selector type")
	}

	// Append the query parameters to the URL
	url += "?" + v
	klog.Infof("URL: %s", url)

	// Make the request
	resp, err := makeRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		klog.Infof("Received OK response status code: %d", resp.StatusCode)
	case http.StatusNoContent:
		klog.Infof("Received No Content response status code: %d", resp.StatusCode)
		return nil, nil
	default:
		return nil, fmt.Errorf("received non-OK response status code: %d", resp.StatusCode)
	}

	// Print the response body
	klog.Infof("Response body: %s", resp.Body)

	if err := json.NewDecoder(resp.Body).Decode(&flavors); err != nil {
		klog.Errorf("Error decoding the response body: %s", err)
		return nil, err
	}

	var flavorCRs []*nodecorev1alpha1.Flavor

	for i := range flavors {
		flavor := &flavors[i]
		flavorCR, err := resourceforge.ForgeFlavorFromObj(flavor)
		if err != nil {
			return nil, err
		}
		flavorCRs = append(flavorCRs, flavorCR)
	}

	return flavorCRs, nil
}

func searchFlavor(ctx context.Context, addr string) ([]*nodecorev1alpha1.Flavor, error) {
	var flavors []models.Flavor

	url := fmt.Sprintf("http://%s%s", addr, Routes.Flavors)

	resp, err := makeRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 (OK)
	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, fmt.Errorf("received non-OK response status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&flavors); err != nil {
		klog.Errorf("Error decoding the response body: %s", err)
		return nil, err
	}

	var flavorCRs []*nodecorev1alpha1.Flavor

	for i := range flavors {
		flavor := &flavors[i]
		flavorCR, err := resourceforge.ForgeFlavorFromObj(flavor)
		if err != nil {
			return nil, err
		}
		flavorCRs = append(flavorCRs, flavorCR)
	}

	return flavorCRs, nil
}

func makeRequest(ctx context.Context, method, url string, body *bytes.Buffer) (*http.Response, error) {
	httpClient := &http.Client{}

	if body == nil {
		body = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		klog.Errorf("Error creating the request: %s", err)
		return nil, err
	}
	req.Close = true

	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		klog.Errorf("Error sending the request: %s", err.Error())
		return nil, err
	}

	return resp, nil
}
