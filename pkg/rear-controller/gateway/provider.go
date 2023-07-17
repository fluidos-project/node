package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
)

// TODO: all these functions should be moved into the REAR Gateway package

type Gateway struct {
	ID *nodecorev1alpha1.NodeIdentity

	// TransactionCache is the current transaction used in the Reservation controller
	//TransactionCache models.Transaction

	// Transactions is a map of Transaction
	Transactions map[string]models.Transaction

	// FlavourPeeringCandidateMap is a map of Flavour and PeeringCandidate
	//FlavourPeeringCandidateMap = make(map[string]string)

	client client.Client
}

func NewGateway(c client.Client) *Gateway {
	return &Gateway{
		client:       c,
		Transactions: make(map[string]models.Transaction),
	}
}

// StartHttpServer starts a new HTTP server
func (g *Gateway) StartHttpServer() {
	// mux creation
	router := mux.NewRouter()

	// routes definition
	router.HandleFunc("/api/listflavours", g.listAllFlavoursHandler).Methods("GET")
	router.HandleFunc("/api/listflavours/{flavourID}", g.getFlavourByIDHandler).Methods("GET")
	router.HandleFunc("/api/listflavours/selector", g.getFlavoursBySelectorHandler).Methods("POST")
	router.HandleFunc("/api/reserveflavour/{flavourID}", g.reserveFlavourHandler).Methods("POST")
	router.HandleFunc("/api/purchaseflavour/{flavourID}", g.purchaseFlavourHandler).Methods("POST")

	// Start server HTTP
	klog.Infof("Starting HTTP server on port 14144")
	// TODO: after the demo recover correct address (14144)
	log.Fatal(http.ListenAndServe(flags.HTTP_PORT, router))

}

// listAllFlavoursHandler gets all the flavours CRs from the cluster
func (g *Gateway) listAllFlavoursHandler(w http.ResponseWriter, r *http.Request) {
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

// SearchFlavour is a function that returns an array of Flavour that fit the Selector by performing a get request to an http server
func (g *Gateway) SearchFlavours(selector nodecorev1alpha1.FlavourSelector) ([]*nodecorev1alpha1.Flavour, error) {
	// Marshal the selector into JSON bytes
	s := parseutil.ParseFlavourSelector(selector)

	// Create the Flavour CR from the first flavour in the array of Flavour
	var flavoursCR []*nodecorev1alpha1.Flavour

	// Send the POST request to all the servers in the list
	for _, ADDRESS := range flags.SERVER_ADDRESSES {
		flavours, err := searchFlavour(s, ADDRESS)
		if err != nil {
			return nil, err
		}
		flavoursCR = append(flavoursCR, flavours...)
	}

	klog.Info("Flavours created", flavoursCR)
	return flavoursCR, nil
}

func searchFlavour(selector models.Selector, addr string) ([]*nodecorev1alpha1.Flavour, error) {
	var flavours []*models.Flavour
	var flavoursCR []*nodecorev1alpha1.Flavour

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

	if err := json.NewDecoder(resp.Body).Decode(&flavours); err != nil {
		return nil, err
	}

	for _, flavour := range flavours {
		klog.Infof("Flavour found: %s", flavour.FlavourID)
		cr := resourceforge.ForgeFlavourFromObj(*flavour)
		flavoursCR = append(flavoursCR, cr)
	}

	return flavoursCR, nil
}

// getFlavourByIDHandler gets the flavour CR from the cluster that matches the flavourID
func (g *Gateway) getFlavourByIDHandler(w http.ResponseWriter, r *http.Request) {
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
func (g *Gateway) getFlavoursBySelectorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	flavours, err := services.GetAllFlavours(g.client)
	if err != nil {
		klog.Errorf("Error getting all the Flavour CRs: %s", err)
		http.Error(w, "Error getting all the Flavour CRs", http.StatusInternalServerError)
		return
	}

	// Filter flavours from array maintaining only the Spec.Available == true
	for i, f := range flavours {
		if !f.Spec.OptionalFields.Availability {
			flavours = append(flavours[:i], flavours[i+1:]...)
		}
	}

	if len(flavours) == 0 {
		http.Error(w, "No Flavours found", http.StatusNotFound)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// build the selector from the request body
	selector, err := buildSelector(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := common.CheckSelector(selector); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	flavoursSelected, err := common.FilterFlavoursBySelector(flavours, selector)
	if err != nil {
		http.Error(w, "Error getting the Flavours by selector", http.StatusInternalServerError)
		return
	}

	if len(flavoursSelected) == 0 {
		http.Error(w, "No Flavours found", http.StatusNotFound)
		return
	}

	klog.Infof("Flavours found that match the selector are: %d", len(flavours))

	// Select the flavour with the max CPU
	max := resource.MustParse("0")
	var selected nodecorev1alpha1.Flavour
	for _, f := range flavoursSelected {
		if f.Spec.Characteristics.Cpu.Cmp(max) == 1 {
			max = f.Spec.Characteristics.Cpu
			selected = f
		}
	}

	// Encode the FlavourList as JSON and write it to the response writer
	encodeResponse(w, parseutil.ParseFlavour(selected))

}

// reserveFlavourHandler reserves a Flavour by its flavourID
func (g *Gateway) reserveFlavourHandler(w http.ResponseWriter, r *http.Request) {
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
		t.StartTime = common.GetTimeNow()
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

// purchaseFlavourHandler is an handler for purchasing a Flavour
func (g *Gateway) purchaseFlavourHandler(w http.ResponseWriter, r *http.Request) {
	// Get the flavourID value from the URL parameters
	params := mux.Vars(r)
	flavourID := params["flavourID"]
	var purchase models.PurchaseRequest

	if err := json.NewDecoder(r.Body).Decode(&purchase); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	klog.Infof("Purchasing flavour %s started", flavourID)

	// Retrieve the transaction from the transactions map
	transaction, err := g.GetTransaction(purchase.TransactionID)
	if err != nil {
		klog.Errorf("Error getting the Transaction: %s", err)
		http.Error(w, "Error getting the Transaction", http.StatusInternalServerError)
		return
	}

	if common.CheckExpiration(transaction.StartTime, flags.EXPIRATION_TRANSACTION) {
		http.Error(w, "Error: transaction Timeout", http.StatusRequestTimeout)
		g.removeTransaction(transaction.TransactionID)
		return
	}

	var contractList *reservationv1alpha1.ContractList
	var contract *reservationv1alpha1.Contract

	// Check if the Contract with the same TransactionID already exists
	if err := g.client.List(context.Background(), contractList, client.MatchingFields{"spec.TransactionID": purchase.TransactionID}); err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Errorf("Error when listing Contracts: %s", err)
			http.Error(w, "Error when listing Contracts", http.StatusInternalServerError)
			return
		}
	}

	if len(contractList.Items) > 0 {
		contract = &contractList.Items[0]
		// Create a contract object to be returned with the response
		contractObject := parseutil.ParseContract(contract)
		// create a response purchase
		responsePurchase := resourceforge.ForgeResponsePurchaseObj(contractObject)
		// Respond with the response purchase as JSON
		encodeResponse(w, responsePurchase)
		return
	}

	klog.Infof("Performing purchase of flavour %s...", flavourID)

	// Remove the transaction from the transactions map
	g.removeTransaction(transaction.TransactionID)

	klog.Infof("Flavour %s purchased!", flavourID)

	// Get the flavour sold for creating the contract
	flavourSold, err := services.GetFlavourByID(flavourID, g.client)
	if err != nil {
		klog.Errorf("Error getting the Flavour by ID: %s", err)
		http.Error(w, "Error getting the Flavour by ID", http.StatusInternalServerError)
		return
	}

	// Create a new contract
	klog.Infof("Creating a new contract...")
	contract = resourceforge.ForgeContract(*flavourSold, transaction)
	err = g.client.Create(context.Background(), contract)
	if err != nil {
		klog.Errorf("Error creating the Contract: %s", err)
		http.Error(w, "Error creating the Contract: "+err.Error(), http.StatusInternalServerError)
		return
	}

	klog.Infof("Contract created!")

	// Create a contract object to be returned with the response
	contractObject := parseutil.ParseContract(contract)
	// create a response purchase
	responsePurchase := resourceforge.ForgeResponsePurchaseObj(contractObject)

	// Respond with the response purchase as JSON
	encodeResponse(w, responsePurchase)

}

// buildSelector builds a selector from a request body
func buildSelector(body []byte) (models.Selector, error) {
	// Parse the request body into the APIRequest struct
	var selector models.Selector
	err := json.Unmarshal(body, &selector)
	if err != nil {
		return models.Selector{}, err
	}
	return selector, nil
}

// addNewTransacion add a new transaction to the transactions map
func (g *Gateway) addNewTransacion(transaction models.Transaction) {
	g.Transactions[transaction.TransactionID] = transaction
}

// removeTransaction removes a transaction from the transactions map
func (g *Gateway) removeTransaction(transactionID string) {
	delete(g.Transactions, transactionID)
}

// getTransaction returns a transaction from the transactions map
func (g *Gateway) GetTransaction(transactionID string) (models.Transaction, error) {
	transaction, exists := g.Transactions[transactionID]
	if !exists {
		return models.Transaction{}, fmt.Errorf("Transaction not found")
	}
	return transaction, nil
}

// SearchTransaction returns a transaction from the transactions map
func (g *Gateway) SearchTransaction(buyerID string, flavourID string) (models.Transaction, bool) {
	for _, t := range g.Transactions {
		if t.Buyer.NodeID == buyerID && t.FlavourID == flavourID {
			return t, true
		}
	}
	return models.Transaction{}, false
}

// handleError handles errors by sending an error response
func handleError(w http.ResponseWriter, err error, statusCode int) {
	http.Error(w, err.Error(), statusCode)
}

// encodeResponse encodes the response as JSON and writes it to the response writer
func encodeResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
	}
}
