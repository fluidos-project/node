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
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
)

//nolint:lll // This is a long line
// clusterRole
//+kubebuilder:webhook:path=/validate/solver,mutating=false,failurePolicy=ignore,groups=nodecore.node.fluidos.io,resources=solvers,verbs=create;update;delete,versions=v1alpha1,name=solver.validate.fluidos.eu,sideEffects=None,admissionReviewVersions={v1,v1beta1}

// SolverValidator is the Solver validator.
type SolverValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewSolverValidator creates a new SolverValidator.
func NewSolverValidator(c client.Client) *SolverValidator {
	return &SolverValidator{client: c, decoder: admission.NewDecoder(runtime.NewScheme())}
}

// Handle manages the validation of the Solver.
func (v *SolverValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case admissionv1.Create:
		return v.HandleCreate(ctx, req)
	default:
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unsupported operation %q", req.Operation))
	}
}

// HandleCreate manages the validation of the Solver creation.
func (v *SolverValidator) HandleCreate(_ context.Context, req admission.Request) admission.Response {
	s, err := v.DecodeSolver(req.Object)
	if err != nil {
		klog.Errorf("Failed to decode solver: %v", err)
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to decode solver: %w", err))
	}
	if !s.Spec.ReserveAndBuy && s.Spec.EnstablishPeering {
		return admission.Denied("solver cannot establish peering without reserve and buy")
	}
	return admission.Allowed("solver is valid")
}

// DecodeSolver decodes the Solver.
func (v *SolverValidator) DecodeSolver(obj runtime.RawExtension) (s *nodecorev1alpha1.Solver, err error) {
	s = &nodecorev1alpha1.Solver{}
	err = v.decoder.DecodeRaw(obj, s)
	return
}
