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

// TODO: Transport this file to the apis section following the kubebuilder conventions

package discoverymanager

import (
	"context"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	advertisementv1alpha1 "github.com/fluidos-project/node/apis/advertisement/v1alpha1"
)

//nolint:lll // This is a long line
// clusterRole
//+kubebuilder:webhook:path=/validate/peeringcandidate,mutating=false,failurePolicy=ignore,groups=advertisement.node.fluidos.io,resources=peeringcandidates,verbs=create;update;delete,versions=v1alpha1,name=pc.validate.fluidos.eu,sideEffects=None,admissionReviewVersions={v1,v1beta1}

// PCValidator is the PeerinCandidate validator.
type PCValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewPCValidator creates a new PCValidator.
func NewPCValidator(c client.Client) *PCValidator {
	return &PCValidator{client: c, decoder: admission.NewDecoder(runtime.NewScheme())}
}

// Handle manages the validation of the PeeringCandidate.
//
//nolint:gocritic // This function cannot be changed
func (v *PCValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case admissionv1.Create:
		return v.HandleCreate(ctx, req)
	case admissionv1.Delete:
		return v.HandleDelete(ctx, req)
	case admissionv1.Update:
		return v.HandleUpdate(ctx, req)
	default:
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unsupported operation %q", req.Operation))
	}
}

// HandleCreate manages the validation of the PeeringCandidate creation.
//
//nolint:gocritic // This function cannot be changed
func (v *PCValidator) HandleCreate(_ context.Context, req admission.Request) admission.Response {
	pc, err := v.DecodePeeringCandidate(req.Object)
	if err != nil {
		klog.Errorf("Failed to decode peering candidate: %v", err)
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to decode peering candidate: %w", err))
	}

	if pc.Spec.Available && len(pc.Spec.InterestedSolverIDs) == 0 {
		return admission.Denied("Can't create a peering candidate wihout a triggering solver")
	}

	if !pc.Spec.Available {
		return admission.Denied("Can't create a peering candidate with Available flag set to false")
	}

	return admission.Allowed("")
}

// HandleDelete manages the validation of the PeeringCandidate deletion.
//
//nolint:gocritic // This function cannot be changed
func (v *PCValidator) HandleDelete(_ context.Context, req admission.Request) admission.Response {
	// Here we could check if the peering candidate is reserved and if so,we need to check if the solver ID
	// matches the one of the solver that is deleting the peering candidate
	// or if the solver ID is empty, we need to check if there is a Contract that is using this peering candidate
	// Maybe this is not the right logic but it need to be discussed and implemented
	_ = req
	return admission.Allowed("")
}

// HandleUpdate manages the validation of the PeeringCandidate update.
//
//nolint:gocritic // This function cannot be changed
func (v *PCValidator) HandleUpdate(_ context.Context, req admission.Request) admission.Response {
	pc, err := v.DecodePeeringCandidate(req.Object)
	if err != nil {
		klog.Errorf("Failed to decode peering candidate: %v", err)
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to decode peering candidate: %w", err))
	}

	pcOld, err := v.DecodePeeringCandidate(req.OldObject)
	if err != nil {
		klog.Errorf("Failed to decode peering old candidate: %v", err)
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to decode peering old candidate: %w", err))
	}

	if !pcOld.Spec.Available && !pc.Spec.Available {
		return admission.Denied("Peering candidate can be updated if Available flag is changed from false to true")
	}

	//nolint:lll // This is a long line
	return admission.Allowed("")
}

// DecodePeeringCandidate decodes the PeeringCandidate.
func (v *PCValidator) DecodePeeringCandidate(obj runtime.RawExtension) (pc *advertisementv1alpha1.PeeringCandidate, err error) {
	pc = &advertisementv1alpha1.PeeringCandidate{}
	err = v.decoder.DecodeRaw(obj, pc)
	return
}
