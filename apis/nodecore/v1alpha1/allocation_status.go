// Copyright 2022-2025 FLUIDOS Project
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

// SetResourceRef sets the resource reference of the allocation.
func (allocation *Allocation) SetResourceRef(resourceRef GenericRef) {
	allocation.Status.ResourceRef = resourceRef
	allocation.Status.LastUpdateTime = tools.GetTimeNow()
}
