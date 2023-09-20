/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package contractmanager

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	advertisementv1alpha1 "fluidos.eu/node/api/advertisement/v1alpha1"
	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	reservationv1alpha1 "fluidos.eu/node/api/reservation/v1alpha1"
	"fluidos.eu/node/pkg/rear-controller/gateway"
	"fluidos.eu/node/pkg/utils/namings"
	"fluidos.eu/node/pkg/utils/resourceforge"
	"fluidos.eu/node/pkg/utils/tools"
)

// ReservationReconciler reconciles a Reservation object
type ReservationReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Gateway *gateway.Gateway
}

//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=reservations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=reservations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=reservations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *ReservationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Info("Reconciling Reservation")
	log := ctrl.LoggerFrom(ctx, "reservation", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	// var contract *reservationv1alpha1.Contract

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

		if err := r.updateReservationStatus(ctx, &reservation); err != nil {
			klog.Errorf("Error when updating Reservation %s status before reconcile: %s", req.NamespacedName, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if reservation.Spec.Reserve {
		reservePhase := reservation.Status.ReservePhase
		switch reservePhase {
		case nodecorev1alpha1.PhaseRunning:
			klog.Infof("Reservation %s: Reserve phase running", reservation.Name)
			flavourID := namings.RetrieveFlavourNameFromPC(reservation.Spec.PeeringCandidate.Name)
			res, err := r.Gateway.ReserveFlavour(ctx, &reservation, flavourID)
			if err != nil {
				klog.Errorf("Error when reserving flavour for Reservation %s: %s", req.NamespacedName, err)
				reservation.SetReserveStatus(nodecorev1alpha1.PhaseFailed)
				reservation.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation failed: error when reserving flavour")
				if err := r.updateReservationStatus(ctx, &reservation); err != nil {
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
			if err := r.updateReservationStatus(ctx, &reservation); err != nil {
				klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil

		case nodecorev1alpha1.PhaseSolved:
			klog.Infof("Reserve %s solved", reservation.Name)
		case nodecorev1alpha1.PhaseFailed:
			klog.Infof("Reserve %s failed", reservation.Name)
			return ctrl.Result{}, nil
		case nodecorev1alpha1.PhaseIdle:
			klog.Infof("Reserve %s idle", reservation.Name)
			reservation.SetReserveStatus(nodecorev1alpha1.PhaseRunning)
			if err := r.updateReservationStatus(ctx, &reservation); err != nil {
				klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		default:
			klog.Infof("Reserve %s unknown phase", reservation.Name)
			reservation.SetReserveStatus(nodecorev1alpha1.PhaseIdle)
			if err := r.updateReservationStatus(ctx, &reservation); err != nil {
				klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

	}

	if reservation.Spec.Purchase && reservation.Status.ReservePhase == nodecorev1alpha1.PhaseSolved {
		purchasePhase := reservation.Status.PurchasePhase
		switch purchasePhase {
		case nodecorev1alpha1.PhaseIdle:
			klog.Infof("Purchase phase for the reservation %s idle, starting...", reservation.Name)
			reservation.SetPurchaseStatus(nodecorev1alpha1.PhaseRunning)
			if err := r.updateReservationStatus(ctx, &reservation); err != nil {
				klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		case nodecorev1alpha1.PhaseRunning:
			if reservation.Status.TransactionID == "" {
				klog.Infof("TransactionID not set for Reservation %s", reservation.Name)
				reservation.SetPurchaseStatus(nodecorev1alpha1.PhaseFailed)
				reservation.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation failed: TransactionID not set")
				if err := r.updateReservationStatus(ctx, &reservation); err != nil {
					klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, nil
			}

			transactionID := reservation.Status.TransactionID
			resPurchase, err := r.Gateway.PurchaseFlavour(ctx, transactionID)
			if err != nil {
				klog.Errorf("Error when purchasing flavour for Reservation %s: %s", req.NamespacedName, err)
				reservation.SetPurchaseStatus(nodecorev1alpha1.PhaseFailed)
				reservation.SetPhase(nodecorev1alpha1.PhaseFailed, "Reservation failed: error when purchasing flavour")
				if err := r.updateReservationStatus(ctx, &reservation); err != nil {
					klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, err
			}

			klog.Infof("Purchase completed with status %s", resPurchase.Status)

			reservation.SetPurchaseStatus(nodecorev1alpha1.PhaseRunning)

			if err := r.Update(ctx, &reservation); err != nil {
				klog.Errorf("Error when updating Reservation %s: %s", reservation.Name, err)
				return ctrl.Result{}, err
			}

			klog.Infof("Reservation %s updated", reservation.Name)

			// Create a contract CR now that the reservation is solved
			contract := resourceforge.ForgeContractFromObj(resPurchase.Contract)
			err = r.Create(ctx, contract)
			if errors.IsAlreadyExists(err) {
				klog.Errorf("Error when creating Contract %s: %s", contract.Name, err)
			} else if err != nil {
				klog.Errorf("Error when creating Contract %s: %s", contract.Name, err)
				return ctrl.Result{}, err
			}
			klog.Infof("Contract %s created", contract.Name)

			reservation.SetPurchaseStatus(nodecorev1alpha1.PhaseSolved)
			reservation.Status.Contract = nodecorev1alpha1.GenericRef{
				Name:      contract.Name,
				Namespace: contract.Namespace,
			}
			reservation.SetPhase(nodecorev1alpha1.PhaseSolved, "Reservation solved")

			if err := r.updateReservationStatus(ctx, &reservation); err != nil {
				klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil

		case nodecorev1alpha1.PhaseFailed:
			klog.Infof("Purchase phase for the reservation %s failed", reservation.Name)
			return ctrl.Result{}, nil

		case nodecorev1alpha1.PhaseSolved:
			klog.Infof("Purchase phase for the reservation %s solved", reservation.Name)
		default:
			klog.Infof("Purchase phase for the reservation %s unknown", reservation.Name)
			reservation.SetPurchaseStatus(nodecorev1alpha1.PhaseIdle)
			if err := r.updateReservationStatus(ctx, &reservation); err != nil {
				klog.Errorf("Error when updating Reservation %s status: %s", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

// updateSolverStatus updates the status of the discovery
func (r *ReservationReconciler) updateReservationStatus(ctx context.Context, reservation *reservationv1alpha1.Reservation) error {
	return r.Status().Update(ctx, reservation)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReservationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&reservationv1alpha1.Reservation{}).
		Complete(r)
}
