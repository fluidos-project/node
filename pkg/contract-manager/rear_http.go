package contractmanager

import (
	"context"
	"encoding/json"
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

// StartHttpServer starts a new HTTP server
func StartHttpServer(cl client.Client) {
	// mux creation
	router := mux.NewRouter()

	// routes definition
	router.HandleFunc("/api/listflavours", listAllFlavoursHandler(cl)).Methods("GET")
	router.HandleFunc("/api/listflavours/{flavourID}", getFlavourByIDHandler(cl)).Methods("GET")
	router.HandleFunc("/api/listflavours/selector", getFlavoursBySelectorHandler(cl)).Methods("POST")
	router.HandleFunc("/api/reserveflavour/{flavourID}", reserveFlavourHandler(cl)).Methods("POST")
	router.HandleFunc("/api/purchaseflavour/{flavourID}", purchaseFlavourHandler(cl)).Methods("POST")

	// Start server HTTP
	klog.Infof("Starting HTTP server on port 14144")
	// TODO: after the demo recover correct address (14144)
	log.Fatal(http.ListenAndServe(flags.HTTP_PORT, router))

}

// listAllFlavoursHandler gets all the flavours CRs from the cluster
func listAllFlavoursHandler(cl client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		flavours, err := services.GetAllFlavours(cl)
		if err != nil {
			klog.Errorf("Error getting all the Flavour CRs: %s", err)
			http.Error(w, "Error getting all the Flavour CRs", http.StatusInternalServerError)
			return
		}

		klog.Infof("Listing all the Flavour CRs: %d", len(flavours))

		// Export the result as a list of Flavour struct (no CRs)
		var flavoursParsed []models.Flavour
		for _, f := range flavours {
			flavoursParsed = append(flavoursParsed, *parseutil.ParseCRToFlavour(f))
		}

		// Encode the FlavourList as JSON and write it to the response writer
		encodeResponse(w, flavoursParsed)
	}
}

// getFlavourByIDHandler gets the flavour CR from the cluster that matches the flavourID
func getFlavourByIDHandler(cl client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Get the flavourID from the URL
		params := mux.Vars(r)
		flavourID := params["flavourID"]

		// Get the Flavour that matches the flavourID
		flavour, err := services.GetFlavourByID(flavourID, cl)
		if err != nil {
			http.Error(w, "Error getting the Flavour by ID", http.StatusInternalServerError)
			return
		}

		if flavour == nil {
			http.Error(w, "No Flavour found", http.StatusNotFound)
			return
		}

		flavourParsed := parseutil.ParseCRToFlavour(*flavour)

		klog.Infof("Flavour found is: %s", flavourParsed.FlavourID)

		// Encode the FlavourList as JSON and write it to the response writer
		encodeResponse(w, flavourParsed)
	}
}

// getFlavourBySelectorHandler gets the flavour CRs from the cluster that match the selector
func getFlavoursBySelectorHandler(cl client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		flavours, err := services.GetAllFlavours(cl)
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

		if err := services.CheckSelector(selector); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		flavoursSelected, err := services.FilterFlavoursBySelector(flavours, selector)
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
		encodeResponse(w, parseutil.ParseCRToFlavour(selected))
	}
}

// reserveFlavourHandler reserves a Flavour by its flavourID
func reserveFlavourHandler(cl client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the flavourID value from the URL parameters
		params := mux.Vars(r)
		flavourID := params["flavourID"]
		var transaction models.Transaction

		var request struct {
			FlavourID string       `json:"flavourID"`
			Buyer     models.Owner `json:"buyer"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if flavourID != request.FlavourID {
			http.Error(w, "Mismatch body & param", http.StatusConflict)
			return
		}

		if models.Transactions == nil {
			models.Transactions = make(map[string]models.Transaction)
		}

		// Check if the Transaction already exists
		found := false
		for _, t := range models.Transactions {
			if t.FlavourID == flavourID && t.Buyer.ID == request.Buyer.ID {
				found = true
				t.StartTime = common.GetTimeNow()
				transaction = t
				addNewTransacion(t)
				break
			}
		}

		if !found {
			klog.Infof("Reserving flavour %s started", flavourID)

			flavour, _ := services.GetFlavourByID(flavourID, cl)
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
			transaction := resourceforge.ForgeTransactionObject(flavourID, transactionID, request.Buyer)

			// Add the transaction to the transactions map
			addNewTransacion(transaction)
		}

		encodeResponse(w, transaction)
	}
}

// purchaseFlavourHandler is an handler for purchasing a Flavour
func purchaseFlavourHandler(cl client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the flavourID value from the URL parameters
		params := mux.Vars(r)
		flavourID := params["flavourID"]
		var purchase models.Purchase

		if err := json.NewDecoder(r.Body).Decode(&purchase); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		klog.Infof("Purchasing flavour %s started", flavourID)

		// Check if the transactions map is nil or not initialized
		if models.Transactions == nil {
			klog.Infof("No active transactions found")
			http.Error(w, "Error: no active transactions found.", http.StatusNotFound)
			return
		}

		// Retrieve the transaction from the transactions map
		transaction, exists := models.Transactions[purchase.TransactionID]
		if !exists {
			klog.Infof("Transaction not found")
			http.Error(w, "Error: transaction not found", http.StatusNotFound)
			return
		}

		if common.CheckExpiration(transaction.StartTime, flags.EXPIRATION_TRANSACTION) {
			http.Error(w, "Error: transaction Timeout", http.StatusRequestTimeout)
			delete(models.Transactions, purchase.TransactionID)
			return
		}

		var contractList *reservationv1alpha1.ContractList
		var contract *reservationv1alpha1.Contract

		// Check if the Contract with the same TransactionID already exists
		if err := cl.List(context.Background(), contractList, client.MatchingFields{"spec.TransactionID": purchase.TransactionID}); err != nil {
			if client.IgnoreNotFound(err) != nil {
				klog.Errorf("Error when listing Contracts: %s", err)
				http.Error(w, "Error when listing Contracts", http.StatusInternalServerError)
				return
			}
		}

		if len(contractList.Items) > 0 {
			contract = &contractList.Items[0]
			// Create a contract object to be returned with the response
			contractObject := resourceforge.ForgeContractObject(contract, purchase.BuyerID, transaction.TransactionID, purchase.Partition)
			// create a response purchase
			responsePurchase := resourceforge.ForgeResponsePurchaseObject(contractObject)
			// Respond with the response purchase as JSON
			encodeResponse(w, responsePurchase)
			return
		}

		klog.Infof("Performing purchase of flavour %s...", flavourID)

		// Remove the transaction from the transactions map
		delete(models.Transactions, purchase.TransactionID)

		klog.Infof("Flavour %s purchased!", flavourID)

		// Get the flavour sold for creating the contract
		flavourSold, err := services.GetFlavourByID(flavourID, cl)
		if err != nil {
			klog.Errorf("Error getting the Flavour by ID: %s", err)
			http.Error(w, "Error getting the Flavour by ID", http.StatusInternalServerError)
			return
		}

		// Create a new contract
		klog.Infof("Creating a new contract...")
		contract = resourceforge.ForgeContractCustomResource(*flavourSold, purchase.BuyerID)
		err = cl.Create(context.Background(), contract)
		if err != nil {
			klog.Errorf("Error creating the Contract: %s", err)
			http.Error(w, "Error creating the Contract: "+err.Error(), http.StatusInternalServerError)
			return
		}

		klog.Infof("Contract created!")

		// Create a contract object to be returned with the response
		contractObject := resourceforge.ForgeContractObject(contract, purchase.BuyerID, transaction.TransactionID, purchase.Partition)
		// create a response purchase
		responsePurchase := resourceforge.ForgeResponsePurchaseObject(contractObject)

		// Respond with the response purchase as JSON
		encodeResponse(w, responsePurchase)

	}
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
func addNewTransacion(transaction models.Transaction) {
	// Save the transaction in the transactions map
	models.Transactions[transaction.TransactionID] = transaction
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
