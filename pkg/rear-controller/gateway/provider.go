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

	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	reservationv1alpha1 "fluidos.eu/node/api/reservation/v1alpha1"
	"fluidos.eu/node/pkg/utils/common"
	"fluidos.eu/node/pkg/utils/flags"
	"fluidos.eu/node/pkg/utils/models"
	"fluidos.eu/node/pkg/utils/namings"
	"fluidos.eu/node/pkg/utils/parseutil"
	"fluidos.eu/node/pkg/utils/resourceforge"
	"fluidos.eu/node/pkg/utils/services"
	"fluidos.eu/node/pkg/utils/tools"
)

// TODO: all these functions should be moved into the REAR Gateway package

// getFlavours gets all the flavours CRs from the cluster
func (g *Gateway) getFlavours(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	flavours, err := services.GetAllFlavours(g.client)
	if err != nil {
		klog.Errorf("Error getting all the Flavour CRs: %s", err)
		http.Error(w, "Error getting all the Flavour CRs", http.StatusInternalServerError)
		return
	}

	klog.Infof("Listing all the Flavour CRs: %d", len(flavours))

	// Export the result as a list of Flavour struct (no CRs)
	var flavoursParsed []models.Flavour
	for _, f := range flavours {
		flavoursParsed = append(flavoursParsed, parseutil.ParseFlavour(f))
	}

	// Encode the FlavourList as JSON and write it to the response writer
	encodeResponse(w, flavoursParsed)

}

// getFlavourByID gets the flavour CR from the cluster that matches the flavourID
func (g *Gateway) getFlavourByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the flavourID from the URL
	params := mux.Vars(r)
	flavourID := params["flavourID"]

	// Get the Flavour that matches the flavourID
	flavour, err := services.GetFlavourByID(flavourID, g.client)
	if err != nil {
		http.Error(w, "Error getting the Flavour by ID", http.StatusInternalServerError)
		return
	}

	if flavour == nil {
		http.Error(w, "No Flavour found", http.StatusNotFound)
		return
	}

	flavourParsed := parseutil.ParseFlavour(*flavour)

	klog.Infof("Flavour found is: %s", flavourParsed.FlavourID)

	// Encode the FlavourList as JSON and write it to the response writer
	encodeResponse(w, flavourParsed)

}

// getFlavourBySelectorHandler gets the flavour CRs from the cluster that match the selector
func (g *Gateway) getFlavoursBySelector(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	klog.Infof("Getting Flavours by selector...")

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

	klog.Infof("Selector is: %s", selector)

	flavours, err := services.GetAllFlavours(g.client)
	if err != nil {
		klog.Errorf("Error getting all the Flavour CRs: %s", err)
		http.Error(w, "Error getting all the Flavour CRs", http.StatusInternalServerError)
		return
	}

	klog.Infof("Listing all the Flavour CRs: %d", len(flavours))

	// Filter flavours from array maintaining only the Spec.Available == true
	for i, f := range flavours {
		if !f.Spec.OptionalFields.Availability {
			flavours = append(flavours[:i], flavours[i+1:]...)
		}
	}

	klog.Infof("Available Flavours: %d", len(flavours))
	if len(flavours) == 0 {
		klog.Infof("No available Flavours found")
		http.Error(w, "No Flavours found", http.StatusNotFound)
		return
	}

	klog.Infof("Checking selector syntax...")
	if err := common.CheckSelector(selector); err != nil {
		klog.Errorf("Error checking the selector syntax: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	klog.Infof("Filtering Flavours by selector...")
	flavoursSelected, err := common.FilterFlavoursBySelector(flavours, selector)
	if err != nil {
		http.Error(w, "Error getting the Flavours by selector", http.StatusInternalServerError)
		return
	}

	klog.Infof("Flavours found that match the selector are: %d", len(flavoursSelected))

	if len(flavoursSelected) == 0 {
		klog.Infof("No matching Flavours found")
		http.Error(w, "No Flavours found", http.StatusNotFound)
		return
	}

	// Select the flavour with the max CPU
	max := resource.MustParse("0")
	var selected nodecorev1alpha1.Flavour
	for _, f := range flavoursSelected {
		if f.Spec.Characteristics.Cpu.Cmp(max) == 1 {
			max = f.Spec.Characteristics.Cpu
			selected = f
		}
	}

	klog.Infof("Flavour %s selected - Parsing...", selected.Name)
	parsed := parseutil.ParseFlavour(selected)

	// Encode the FlavourList as JSON and write it to the response writer
	encodeResponse(w, parsed)
}

// reserveFlavour reserves a Flavour by its flavourID
func (g *Gateway) reserveFlavour(w http.ResponseWriter, r *http.Request) {
	// Get the flavourID value from the URL parameters
	params := mux.Vars(r)
	flavourID := params["flavourID"]
	var transaction models.Transaction
	var request models.ReserveRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if flavourID != request.FlavourID {
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
		transaction := resourceforge.ForgeTransactionObj(transactionID, request)

		// Add the transaction to the transactions map
		g.addNewTransacion(transaction)
	}

	encodeResponse(w, transaction)
}

// purchaseFlavour is an handler for purchasing a Flavour
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

	if tools.CheckExpiration(transaction.StartTime, flags.EXPIRATION_TRANSACTION) {
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

	// Create a new contract
	klog.Infof("Creating a new contract...")
	contract = *resourceforge.ForgeContract(*flavourSold, transaction)
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

	// Respond with the response purchase as JSON
	encodeResponse(w, responsePurchase)
}
