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
	"strings"

	"k8s.io/klog/v2"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/parseutil"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
)

// ReserveFlavor reserves a flavor with the given flavorID.
func (g *Gateway) ReserveFlavor(ctx context.Context,
	reservation *reservationv1alpha1.Reservation, flavor *nodecorev1alpha1.Flavor) (*models.Transaction, error) {
	err := checkLiqoReadiness(g.LiqoReady)
	if err != nil {
		return nil, err
	}

	liqoCredentials, err := getters.GetLiqoCredentials(ctx, g.client)
	if err != nil {
		klog.Errorf("Error when getting Liqo credentials: %s", err)
		return nil, err
	}

	var transaction models.Transaction

	flavorID := flavor.Name

	body := models.ReserveRequest{
		FlavorID: flavorID,
		Buyer: models.NodeIdentity{
			NodeID: g.ID.NodeID,
			IP:     g.ID.IP,
			Domain: g.ID.Domain,
			AdditionalInformation: &models.NodeIdentityAdditionalInfo{
				LiqoID: liqoCredentials.ClusterID,
			},
		},
		Configuration: func() *models.Configuration {
			klog.Infof("Reservation configuration is %v", reservation.Spec.Configuration)
			if reservation.Spec.Configuration != nil {
				klog.Infof("Configuration found in the reservation %s", reservation.Name)
				configuration, err := parseutil.ParseConfiguration(reservation.Spec.Configuration, flavor)
				if err != nil {
					klog.Errorf("Error when parsing configuration: %s", err)
					return nil
				}
				return configuration
			}
			klog.Infof("No configuration found in the reservation %s", reservation.Name)
			return nil
		}(),
	}

	klog.Infof("Reserve request: %v", body)

	klog.Infof("Reservation %s for flavor %s", reservation.Name, flavorID)

	selectorBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// TODO: this url should be taken from the nodeIdentity of the flavor
	bodyBytes := bytes.NewBuffer(selectorBytes)
	url := fmt.Sprintf("http://%s%s", reservation.Spec.Seller.IP, Routes.Reserve)

	klog.Infof("Sending request to %s", url)

	resp, err := makeRequest(ctx, "POST", url, bodyBytes)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 (OK)
	switch resp.StatusCode {
	case http.StatusOK:
		klog.Infof("Received OK response status code: %v", resp)
	case http.StatusNotFound:
		klog.Errorf("Received NOT FOUND response status code: %v", resp)
		return nil, fmt.Errorf("received NOT FOUND response status code: %d", resp.StatusCode)
	default:
		klog.Errorf("Received non-OK response status code: %v", resp)
		return nil, fmt.Errorf("received non-OK response status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&transaction); err != nil {
		return nil, err
	}

	// check if transaction is not correctly set
	// This behaviuor is a possible bug of the rear-controller
	// TODO: check it
	if transaction.TransactionID == "" {
		klog.Errorf("TransactionID received is empty")
		return &transaction, fmt.Errorf("transactionID is empty")
	}

	klog.Infof("Flavor %s reserved: transaction ID %s", flavorID, transaction.TransactionID)

	g.addNewTransaction(&transaction)

	return &transaction, nil
}

// PurchaseFlavor purchases a flavor with the given flavorID.
func (g *Gateway) PurchaseFlavor(ctx context.Context, transactionID string,
	seller nodecorev1alpha1.NodeIdentity, buyerLiqoCredentials *nodecorev1alpha1.LiqoCredentials) (*models.Contract, error) {
	err := checkLiqoReadiness(g.LiqoReady)
	if err != nil {
		return nil, err
	}

	var contract models.Contract

	// Check if the transaction exists
	transaction, err := g.GetTransaction(transactionID)
	if err != nil {
		return nil, err
	}

	// Get Reservation from the transaction
	reservation, err := getters.GetReservationByTransactionID(ctx, g.client, transactionID)
	if err != nil {
		return nil, err
	}

	// Parse the IngressTelemetryEndpoint
	ite := parseutil.ParseTelemetryServer(reservation.Spec.IngressTelemetryEndpoint)

	klog.Infof("Transaction %s for flavor %s", transactionID, transaction.FlavorID)

	var liqoCredentials *models.LiqoCredentials

	if buyerLiqoCredentials != nil {
		liqoCredentials, err = resourceforge.ForgeLiqoCredentialsObj(buyerLiqoCredentials)
		if err != nil {
			return nil, err
		}
	}

	body := models.PurchaseRequest{
		LiqoCredentials:          liqoCredentials,
		IngressTelemetryEndpoint: ite,
	}

	selectorBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyBytes := bytes.NewBuffer(selectorBytes)
	apiPath := strings.Replace(Routes.Purchase, "{transactionID}", transactionID, 1)
	url := fmt.Sprintf("http://%s%s", seller.IP, apiPath)

	resp, err := makeRequest(ctx, "POST", url, bodyBytes)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 (OK)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&contract); err != nil {
		return nil, err
	}

	return &contract, nil
}

// DiscoverFlavors is a function that returns an array of Flavor that fit the Selector by performing a get request to an http server.
func (g *Gateway) DiscoverFlavors(ctx context.Context, selector *nodecorev1alpha1.Selector) ([]*nodecorev1alpha1.Flavor, error) {
	klog.Info("Discovering flavors")

	// Check if Liqo is ready
	err := checkLiqoReadiness(g.LiqoReady)
	if err != nil {
		return nil, err
	}

	var s models.Selector
	var flavorsCR []*nodecorev1alpha1.Flavor

	klog.Info("Checking selector")
	if selector != nil {
		s, err = parseutil.ParseFlavorSelector(selector)
		klog.Infof("Selector parsed: %v", s)
		if err != nil {
			return nil, err
		}
	}

	klog.Info("Getting local providers")
	providers := getters.GetLocalProviders(context.Background(), g.client)

	// Send the GET request to all the servers in the list
	for _, provider := range providers {
		klog.Infof("Provider: %s", provider)
		flavors, err := discover(ctx, s, provider)
		if err != nil {
			klog.Errorf("Error when searching Flavor: %s", err)
			return nil, err
		}
		// Check if the flavor is nil
		if len(flavors) == 0 {
			klog.Infof("No Flavors found for provider %s", provider)
		} else {
			klog.Infof("Flavors found for provider %s: %d", provider, len(flavors))
			flavorsCR = append(flavorsCR, flavors...)
		}
	}

	klog.Infof("Found %d flavors", len(flavorsCR))
	return flavorsCR, nil
}

func discover(ctx context.Context, s models.Selector, provider string) ([]*nodecorev1alpha1.Flavor, error) {
	if s != nil {
		klog.Infof("Searching Flavor with selector %v", s)
		return searchFlavorWithSelector(ctx, s, provider)
	}
	klog.Infof("Searching Flavor with no selector")
	return searchFlavor(ctx, provider)
}

func checkLiqoReadiness(b bool) error {
	if !b {
		klog.Errorf("Liqo is not ready, please check or wait for the Liqo installation")
		return fmt.Errorf("liqo is not ready, please check or wait for the Liqo installation")
	}
	return nil
}
