package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	"fluidos.eu/node/pkg/utils/flags"
	"fluidos.eu/node/pkg/utils/models"
	"fluidos.eu/node/pkg/utils/parseutil"
	"fluidos.eu/node/pkg/utils/resourceforge"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetAllFlavours returns all the Flavours in the cluster
func GetAllFlavours(cl client.Client) ([]nodecorev1alpha1.Flavour, error) {

	var flavourList nodecorev1alpha1.FlavourList

	// List all Flavour CRs
	err := cl.List(context.Background(), &flavourList)
	if err != nil {
		return nil, err
	}

	return flavourList.Items, nil
}

// GetFlavourByID returns the entire Flavour CR (not only spec) in the cluster that matches the flavourID
func GetFlavourByID(flavourID string, cl client.Client) (*nodecorev1alpha1.Flavour, error) {

	// Get the flavour with the given ID (that is the name of the CR)
	flavour := nodecorev1alpha1.Flavour{}
	err := cl.Get(context.Background(), client.ObjectKey{
		Namespace: flags.FLAVOUR_DEFAULT_NAMESPACE,
		Name:      flavourID,
	}, &flavour)
	if err != nil {
		klog.Errorf("Error when getting Flavour %s: %s", flavourID, err)
		return nil, err
	}

	return &flavour, nil
}

// FilterFlavoursBySelector returns the Flavour CRs in the cluster that match the selector
func FilterFlavoursBySelector(flavours []nodecorev1alpha1.Flavour, selector models.Selector) ([]nodecorev1alpha1.Flavour, error) {
	var flavoursSelected []nodecorev1alpha1.Flavour

	// Get the Flavours that match the selector
	for _, f := range flavours {
		if string(f.Spec.Type) == selector.FlavourType {
			// filter function
			if filterFlavour(selector, f) {
				flavoursSelected = append(flavoursSelected, f)
			}

		}
	}

	return flavoursSelected, nil
}

// filterFlavour filters the Flavour CRs in the cluster that match the selector
func filterFlavour(selector models.Selector, f nodecorev1alpha1.Flavour) bool {
	// TODO: in the future, we must add more filter conditions (remember also that for now REAR works just with int values)

	if f.Spec.Characteristics.Architecture != selector.Architecture {
		return false
	}

	if selector.Cpu == 0 && f.Spec.Characteristics.Cpu.CmpInt64(int64(selector.Cpu)) != 0 {
		return false
	}

	if selector.Memory == 0 && f.Spec.Characteristics.Memory.CmpInt64(int64(selector.Memory)) != 0 {
		return false
	}

	if selector.EphemeralStorage == 0 && f.Spec.Characteristics.EphemeralStorage.CmpInt64(int64(selector.EphemeralStorage)) < 0 {
		return false
	}

	if selector.MoreThanCpu != 0 && f.Spec.Characteristics.Cpu.CmpInt64(int64(selector.MoreThanCpu)) < 0 {
		return false
	}

	if selector.MoreThanMemory != 0 && f.Spec.Characteristics.Memory.CmpInt64(int64(selector.MoreThanMemory)) < 0 {
		return false
	}

	if selector.MoreThanEph != 0 && f.Spec.Characteristics.EphemeralStorage.CmpInt64(int64(selector.MoreThanEph)) < 0 {
		return false
	}

	if selector.LessThanCpu != 0 && f.Spec.Characteristics.Cpu.CmpInt64(int64(selector.LessThanCpu)) > 0 {
		return false
	}

	if selector.LessThanMemory != 0 && f.Spec.Characteristics.Memory.CmpInt64(int64(selector.LessThanMemory)) > 0 {
		return false
	}

	if selector.LessThanEph != 0 && f.Spec.Characteristics.EphemeralStorage.CmpInt64(int64(selector.LessThanEph)) > 0 {
		return false
	}

	return true
}

// SearchFlavour is a function that returns an array of Flavour that fit the Selector by performing a get request to an http server
func SearchFlavour(selector nodecorev1alpha1.FlavourSelector) ([]*nodecorev1alpha1.Flavour, error) {
	// Marshal the selector into JSON bytes
	body := parseutil.ParseSelectorValues(selector)

	klog.Info("Selector is ", body)
	selectorBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Create the Flavour CR from the first flavour in the array of Flavour
	var flavoursCR []*nodecorev1alpha1.Flavour

	// Send the POST request to all the servers in the list
	for _, ADDRESS := range flags.SERVER_ADDRESSES {
		resp, err := http.Post(ADDRESS+"/listflavours/selector", "application/json", bytes.NewBuffer(selectorBytes))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Check if the response status code is 200 (OK)
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("received non-OK response status code: %d", resp.StatusCode)
		}

		// Read the response body
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// Unmarshal the response JSON into an array of Flavour Object
		var flavours []*models.Flavour
		if err := json.Unmarshal(respBody, &flavours); err != nil {
			return nil, err
		}

		for _, flavour := range flavours {
			klog.Infof("Flavour found: %s", flavour.FlavourID)
			cr := resourceforge.ForgeFlavourCustomResource(*flavour)
			flavoursCR = append(flavoursCR, cr)
		}

	}
	klog.Info("Flavours created", flavoursCR)
	return flavoursCR, nil

}

// CheckSelector ia a func to check if the syntax of the Selector is right.
// Strict and range syntax cannot be used together
func CheckSelector(selector models.Selector) error {
	if selector.Cpu != 0 || selector.Memory != 0 || selector.EphemeralStorage != 0 {
		if selector.MoreThanCpu != 0 || selector.MoreThanMemory != 0 || selector.MoreThanEph != 0 || selector.LessThanEph != 0 || selector.LessThanCpu != 0 || selector.LessThanMemory != 0 {
			return fmt.Errorf("selector syntax error: strict and range syntax cannot be used together")
		}
	}
	return nil
}
