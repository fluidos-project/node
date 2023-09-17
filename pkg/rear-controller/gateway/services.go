package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"k8s.io/klog/v2"

	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	"fluidos.eu/node/pkg/utils/models"
	"fluidos.eu/node/pkg/utils/resourceforge"
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
