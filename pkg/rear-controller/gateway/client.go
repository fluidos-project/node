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
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"k8s.io/klog/v2"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/parseutil"
)

// TODO: move this function into the REAR Gateway package
// reserveFlavour reserves a flavour with the given flavourID
func (g *Gateway) ReserveFlavour(ctx context.Context, reservation *reservationv1alpha1.Reservation, flavourID string) (*models.Transaction, error) {
	var transaction models.Transaction

	body := models.ReserveRequest{
		FlavourID: flavourID,
		Buyer: models.NodeIdentity{
			NodeID: reservation.Spec.Buyer.NodeID,
			IP:     reservation.Spec.Buyer.IP,
			Domain: reservation.Spec.Buyer.Domain,
		},
		Partition: parseutil.ParsePartition(reservation.Spec.Partition),
	}

	klog.Infof("Reservation %s for flavour %s", reservation.Name, flavourID)

	selectorBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// TODO: this url should be taken from the nodeIdentity of the flavour
	url := flags.SERVER_ADDR + "/reserveflavour/" + flavourID
	klog.Infof("Sending POST request to %s", url)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(selectorBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 (OK)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&transaction); err != nil {
		return nil, err
	}

	klog.Infof("Flavour %s reserved: transaction %v", flavourID, transaction)

	g.addNewTransacion(transaction)

	return &transaction, nil
}

// PurchaseFlavour purchases a flavour with the given flavourID
func (g *Gateway) PurchaseFlavour(ctx context.Context, transactionID string) (*models.ResponsePurchase, error) {
	var purchase models.ResponsePurchase

	// Check if the transaction exists
	transaction, err := g.GetTransaction(transactionID)
	if err != nil {
		return nil, err
	}

	body := models.PurchaseRequest{
		TransactionID: transaction.TransactionID,
	}

	selectorBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// TODO: this url should be taken from the nodeIdentity of the flavour
	// Send the POST request to the server
	resp, err := http.Post(flags.SERVER_ADDR+"/purchaseflavour/"+transactionID, "application/json", bytes.NewBuffer(selectorBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 (OK)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&purchase); err != nil {
		return nil, err
	}

	return &purchase, nil
}

// SearchFlavour is a function that returns an array of Flavour that fit the Selector by performing a get request to an http server
func (g *Gateway) DiscoverFlavours(selector nodecorev1alpha1.FlavourSelector) ([]*nodecorev1alpha1.Flavour, error) {
	// Marshal the selector into JSON bytes
	s := parseutil.ParseFlavourSelector(selector)

	// Create the Flavour CR from the first flavour in the array of Flavour
	var flavoursCR []*nodecorev1alpha1.Flavour

	// Send the POST request to all the servers in the list
	for _, ADDRESS := range flags.SERVER_ADDRESSES {
		flavour, err := searchFlavour(s, ADDRESS)
		if err != nil {
			klog.Errorf("Error when searching Flavour: %s", err)
			return nil, err
		}
		flavoursCR = append(flavoursCR, flavour)
	}

	klog.Info("Flavours created", flavoursCR)
	return flavoursCR, nil
}
