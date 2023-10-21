// Copyright 2022-2023 FLUIDOS Project
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

import "github.com/fluidos-project/node/pkg/utils/tools"

// SetStatus sets the status of the allocation.
func (allocation *Allocation) SetStatus(status Status, msg string) {
	allocation.Status.Status = status
	allocation.Status.LastUpdateTime = tools.GetTimeNow()
	allocation.Status.Message = msg
}

/*
// SetPurchasePhase sets the ReserveAndBuy phase of the solver
func (allocation *Allocation) SetReserveAndBuyStatus(phase Phase) {
	solver.Status.ReserveAndBuy = phase
	solver.Status.SolverPhase.LastChangeTime = tools.GetTimeNow()
}

// SetFindCandidateStatus sets the FindCandidate phase of the solver
func (allocation *Allocation) SetFindCandidateStatus(phase Phase) {
	solver.Status.FindCandidate = phase
	solver.Status.SolverPhase.LastChangeTime = tools.GetTimeNow()
}

// SetDiscoveryStatus sets the discovery phase of the solver
func (allocation *Allocation) SetDiscoveryStatus(phase Phase) {
	solver.Status.DiscoveryPhase = phase
	solver.Status.SolverPhase.LastChangeTime = tools.GetTimeNow()
}

// SetReservationStatus sets the reservation phase of the solver
func (allocation *Allocation) SetReservationStatus(phase Phase) {
	solver.Status.ReservationPhase = phase
	solver.Status.SolverPhase.LastChangeTime = tools.GetTimeNow()
}
*/
