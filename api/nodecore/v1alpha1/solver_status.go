package v1alpha1

import (
	"fluidos.eu/node/pkg/utils/tools"
)

func (solver *Solver) SetPhase(phase Phase, msg string) {
	t := tools.GetTimeNow()
	solver.Status.SolverPhase.Phase = phase
	solver.Status.SolverPhase.LastChangeTime = t
	solver.Status.SolverPhase.Message = msg
	solver.Status.SolverPhase.EndTime = t
}

// SetPurchasePhase sets the ReserveAndBuy phase of the solver
func (solver *Solver) SetReserveAndBuyStatus(phase Phase) {
	solver.Status.ReserveAndBuy = phase
	solver.Status.SolverPhase.LastChangeTime = tools.GetTimeNow()
}

// SetFindCandidateStatus sets the FindCandidate phase of the solver
func (solver *Solver) SetFindCandidateStatus(phase Phase) {
	solver.Status.FindCandidate = phase
	solver.Status.SolverPhase.LastChangeTime = tools.GetTimeNow()
}

// SetDiscoveryStatus sets the discovery phase of the solver
func (solver *Solver) SetDiscoveryStatus(phase Phase) {
	solver.Status.DiscoveryPhase = phase
	solver.Status.SolverPhase.LastChangeTime = tools.GetTimeNow()
}

// SetReservationStatus sets the reservation phase of the solver
func (solver *Solver) SetReservationStatus(phase Phase) {
	solver.Status.ReservationPhase = phase
	solver.Status.SolverPhase.LastChangeTime = tools.GetTimeNow()
}
