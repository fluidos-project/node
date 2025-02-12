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
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/common"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/namings"
	"github.com/fluidos-project/node/pkg/utils/parseutil"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
	"github.com/fluidos-project/node/pkg/utils/services"
	"github.com/fluidos-project/node/pkg/utils/tools"
)

// getFlavors gets all the flavors CRs from the cluster.
func (g *Gateway) getFlavors(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	klog.Infof("Processing request for getting all Flavors...")

	flavors, err := services.GetAllFlavors(g.client)
	if err != nil {
		klog.Errorf("Error getting all the Flavor CRs: %s", err)
		http.Error(w, "Error getting all the Flavor CRs", http.StatusInternalServerError)
		return
	}

	klog.Infof("Found %d Flavors in the cluster", len(flavors))

	availableFlavors := make([]nodecorev1alpha1.Flavor, 0)

	// Filtering only the available flavors
	for i := range flavors {
		if flavors[i].Spec.Availability {
			availableFlavors = append(availableFlavors, flavors[i])
		}
	}

	klog.Infof("Available Flavors: %d", len(availableFlavors))
	if len(availableFlavors) == 0 {
		klog.Infof("No available Flavors found")
		// Return content for empty list
		emptyList := make([]*nodecorev1alpha1.Flavor, 0)
		encodeResponseStatusCode(w, emptyList, http.StatusNoContent)
		return
	}

	// Parse the flavors CR to the models.Flavor struct
	flavorsParsed := make([]models.Flavor, 0)
	for i := range availableFlavors {
		parsedFlavor := parseutil.ParseFlavor(&availableFlavors[i])
		if parsedFlavor == nil {
			klog.Errorf("Error parsing the Flavor: %s", err)
			continue
		}
		flavorsParsed = append(flavorsParsed, *parsedFlavor)
	}

	// Encode the Flavor as JSON and write it to the response writer
	encodeResponse(w, flavorsParsed)
}

// getFlavorBySelectorHandler gets the flavor CRs from the cluster that match the selector.
func (g *Gateway) getK8SliceFlavorsBySelector(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	klog.Infof("Processing request for getting K8Slice Flavors by selector...")

	klog.Infof("URL: %s", r.URL.String())

	// build the selector from the url query parameters
	selector, err := queryParamToSelector(r.URL.Query(), models.K8SliceNameDefault)
	if err != nil {
		klog.Errorf("Error building the selector from the URL query parameters: %s", err)
		http.Error(w, "Error building the selector from the URL query parameters", http.StatusBadRequest)
		return
	}

	// Print the selector information parsing it:
	klog.Infof("Selector type: %s", selector.GetSelectorType())

	flavors, err := services.GetAvailableFlavors(g.client)
	if err != nil {
		klog.Errorf("Error getting all the Flavor CRs: %s", err)
		http.Error(w, "Error getting all the Flavor CRs", http.StatusInternalServerError)
		return
	}

	klog.Infof("Checking selector syntax...")
	if err := common.CheckSelector(selector); err != nil {
		klog.Errorf("Error checking the selector syntax: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	klog.Infof("Filtering Flavors by selector...")
	flavorsSelected, err := common.FilterFlavorsBySelector(flavors, selector)
	if err != nil {
		http.Error(w, "Error getting the Flavors by selector", http.StatusInternalServerError)
		return
	}

	klog.Infof("Flavors found that match the selector are: %d", len(flavorsSelected))

	if len(flavorsSelected) == 0 {
		klog.Infof("No matching Flavors found")
		// Return content for empty list
		emptyList := make([]models.Flavor, 0)
		encodeResponse(w, emptyList)
		return
	}

	// Parse the flavors CR to the models.Flavor struct
	flavorsParsed := make([]models.Flavor, 0)
	for i := range flavorsSelected {
		parsedFlavor := parseutil.ParseFlavor(&flavors[i])
		if parsedFlavor == nil {
			klog.Errorf("Error parsing the Flavor: %s", err)
			continue
		}
		flavorsParsed = append(flavorsParsed, *parsedFlavor)
	}

	// Encode the Flavor as JSON and write it to the response writer
	encodeResponse(w, flavorsParsed)
}

// getServiceFlavorsBySelector gets the flavor CRs from the cluster that match the selector.
func (g *Gateway) getServiceFlavorsBySelector(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	klog.Infof("Processing request for getting Service Flavors by selector...")

	klog.Infof("URL: %s", r.URL.String())

	// build the selector from the url query parameters
	selector, err := queryParamToSelector(r.URL.Query(), models.ServiceNameDefault)
	if err != nil {
		klog.Errorf("Error building the selector from the URL query parameters: %s", err)
		http.Error(w, "Error building the selector from the URL query parameters", http.StatusBadRequest)
		return
	}

	// Print the selector information parsing it:
	klog.Infof("Selector type: %s", selector.GetSelectorType())

	// Get the available flavors
	flavors, err := services.GetAvailableFlavors(g.client)
	if err != nil {
		klog.Errorf("Error getting the available Flavors: %s", err)
		http.Error(w, "Error getting the available Flavors", http.StatusInternalServerError)
		return
	}

	klog.Infof("Checking selector syntax...")
	if err := common.CheckSelector(selector); err != nil {
		klog.Errorf("Error checking the selector syntax: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	klog.Infof("Filtering Flavors by selector...")
	flavorsSelected, err := common.FilterFlavorsBySelector(flavors, selector)
	if err != nil {
		http.Error(w, "Error getting the Flavors by selector", http.StatusInternalServerError)
		return
	}

	klog.Infof("Flavors found that match the selector are: %d", len(flavorsSelected))

	if len(flavorsSelected) == 0 {
		klog.Infof("No matching Flavors found")
		// Return content for empty list
		emptyList := make([]models.Flavor, 0)
		encodeResponse(w, emptyList)
		return
	}

	// Parse the flavors CR to the models.Flavor struct
	flavorsParsed := make([]models.Flavor, 0)
	for i := range flavorsSelected {
		parsedFlavor := parseutil.ParseFlavor(&flavorsSelected[i])
		if parsedFlavor == nil {
			klog.Errorf("Error parsing the Flavor: %s", err)
			continue
		}
		flavorsParsed = append(flavorsParsed, *parsedFlavor)
	}

	// Encode the Flavor as JSON and write it to the response writer
	encodeResponse(w, flavorsParsed)
}

// reserveFlavor reserves a Flavor by its flavorID.
func (g *Gateway) reserveFlavor(w http.ResponseWriter, r *http.Request) {
	// Get the flavorID value from the URL parameters
	var transaction *models.Transaction
	var request models.ReserveRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		klog.Errorf("Error decoding the ReserveRequest: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	flavorID := request.FlavorID

	// Get the flavor by ID
	flavor, err := services.GetFlavorByID(flavorID, g.client)
	if err != nil {
		klog.Errorf("Error getting the Flavor by ID: %s", err)
		http.Error(w, "Error getting the Flavor by ID", http.StatusInternalServerError)
		return
	}
	if flavor == nil {
		klog.Errorf("Flavor %s not found", flavorID)
		http.Error(w, "Flavor not found", http.StatusNotFound)
		return
	}
	// Get the flavor type
	flavorTypeIdentifier, flavorData, err := nodecorev1alpha1.ParseFlavorType(flavor)
	if err != nil {
		klog.Errorf("Error parsing the Flavor type: %s", err)
		http.Error(w, "Error parsing the Flavor type", http.StatusInternalServerError)
		return
	}
	// Check if configuration is valid, based on the Flavor the client wants to reserve
	switch flavorTypeIdentifier {
	case nodecorev1alpha1.TypeK8Slice:
		// The configuration is not necessary for the K8Slice flavor
		if request.Configuration == nil {
			klog.Info("No configuration provided for K8Slice flavor")
		} else {
			// Configuration type is the K8SliceConfiguration
			if request.Configuration.Type != models.K8SliceNameDefault {
				klog.Errorf("Configuration type %s not supported", request.Configuration.Type)
				http.Error(w, "Configuration type not supported", http.StatusBadRequest)
				return
			}
			// Forge the configuration object
			configuration, err := resourceforge.ForgeConfigurationFromObj(*request.Configuration)
			if err != nil {
				klog.Errorf("Error forging the configuration object: %s", err)
				http.Error(w, "Error forging the configuration object", http.StatusInternalServerError)
				return
			}
			// Parse the configuration
			_, _, err = nodecorev1alpha1.ParseConfiguration(configuration, flavor)
			if err != nil {
				klog.Errorf("Error parsing the configuration: %s", err)
				http.Error(w, "Error parsing the configuration", http.StatusInternalServerError)
				return
			}
			// No further checks are needed for the K8Slice flavor configuration
			// TODO(K8Slice): Implement the K8Slice flavor configuration checks if needed
		}
	case nodecorev1alpha1.TypeVM:
		// TODO (VM): Implement the VM flavor configuration
		klog.Errorf("Flavor type %s not supported", flavorTypeIdentifier)
		http.Error(w, "Flavor type not supported", http.StatusBadRequest)
		return
	case nodecorev1alpha1.TypeService:
		if request.Configuration == nil {
			klog.Errorf("No configuration provided for Service flavor")

			// Need to check that the flavor can be used without a configuration
			// Create fake empty service configuration
			emptyServiceConfiguration := nodecorev1alpha1.ServiceConfiguration{
				HostingPolicy: nil,
				ConfigurationData: runtime.RawExtension{
					Raw: []byte("{}"),
				},
			}

			// Force casting the flavor data to the ServiceFlavor type
			serviceFlavor, ok := flavorData.(*nodecorev1alpha1.ServiceFlavor)
			if !ok {
				klog.Errorf("Error casting the flavor data to ServiceFlavor")
				http.Error(w, "Error casting the flavor data to ServiceFlavor", http.StatusInternalServerError)
				return
			}

			// Validate the empty service configuration
			if err := emptyServiceConfiguration.Validate(serviceFlavor); err != nil {
				// The flavor cannot be used without a configuration
				klog.Errorf("Error validating the flavor without a configuration: %s", err)
				http.Error(w, "Error validating the flavor without a configuration. A configuration is required", http.StatusBadRequest)
				return
			}
		} else {
			// Configuration type is the ServiceConfiguration
			if request.Configuration.Type != models.ServiceNameDefault {
				klog.Errorf("Configuration type %s not supported", request.Configuration.Type)
				http.Error(w, "Configuration type not supported", http.StatusBadRequest)
				return
			}
			// Forge the configuration object
			configuration, err := resourceforge.ForgeConfigurationFromObj(*request.Configuration)
			if err != nil {
				klog.Errorf("Error forging the configuration object: %s", err)
				http.Error(w, "Error forging the configuration object", http.StatusInternalServerError)
				return
			}
			// Parse the configuration and validate it over the ServiceFlavor
			configurationType, configurationData, err := nodecorev1alpha1.ParseConfiguration(configuration, flavor)
			if err != nil {
				klog.Errorf("Error parsing the configuration: %s", err)
				http.Error(w, "Error parsing the configuration", http.StatusInternalServerError)
				return
			}
			if configurationType != nodecorev1alpha1.TypeService {
				klog.Errorf("Configuration type %s not supported", configurationType)
				http.Error(w, "Configuration type not supported", http.StatusBadRequest)
				return
			}
			// Force cast the configuration data to the ServiceConfiguration type
			serviceConfiguration, ok := configurationData.(nodecorev1alpha1.ServiceConfiguration)
			if !ok {
				klog.Errorf("Error casting the configuration data to ServiceConfiguration")
				http.Error(w, "Error casting the configuration data to ServiceConfiguration", http.StatusInternalServerError)
				return
			}

			// Check the hosting policy is supported by the flavor
			if serviceConfiguration.HostingPolicy != nil {
				// Force cast the flavor data to the ServiceFlavor type
				serviceFlavor, ok := flavorData.(*nodecorev1alpha1.ServiceFlavor)
				if !ok {
					klog.Errorf("Error casting the flavor data to ServiceFlavor")
					http.Error(w, "Error casting the flavor data to ServiceFlavor", http.StatusInternalServerError)
					return
				}

				// Check if the hosting policy is included in the supported hosting policies by the service flavor
				supported := false
				for _, hp := range serviceFlavor.HostingPolicies {
					if hp == *serviceConfiguration.HostingPolicy {
						supported = true
						break
					}
				}

				if !supported {
					klog.Errorf("Hosting policy %s not supported by the flavor", serviceConfiguration.HostingPolicy)
					http.Error(w, "Hosting policy not supported by the flavor", http.StatusBadRequest)
					return
				}
			}

			klog.Info("Service flavor configuration validated")
		}
	case nodecorev1alpha1.TypeSensor:
		// TODO (Sensor): Implement the Sensor flavor configuration
		klog.Errorf("Flavor type %s not supported", flavorTypeIdentifier)
		http.Error(w, "Flavor type not supported", http.StatusBadRequest)
		return
	default:
		klog.Errorf("Flavor type %s not supported", flavorTypeIdentifier)
		http.Error(w, "Flavor type not supported", http.StatusBadRequest)
		return
	}

	// Check if the Transaction already exists
	t, found := g.SearchTransaction(request.Buyer.NodeID, flavorID)
	if found {
		t.ExpirationTime = tools.GetExpirationTime(1, 0, 0)
		transaction = t
		g.addNewTransaction(t)
	}

	if !found {
		klog.Infof("Reserving flavor %s started", flavorID)

		// Create a new transaction ID
		transactionID, err := namings.ForgeTransactionID()
		if err != nil {
			http.Error(w, "Error generating transaction ID", http.StatusInternalServerError)
			return
		}

		// Check the consumer communicated the LiqoID in the optional AdditionalInformation field
		if request.Buyer.AdditionalInformation == nil || request.Buyer.AdditionalInformation.LiqoID == "" {
			http.Error(w, "Error: LiqoID not provided", http.StatusBadRequest)
			return
		}

		// Create a new transaction
		transaction = resourceforge.ForgeTransactionObj(transactionID, &request)

		// Add the transaction to the transactions map
		g.addNewTransaction(transaction)
	}

	klog.Infof("Transaction %s reserved", transaction.TransactionID)

	encodeResponse(w, transaction)
}

// purchaseFlavor is an handler for purchasing a Flavor.
func (g *Gateway) purchaseFlavor(w http.ResponseWriter, r *http.Request) {
	// Get the flavorID value from the URL parameters
	params := mux.Vars(r)
	transactionID := params["transactionID"]
	var purchase models.PurchaseRequest

	if err := json.NewDecoder(r.Body).Decode(&purchase); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	klog.Infof("Purchasing request for transaction %s", transactionID)

	// Retrieve the transaction from the transactions map
	transaction, err := g.GetTransaction(transactionID)
	if err != nil {
		klog.Errorf("Error getting the Transaction: %s", err)
		http.Error(w, "Error getting the Transaction", http.StatusInternalServerError)
		return
	}

	klog.Infof("Flavor requested: %s", transaction.FlavorID)

	if tools.CheckExpiration(transaction.ExpirationTime) {
		klog.Infof("Transaction %s expired", transaction.TransactionID)
		http.Error(w, "Error: transaction Timeout", http.StatusRequestTimeout)
		g.removeTransaction(transaction.TransactionID)
		return
	}

	var contractList reservationv1alpha1.ContractList
	var contract reservationv1alpha1.Contract

	// Check if the Contract with the same TransactionID already exists
	if err := g.client.List(context.Background(), &contractList, client.MatchingFields{"spec.transactionID": transactionID}); err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Errorf("Error when listing Contracts: %s", err)
			http.Error(w, "Error when listing Contracts", http.StatusInternalServerError)
			return
		}
	}

	if len(contractList.Items) > 0 {
		klog.Infof("Contract already exists for transaction %s", transactionID)
		contract = contractList.Items[0]
		// Create a contract object to be returned with the response
		contractObject := parseutil.ParseContract(&contract)
		// Respond with the response purchase as JSON
		encodeResponse(w, contractObject)
		return
	}

	klog.Infof("Performing purchase of flavor %s...", transaction.FlavorID)

	// Remove the transaction from the transactions map
	g.removeTransaction(transaction.TransactionID)

	klog.Infof("Flavor %s successfully purchased!", transaction.FlavorID)

	// Get the flavor sold for creating the contract
	flavorSold, err := services.GetFlavorByID(transaction.FlavorID, g.client)
	if err != nil {
		klog.Errorf("Error getting the Flavor by ID: %s", err)
		http.Error(w, "Error getting the Flavor by ID", http.StatusInternalServerError)
		return
	}

	var liqoCredentials *nodecorev1alpha1.LiqoCredentials

	// According to the flavor type, create the contract with the right liqo credentials
	switch flavorSold.Spec.FlavorType.TypeIdentifier {
	case nodecorev1alpha1.TypeK8Slice:
		// Create a new Liqo credentials for the K8Slice flavor based on the ones provided by the provider
		liqoCredentials, err = getters.GetLiqoCredentials(context.Background(), g.client)
		if err != nil {
			klog.Errorf("Error getting Liqo Credentials: %s", err)
			http.Error(w, "Error getting Liqo Credentials", http.StatusInternalServerError)
			return
		}
	case nodecorev1alpha1.TypeVM:
		// TODO (VM): Implement the VM flavor contract
		klog.Errorf("Flavor type %s not supported", flavorSold.Spec.FlavorType.TypeIdentifier)
		http.Error(w, "Flavor type not supported", http.StatusBadRequest)
		return
	case nodecorev1alpha1.TypeService:
		// Check client sent its liqo credentials
		if purchase.LiqoCredentials == nil {
			klog.Errorf("Error: Liqo credentials not provided")
			http.Error(w, "Error: Liqo credentials not provided", http.StatusBadRequest)
			return
		}
		// Override the Liqo credentials with the ones sent by the client
		liqoCredentials, err = resourceforge.ForgeLiqoCredentialsFromObj(purchase.LiqoCredentials)
		if err != nil {
			klog.Errorf("Error forging the Liqo credentials: %s", err)
			http.Error(w, "Error forging the Liqo credentials", http.StatusInternalServerError)
			return
		}
	case nodecorev1alpha1.TypeSensor:
		// TODO (Sensor): Implement the Sensor flavor contract
		klog.Errorf("Flavor type %s not supported", flavorSold.Spec.FlavorType.TypeIdentifier)
		http.Error(w, "Flavor type not supported", http.StatusBadRequest)
		return
	default:
		klog.Errorf("Flavor type %s not supported", flavorSold.Spec.FlavorType.TypeIdentifier)
		http.Error(w, "Flavor type not supported", http.StatusBadRequest)
		return
	}

	// Obtaining the Seller Liqo Cluster ID
	sellerLiqoCredentials, err := getters.GetLiqoCredentials(context.Background(), g.client)
	if err != nil {
		klog.Errorf("Error getting Liqo Credentials: %s", err)
		http.Error(w, "Error getting Liqo Credentials", http.StatusInternalServerError)
		return
	}

	// Create a new contract
	klog.Infof("Creating a new contract...")
	// Forge the contract object
	contract = *resourceforge.ForgeContract(
		flavorSold,
		&transaction,
		liqoCredentials,
		sellerLiqoCredentials.ClusterID,
		purchase.IngressTelemetryEndpoint,
	)
	err = g.client.Create(context.Background(), &contract)
	if err != nil {
		klog.Errorf("Error creating the Contract: %s", err)
		http.Error(w, "Error creating the Contract: "+err.Error(), http.StatusInternalServerError)
		return
	}

	klog.Infof("Contract created!")

	// Create a contract object to be returned with the response
	contractObject := parseutil.ParseContract(&contract)

	if contractObject.Configuration != nil {
		klog.Infof("Contract %v", *contractObject.Configuration)
	} else {
		klog.Infof("No configuration found in the contract")
	}

	// Create allocation
	klog.Infof("Creating allocation...")
	allocation := *resourceforge.ForgeAllocation(&contract)
	err = g.client.Create(context.Background(), &allocation)
	if err != nil {
		klog.Errorf("Error creating the Allocation: %s", err)
		http.Error(w, "Contract created but we ran into an error while allocating the resources", http.StatusInternalServerError)
		return
	}

	klog.Infof("Contract %s successfully created and now sending to the client!", contract.Name)

	// Respond with the response purchase as JSON
	encodeResponse(w, contractObject)
}
