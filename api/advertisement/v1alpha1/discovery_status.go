package v1alpha1

import (
	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
	"fluidos.eu/node/pkg/utils/tools"
)

// SetPhase sets the phase of the discovery
func (d *Discovery) SetPhase(phase nodecorev1alpha1.Phase, msg string) {
	d.Status.Phase.Phase = phase
	d.Status.Phase.LastChangeTime = tools.GetTimeNow()
	d.Status.Phase.Message = msg
}
