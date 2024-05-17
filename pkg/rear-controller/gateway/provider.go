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
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/common"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/namings"
	"github.com/fluidos-project/node/pkg/utils/parseutil"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
	"github.com/fluidos-project/node/pkg/utils/services"
	"github.com/fluidos-project/node/pkg/utils/tools"
)

// getFlavours gets all the flavours CRs from the cluster.
func (g *Gateway) getFlavours(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	klog.Infof("Processing request for getting all Flavours...")

	flavours, err := services.GetAllFlavours(g.client)
	if err != nil {
		klog.Errorf("Error getting all the Flavour CRs: %s", err)
		http.Error(w, "Error getting all the Flavour CRs", http.StatusInternalServerError)
		return
	}

	klog.Infof("Found %d Flavours in the cluster", len(flavours))

	availableFlavours := make([]nodecorev1alpha1.Flavour, 0)

	// Filtering only the available flavours
	for i := range flavours {
		if !flavours[i].Spec.OptionalFields.Availability {
			availableFlavours = append(availableFlavours, flavours[i])
		}
	}

	klog.Infof("Available Flavours: %d", len(availableFlavours))
	if len(availableFlavours) == 0 {
		klog.Infof("No available Flavours found")
		// Return content for empty list
		emptyList := make([]*nodecorev1alpha1.Flavour, 0)
		encodeResponseStatusCode(w, emptyList, http.StatusNoContent)
		return
	}

	// Select the flavour with the max CPU
	max := resource.MustParse("0")
	index := 0
	for i := range availableFlavours {
		if availableFlavours[i].Spec.Characteristics.Cpu.Cmp(max) == 1 {
			max = availableFlavours[i].Spec.Characteristics.Cpu
			index = i
		}
	}

	selected := *flavours[index].DeepCopy()

	klog.Infof("Flavour %s selected - Parsing...", selected.Name)
	parsed := parseutil.ParseFlavour(&selected)

	klog.Infof("Flavour parsed: %v", parsed)

	// Encode the Flavour as JSON and write it to the response writer
	encodeResponse(w, parsed)
}

// getFlavourBySelectorHandler gets the flavour CRs from the cluster that match the selector.
func (g *Gateway) getFlavoursBySelector(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	klog.Infof("Processing request for getting Flavours by selector...")

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// build the selector from the request body
	selector, err := buildSelector(body)
	if err != nil {
		klog.Errorf("Error building the selector: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	flavours, err := services.GetAllFlavours(g.client)
	if err != nil {
		klog.Errorf("Error getting all the Flavour CRs: %s", err)
		http.Error(w, "Error getting all the Flavour CRs", http.StatusInternalServerError)
		return
	}

	klog.Infof("Found %d Flavours in the cluster", len(flavours))

	availableFlavours := make([]nodecorev1alpha1.Flavour, 0)

	// Filtering only the available flavours
	for i := range flavours {
		if flavours[i].Spec.OptionalFields.Availability {
			availableFlavours = append(availableFlavours, flavours[i])
		}
	}

	klog.Infof("Available Flavours: %d", len(availableFlavours))
	if len(availableFlavours) == 0 {
		klog.Infof("No available Flavours found")
		// Return content for empty list
		emptyList := make([]*nodecorev1alpha1.Flavour, 0)
		encodeResponseStatusCode(w, emptyList, http.StatusNoContent)
		return
	}

	klog.Infof("Checking selector syntax...")
	if err := common.CheckSelector(selector); err != nil {
		klog.Errorf("Error checking the selector syntax: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	klog.Infof("Filtering Flavours by selector...")
	flavoursSelected, err := common.FilterFlavoursBySelector(availableFlavours, selector)
	if err != nil {
		http.Error(w, "Error getting the Flavours by selector", http.StatusInternalServerError)
		return
	}

	klog.Infof("Flavours found that match the selector are: %d", len(flavoursSelected))

	if len(flavoursSelected) == 0 {
		klog.Infof("No matching Flavours found")
		// Return content for empty list
		emptyList := make([]*nodecorev1alpha1.Flavour, 0)
		encodeResponse(w, emptyList)
		return
	}

	// Select the flavour with the max CPU
	max := resource.MustParse("0")
	index := 0

	for i := range flavoursSelected {
		if flavoursSelected[i].Spec.Characteristics.Cpu.Cmp(max) == 1 {
			max = flavoursSelected[i].Spec.Characteristics.Cpu
			index = i
		}
	}

	selected := *flavoursSelected[index].DeepCopy()

	klog.Infof("Flavour %s selected - Parsing...", selected.Name)
	parsed := parseutil.ParseFlavour(&selected)

	klog.Infof("Flavour parsed: %v", parsed)

	// Encode the Flavour as JSON and write it to the response writer
	encodeResponse(w, parsed)
}

// reserveFlavour reserves a Flavour by its flavourID.
func (g *Gateway) reserveFlavour(w http.ResponseWriter, r *http.Request) {
	// Get the flavourID value from the URL parameters
	params := mux.Vars(r)
	flavourID := params["flavourID"]
	var transaction *models.Transaction
	var request models.ReserveRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		klog.Errorf("Error decoding the ReserveRequest: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	klog.Infof("Partition: %v", *request.Partition)

	if flavourID != request.FlavourID {
		klog.Infof("Mismatch body & param: %s != %s", flavourID, request.FlavourID)
		http.Error(w, "Mismatch body & param", http.StatusConflict)
		return
	}

	// Check if the Transaction already exists
	t, found := g.SearchTransaction(request.Buyer.NodeID, flavourID)
	if found {
		t.StartTime = tools.GetTimeNow()
		transaction = t
		g.addNewTransacion(t)
	}

	if !found {
		klog.Infof("Reserving flavour %s started", flavourID)

		flavour, _ := services.GetFlavourByID(flavourID, g.client)
		if flavour == nil {
			http.Error(w, "Flavour not found", http.StatusNotFound)
			return
		}

		// Create a new transaction ID
		transactionID, err := namings.ForgeTransactionID()
		if err != nil {
			http.Error(w, "Error generating transaction ID", http.StatusInternalServerError)
			return
		}

		// Create a new transaction
		transaction = resourceforge.ForgeTransactionObj(transactionID, &request)

		// Add the transaction to the transactions map
		g.addNewTransacion(transaction)
	}

	klog.Infof("Transaction %s reserved", transaction.TransactionID)

	encodeResponse(w, transaction)
}

// purchaseFlavour is an handler for purchasing a Flavour.
func (g *Gateway) purchaseFlavour(w http.ResponseWriter, r *http.Request) {
	// Get the flavourID value from the URL parameters
	params := mux.Vars(r)
	transactionID := params["transactionID"]
	var purchase models.PurchaseRequest

	if err := json.NewDecoder(r.Body).Decode(&purchase); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if transactionID != purchase.TransactionID {
		klog.Infof("Mismatch body & param")
		http.Error(w, "Mismatch body & param", http.StatusConflict)
		return
	}

	klog.Infof("Purchasing request for transaction %s", purchase.TransactionID)

	// Retrieve the transaction from the transactions map
	transaction, err := g.GetTransaction(purchase.TransactionID)
	if err != nil {
		klog.Errorf("Error getting the Transaction: %s", err)
		http.Error(w, "Error getting the Transaction", http.StatusInternalServerError)
		return
	}

	klog.Infof("Flavour requested: %s", transaction.FlavourID)

	if tools.CheckExpiration(transaction.StartTime, flags.ExpirationTransaction) {
		klog.Infof("Transaction %s expired", transaction.TransactionID)
		http.Error(w, "Error: transaction Timeout", http.StatusRequestTimeout)
		g.removeTransaction(transaction.TransactionID)
		return
	}

	var contractList reservationv1alpha1.ContractList
	var contract reservationv1alpha1.Contract

	// Check if the Contract with the same TransactionID already exists
	if err := g.client.List(context.Background(), &contractList, client.MatchingFields{"spec.transactionID": purchase.TransactionID}); err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Errorf("Error when listing Contracts: %s", err)
			http.Error(w, "Error when listing Contracts", http.StatusInternalServerError)
			return
		}
	}

	if len(contractList.Items) > 0 {
		klog.Infof("Contract already exists for transaction %s", purchase.TransactionID)
		contract = contractList.Items[0]
		// Create a contract object to be returned with the response
		contractObject := parseutil.ParseContract(&contract)
		// create a response purchase
		responsePurchase := resourceforge.ForgeResponsePurchaseObj(contractObject)
		// Respond with the response purchase as JSON
		encodeResponse(w, responsePurchase)
		return
	}

	klog.Infof("Performing purchase of flavour %s...", transaction.FlavourID)

	// Remove the transaction from the transactions map
	g.removeTransaction(transaction.TransactionID)

	klog.Infof("Flavour %s successfully purchased!", transaction.FlavourID)

	// Get the flavour sold for creating the contract
	flavourSold, err := services.GetFlavourByID(transaction.FlavourID, g.client)
	if err != nil {
		klog.Errorf("Error getting the Flavour by ID: %s", err)
		http.Error(w, "Error getting the Flavour by ID", http.StatusInternalServerError)
		return
	}

	liqoCredentials, err := getters.GetLiqoCredentials(context.Background(), g.client)
	if err != nil {
		klog.Errorf("Error getting Liqo Credentials: %s", err)
		http.Error(w, "Error getting Liqo Credentials", http.StatusInternalServerError)
		return
	}

	// Create a new contract
	klog.Infof("Creating a new contract...")
	contract = *resourceforge.ForgeContract(flavourSold, &transaction, liqoCredentials)
	err = g.client.Create(context.Background(), &contract)
	if err != nil {
		klog.Errorf("Error creating the Contract: %s", err)
		http.Error(w, "Error creating the Contract: "+err.Error(), http.StatusInternalServerError)
		return
	}

	klog.Infof("Contract created!")

	// Create a contract object to be returned with the response
	contractObject := parseutil.ParseContract(&contract)
	// create a response purchase
	responsePurchase := resourceforge.ForgeResponsePurchaseObj(contractObject)

	klog.Infof("Contract %v", *contractObject.Partition)

	// Create allocation
	klog.Infof("Creating allocation...")
	workerName := contract.Spec.Flavour.Spec.OptionalFields.WorkerID
	allocation := *resourceforge.ForgeAllocation(&contract, "", workerName, nodecorev1alpha1.Remote, nodecorev1alpha1.Node)
	err = g.client.Create(context.Background(), &allocation)
	if err != nil {
		klog.Errorf("Error creating the Allocation: %s", err)
		http.Error(w, "Contract created but we ran into an error while allocating the resources", http.StatusInternalServerError)
		return
	}

	klog.Infof("Response purchase %v", *responsePurchase.Contract.Partition)

	// Respond with the response purchase as JSON
	encodeResponse(w, responsePurchase)
}
