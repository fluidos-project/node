package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"k8s.io/klog/v2"

	reservationv1alpha1 "fluidos.eu/node/api/reservation/v1alpha1"
	"fluidos.eu/node/pkg/utils/flags"
	"fluidos.eu/node/pkg/utils/models"
	"fluidos.eu/node/pkg/utils/parseutil"
)

// TODO: move this function into the REAR Gateway package
// reserveFlavour reserves a flavour with the given flavourID
func (g *Gateway) ReserveFlavour(ctx context.Context, reservation *reservationv1alpha1.Reservation, flavourID string) (*models.Transaction, error) {
	var transaction models.Transaction

	body := models.ReserveRequest{
		FlavourID: flavourID,
		Buyer: models.Owner{
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

	return &transaction, nil
}

// TODO: move this function into the REAR Gateway package
// purchaseFlavour purchases a flavour with the given flavourID
func (g *Gateway) PurchaseFlavour(ctx context.Context, transactionID string) (*models.ResponsePurchase, error) {
	var purchase models.ResponsePurchase

	// Check if the transaction exists
	transaction, err := g.GetTransaction(transactionID)
	if err != nil {
		return nil, err
	}

	body := models.PurchaseRequest{
		TransactionID: transaction.TransactionID,
		FlavourID:     transaction.FlavourID,
		BuyerID:       transaction.Buyer.NodeID,
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
