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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	resourceLib "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"

	"github.com/fluidos-project/node/pkg/utils/models"
)

// selectorToQueryParams converts a selector to a query string.
func selectorToQueryParams(selector models.Selector) (string, error) {
	switch selector.GetSelectorType() {
	case models.K8SliceNameDefault:
		k8sliceSelector := selector.(models.K8SliceSelector)
		return encodeK8SliceSelector(k8sliceSelector)
	case models.VMNameDefault:
		// TODO (VM): Implement the VM selector type
		return "", fmt.Errorf("unsupported selector type %s", selector.GetSelectorType())
	case models.ServiceNameDefault:
		serviceSelector := selector.(models.ServiceSelector)
		return encodeServiceSelector(serviceSelector)
	case models.SensorNameDefault:
		// TODO (Sensor): Implement the Sensor selector type
		return "", fmt.Errorf("unsupported selector type %s", selector.GetSelectorType())
	default:
		return "", fmt.Errorf("unsupported selector type")
	}
}

// queryParamToSelector converts a query string to a selector.
func queryParamToSelector(queryValues url.Values, selectorType models.FlavorTypeName) (models.Selector, error) {
	klog.Infof("queryValues: %v", queryValues)

	// Print each query parameter
	for key, value := range queryValues {
		klog.Infof("key: %s, value: %s", key, value)
	}

	switch selectorType {
	case models.K8SliceNameDefault:
		k8sliceSelector, err := decodeK8SliceSelector(queryValues)
		if err != nil {
			return nil, err
		}
		return *k8sliceSelector, nil
	case models.VMNameDefault:
		// TODO (VM): Implement the VM query param to selector conversion
		return nil, fmt.Errorf("unsupported selector type %s", selectorType)
	case models.ServiceNameDefault:
		serviceSelector, err := decodeServiceSelector(queryValues)
		if err != nil {
			return nil, err
		}
		return *serviceSelector, nil
	case models.SensorNameDefault:
		// TODO (Sensor): Implement the Sensor query param to selector conversion
		return nil, fmt.Errorf("unsupported selector type %s", selectorType)
	default:
		return nil, fmt.Errorf("unsupported selector type")
	}
}

// GetTransaction returns a transaction from the transactions map.
func (g *Gateway) GetTransaction(transactionID string) (models.Transaction, error) {
	transaction, exists := g.Transactions[transactionID]
	if !exists {
		return models.Transaction{}, fmt.Errorf("transaction not found")
	}
	return *transaction, nil
}

// SearchTransaction returns a transaction from the transactions map.
func (g *Gateway) SearchTransaction(buyerID, flavorID string) (*models.Transaction, bool) {
	for _, t := range g.Transactions {
		if t.Buyer.NodeID == buyerID && t.FlavorID == flavorID {
			return t, true
		}
	}
	return &models.Transaction{}, false
}

// addNewTransaction add a new transaction to the transactions map.
func (g *Gateway) addNewTransaction(transaction *models.Transaction) {
	g.Transactions[transaction.TransactionID] = transaction
}

// removeTransaction removes a transaction from the transactions map.
func (g *Gateway) removeTransaction(transactionID string) {
	delete(g.Transactions, transactionID)
}

// handleError handles errors by sending an error response.
func handleError(w http.ResponseWriter, err error, statusCode int) {
	http.Error(w, err.Error(), statusCode)
}

// encodeResponse encodes the response as JSON and writes it to the response writer.
func encodeResponse(w http.ResponseWriter, data interface{}) {
	encodeResponseStatusCode(w, data, http.StatusOK)
}

func encodeResponseStatusCode(w http.ResponseWriter, data interface{}, statusCode int) {
	resp, err := json.Marshal(data)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(resp)
}

func encodeK8SliceSelector(selector models.K8SliceSelector) (string, error) {
	var values string

	if selector.Architecture != nil {
		klog.Info("Encoding Architecture")
		architectureEncoded := encodeStringFilter(reflect.ValueOf(selector.Architecture))
		for key, value := range architectureEncoded {
			values += fmt.Sprintf("filter[architecture]%s=%s&", key, value)
		}
	}
	if selector.CPU != nil {
		klog.Info("Encoding CPU")
		cpuEncoded := encoderResourceQuantityFilter(reflect.ValueOf(selector.CPU))
		for key, value := range cpuEncoded {
			values += fmt.Sprintf("filter[cpu]%s=%s&", key, value)
		}
	}
	if selector.Memory != nil {
		klog.Info("Encoding Memory")
		memoryEncoded := encoderResourceQuantityFilter(reflect.ValueOf(selector.Memory))
		for key, value := range memoryEncoded {
			values += fmt.Sprintf("filter[memory]%s=%s&", key, value)
		}
	}
	if selector.Pods != nil {
		klog.Info("Encoding Pods")
		podsEncoded := encoderResourceQuantityFilter(reflect.ValueOf(selector.Pods))
		for key, value := range podsEncoded {
			values += fmt.Sprintf("filter[pods]%s=%s&", key, value)
		}
	}
	if selector.Storage != nil {
		klog.Info("Encoding Storage")
		storageEncoded := encoderResourceQuantityFilter(reflect.ValueOf(selector.Storage))
		for key, value := range storageEncoded {
			values += fmt.Sprintf("[storage]%s=%s&", key, value)
		}
	}

	// Remove trailing "&" if present
	if values != "" && values[len(values)-1] == '&' {
		values = values[:len(values)-1]
	}

	return values, nil
}

func encodeServiceSelector(selector models.ServiceSelector) (string, error) {
	var values string

	if selector.Category != nil {
		klog.Info("Encoding Category")
		categoryEncoded := encodeStringFilter(reflect.ValueOf(selector.Category))
		for key, value := range categoryEncoded {
			values += fmt.Sprintf("filter[category]%s=%s&", key, value)
		}
	}

	if selector.Tags != nil {
		klog.Info("Encoding Tags")
		tagsEncoded := encodeStringFilter(reflect.ValueOf(selector.Tags))
		for key, value := range tagsEncoded {
			values += fmt.Sprintf("filter[tags]%s=%s&", key, value)
		}
	}

	// TODO(Service): Add more filters encoding here

	// Remove trailing "&" if present
	if values != "" && values[len(values)-1] == '&' {
		values = values[:len(values)-1]
	}

	return values, nil
}

func encoderResourceQuantityFilter(v reflect.Value) map[string]string {
	filter, ok := v.Interface().(*models.ResourceQuantityFilter)
	if filter == nil {
		return nil
	}
	if !ok {
		klog.Warning("Not a ResourceQuantityFilter")
		return nil
	}

	var values = map[string]string{}

	switch filter.Name {
	case models.MatchFilter:
		var matchFilter models.ResourceQuantityMatchFilter
		klog.Info("MatchFilter")
		if err := json.Unmarshal(filter.Data, &matchFilter); err != nil {
			return nil
		}
		value := matchFilter.Value
		values["[match]"] = value.String()
	case models.RangeFilter:
		var rangeFilter models.ResourceQuantityRangeFilter
		klog.Info("RangeFilter")
		if err := json.Unmarshal(filter.Data, &rangeFilter); err != nil {
			klog.Warningf("Error unmarshaling RangeFilter: %v", err)
			return nil
		}
		klog.Info("RangeFilter: ", rangeFilter)
		if rangeFilter.Min != nil {
			minValue := *rangeFilter.Min
			values["[range][min]"] = minValue.String()
		}
		if rangeFilter.Max != nil {
			maxValue := *rangeFilter.Max
			values["[range][max]"] = maxValue.String()
		}
	default:
		klog.Warningf("Unsupported filter type: %s", filter.Name)
		return nil
	}
	return values
}

func encodeStringFilter(v reflect.Value) map[string]string {
	filter, ok := v.Interface().(*models.StringFilter)
	if filter == nil {
		return nil
	}
	if !ok {
		klog.Warning("Not a StringFilter")
		return nil
	}

	var values = map[string]string{}

	switch filter.Name {
	case models.MatchFilter:
		var matchFilter models.StringMatchFilter
		klog.Info("MatchFilter")
		if err := json.Unmarshal(filter.Data, &matchFilter); err != nil {
			return nil
		}
		value := matchFilter.Value
		values["[match]"] = value
	default:
		klog.Warningf("Unsupported filter type: %s", filter.Name)
		return nil
	}
	return values
}

func decodeStringFilter(filterTypeName models.FilterType,
	original *models.StringFilter, value []string) (*models.StringFilter, error) {
	var result = models.StringFilter{}

	switch filterTypeName {
	case models.MatchFilter:
		// Check you can't have multiple match filters for the same resource in the same selector
		if original != nil {
			return nil, fmt.Errorf("multiple match filters for the same resource")
		}
		// Set the match filter
		matchFilter := models.StringMatchFilter{
			Value: value[0],
		}
		// Encode the match filter to a json.RawMessage
		matchFilterBytes, err := json.Marshal(matchFilter)
		if err != nil {
			return nil, err
		}

		result = models.StringFilter{
			Name: models.MatchFilter,
			Data: matchFilterBytes,
		}
	default:
		return nil, fmt.Errorf("invalid filter type: %s", filterTypeName)
	}

	klog.Infof("Decoded filter: %v", result)
	klog.Infof("Decoded filter type: %v", result.Name)

	return &result, nil
}

func decodeResourceQuantityFilter(filterTypeName models.FilterType,
	original *models.ResourceQuantityFilter, parts, value []string) (*models.ResourceQuantityFilter, error) {
	var result = models.ResourceQuantityFilter{}

	switch filterTypeName {
	case models.MatchFilter:
		// Check you can't have multiple match filters for the same resource in the same selector
		if original != nil {
			return nil, fmt.Errorf("multiple match filters for the same resource")
		}
		// Set the match filter
		var matchFilter models.ResourceQuantityMatchFilter
		match, err := resourceLib.ParseQuantity(value[0])
		if err != nil {
			return nil, err
		}
		matchFilter.Value = match
		// Encode the match filter to a json.RawMessage
		matchFilterBytes, err := json.Marshal(matchFilter)
		if err != nil {
			return nil, err
		}

		result = models.ResourceQuantityFilter{
			Name: models.MatchFilter,
			Data: matchFilterBytes,
		}
	case models.RangeFilter:
		// You can have multiple range filters for the same resource in the same selector, but they must be different
		// Check if the filter is already set
		if original != nil {
			// FILTER ALREADY SET

			result = *original

			// Check if the filter is a range filter
			if result.Name != models.RangeFilter {
				return nil, fmt.Errorf("cannot have range filter for a resource with another filter type")
			}
			// The filter is a range filter
			// Get the range filter
			var rangeFilter models.ResourceQuantityRangeFilter
			if err := json.Unmarshal(result.Data, &rangeFilter); err != nil {
				return nil, err
			}
			// Check the value of the filter is not already set
			switch parts[3][:len(parts[3])-1] {
			case models.RangeMinKey:
				if rangeFilter.Min != nil {
					return nil, fmt.Errorf("min value already set")
				}
				// Set the minValue value
				minValue, err := resourceLib.ParseQuantity(value[0])
				if err != nil {
					return nil, err
				}
				rangeFilter.Min = &minValue
			case models.RangeMaxKey:
				if rangeFilter.Max != nil {
					return nil, fmt.Errorf("max value already set")
				}
				// Set the maxValue value
				maxValue, err := resourceLib.ParseQuantity(value[0])
				if err != nil {
					return nil, err
				}
				rangeFilter.Max = &maxValue
			default:
				return nil, fmt.Errorf("invalid key format: %s", parts[3])
			}
			// Encode the range filter to a json.RawMessage
			rangeFilterBytes, err := json.Marshal(rangeFilter)
			if err != nil {
				return nil, err
			}

			result.Data = rangeFilterBytes
		} else {
			// FILTER NOT SET

			// Set the range filter
			var rangeFilter models.ResourceQuantityRangeFilter
			switch parts[3][:len(parts[3])-1] {
			case models.RangeMinKey:
				minValue, err := resourceLib.ParseQuantity(value[0])
				if err != nil {
					return nil, err
				}
				rangeFilter.Min = &minValue
			case models.RangeMaxKey:
				maxValue, err := resourceLib.ParseQuantity(value[0])
				if err != nil {
					return nil, err
				}
				rangeFilter.Max = &maxValue
			default:
				return nil, fmt.Errorf("invalid key format: %s", parts[3])
			}
			// Encode the range filter to a json.RawMessage
			rangeFilterBytes, err := json.Marshal(rangeFilter)
			if err != nil {
				return nil, err
			}

			result = models.ResourceQuantityFilter{
				Name: models.RangeFilter,
				Data: rangeFilterBytes,
			}
		}
	default:
		return nil, fmt.Errorf("invalid filter type: %s", filterTypeName)
	}

	klog.Infof("Decoded filter: %v", result)
	klog.Infof("Decoded filter type: %v", result.Name)

	return &result, nil
}

func decodeK8SliceSelector(values url.Values) (*models.K8SliceSelector, error) {
	selector := models.K8SliceSelector{}

	// Define the regex pattern for the expected keys
	keyPattern := regexp.MustCompile(`^filter\[(architecture|cpu|memory|pods|storage)\]\[(match|range)\](?:\[(min|max|regex)\])?$`)

	for key, value := range values {
		// Check if the key matches the expected pattern
		if !keyPattern.MatchString(key) {
			return nil, fmt.Errorf("invalid key format: %s", key)
		}

		klog.Infof("Decoding key: %s", key)

		parts := strings.Split(key, "[")
		resource := parts[1][:len(parts[1])-1]
		filterType := parts[2][:len(parts[2])-1]

		// Convert filterType to FilterType
		var filterTypeName models.FilterType
		switch filterType {
		case "match":
			filterTypeName = models.MatchFilter
		case "range":
			filterTypeName = models.RangeFilter
		default:
			return nil, fmt.Errorf("invalid filter type: %s", filterType)
		}

		switch resource {
		case "architecture":
			filter, err := decodeStringFilter(filterTypeName, selector.Architecture, value)
			if err != nil {
				return nil, err
			}
			selector.Architecture = filter
			klog.Infof("Selector Architecture: %v", selector.Architecture)
		case "cpu":
			filter, err := decodeResourceQuantityFilter(filterTypeName, selector.CPU, parts, value)
			if err != nil {
				return nil, err
			}
			selector.CPU = filter
			klog.Infof("Selector CPU: %v", selector.CPU)
		case "memory":
			filter, err := decodeResourceQuantityFilter(filterTypeName, selector.Memory, parts, value)
			if err != nil {
				return nil, err
			}
			selector.Memory = filter
			klog.Infof("Selector Memory: %v", selector.Memory)
		case "pods":
			filter, err := decodeResourceQuantityFilter(filterTypeName, selector.Pods, parts, value)
			if err != nil {
				return nil, err
			}
			selector.Pods = filter
			klog.Infof("Selector Pods: %v", selector.Pods)
		case "storage":
			filter, err := decodeResourceQuantityFilter(filterTypeName, selector.Storage, parts, value)
			if err != nil {
				return nil, err
			}
			selector.Storage = filter
			klog.Infof("Selector Storage: %v", selector.Storage)
		default:
			return nil, fmt.Errorf("invalid resource: %s", resource)
		}
	}

	return &selector, nil
}

func decodeServiceSelector(values url.Values) (*models.ServiceSelector, error) {
	selector := models.ServiceSelector{}

	// Define the regex pattern for the expected keys
	keyPattern := regexp.MustCompile(`^filter\[(category|tags)\]\[(match)\]$`)

	for key, value := range values {
		// Check if the key matches the expected pattern
		if !keyPattern.MatchString(key) {
			return nil, fmt.Errorf("invalid key format: %s", key)
		}

		klog.Infof("Decoding key: %s", key)

		parts := strings.Split(key, "[")
		resource := parts[1][:len(parts[1])-1]
		filterType := parts[2][:len(parts[2])-1]

		// Convert filterType to FilterType
		var filterTypeName models.FilterType
		switch filterType {
		case "match":
			filterTypeName = models.MatchFilter
		default:
			return nil, fmt.Errorf("invalid filter type: %s", filterType)
		}

		switch resource {
		case "category":
			filter, err := decodeStringFilter(filterTypeName, selector.Category, value)
			if err != nil {
				return nil, err
			}
			selector.Category = filter
			klog.Infof("Selector Category: %v", selector.Category)
		case "tags":
			filter, err := decodeStringFilter(filterTypeName, selector.Tags, value)
			if err != nil {
				return nil, err
			}
			selector.Tags = filter
			klog.Infof("Selector Tags: %v", selector.Tags)
		default:
			return nil, fmt.Errorf("invalid resource: %s", resource)
		}
	}

	return &selector, nil
}
