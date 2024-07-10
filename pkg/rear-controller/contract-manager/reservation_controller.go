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

package contractmanager

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	advertisementv1alpha1 "github.com/fluidos-project/node/apis/advertisement/v1alpha1"
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	"github.com/fluidos-project/node/pkg/rear-controller/gateway"
	"github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/namings"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
	"github.com/fluidos-project/node/pkg/utils/tools"
)

// ReservationReconciler reconciles a Reservation object.
type ReservationReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Gateway *gateway.Gateway
}

// clusterRole
//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=reservations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=reservations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=reservations/finalizers,verbs=update
//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=contracts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=contracts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=contracts/finalizers,verbs=update
//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=transactions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=transactions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=transactions/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *ReservationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "reservation", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	var reservation reservationv1alpha1.Reservation
	if err := r.Get(ctx, req.NamespacedName, &reservation); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting Reservation %s before reconcile: %s", req.NamespacedName, err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("Reservation %s not found, probably deleted", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	var peeringCandidate advertisementv1alpha1.PeeringCandidate
	if err := r.Get(ctx, client.ObjectKey{
		Name:      reservation.Spec.PeeringCandidate.Name,
		Namespace: reservation.Spec.PeeringCandidate.Namespace,
	}, &peeringCandidate); err != nil {
		klog.Errorf("Error when getting PeeringCandidate %s before reconcile: %s", req.NamespacedName, err)
		reservation.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation failed: error when getting PeeringCandidate")
		if err := r.updateReservationStatus(ctx, &reservation); err != nil {
			klog.Errorf("Error when updating Reservation %s status before reconcile: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	if checkInitialStatus(&reservation) {
		if err := r.updateReservationStatus(ctx, &reservation); err != nil {
			klog.Errorf("Error when updating Reservation %s status before reconcile: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if reservation.Status.Phase.Phase == nodecorev1alpha1.PhaseSolved ||
		reservation.Status.Phase.Phase == nodecorev1alpha1.PhaseTimeout ||
		reservation.Status.Phase.Phase == nodecorev1alpha1.PhaseFailed {
		return ctrl.Result{}, nil
	}

	if reservation.Spec.Reserve {
		if reservation.Status.ReservePhase != nodecorev1alpha1.PhaseSolved {
			return r.handleReserve(ctx, req, &reservation, &peeringCandidate)
		}
		klog.Infof("Reservation %s: Reserve phase solved", reservation.Name)
	}

	if reservation.Spec.Purchase && reservation.Status.ReservePhase == nodecorev1alpha1.PhaseSolved {
		if reservation.Status.PurchasePhase != nodecorev1alpha1.PhaseSolved {
			return r.handlePurchase(ctx, req, &reservation)
		}
		klog.Infof("Reservation %s: Purchase phase solved", reservation.Name)
	}

	return ctrl.Result{}, nil
}

// updateSolverStatus updates the status of the discovery.
func (r *ReservationReconciler) updateReservationStatus(ctx context.Context, reservation *reservationv1alpha1.Reservation) error {
	return r.Status().Update(ctx, reservation)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReservationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&reservationv1alpha1.Reservation{}).
		Complete(r)
}

func checkInitialStatus(reservation *reservationv1alpha1.Reservation) bool {
	if reservation.Status.Phase.Phase != nodecorev1alpha1.PhaseSolved &&
		reservation.Status.Phase.Phase != nodecorev1alpha1.PhaseTimeout &&
		reservation.Status.Phase.Phase != nodecorev1alpha1.PhaseFailed &&
		reservation.Status.Phase.Phase != nodecorev1alpha1.PhaseRunning &&
		reservation.Status.Phase.Phase != nodecorev1alpha1.PhaseIdle {
		klog.Infof("Reservation %s started", reservation.Name)
		reservation.Status.Phase.StartTime = tools.GetTimeNow()
		reservation.SetPhase(nodecorev1alpha1.PhaseRunning, "Reservation started")
		reservation.SetReserveStatus(nodecorev1alpha1.PhaseIdle)
		reservation.SetPurchaseStatus(nodecorev1alpha1.PhaseIdle)

		return true
	}
	return false
}

func (r *ReservationReconciler) handleReserve(ctx context.Context,
	req ctrl.Request, reservation *reservationv1alpha1.Reservation, peeringCandidate *advertisementv1alpha1.PeeringCandidate) (ctrl.Result, error) {
	reservePhase := reservation.Status.ReservePhase
	switch reservePhase {
	case nodecorev1alpha1.PhaseRunning:
		klog.Infof("Reservation %s: Reserve phase running", reservation.Name)

		// Check if the peering candidate is available
		if !peeringCandidate.Spec.Available {
			klog.Infof("PeeringCandidate %s not available", peeringCandidate.Name)
			reservation.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation failed: PeeringCandidate not available")
			if err := r.updateReservationStatus(ctx, reservation); err != nil {
				klog.Errorf("Error when updating Reservation %s status before reconcile: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		// Set the peering candidate as not available
		peeringCandidate.Spec.Available = false
		peeringCandidate.Spec.SolverID = reservation.Spec.SolverID
		if err := r.Update(ctx, peeringCandidate); err != nil {
			klog.Errorf("Error when updating PeeringCandidate %s status before reconcile: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		peeringCandidate.Status.LastUpdateTime = tools.GetTimeNow()

		if err := r.Status().Update(ctx, peeringCandidate); err != nil {
			klog.Errorf("Error when updating PeeringCandidate %s status before reconcile: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}

		flavorID := namings.RetrieveFlavorNameFromPC(reservation.Spec.PeeringCandidate.Name)
		res, err := r.Gateway.ReserveFlavor(ctx, reservation, flavorID)
		if err != nil {
			if res != nil {
				klog.Infof("Transaction is non correctly set, Retrying...")
				return ctrl.Result{Requeue: true}, nil
			}
			klog.Errorf("Error when reserving flavor for Reservation %s: %s", req.NamespacedName, err)

			// Set the peering candidate as available again
			peeringCandidate.Spec.Available = true
			peeringCandidate.Spec.SolverID = ""
			if err := r.Update(ctx, peeringCandidate); err != nil {
				klog.Errorf("Error when updating PeeringCandidate %s status before reconcile: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			peeringCandidate.Status.LastUpdateTime = tools.GetTimeNow()

			if err := r.Status().Update(ctx, peeringCandidate); err != nil {
				klog.Errorf("Error when updating PeeringCandidate %s status before reconcile: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}

			// Set the reservation as failed
			reservation.SetReserveStatus(nodecorev1alpha1.PhaseFailed)
			reservation.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation failed: error when reserving flavor")
			if err := r.updateReservationStatus(ctx, reservation); err != nil {
				klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}

		klog.Infof("Transaction: %v", res)

		// Create a Transaction CR starting from the transaction object
		transaction := resourceforge.ForgeTransactionFromObj(res)

		if err := r.Create(ctx, transaction); err != nil {
			klog.Errorf("Error when creating Transaction %s: %s", transaction.Name, err)
			return ctrl.Result{}, err
		}

		klog.Infof("Transaction %s created", transaction.Name)
		reservation.Status.TransactionID = res.TransactionID
		reservation.SetReserveStatus(nodecorev1alpha1.PhaseSolved)

		// Update the status for reconcile
		if err := r.updateReservationStatus(ctx, reservation); err != nil {
			klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	case nodecorev1alpha1.PhaseFailed:
		klog.Infof("Reserve %s failed", reservation.Name)
		reservation.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation failed during the 'Reserve' phase")
		if err := r.updateReservationStatus(ctx, reservation); err != nil {
			klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	case nodecorev1alpha1.PhaseIdle:
		klog.Infof("Reserve %s idle", reservation.Name)
		reservation.SetReserveStatus(nodecorev1alpha1.PhaseRunning)
		if err := r.updateReservationStatus(ctx, reservation); err != nil {
			klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	default:
		klog.Infof("Reserve %s unknown phase", reservation.Name)
		reservation.SetReserveStatus(nodecorev1alpha1.PhaseIdle)
		if err := r.updateReservationStatus(ctx, reservation); err != nil {
			klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
}

func (r *ReservationReconciler) handlePurchase(ctx context.Context,
	req ctrl.Request, reservation *reservationv1alpha1.Reservation) (ctrl.Result, error) {
	purchasePhase := reservation.Status.PurchasePhase

	// Get the flavor from the peering candidate of the reservation
	var peeringCandidate advertisementv1alpha1.PeeringCandidate
	if err := r.Get(ctx, client.ObjectKey{
		Name:      reservation.Spec.PeeringCandidate.Name,
		Namespace: reservation.Spec.PeeringCandidate.Namespace,
	}, &peeringCandidate); err != nil {
		klog.Errorf("Error when getting PeeringCandidate %s before reconcile: %s", req.NamespacedName, err)
		reservation.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation failed: error when getting PeeringCandidate")
		if err := r.updateReservationStatus(ctx, reservation); err != nil {
			klog.Errorf("Error when updating Reservation %s status before reconcile: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	// Get the flavor type identifier from the flavor
	flavorTypeIdentifier, _, err := nodecorev1alpha1.ParseFlavorType(&peeringCandidate.Spec.Flavor)
	if err != nil {
		klog.Errorf("Error when parsing Flavor %s type: %s", req.NamespacedName, err)
		reservation.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation failed: error when parsing Flavor type")
		if err := r.updateReservationStatus(ctx, reservation); err != nil {
			klog.Errorf("Error when updating Reservation %s status before reconcile: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	switch purchasePhase {
	case nodecorev1alpha1.PhaseIdle:
		klog.Infof("Purchase phase for the reservation %s idle, starting...", reservation.Name)
		reservation.SetPurchaseStatus(nodecorev1alpha1.PhaseRunning)
		if err := r.updateReservationStatus(ctx, reservation); err != nil {
			klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	case nodecorev1alpha1.PhaseRunning:
		if reservation.Status.TransactionID == "" {
			klog.Infof("TransactionID not set for Reservation %s", reservation.Name)
			reservation.SetPurchaseStatus(nodecorev1alpha1.PhaseFailed)
			reservation.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation failed: TransactionID not set")
			if err := r.updateReservationStatus(ctx, reservation); err != nil {
				klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		transactionID := reservation.Status.TransactionID
		var contract *models.Contract
		var err error

		// Based on the flavorTypeIdentifier, purchase flavor action may need buyer credentials
		switch flavorTypeIdentifier {
		case nodecorev1alpha1.TypeK8Slice:
			contract, err = r.Gateway.PurchaseFlavor(ctx, transactionID, reservation.Spec.Seller, nil)
			// TODO: Implement other flavor types if purchase request needs buyer credentials
		default:
			klog.Errorf("Flavor type %s not supported", flavorTypeIdentifier)
		}

		if err != nil {
			klog.Errorf("Error when purchasing flavor for Reservation %s: %s", req.NamespacedName, err)

			// Set the PeerCandidate as available again
			var peeringCandidate advertisementv1alpha1.PeeringCandidate
			if err := r.Get(ctx, client.ObjectKey{
				Name:      reservation.Spec.PeeringCandidate.Name,
				Namespace: reservation.Spec.PeeringCandidate.Namespace,
			}, &peeringCandidate); err != nil {
				klog.Errorf("Error when getting PeeringCandidate %s before reconcile: %s", req.NamespacedName, err)
				reservation.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation failed: error when getting PeeringCandidate")
				if err := r.updateReservationStatus(ctx, reservation); err != nil {
					klog.Errorf("Error when updating Reservation %s status before reconcile: %s", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, err
			}

			peeringCandidate.Spec.Available = true
			peeringCandidate.Spec.SolverID = ""
			if err := r.Update(ctx, &peeringCandidate); err != nil {
				klog.Errorf("Error when updating PeeringCandidate %s status before reconcile: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			peeringCandidate.Status.LastUpdateTime = tools.GetTimeNow()

			if err := r.Status().Update(ctx, &peeringCandidate); err != nil {
				klog.Errorf("Error when updating PeeringCandidate %s status before reconcile: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}

			// Set the reservation as failed
			reservation.SetPurchaseStatus(nodecorev1alpha1.PhaseFailed)
			reservation.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation failed: error when purchasing flavor")
			if err := r.updateReservationStatus(ctx, reservation); err != nil {
				klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}

		klog.Infof("Purchase completed, got contract ID %s", contract.ContractID)

		// Create a contract CR now that the reservation is solved
		contractCR, err := resourceforge.ForgeContractFromObj(contract)
		if err != nil {
			klog.Errorf("Error when forging Contract %s: %s", contractCR.Name, err)
			return ctrl.Result{}, err
		}
		err = r.Create(ctx, contractCR)
		if errors.IsAlreadyExists(err) {
			klog.Errorf("Error when creating Contract %s: %s", contractCR.Name, err)
		} else if err != nil {
			klog.Errorf("Error when creating Contract %s: %s", contractCR.Name, err)
			return ctrl.Result{}, err
		}
		klog.Infof("Contract %s created", contractCR.Name)

		reservation.Status.Contract = nodecorev1alpha1.GenericRef{
			Name:      contractCR.Name,
			Namespace: contractCR.Namespace,
		}
		reservation.SetPurchaseStatus(nodecorev1alpha1.PhaseSolved)
		reservation.SetPhase(nodecorev1alpha1.PhaseSolved, "Reservation solved")

		if err := r.updateReservationStatus(ctx, reservation); err != nil {
			klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	case nodecorev1alpha1.PhaseFailed:
		klog.Infof("Purchase phase for the reservation %s failed", reservation.Name)
		reservation.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation failed during the 'Purchase' phase")
		if err := r.updateReservationStatus(ctx, reservation); err != nil {
			klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	default:
		klog.Infof("Purchase phase for the reservation %s unknown", reservation.Name)
		reservation.SetPurchaseStatus(nodecorev1alpha1.PhaseIdle)
		if err := r.updateReservationStatus(ctx, reservation); err != nil {
			klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
}
