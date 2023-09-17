package services

import (
	"context"

	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	"fluidos.eu/node/pkg/utils/flags"
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
		Namespace: flags.FLAVOUR_DEFAULT_NAMESPACE,
		Name:      flavourID,
	}, flavour)
	if err != nil {
		klog.Errorf("Error when getting Flavour %s: %s", flavourID, err)
		return nil, err
	}

	return flavour, nil
}
