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

package rearmanager

import (
	"context"

	//admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
)

type Validator struct {
	client  client.Client
	decoder *admission.Decoder
}

func NewValidator(client client.Client) *Validator {
	return &Validator{client: client, decoder: admission.NewDecoder(runtime.NewScheme())}
}

func (v *Validator) Handle(ctx context.Context, req admission.Request) admission.Response {

	return admission.Allowed("allowed")
}

func (v *Validator) HandleCreate(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("allowed")
}

func (v *Validator) HandleDelete(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("allowed")
}

func (v *Validator) HandleUpdate(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("allowed")
}

/* func (v *Validator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
} */

func (v *Validator) DecodeAllocation(obj runtime.RawExtension) (pc *nodecorev1alpha1.Allocation, err error) {
	pc = &nodecorev1alpha1.Allocation{}
	err = v.decoder.DecodeRaw(obj, pc)
	return
}
