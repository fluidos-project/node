package v1alpha1

import (
	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	"fluidos.eu/node/pkg/utils/tools"
)

// SetPhase sets the phase of the discovery
func (r *Reservation) SetPhase(phase nodecorev1alpha1.Phase, msg string) {
	r.Status.Phase.Phase = phase
	r.Status.Phase.LastChangeTime = tools.GetTimeNow()
	r.Status.Phase.Message = msg
}

// SetReserveStatus sets the status of the reserve (if it is a reserve)
func (r *Reservation) SetReserveStatus(status nodecorev1alpha1.Phase) {
	r.Status.ReservePhase = status
}

// SetPurchaseStatus sets the status of the purchase (if it is a purchase)
func (r *Reservation) SetPurchaseStatus(status nodecorev1alpha1.Phase) {
	r.Status.PurchasePhase = status
}
