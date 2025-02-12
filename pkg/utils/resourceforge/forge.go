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

package resourceforge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/Masterminds/sprig"
	offloadingv1alpha1 "github.com/liqotech/liqo/apis/offloading/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	advertisementv1alpha1 "github.com/fluidos-project/node/apis/advertisement/v1alpha1"
	networkv1alpha1 "github.com/fluidos-project/node/apis/network/v1alpha1"
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/consts"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/namings"
	"github.com/fluidos-project/node/pkg/utils/parseutil"
	"github.com/fluidos-project/node/pkg/utils/tools"
)

// ForgeDiscovery creates a Discovery CR from a FlavorSelector and a solverID.
func ForgeDiscovery(selector *nodecorev1alpha1.Selector, solverID string) *advertisementv1alpha1.Discovery {
	return &advertisementv1alpha1.Discovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeDiscoveryName(solverID),
			Namespace: flags.FluidosNamespace,
		},
		Spec: advertisementv1alpha1.DiscoverySpec{
			Selector: func() *nodecorev1alpha1.Selector {
				if selector != nil {
					return selector
				}
				return nil
			}(),
			SolverID:  solverID,
			Subscribe: false,
		},
	}
}

// ForgePeeringCandidate creates a PeeringCandidate CR from a Flavor and a Discovery.
func ForgePeeringCandidate(flavorPeeringCandidate *nodecorev1alpha1.Flavor,
	solverID string, available bool) (pc *advertisementv1alpha1.PeeringCandidate) {
	pc = &advertisementv1alpha1.PeeringCandidate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgePeeringCandidateName(flavorPeeringCandidate.Name),
			Namespace: flags.FluidosNamespace,
		},
		Spec: advertisementv1alpha1.PeeringCandidateSpec{
			Flavor: nodecorev1alpha1.Flavor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      flavorPeeringCandidate.Name,
					Namespace: flavorPeeringCandidate.Namespace,
				},
				Spec: flavorPeeringCandidate.Spec,
			},
			Available: available,
		},
	}
	pc.Spec.InterestedSolverIDs = append(pc.Spec.InterestedSolverIDs, solverID)
	return
}

// ForgeReservation creates a Reservation CR from a PeeringCandidate.
func ForgeReservation(pc *advertisementv1alpha1.PeeringCandidate,
	configuration *nodecorev1alpha1.Configuration,
	ni nodecorev1alpha1.NodeIdentity,
	reservingSolver string) *reservationv1alpha1.Reservation {
	solverID := reservingSolver
	reservation := &reservationv1alpha1.Reservation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeReservationName(solverID),
			Namespace: flags.FluidosNamespace,
		},
		Spec: reservationv1alpha1.ReservationSpec{
			SolverID: solverID,
			Buyer:    ni,
			Seller: nodecorev1alpha1.NodeIdentity{
				Domain: pc.Spec.Flavor.Spec.Owner.Domain,
				NodeID: pc.Spec.Flavor.Spec.Owner.NodeID,
				IP:     pc.Spec.Flavor.Spec.Owner.IP,
			},
			PeeringCandidate: nodecorev1alpha1.GenericRef{
				Name:      pc.Name,
				Namespace: pc.Namespace,
			},
			Reserve:  true,
			Purchase: true,
			Configuration: func() *nodecorev1alpha1.Configuration {
				if configuration != nil {
					return configuration
				}
				return nil
			}(),
		},
	}
	if configuration != nil {
		reservation.Spec.Configuration = configuration
	}
	return reservation
}

// ForgeTelemetryServer creates a TelemetryServer CR from a TelemetryServer model.
func ForgeTelemetryServer(telemetryServer *models.TelemetryServer) *reservationv1alpha1.TelemetryServer {
	if telemetryServer == nil {
		return nil
	}
	return &reservationv1alpha1.TelemetryServer{
		Endpoint: telemetryServer.Endpoint,
		Intents:  telemetryServer.Intents,
	}
}

// ForgeContract creates a Contract CR.
func ForgeContract(
	flavor *nodecorev1alpha1.Flavor,
	transaction *models.Transaction,
	peeringTargetLiqoCredentials *nodecorev1alpha1.LiqoCredentials,
	sellerLiqoID string,
	ingressTelemetryEndpoint *models.TelemetryServer) *reservationv1alpha1.Contract {
	return &reservationv1alpha1.Contract{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeContractName(flavor.Name),
			Namespace: flags.FluidosNamespace,
		},
		Spec: reservationv1alpha1.ContractSpec{
			Flavor: *flavor,
			Buyer: nodecorev1alpha1.NodeIdentity{
				Domain: transaction.Buyer.Domain,
				IP:     transaction.Buyer.IP,
				NodeID: transaction.Buyer.NodeID,
				AdditionalInformation: &nodecorev1alpha1.NodeIdentityAdditionalInfo{
					LiqoID: transaction.ClusterID,
				},
			},
			BuyerClusterID: transaction.ClusterID,
			Seller: func() nodecorev1alpha1.NodeIdentity {
				return nodecorev1alpha1.NodeIdentity{
					Domain: flavor.Spec.Owner.Domain,
					NodeID: flavor.Spec.Owner.NodeID,
					IP:     flavor.Spec.Owner.IP,
					AdditionalInformation: &nodecorev1alpha1.NodeIdentityAdditionalInfo{
						LiqoID: sellerLiqoID,
					},
				}
			}(),
			PeeringTargetCredentials: *peeringTargetLiqoCredentials,
			TransactionID:            transaction.TransactionID,
			Configuration: func() *nodecorev1alpha1.Configuration {
				if transaction.Configuration != nil {
					configuration, err := ForgeConfigurationFromObj(*transaction.Configuration)
					if err != nil {
						klog.Errorf("Error when parsing configuration: %s", err)
						return nil
					}
					return configuration
				}
				return nil
			}(),
			ExpirationTime:   time.Now().Add(flags.ExpirationContract).Format(time.RFC3339),
			ExtraInformation: nil,
			// TODO: Add logic to network requests
			NetworkRequests:          "",
			IngressTelemetryEndpoint: ForgeTelemetryServer(ingressTelemetryEndpoint),
		},
		Status: reservationv1alpha1.ContractStatus{
			Phase: nodecorev1alpha1.PhaseStatus{
				Phase:     nodecorev1alpha1.PhaseActive,
				StartTime: tools.GetTimeNow(),
			},
		},
	}
}

// ForgeK8SliceFlavorFromMetrics creates a new flavor custom resource from the metrics of the node.
func ForgeK8SliceFlavorFromMetrics(node *models.NodeInfo, ni nodecorev1alpha1.NodeIdentity,
	ownerReferences []metav1.OwnerReference) (flavor *nodecorev1alpha1.Flavor) {
	k8SliceType := nodecorev1alpha1.K8Slice{
		Characteristics: nodecorev1alpha1.K8SliceCharacteristics{
			Architecture: node.Architecture,
			CPU:          node.ResourceMetrics.CPUAvailable,
			Memory:       node.ResourceMetrics.MemoryAvailable,
			Pods:         node.ResourceMetrics.PodsAvailable,
			Storage:      &node.ResourceMetrics.EphemeralStorage,
			Gpu: &nodecorev1alpha1.GPU{
				Model:  node.ResourceMetrics.GPU.Model,
				Cores:  node.ResourceMetrics.GPU.CoresAvailable,
				Memory: node.ResourceMetrics.GPU.MemoryAvailable,
			},
		},
		Properties: nodecorev1alpha1.Properties{},
		Policies: nodecorev1alpha1.Policies{
			Partitionability: nodecorev1alpha1.Partitionability{
				CPUMin:     parseutil.ParseQuantityFromString(flags.CPUMin),
				MemoryMin:  parseutil.ParseQuantityFromString(flags.MemoryMin),
				PodsMin:    parseutil.ParseQuantityFromString(flags.PodsMin),
				CPUStep:    parseutil.ParseQuantityFromString(flags.CPUStep),
				MemoryStep: parseutil.ParseQuantityFromString(flags.MemoryStep),
				PodsStep:   parseutil.ParseQuantityFromString(flags.PodsStep),
			},
		},
	}

	// Serialize K8SliceType to JSON
	k8SliceTypeJSON, err := json.Marshal(k8SliceType)
	if err != nil {
		klog.Errorf("Error when marshaling K8SliceType: %s", err)
		return nil
	}

	return &nodecorev1alpha1.Flavor{
		ObjectMeta: metav1.ObjectMeta{
			Name:            namings.ForgeFlavorName(string(nodecorev1alpha1.TypeK8Slice), ni.Domain),
			Namespace:       flags.FluidosNamespace,
			OwnerReferences: ownerReferences,
		},
		Spec: nodecorev1alpha1.FlavorSpec{
			ProviderID: ni.NodeID,
			FlavorType: nodecorev1alpha1.FlavorType{
				TypeIdentifier: nodecorev1alpha1.TypeK8Slice,
				TypeData:       runtime.RawExtension{Raw: k8SliceTypeJSON},
			},
			Owner: ni,
			Price: nodecorev1alpha1.Price{
				Amount:   flags.AMOUNT,
				Currency: flags.CURRENCY,
				Period:   flags.PERIOD,
			},
			Availability: true,
			// FIXME: NetworkPropertyType should be taken in a smarter way
			NetworkPropertyType: "networkProperty",
			// FIXME: Location should be taken in a smarter way
			Location: &nodecorev1alpha1.Location{
				Latitude:        "10",
				Longitude:       "58",
				Country:         "Italy",
				City:            "Turin",
				AdditionalNotes: "None",
			},
		},
	}
}

// ForgeServiceFlavorFromBlueprint creates a new flavor custom resource from a ServiceBlueprint.
func ForgeServiceFlavorFromBlueprint(serviceBlueprint *nodecorev1alpha1.ServiceBlueprint, ni *nodecorev1alpha1.NodeIdentity,
	ownerReferences []metav1.OwnerReference) (flavor *nodecorev1alpha1.Flavor) {
	configurationTemplate, err := forgeServiceConfigurationTemplateFromCategory(models.MapToServiceCategory(serviceBlueprint.Spec.Category))
	if err != nil {
		klog.Errorf("Error when forging configuration template: %s", err)
		return nil
	}

	serviceFlavor := &nodecorev1alpha1.ServiceFlavor{
		Name:                  serviceBlueprint.Spec.Name,
		Description:           serviceBlueprint.Spec.Description,
		Category:              serviceBlueprint.Spec.Category,
		Tags:                  serviceBlueprint.Spec.Tags,
		HostingPolicies:       serviceBlueprint.Spec.HostingPolicies,
		ConfigurationTemplate: runtime.RawExtension{Raw: []byte(configurationTemplate)},
	}

	serviceFlavorJSON, err := json.Marshal(serviceFlavor)
	if err != nil {
		klog.Errorf("Error when marshaling service flavor: %s", err)
		return nil
	}

	return &nodecorev1alpha1.Flavor{
		ObjectMeta: metav1.ObjectMeta{
			Name:            namings.ForgeFlavorName(string(nodecorev1alpha1.TypeService), ni.Domain),
			Namespace:       flags.FluidosNamespace,
			OwnerReferences: ownerReferences,
		},
		Spec: nodecorev1alpha1.FlavorSpec{
			ProviderID: ni.NodeID,
			FlavorType: nodecorev1alpha1.FlavorType{
				TypeIdentifier: nodecorev1alpha1.TypeService,
				TypeData:       runtime.RawExtension{Raw: serviceFlavorJSON},
			},
			Owner: *ni,
			Price: nodecorev1alpha1.Price{
				Amount:   flags.AMOUNT,
				Currency: flags.CURRENCY,
				Period:   flags.PERIOD,
			},
			Availability: true,
			// FIXME: NetworkPropertyType should be taken in a smarter way
			NetworkPropertyType: "networkProperty",
			// FIXME: Location should be taken in a smarter way
			Location: &nodecorev1alpha1.Location{
				Latitude:        "10",
				Longitude:       "58",
				Country:         "Italy",
				City:            "Turin",
				AdditionalNotes: "None",
			},
		},
	}
}

// forgeServiceConfigurationTemplateFromCategory creates a JSON schema for the configuration template of a service.
func forgeServiceConfigurationTemplateFromCategory(category consts.ServiceCategory) (configurationTemplate string, err error) {
	var JSONtemplate string

	switch category {
	case consts.Database:
		JSONtemplate = `{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type": "object",
			"properties": {
				"username": {
					"type": "string"
				},
				"password": {
					"type": "string"
				},
				"database": {
					"type": "string"
				}
			},
			"required": ["username", "password", "database"]
		}`
	case consts.MessageQueue:
		JSONtemplate = `{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type": "object",
			"properties": {
				"username": {
					"type": "string"
				},
				"password": {
					"type": "string"
				}
			},
			"required": ["username", "password"]
		}`
	// TODO (Service): Implement more categories based on ontology
	default:
		klog.Errorf("Category not recognized")
		return "", fmt.Errorf("category not recognized")
	}

	return JSONtemplate, nil
}

// forgeServiceConfigurationDefaultFromCategory creates a default configuration for a service based on the category.
func forgeServiceConfigurationDefaultFromCategory(category consts.ServiceCategory) (configuration string, err error) {
	var JSONtemplate string

	switch category {
	case consts.Database:
		JSONtemplate = `{
			"username": "admin",
			"password": "admin",
			"database": "mydb"
		}`
	case consts.MessageQueue:
		JSONtemplate = `{
			"username": "admin",
			"password": "adminpassword"
		}`
	// TODO (Service): Implement more categories based on ontology
	default:
		klog.Errorf("Category not recognized")
		return "", fmt.Errorf("category not recognized")
	}

	return JSONtemplate, nil
}

// ForgeFlavorFromRef creates a new flavor starting from a Reference Flavor and the new Characteristics.
func ForgeFlavorFromRef(f *nodecorev1alpha1.Flavor, newFlavorType *nodecorev1alpha1.FlavorType) (flavor *nodecorev1alpha1.Flavor) {
	return &nodecorev1alpha1.Flavor{
		ObjectMeta: metav1.ObjectMeta{
			Name:            namings.ForgeFlavorName(string(f.Spec.FlavorType.TypeIdentifier), f.Spec.Owner.Domain),
			Namespace:       flags.FluidosNamespace,
			OwnerReferences: f.GetOwnerReferences(),
		},
		Spec: nodecorev1alpha1.FlavorSpec{
			ProviderID:          f.Spec.ProviderID,
			FlavorType:          *newFlavorType,
			Owner:               f.Spec.Owner,
			Price:               f.Spec.Price,
			Availability:        true,
			NetworkPropertyType: f.Spec.NetworkPropertyType,
			Location:            f.Spec.Location,
		},
	}
}

// FORGER FUNCTIONS FROM OBJECTS

// ForgeTransactionObj creates a new Transaction object.
func ForgeTransactionObj(id string, req *models.ReserveRequest) *models.Transaction {
	return &models.Transaction{
		TransactionID: id,
		Buyer:         req.Buyer,
		ClusterID:     req.Buyer.AdditionalInformation.LiqoID,
		FlavorID:      req.FlavorID,
		Configuration: func() *models.Configuration {
			if req.Configuration != nil {
				return req.Configuration
			}
			return nil
		}(),
		ExpirationTime: tools.GetExpirationTime(1, 0, 0),
	}
}

// ForgeContractObj creates a new Contract object.
func ForgeContractObj(contract *reservationv1alpha1.Contract) models.Contract {
	return models.Contract{
		ContractID:     contract.Name,
		Flavor:         *parseutil.ParseFlavor(&contract.Spec.Flavor),
		Buyer:          parseutil.ParseNodeIdentity(contract.Spec.Buyer),
		BuyerClusterID: contract.Spec.BuyerClusterID,
		Seller:         parseutil.ParseNodeIdentity(contract.Spec.Seller),
		PeeringTargetCredentials: models.LiqoCredentials{
			ClusterID:   contract.Spec.PeeringTargetCredentials.ClusterID,
			ClusterName: contract.Spec.PeeringTargetCredentials.ClusterName,
			Token:       contract.Spec.PeeringTargetCredentials.Token,
			Endpoint:    contract.Spec.PeeringTargetCredentials.Endpoint,
		},
		Configuration: func() *models.Configuration {
			if contract.Spec.Configuration != nil {
				configuration, err := parseutil.ParseConfiguration(contract.Spec.Configuration, &contract.Spec.Flavor)
				if err != nil {
					klog.Errorf("Error when parsing configuration: %s", err)
					return nil
				}
				return configuration
			}
			return nil
		}(),
		TransactionID:  contract.Spec.TransactionID,
		ExpirationTime: contract.Spec.ExpirationTime,
		ExtraInformation: func() map[string]string {
			if contract.Spec.ExtraInformation != nil {
				return contract.Spec.ExtraInformation
			}
			return nil
		}(),
	}
}

// ForgeNodeIdentitiesFromObj creates a NodeIdentity CR from a NodeIdentity Object.
func ForgeNodeIdentitiesFromObj(nodeIdentity *models.NodeIdentity) *nodecorev1alpha1.NodeIdentity {
	return &nodecorev1alpha1.NodeIdentity{
		NodeID: nodeIdentity.NodeID,
		IP:     nodeIdentity.IP,
		Domain: nodeIdentity.Domain,
		AdditionalInformation: func() *nodecorev1alpha1.NodeIdentityAdditionalInfo {
			if nodeIdentity.AdditionalInformation != nil {
				return &nodecorev1alpha1.NodeIdentityAdditionalInfo{
					LiqoID: nodeIdentity.AdditionalInformation.LiqoID,
				}
			}
			return nil
		}(),
	}
}

// ForgeContractFromObj creates a Contract from a reservation.
func ForgeContractFromObj(contract *models.Contract) (*reservationv1alpha1.Contract, error) {
	// Forge flavorCR
	flavorCR, err := ForgeFlavorFromObj(&contract.Flavor)
	if err != nil {
		return nil, err
	}
	return &reservationv1alpha1.Contract{
		ObjectMeta: metav1.ObjectMeta{
			Name:      contract.ContractID,
			Namespace: flags.FluidosNamespace,
		},
		Spec: reservationv1alpha1.ContractSpec{
			Flavor:         *flavorCR,
			Buyer:          *ForgeNodeIdentitiesFromObj(&contract.Buyer),
			BuyerClusterID: contract.BuyerClusterID,
			Seller:         *ForgeNodeIdentitiesFromObj(&contract.Seller),
			PeeringTargetCredentials: nodecorev1alpha1.LiqoCredentials{
				ClusterID:   contract.PeeringTargetCredentials.ClusterID,
				ClusterName: contract.PeeringTargetCredentials.ClusterName,
				Token:       contract.PeeringTargetCredentials.Token,
				Endpoint:    contract.PeeringTargetCredentials.Endpoint,
			},
			TransactionID: contract.TransactionID,
			Configuration: func() *nodecorev1alpha1.Configuration {
				if contract.Configuration != nil {
					configuration, err := ForgeConfigurationFromObj(*contract.Configuration)
					if err != nil {
						klog.Errorf("Error when parsing configuration: %s", err)
						return nil
					}
					return configuration
				}
				return nil
			}(),
			ExpirationTime: contract.ExpirationTime,
			ExtraInformation: func() map[string]string {
				if contract.ExtraInformation != nil {
					return contract.ExtraInformation
				}
				return nil
			}(),
			NetworkRequests:          contract.NetworkRequests,
			IngressTelemetryEndpoint: ForgeTelemetryServer(contract.IngressTelemetryEndpoint),
		},
		Status: reservationv1alpha1.ContractStatus{
			Phase: nodecorev1alpha1.PhaseStatus{
				Phase:     nodecorev1alpha1.PhaseActive,
				StartTime: tools.GetTimeNow(),
			},
		},
	}, nil
}

// ForgeTransactionFromObj creates a transaction from a Transaction object.
func ForgeTransactionFromObj(transaction *models.Transaction) *reservationv1alpha1.Transaction {
	return &reservationv1alpha1.Transaction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      transaction.TransactionID,
			Namespace: flags.FluidosNamespace,
		},
		Spec: reservationv1alpha1.TransactionSpec{
			FlavorID:       transaction.FlavorID,
			ExpirationTime: transaction.ExpirationTime,
			Buyer: nodecorev1alpha1.NodeIdentity{
				Domain: transaction.Buyer.Domain,
				IP:     transaction.Buyer.IP,
				NodeID: transaction.Buyer.NodeID,
				AdditionalInformation: func() *nodecorev1alpha1.NodeIdentityAdditionalInfo {
					if transaction.Buyer.AdditionalInformation != nil {
						return &nodecorev1alpha1.NodeIdentityAdditionalInfo{
							LiqoID: transaction.Buyer.AdditionalInformation.LiqoID,
						}
					}
					return nil
				}(),
			},
			Configuration: func() *nodecorev1alpha1.Configuration {
				if transaction.Configuration != nil {
					configuration, err := ForgeConfigurationFromObj(*transaction.Configuration)
					if err != nil {
						klog.Errorf("Error when parsing configuration: %s", err)
						return nil
					}
					return configuration
				}
				return nil
			}(),
		},
	}
}

// ForgeConfigurationFromObj creates a Configuration CR from a Configuration object.
func ForgeConfigurationFromObj(configuration models.Configuration) (*nodecorev1alpha1.Configuration, error) {
	// Parse the Configuration
	switch configuration.Type {
	case models.K8SliceNameDefault:
		// Force casting of configurationStruct to K8Slice
		var configurationStruct models.K8SliceConfiguration
		err := json.Unmarshal(configuration.Data, &configurationStruct)
		if err != nil {
			return nil, err
		}
		k8SliceConfiguration := &nodecorev1alpha1.K8SliceConfiguration{
			CPU:    configurationStruct.CPU,
			Memory: configurationStruct.Memory,
			Pods:   configurationStruct.Pods,
			Gpu: func() *nodecorev1alpha1.GPU {
				if configurationStruct.Gpu != nil {
					return &nodecorev1alpha1.GPU{
						Model:  configurationStruct.Gpu.Model,
						Cores:  configurationStruct.Gpu.Cores,
						Memory: configurationStruct.Gpu.Memory,
					}
				}
				return nil
			}(),
			Storage: configurationStruct.Storage,
		}

		// Marshal the K8Slice configuration to JSON
		configurationData, err := json.Marshal(k8SliceConfiguration)
		if err != nil {
			return nil, err
		}

		return &nodecorev1alpha1.Configuration{
			ConfigurationTypeIdentifier: nodecorev1alpha1.TypeK8Slice,
			ConfigurationData:           runtime.RawExtension{Raw: configurationData},
		}, nil
	case models.VMNameDefault:
		// TODO (VM): Implement VM configuration
		return nil, fmt.Errorf("vm configuration not implemented")
	case models.ServiceNameDefault:
		// Force casting of configurationStruct to Service
		var configurationStruct models.ServiceConfiguration
		err := json.Unmarshal(configuration.Data, &configurationStruct)
		if err != nil {
			return nil, err
		}
		// Create ServiceConfiguration nodecorev1alpha1
		serviceConfigurationCR := nodecorev1alpha1.ServiceConfiguration{
			HostingPolicy: func() *nodecorev1alpha1.HostingPolicy {
				if configurationStruct.HostingPolicy != nil {
					hp := models.MapFromModelHostingPolicy(*configurationStruct.HostingPolicy)
					return &hp
				}
				return nil
			}(),
			ConfigurationData: runtime.RawExtension{
				Raw: configurationStruct.ConfigurationData,
			},
		}
		// Marshal ServiceConfiguration to JSON
		configurationData, err := json.Marshal(serviceConfigurationCR)
		if err != nil {
			return nil, err
		}
		return &nodecorev1alpha1.Configuration{
			ConfigurationTypeIdentifier: nodecorev1alpha1.TypeService,
			ConfigurationData:           runtime.RawExtension{Raw: configurationData},
		}, nil
	case models.SensorNameDefault:
		// TODO (Sensor): Implement Sensor configuration
		return nil, fmt.Errorf("sensor configuration not implemented")
	default:
		return nil, fmt.Errorf("unknown configuration type")
	}
}

// ForgeConfigurationObj creates a Configuration object from a Configuration CR.
func ForgeConfigurationObj(configuration *nodecorev1alpha1.Configuration) (*models.Configuration, error) {
	var data json.RawMessage
	switch configuration.ConfigurationTypeIdentifier {
	case nodecorev1alpha1.TypeK8Slice:
		// Force casting of configurationStruct to K8Slice
		var configurationStruct nodecorev1alpha1.K8SliceConfiguration
		err := json.Unmarshal(configuration.ConfigurationData.Raw, &configurationStruct)
		if err != nil {
			return nil, err
		}
		k8SliceConfiguration := models.K8SliceConfiguration{
			CPU:    configurationStruct.CPU,
			Memory: configurationStruct.Memory,
			Pods:   configurationStruct.Pods,
			Gpu: func() *models.GpuCharacteristics {
				if configurationStruct.Gpu != nil {
					return &models.GpuCharacteristics{
						Model:  configurationStruct.Gpu.Model,
						Cores:  configurationStruct.Gpu.Cores,
						Memory: configurationStruct.Gpu.Memory,
					}
				}
				return nil
			}(),
			Storage: configurationStruct.Storage,
		}

		// Marshal the K8Slice configuration to JSON
		data, err = json.Marshal(k8SliceConfiguration)
		if err != nil {
			return nil, err
		}
	case nodecorev1alpha1.TypeService:
		// Force casting of configurationStruct to Service
		var configurationStruct nodecorev1alpha1.ServiceConfiguration
		err := json.Unmarshal(configuration.ConfigurationData.Raw, &configurationStruct)
		if err != nil {
			return nil, err
		}
		// Convert ConfigurationData to json.RawMessage
		configurationData, err := json.Marshal(configurationStruct.ConfigurationData)
		if err != nil {
			return nil, err
		}
		// Create ServiceConfiguration nodecorev1alpha1
		serviceConfiguration := models.ServiceConfiguration{
			HostingPolicy: func() *models.HostingPolicy {
				if configurationStruct.HostingPolicy != nil {
					hp := models.MapToModelHostingPolicy(*configurationStruct.HostingPolicy)
					return &hp
				}
				return nil
			}(),
			ConfigurationData: json.RawMessage(configurationData),
		}

		// Marshal ServiceConfiguration to JSON
		data, err = json.Marshal(serviceConfiguration)
		if err != nil {
			return nil, err
		}
	case nodecorev1alpha1.TypeSensor:
		// TODO (Sensor): Implement Sensor configuration
		return nil, fmt.Errorf("sensor configuration not implemented")
	case nodecorev1alpha1.TypeVM:
		// TODO (VM): Implement VM configuration
		return nil, fmt.Errorf("vm configuration not implemented")
	default:
		return nil, fmt.Errorf("unknown configuration type")
	}
	modelConf := models.Configuration{
		Type: models.MapToFlavorTypeName(configuration.ConfigurationTypeIdentifier),
		Data: data,
	}

	return &modelConf, nil
}

// ForgeResourceSelectorFromObj creates a ResourceSelector CR from a ResourceSelector Object.
func ForgeResourceSelectorFromObj(resourceSelector *models.ResourceSelector) *nodecorev1alpha1.ResourceSelector {
	// Parse ResourceSelector
	switch resourceSelector.TypeIdentifier {
	case models.CIDRSelectorType:
		// unmarshal CIDRSelector
		var resourceSelectorStruct models.CIDRSelector
		err := json.Unmarshal(resourceSelector.Selector, &resourceSelectorStruct)
		if err != nil {
			klog.Errorf("Error when unmarshaling CIDRSelector: %s", err)
			return nil
		}
		// Create CIDRSelector nodecorev1alpha1
		cidrSelectorCR := nodecorev1alpha1.CIDRSelector(resourceSelectorStruct)
		// Marshal CIDRSelector to JSON
		resourceSelectorData, err := json.Marshal(cidrSelectorCR)
		if err != nil {
			klog.Errorf("Error when marshaling CIDRSelector: %s", err)
			return nil
		}
		return &nodecorev1alpha1.ResourceSelector{
			TypeIdentifier: nodecorev1alpha1.CIDRSelectorType,
			Selector:       runtime.RawExtension{Raw: resourceSelectorData},
		}
	case models.PodNamespaceSelectorType:
		// Force casting of resourceSelector to PodNamespaceSelector type
		var resourceSelectorStruct models.PodNamespaceSelector
		err := json.Unmarshal(resourceSelector.Selector, &resourceSelectorStruct)
		if err != nil {
			klog.Errorf("Error when unmarshaling PodNamespaceSelector: %s", err)
			return nil
		}
		// Create PodNamespaceSelector nodecorev1alpha1
		podNamespaceSelectorCR := nodecorev1alpha1.PodNamespaceSelector{
			// Copy map of models.PodNamespaceSelector.Pod to nodecorev1alpha1.PodNamespaceSelector.Pod
			Pod: func() map[string]string {
				podMap := make(map[string]string)
				for i := range resourceSelectorStruct.Pod {
					keyValuePair := resourceSelectorStruct.Pod[i]
					podMap[keyValuePair.Key] = keyValuePair.Value
				}
				return podMap
			}(),
			// Copy map of models.PodNamespaceSelector.Namespace to nodecorev1alpha1.PodNamespaceSelector.Namespace
			Namespace: func() map[string]string {
				namespaceMap := make(map[string]string)
				for i := range resourceSelectorStruct.Namespace {
					keyValuePair := resourceSelectorStruct.Namespace[i]
					namespaceMap[keyValuePair.Key] = keyValuePair.Value
				}
				return namespaceMap
			}(),
		}
		// Marshal PodNamespaceSelector to JSON
		resourceSelectorData, err := json.Marshal(podNamespaceSelectorCR)
		if err != nil {
			klog.Errorf("Error when marshaling PodNamespaceSelector: %s", err)
			return nil
		}
		return &nodecorev1alpha1.ResourceSelector{
			TypeIdentifier: nodecorev1alpha1.PodNamespaceSelectorType,
			Selector:       runtime.RawExtension{Raw: resourceSelectorData},
		}
	default:
		klog.Errorf("Resource selector type not recognized")
		return nil
	}
}

// ForgeSourceDestinationFromObj creates a SourceDestination CR from a SourceDestination Object.
func ForgeSourceDestinationFromObj(sourceDestination *models.SourceDestination) *nodecorev1alpha1.SourceDestination {
	// Parse ResourceSelector
	resourceSelector := ForgeResourceSelectorFromObj(&sourceDestination.ResourceSelector)
	if resourceSelector == nil {
		klog.Errorf("Error when parsing resource selector from source destination")
		return nil
	}
	return &nodecorev1alpha1.SourceDestination{
		IsHotCluster:     sourceDestination.IsHotCluster,
		ResourceSelector: *resourceSelector,
	}
}

// ForgeNetworkIntentFromObj creates a NetworkIntent CR from a NetworkIntent Object.
func ForgeNetworkIntentFromObj(networkIntent *models.NetworkIntent) *nodecorev1alpha1.NetworkIntent {
	// Parse NetworkIntent
	source := ForgeSourceDestinationFromObj(&networkIntent.Source)
	if source == nil {
		klog.Errorf("Error when parsing source from network intent")
		return nil
	}
	destination := ForgeSourceDestinationFromObj(&networkIntent.Destination)
	if destination == nil {
		klog.Errorf("Error when parsing destination from network intent")
		return nil
	}
	return &nodecorev1alpha1.NetworkIntent{
		Name:            networkIntent.Name,
		Source:          *source,
		Destination:     *destination,
		DestinationPort: networkIntent.DestinationPort,
		ProtocolType:    networkIntent.ProtocolType,
	}
}

// ForgeNetworkAuthorizationsFromObj creates a NetworkAuthorizations CR from a NetworkAuthorizations Object.
func ForgeNetworkAuthorizationsFromObj(networkAuthorizations *models.NetworkAuthorizations) *nodecorev1alpha1.NetworkAuthorizations {
	// DeniedCommunications
	var deniedCommunicationsModel []nodecorev1alpha1.NetworkIntent
	var mandatoryCommunicationsModel []nodecorev1alpha1.NetworkIntent
	for i := range networkAuthorizations.DeniedCommunications {
		deniedCommunication := networkAuthorizations.DeniedCommunications[i]
		// Parse the DeniedCommunication
		ni := ForgeNetworkIntentFromObj(&deniedCommunication)
		if ni == nil {
			klog.Errorf("Error when parsing denied communication from network authorizations")
		} else {
			deniedCommunicationsModel = append(deniedCommunicationsModel, *ni)
		}
	}
	// MandatoryCommunications
	for i := range networkAuthorizations.MandatoryCommunications {
		mandatoryCommunication := networkAuthorizations.MandatoryCommunications[i]
		// Parse the MandatoryCommunication
		ni := ForgeNetworkIntentFromObj(&mandatoryCommunication)
		if ni == nil {
			klog.Errorf("Error when parsing mandatory communication from network authorizations")
		} else {
			mandatoryCommunicationsModel = append(mandatoryCommunicationsModel, *ni)
		}
	}
	return &nodecorev1alpha1.NetworkAuthorizations{
		DeniedCommunications:    deniedCommunicationsModel,
		MandatoryCommunications: mandatoryCommunicationsModel,
	}
}

// ForgeFlavorFromObj creates a Flavor CR from a Flavor Object (REAR).
func ForgeFlavorFromObj(flavor *models.Flavor) (*nodecorev1alpha1.Flavor, error) {
	var flavorType nodecorev1alpha1.FlavorType

	switch flavor.Type.Name {
	case models.K8SliceNameDefault:
		// Unmarshal K8SliceType
		var flavorTypeDataModel models.K8Slice
		err := json.Unmarshal(flavor.Type.Data, &flavorTypeDataModel)
		if err != nil {
			klog.Errorf("Error when unmarshalling K8SliceType: %s", err)
			return nil, err
		}
		flavorTypeData := nodecorev1alpha1.K8Slice{
			Characteristics: nodecorev1alpha1.K8SliceCharacteristics{
				Architecture: flavorTypeDataModel.Characteristics.Architecture,
				CPU:          flavorTypeDataModel.Characteristics.CPU,
				Memory:       flavorTypeDataModel.Characteristics.Memory,
				Pods:         flavorTypeDataModel.Characteristics.Pods,
				Storage:      flavorTypeDataModel.Characteristics.Storage,
				Gpu: func() *nodecorev1alpha1.GPU {
					if flavorTypeDataModel.Characteristics.Gpu != nil {
						return &nodecorev1alpha1.GPU{
							Model:  flavorTypeDataModel.Characteristics.Gpu.Model,
							Cores:  flavorTypeDataModel.Characteristics.Gpu.Cores,
							Memory: flavorTypeDataModel.Characteristics.Gpu.Memory,
						}
					}
					return nil
				}(),
			},
			Properties: nodecorev1alpha1.Properties{
				Latency:           flavorTypeDataModel.Properties.Latency,
				SecurityStandards: flavorTypeDataModel.Properties.SecurityStandards,
				CarbonFootprint: func() *nodecorev1alpha1.CarbonFootprint {
					if flavorTypeDataModel.Properties.CarbonFootprint != nil {
						return &nodecorev1alpha1.CarbonFootprint{
							Embodied:    flavorTypeDataModel.Properties.CarbonFootprint.Embodied,
							Operational: flavorTypeDataModel.Properties.CarbonFootprint.Operational,
						}
					}
					return nil
				}(),
				NetworkAuthorizations: func() *nodecorev1alpha1.NetworkAuthorizations {
					if flavorTypeDataModel.Properties.NetworkAuthorizations != nil {
						return ForgeNetworkAuthorizationsFromObj(flavorTypeDataModel.Properties.NetworkAuthorizations)
					}
					return nil
				}(),
			},
			Policies: nodecorev1alpha1.Policies{
				Partitionability: nodecorev1alpha1.Partitionability{
					CPUMin:     flavorTypeDataModel.Policies.Partitionability.CPUMin,
					MemoryMin:  flavorTypeDataModel.Policies.Partitionability.MemoryMin,
					PodsMin:    flavorTypeDataModel.Policies.Partitionability.PodsMin,
					CPUStep:    flavorTypeDataModel.Policies.Partitionability.CPUStep,
					MemoryStep: flavorTypeDataModel.Policies.Partitionability.MemoryStep,
					PodsStep:   flavorTypeDataModel.Policies.Partitionability.PodsStep,
				},
			},
		}

		if err := forgeK8SlicePropertyAdditionalPropertiesFromObj(&flavorTypeDataModel.Properties, &flavorTypeData.Properties); err != nil {
			klog.Errorf("Error when forging K8Slice additional properties: %s", err)
			return nil, err
		}

		flavorTypeDataJSON, err := json.Marshal(flavorTypeData)
		if err != nil {
			klog.Errorf("Error when marshaling K8SliceType: %s", err)
			return nil, err
		}
		flavorType = nodecorev1alpha1.FlavorType{
			TypeIdentifier: nodecorev1alpha1.TypeK8Slice,
			TypeData:       runtime.RawExtension{Raw: flavorTypeDataJSON},
		}
	case models.VMNameDefault:
		// TODO (VM): Implement VM flavor
		return nil, fmt.Errorf("VM flavor not implemented")
	case models.ServiceNameDefault:
		// Unmarshal ServiceFlavorType
		var flavorTypeDataModel models.ServiceFlavor
		err := json.Unmarshal(flavor.Type.Data, &flavorTypeDataModel)
		if err != nil {
			klog.Errorf("Error when unmarshalling ServiceType: %s", err)
			return nil, err
		}
		flavorTypeData := nodecorev1alpha1.ServiceFlavor{
			Name:                  flavorTypeDataModel.Name,
			Description:           flavorTypeDataModel.Description,
			Category:              flavorTypeDataModel.Category,
			Tags:                  flavorTypeDataModel.Tags,
			ConfigurationTemplate: runtime.RawExtension{Raw: flavorTypeDataModel.ConfigurationTemplate},
		}
		flavorTypeDataJSON, err := json.Marshal(flavorTypeData)
		if err != nil {
			klog.Errorf("Error when marshaling ServiceType: %s", err)
			return nil, err
		}
		flavorType = nodecorev1alpha1.FlavorType{
			TypeIdentifier: nodecorev1alpha1.TypeService,
			TypeData:       runtime.RawExtension{Raw: flavorTypeDataJSON},
		}
	case models.SensorNameDefault:
		// TODO (Sensor): Implement Sensor flavor
		return nil, fmt.Errorf("sensor flavor not implemented")
	default:
		klog.Errorf("Flavor type not recognized")
		return nil, fmt.Errorf("flavor type not recognized")
	}
	f := &nodecorev1alpha1.Flavor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      flavor.FlavorID,
			Namespace: flags.FluidosNamespace,
		},
		Spec: nodecorev1alpha1.FlavorSpec{
			ProviderID: flavor.Owner.NodeID,
			FlavorType: flavorType,
			Owner: nodecorev1alpha1.NodeIdentity{
				Domain: flavor.Owner.Domain,
				IP:     flavor.Owner.IP,
				NodeID: flavor.Owner.NodeID,
			},
			Price: nodecorev1alpha1.Price{
				Amount:   flavor.Price.Amount,
				Currency: flavor.Price.Currency,
				Period:   flavor.Price.Period,
			},
			Availability:        flavor.Availability,
			NetworkPropertyType: flavor.NetworkPropertyType,
			Location: func() *nodecorev1alpha1.Location {
				if flavor.Location != nil {
					return &nodecorev1alpha1.Location{
						Latitude:        flavor.Location.Latitude,
						Longitude:       flavor.Location.Longitude,
						Country:         flavor.Location.Country,
						City:            flavor.Location.City,
						AdditionalNotes: flavor.Location.AdditionalNotes,
					}
				}
				return nil
			}(),
		},
	}
	return f, nil
}

func forgeK8SlicePropertyAdditionalPropertiesFromObj(k8SliceObjProperties *models.K8SliceProperties,
	k8SliceProperties *nodecorev1alpha1.Properties) error {
	if k8SliceObjProperties == nil {
		klog.Info("K8Slice properties not found")
		return nil
	}

	if k8SliceProperties == nil {
		klog.Info("K8Slice additional properties not found")
		return nil
	}

	// Check additional properties
	if k8SliceObjProperties.AdditionalProperties == nil {
		klog.Info("K8Slice additional properties not found")
		return nil
	}

	// Initialize additional properties
	k8SliceProperties.AdditionalProperties = make(map[string]runtime.RawExtension)

	// Forge additional properties
	for key, value := range k8SliceObjProperties.AdditionalProperties {
		// Check for empty value
		if value == nil {
			klog.Errorf("Empty value for additional property %s", key)
			return fmt.Errorf("empty value for additional property %s", key)
		}
		// Convert json.RawMessage to runtime.RawExtension
		rawExtension := runtime.RawExtension{Raw: value}
		// Add additional property to K8SliceProperties
		k8SliceProperties.AdditionalProperties[key] = rawExtension
	}

	return nil
}

// ForgeK8SliceConfiguration creates a Configuration from a FlavorSelector.
func ForgeK8SliceConfiguration(selector nodecorev1alpha1.K8SliceSelector, flavor *nodecorev1alpha1.K8Slice) *nodecorev1alpha1.K8SliceConfiguration {
	var cpu, memory, pods resource.Quantity
	var gpu *nodecorev1alpha1.GPU
	var storage *resource.Quantity

	klog.Info("Parsing K8Slice selector")

	if selector.CPUFilter != nil {
		// Parse CPU filter
		klog.Info("Parsing CPU filter")
		cpuFilterType, cpuFilterData, err := nodecorev1alpha1.ParseResourceQuantityFilter(selector.CPUFilter)
		if err != nil {
			klog.Errorf("Error when parsing CPU filter: %s", err)
			return nil
		}
		// Define configuration value based on filter type
		switch cpuFilterType {
		// Match Filter
		case nodecorev1alpha1.TypeMatchFilter:
			cpu = cpuFilterData.(nodecorev1alpha1.ResourceMatchSelector).Value
		// Range Filter
		case nodecorev1alpha1.TypeRangeFilter:
			// Check if min value is set
			if cpuFilterData.(nodecorev1alpha1.ResourceRangeSelector).Min != nil {
				rrs := cpuFilterData.(nodecorev1alpha1.ResourceRangeSelector)
				cpu = *rrs.Min
			}

		// Default
		default:
			klog.Errorf("CPU filter type not recognized")
			return nil
		}
	} else {
		cpu = flavor.Characteristics.CPU
	}

	if selector.MemoryFilter != nil {
		// Parse Memory filter
		klog.Info("Parsing Memory filter")
		memoryFilterType, memoryFilterData, err := nodecorev1alpha1.ParseResourceQuantityFilter(selector.MemoryFilter)
		if err != nil {
			klog.Errorf("Error when parsing Memory filter: %s", err)
			return nil
		}
		// Define configuration value based on filter type
		switch memoryFilterType {
		// Match Filter
		case nodecorev1alpha1.TypeMatchFilter:
			memory = memoryFilterData.(nodecorev1alpha1.ResourceMatchSelector).Value
		// Range Filter
		case nodecorev1alpha1.TypeRangeFilter:
			// Check if min value is set
			if memoryFilterData.(nodecorev1alpha1.ResourceRangeSelector).Min != nil {
				rrs := memoryFilterData.(nodecorev1alpha1.ResourceRangeSelector)
				memory = *rrs.Min
			}
		// Default
		default:
			klog.Errorf("Memory filter type not recognized")
			return nil
		}
	} else {
		memory = flavor.Characteristics.Memory
	}

	if selector.PodsFilter != nil {
		// Parse Pods filter
		klog.Info("Parsing Pods filter")
		podsFilterType, podsFilterData, err := nodecorev1alpha1.ParseResourceQuantityFilter(selector.PodsFilter)
		if err != nil {
			klog.Errorf("Error when parsing Pods filter: %s", err)
			return nil
		}
		// Define configuration value based on filter type
		switch podsFilterType {
		// Match Filter
		case nodecorev1alpha1.TypeMatchFilter:
			pods = podsFilterData.(nodecorev1alpha1.ResourceMatchSelector).Value
		// Range Filter
		case nodecorev1alpha1.TypeRangeFilter:
			// Check if min value is set
			if podsFilterData.(nodecorev1alpha1.ResourceRangeSelector).Min != nil {
				rrs := podsFilterData.(nodecorev1alpha1.ResourceRangeSelector)
				pods = *rrs.Min
			}

		// Default
		default:
			klog.Errorf("Pods filter type not recognized")
			return nil
		}
	} else {
		pods = flavor.Characteristics.Pods
	}

	if selector.StorageFilter != nil {
		// Parse Storage filter
		klog.Info("Parsing Storage filter")
		storageFilterType, storageFilterData, err := nodecorev1alpha1.ParseResourceQuantityFilter(selector.StorageFilter)
		if err != nil {
			klog.Errorf("Error when parsing Storage filter: %s", err)
			return nil
		}
		// Define configuration value based on filter type
		switch storageFilterType {
		// Match Filter
		case nodecorev1alpha1.TypeMatchFilter:
			value := storageFilterData.(nodecorev1alpha1.ResourceMatchSelector).Value
			storage = &value
		// Range Filter
		case nodecorev1alpha1.TypeRangeFilter:
			// Check if min value is set
			if storageFilterData.(nodecorev1alpha1.ResourceRangeSelector).Min != nil {
				rrs := storageFilterData.(nodecorev1alpha1.ResourceRangeSelector)
				if rrs.Min != nil {
					storage = rrs.Min
				}
			}
		// Default
		default:
			klog.Errorf("Storage filter type not recognized")
			return nil
		}
	}

	// Compose configuration based on values gathered from filters
	return &nodecorev1alpha1.K8SliceConfiguration{
		CPU:     cpu,
		Memory:  memory,
		Pods:    pods,
		Gpu:     gpu,
		Storage: storage,
	}
}

// ForgeAllocation creates an Allocation from a Contract.
func ForgeAllocation(contract *reservationv1alpha1.Contract) *nodecorev1alpha1.Allocation {
	return &nodecorev1alpha1.Allocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeAllocationName(contract.Spec.Flavor.Name),
			Namespace: flags.FluidosNamespace,
		},
		Spec: nodecorev1alpha1.AllocationSpec{
			Forwarding: false,
			Contract: nodecorev1alpha1.GenericRef{
				Name:      contract.Name,
				Namespace: contract.Namespace,
			},
		},
	}
}

// ForgeDefaultServiceConfiguration forges the default service configuration based on the category of the service.
func ForgeDefaultServiceConfiguration(serviceFlavor *nodecorev1alpha1.ServiceFlavor) (*nodecorev1alpha1.ServiceConfiguration, error) {
	var defaultServiceConfiguration *nodecorev1alpha1.ServiceConfiguration

	var configurationJSON string

	configurationJSON, err := forgeServiceConfigurationDefaultFromCategory(models.MapToServiceCategory(serviceFlavor.Category))
	if err != nil {
		klog.Errorf("Error when forging default service configuration: %s", err)
		return nil, err
	}

	// Marshal the default service configuration to JSON
	defaultServiceConfiguration = &nodecorev1alpha1.ServiceConfiguration{
		ConfigurationData: runtime.RawExtension{Raw: []byte(configurationJSON)},
		// TODO(Service): default hosting policy is to omit it, therefore use the default one
	}

	return defaultServiceConfiguration, nil
}

// ForgeLiqoCredentialsObj creates a LiqoCredentials object from a LiqoCredentials CR.
func ForgeLiqoCredentialsObj(liqoCredentials *nodecorev1alpha1.LiqoCredentials) (*models.LiqoCredentials, error) {
	return &models.LiqoCredentials{
		ClusterID:   liqoCredentials.ClusterID,
		ClusterName: liqoCredentials.ClusterName,
		Token:       liqoCredentials.Token,
		Endpoint:    liqoCredentials.Endpoint,
	}, nil
}

// ForgeLiqoCredentialsFromObj creates a LiqoCredentials CR from a LiqoCredentials object.
func ForgeLiqoCredentialsFromObj(liqoCredentials *models.LiqoCredentials) (*nodecorev1alpha1.LiqoCredentials, error) {
	return &nodecorev1alpha1.LiqoCredentials{
		ClusterID:   liqoCredentials.ClusterID,
		ClusterName: liqoCredentials.ClusterName,
		Token:       liqoCredentials.Token,
		Endpoint:    liqoCredentials.Endpoint,
	}, nil
}

// ForgePodOffloadingStrategy creates a PodOffloadingStrategy CR from a nodecorev1alpha1.HostingPolicy.
func ForgePodOffloadingStrategy(hostingPolicy *nodecorev1alpha1.HostingPolicy) (offloadingv1alpha1.PodOffloadingStrategyType, error) {
	switch *hostingPolicy {
	case nodecorev1alpha1.HostingPolicyProvider:
		return offloadingv1alpha1.LocalPodOffloadingStrategyType, nil
	case nodecorev1alpha1.HostingPolicyConsumer:
		return offloadingv1alpha1.RemotePodOffloadingStrategyType, nil
	case nodecorev1alpha1.HostingPolicyShared:
		return offloadingv1alpha1.LocalAndRemotePodOffloadingStrategyType, nil
	default:
		return "", fmt.Errorf("hosting policy not recognized")
	}
}

// ForgeServiceManifests creates YAML Kubernetes manifests from a reservationv1alpha1.Contract.
func ForgeServiceManifests(ctx context.Context, c client.Client, contract *reservationv1alpha1.Contract) ([]string, error) {
	var manifests []string

	// Retrieve the Service blueprint, owner of the Flavor specified in the contract
	serviceBlueprint, err := getters.GetBlueprint(ctx, c, contract.Spec.Flavor.Name)
	if err != nil {
		klog.Errorf("Error when getting blueprint: %s", err)
		return nil, err
	}

	// Retrieve the Configuration of the Service from the Contract
	_, configurationData, err := nodecorev1alpha1.ParseConfiguration(contract.Spec.Configuration, &contract.Spec.Flavor)
	if err != nil {
		klog.Errorf("Error when parsing configuration: %s", err)
		return nil, err
	}

	// Force casting of configurationData to ServiceConfiguration
	serviceConfiguration, ok := configurationData.(nodecorev1alpha1.ServiceConfiguration)
	if !ok {
		klog.Errorf("Error when casting configuration data to ServiceConfiguration")
		return nil, fmt.Errorf("error when casting configuration data to ServiceConfiguration")
	}

	// Render templates based on configuration

	// Obtain map[string]interface{} from the raw extension
	configurationDataMap := make(map[string]interface{})
	err = json.Unmarshal(serviceConfiguration.ConfigurationData.Raw, &configurationDataMap)
	if err != nil {
		klog.Errorf("Error when unmarshalling configuration data: %s", err)
		return nil, err
	}

	for _, template := range serviceBlueprint.Spec.Templates {
		// Obtain the template string from the raw extension
		templateString := string(template.ServiceTemplateData.Raw)
		klog.Infof("Template: %s", templateString)
		// Render the template
		renderedTemplate, err := RenderTemplate(templateString, configurationDataMap)
		if err != nil {
			klog.Errorf("Error when rendering template: %s", err)
			return nil, err
		}
		manifests = append(manifests, renderedTemplate)
	}

	// TODO (Service): Implement the creation of the service manifests

	return manifests, nil
}

// RenderTemplate renders a template with the given data.
func RenderTemplate(yamlTemplate string, data map[string]interface{}) (string, error) {
	// Parse the template
	tmpl, err := template.New("k8sTemplate").Funcs(sprig.TxtFuncMap()).Parse(yamlTemplate)
	if err != nil {
		klog.Errorf("Error when parsing template: %s", err)
		return "", err
	}

	var rendered bytes.Buffer
	// Execute the template with the data
	err = tmpl.Execute(&rendered, data)
	if err != nil {
		klog.Errorf("Error when executing template: %s", err)
		return "", err
	}

	return rendered.String(), nil
}

// ForgeHostingPolicyFromContract creates a HostingPolicy from a Contract.
func ForgeHostingPolicyFromContract(contract *reservationv1alpha1.Contract, cl client.Client) (nodecorev1alpha1.HostingPolicy, error) {
	if contract.Spec.Configuration == nil {
		klog.Infof("Configuration not found in the contract, proceeding with default hosting policy")

		return forgeDefaultHostingPolicyFromFlavor(&contract.Spec.Flavor, cl)
	}
	klog.Info("Configuration found in the contract, proceeding with the hosting policy specified in the configuration")
	// Retrieve the configuration data from the contract
	configurationType, configurationData, err := nodecorev1alpha1.ParseConfiguration(contract.Spec.Configuration, &contract.Spec.Flavor)
	if err != nil {
		klog.Errorf("Error when parsing configuration: %s", err)
		return "", err
	}

	// Check the configuration type
	if configurationType != nodecorev1alpha1.TypeService {
		klog.Errorf("Configuration type is not Service")
		return "", fmt.Errorf("configuration type is not Service")
	}

	// Force casting the configuration data to ServiceConfiguration
	serviceConfiguration, ok := configurationData.(nodecorev1alpha1.ServiceConfiguration)
	if !ok {
		klog.Errorf("Error when casting configuration data to ServiceConfiguration")
		return "", fmt.Errorf("error when casting configuration data to ServiceConfiguration")
	}

	if serviceConfiguration.HostingPolicy != nil {
		return *serviceConfiguration.HostingPolicy, nil
	}
	klog.Errorf("Hosting policy not found in the configuration, proceeding with default hosting policy")
	return forgeDefaultHostingPolicyFromFlavor(&contract.Spec.Flavor, cl)
}

func forgeDefaultHostingPolicyFromFlavor(flavor *nodecorev1alpha1.Flavor, cl client.Client) (nodecorev1alpha1.HostingPolicy, error) {
	klog.Info("Default hosting policy requested, proceeding with the hosting policy specified in the flavor")
	// Get the service blueprint associated with the flavor specified in the contract
	serviceBlueprint, err := getters.GetBlueprint(context.Background(), cl, flavor.Name)
	if err != nil {
		klog.Errorf("Error when getting blueprint: %s", err)
		return "", err
	}

	if len(serviceBlueprint.Spec.HostingPolicies) == 0 {
		// No hosting policies found in the blueprint, default to Provider
		return nodecorev1alpha1.HostingPolicyProvider, nil
	}
	// Get the first hosting policy found in the blueprint
	return serviceBlueprint.Spec.HostingPolicies[0], nil
}

// ForgeSecretForService creates a Secret based on a contract for the service going to be created,
// following default behaviors based on the service category.
func ForgeSecretForService(contract *reservationv1alpha1.Contract,
	serviceEndpoint *corev1.Service) (*corev1.Secret, error) {
	// Parse the flavor type
	flavorTypeIdentifier, flavorTypeData, err := nodecorev1alpha1.ParseFlavorType(&contract.Spec.Flavor)
	if err != nil {
		klog.Errorf("Error when parsing flavor type: %s", err)
		return nil, err
	}
	if flavorTypeIdentifier != nodecorev1alpha1.TypeService {
		klog.Errorf("Flavor type is not Service")
		return nil, fmt.Errorf("flavor type is not Service")
	}
	// Force casting
	serviceFlavor, ok := flavorTypeData.(nodecorev1alpha1.ServiceFlavor)
	if !ok {
		klog.Errorf("Error when casting flavor type data to ServiceFlavor")
		return nil, fmt.Errorf("error when casting flavor type data to ServiceFlavor")
	}

	// Parse configuration data
	_, configurationData, err := nodecorev1alpha1.ParseConfiguration(contract.Spec.Configuration, &contract.Spec.Flavor)
	if err != nil {
		klog.Errorf("Error when parsing configuration: %s", err)
		return nil, err
	}

	// Force casting of configurationData to ServiceConfiguration
	serviceConfiguration, ok := configurationData.(nodecorev1alpha1.ServiceConfiguration)
	if !ok {
		klog.Errorf("Error when casting configuration data to ServiceConfiguration")
		return nil, fmt.Errorf("error when casting configuration data to ServiceConfiguration")
	}

	// Cast serviceConfiguration.ConfigurationData to map[string]interface{}
	configurationDataMap := make(map[string]interface{})
	err = json.Unmarshal(serviceConfiguration.ConfigurationData.Raw, &configurationDataMap)
	if err != nil {
		klog.Errorf("Error when unmarshalling configuration data: %s", err)
		return nil, err
	}

	// Endpoints generation
	var endpoints = make([]string, 0)
	if serviceEndpoint != nil {
		// Generate service dns name with default k8s dns resolution
		endpointsHost := serviceEndpoint.Name + "." + serviceEndpoint.Namespace + ".svc.cluster.local"
		for _, port := range serviceEndpoint.Spec.Ports {
			endpoints = append(endpoints, fmt.Sprintf("%s:%d", endpointsHost, port.Port))
		}
	} else {
		klog.Infof("Service endpoint not found")
	}

	// Convert endpoints to string
	stringEndpoint := strings.Join(endpoints, ",")

	var secretCredentials *corev1.Secret

	switch serviceFlavor.Category {
	case string(consts.Database):
		// Create the secret
		secretCredentials = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "credentials-" + contract.Name,
			},
			StringData: map[string]string{
				"endpoints": stringEndpoint,
				"username":  configurationDataMap["username"].(string),
				"password":  configurationDataMap["password"].(string),
				"database":  configurationDataMap["database"].(string),
			},
		}
	case string(consts.MessageQueue):
		// Create the secret
		secretCredentials = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "credentials-" + contract.Name,
			},
			StringData: map[string]string{
				"endpoints": stringEndpoint,
				"username":  configurationDataMap["username"].(string),
				"password":  configurationDataMap["password"].(string),
			},
		}
	// TODO (Service): Add more service categories here	for secret creation
	default:
		klog.Infof("Service category %s not supported", serviceFlavor.Category)
		return nil, fmt.Errorf("service category %s not supported", serviceFlavor.Category)
	}

	// Set labels
	if secretCredentials.Labels == nil {
		secretCredentials.Labels = make(map[string]string)
	}

	secretCredentials.Labels[consts.FluidosServiceCredentials] = "true"

	return secretCredentials, nil
}

// ForgeKnownCluster creates a KnownCluster from cluster ID and IP address.
func ForgeKnownCluster(id, address string) *networkv1alpha1.KnownCluster {
	return &networkv1alpha1.KnownCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namings.ForgeKnownClusterName(id),
			Namespace: flags.FluidosNamespace,
		},
		Spec: networkv1alpha1.KnownClusterSpec{
			Address: address,
		},
		Status: networkv1alpha1.KnownClusterStatus{
			ExpirationTime: tools.GetExpirationTime(0, 0, 10),
			LastUpdateTime: tools.GetTimeNow(),
		},
	}
}
