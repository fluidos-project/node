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

package rearmanager

import (
	"context"

	// admissionv1 "k8s.io/api/admission/v1".
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
)

// Validator is the allocation webhook validator.
type Validator struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewValidator creates a new allocation webhook validator.
func NewValidator(c client.Client) *Validator {
	return &Validator{client: c, decoder: admission.NewDecoder(runtime.NewScheme())}
}

// Handle manages the validation of the Allocation.
//
//nolint:gocritic // This function cannot be changed
func (v *Validator) Handle(ctx context.Context, req admission.Request) admission.Response {
	_ = ctx
	_ = req
	return admission.Allowed("allowed")
}

// HandleCreate manages the validation of the Allocation creation.
//
//nolint:gocritic // This function cannot be changed
func (v *Validator) HandleCreate(ctx context.Context, req admission.Request) admission.Response {
	_ = ctx
	_ = req
	return admission.Allowed("allowed")
}

// HandleDelete manages the validation of the Allocation deletion.
//
//nolint:gocritic // This function cannot be changed
func (v *Validator) HandleDelete(ctx context.Context, req admission.Request) admission.Response {
	_ = ctx
	_ = req
	return admission.Allowed("allowed")
}

// HandleUpdate manages the validation of the Allocation update.
//
//nolint:gocritic // This function cannot be changed
func (v *Validator) HandleUpdate(ctx context.Context, req admission.Request) admission.Response {
	_ = ctx
	_ = req
	return admission.Allowed("allowed")
}

// DecodeAllocation decodes the Allocation from the raw extension.
func (v *Validator) DecodeAllocation(obj runtime.RawExtension) (pc *nodecorev1alpha1.Allocation, err error) {
	pc = &nodecorev1alpha1.Allocation{}
	err = v.decoder.DecodeRaw(obj, pc)
	return
}
