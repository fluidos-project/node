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

package v1alpha1

import (
	"github.com/fluidos-project/node/pkg/utils/tools"
)

// SetPhase sets the phase of the solver.
func (r *Solver) SetPhase(phase Phase, msg string) {
	t := tools.GetTimeNow()
	r.Status.SolverPhase.Phase = phase
	r.Status.SolverPhase.LastChangeTime = t
	r.Status.SolverPhase.Message = msg
	r.Status.SolverPhase.EndTime = t
}

// SetPeeringStatus sets the Peering phase of the solver.
func (r *Solver) SetPeeringStatus(phase Phase) {
	r.Status.Peering = phase
	r.Status.SolverPhase.LastChangeTime = tools.GetTimeNow()
}

// SetReserveAndBuyStatus sets the ReserveAndBuy phase of the solver.
func (r *Solver) SetReserveAndBuyStatus(phase Phase) {
	r.Status.ReserveAndBuy = phase
	r.Status.SolverPhase.LastChangeTime = tools.GetTimeNow()
}

// SetFindCandidateStatus sets the FindCandidate phase of the solver.
func (r *Solver) SetFindCandidateStatus(phase Phase) {
	r.Status.FindCandidate = phase
	r.Status.SolverPhase.LastChangeTime = tools.GetTimeNow()
}

// SetDiscoveryStatus sets the discovery phase of the solver.
func (r *Solver) SetDiscoveryStatus(phase Phase) {
	r.Status.DiscoveryPhase = phase
	r.Status.SolverPhase.LastChangeTime = tools.GetTimeNow()
}

// SetReservationStatus sets the reservation phase of the solver.
func (r *Solver) SetReservationStatus(phase Phase) {
	r.Status.ReservationPhase = phase
	r.Status.SolverPhase.LastChangeTime = tools.GetTimeNow()
}
