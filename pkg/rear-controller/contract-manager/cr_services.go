package contractmanager

import (
	"context"

	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	reservationv1alpha1 "fluidos.eu/node/api/reservation/v1alpha1"
	"fluidos.eu/node/pkg/utils/flags"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// getAllFlavoursCRsSpec returns all the Flavour CRs Spec in the cluster
func getAllFlavoursCRsSpec(cl client.Client) ([]nodecorev1alpha1.FlavourSpec, error) {
	flavourList, err := getAllFlavoursCRs(cl)
	if err != nil {
		klog.Errorf("Error getting all the Flavour CRs: %s", err)
		return nil, err
	}

	var flavoursSpec []nodecorev1alpha1.FlavourSpec

	// Get the Spec of each Flavour
	for _, flavour := range flavourList {
		flavoursSpec = append(flavoursSpec, flavour.Spec)
	}

	return flavoursSpec, nil
}

// getAllFlavoursCRs returns all the Flavour CRs in the cluster
func getAllFlavoursCRs(cl client.Client) ([]nodecorev1alpha1.Flavour, error) {

	var flavourList nodecorev1alpha1.FlavourList

	// List all Flavour CRs
	err := cl.List(context.Background(), &flavourList)
	if err != nil {
		return nil, err
	}

	var flavours []nodecorev1alpha1.Flavour
	flavours = append(flavours, flavourList.Items...)

	return flavours, nil
}

// getFlavourByID returns the Flavour CR in the cluster that matches the flavourID
func getFlavourByID(flavourID string, cl client.Client) (*nodecorev1alpha1.FlavourSpec, error) {
	flavoursSpec, err := getAllFlavoursCRsSpec(cl)
	if err != nil {
		klog.Errorf("Error getting all the Flavour CRs: %s", err)
		return nil, err
	}

	// Get the Flavour that matches the flavourID
	for _, flavour := range flavoursSpec {
		if flavour.FlavourID == flavourID {
			return &flavour, nil
		}
	}

	return nil, nil
}

// getFlavourByIDCR returns the entire Flavour CR (not only spec) in the cluster that matches the flavourID
func getFlavourByIDCR(flavourID string, cl client.Client) (*nodecorev1alpha1.Flavour, error) {
	flavoursSpec, err := getAllFlavoursCRs(cl)
	if err != nil {
		klog.Errorf("Error getting all the Flavour CRs: %s", err)
		return nil, err
	}

	// Get the Flavour that matches the flavourID
	for _, flavour := range flavoursSpec {
		if flavour.Spec.FlavourID == flavourID {
			return &flavour, nil
		}
	}

	return nil, nil
}

// getFlavoursBySelector returns the Flavour CRs in the cluster that match the selector
func getFlavoursBySelector(flavoursSpec []nodecorev1alpha1.FlavourSpec, selector Selector) ([]nodecorev1alpha1.FlavourSpec, error) {
	var flavours []nodecorev1alpha1.FlavourSpec

	// Get the Flavours that match the selector
	for _, flavour := range flavoursSpec {
		cpuInt, _ := flavour.Characteristics.Cpu.AsInt64()
		ramInt, _ := flavour.Characteristics.Memory.AsInt64()
		if string(flavour.Type) == selector.FlavourType &&
			int(cpuInt) == selector.CPU &&
			int(ramInt) == selector.RAM {
			flavours = append(flavours, flavour)
		}
	}

	return flavours, nil
}

// createContractCustomResourceSeller creates a Contract CR
func createContractCustomResourceSeller(ctx context.Context, flavour nodecorev1alpha1.Flavour, buyerID string) *reservationv1alpha1.Contract {
	contract := &reservationv1alpha1.Contract{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "contract-" + flavour.Spec.FlavourID,
			Namespace: flags.DISCOVERY_DEFAULT_NAMESPACE,
		},
		Spec: reservationv1alpha1.ContractSpec{
			Flavour: flavour,
			Buyer: nodecorev1alpha1.NodeIdentity{
				NodeID: buyerID,
			},
			Seller: flavour.Spec.Owner,
		},
		Status: reservationv1alpha1.ContractStatus{
			Phase: nodecorev1alpha1.PhaseStatus{
				Phase: nodecorev1alpha1.PhaseSolved,
			},
		},
	}

	return contract
}

// createContract creates a Contract CR in the cluster
func createContract(contract *reservationv1alpha1.Contract, cl client.Client) error {
	return cl.Create(context.Background(), contract)
}
