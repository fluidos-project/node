package rearmanager

import (
	"context"

	//admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
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
