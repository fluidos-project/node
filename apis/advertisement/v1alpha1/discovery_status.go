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
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/tools"
)

// SetPhase sets the phase of the discovery.
func (d *Discovery) SetPhase(phase nodecorev1alpha1.Phase, msg string) {
	d.Status.Phase.Phase = phase
	d.Status.Phase.LastChangeTime = tools.GetTimeNow()
	d.Status.Phase.Message = msg
}
