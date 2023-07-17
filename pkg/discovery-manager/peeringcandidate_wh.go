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

	advertisementv1alpha1 "fluidos.eu/node/api/advertisement/v1alpha1"
)

//+kubebuilder:webhook:path=/validate/peeringcandidate,mutating=false,failurePolicy=ignore,groups=advertisement.node.fluidos.io,resources=peeringcandidates,verbs=create;update;delete,versions=v1alpha1,name=pc.validate.fluidos.eu,sideEffects=None,admissionReviewVersions={v1,v1beta1}


type PCValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

func NewPCValidator(client client.Client) *PCValidator {
	return &PCValidator{client: client, decoder: admission.NewDecoder(runtime.NewScheme())}
}

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

func (v *PCValidator) HandleCreate(ctx context.Context, req admission.Request) admission.Response {
	pc, err := v.DecodePeeringCandidate(req.Object)
	if err != nil {
		klog.Errorf("Failed to decode peering candidate: %v", err)
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to decode peering candidate: %v", err))
	}

	if pc.Spec.Reserved && pc.Spec.SolverID == "" {
		return admission.Denied("If peering candidate is reserved, solver ID must be set")
	}

	if !pc.Spec.Reserved && pc.Spec.SolverID != "" {
		return admission.Denied("If peering candidate is not reserved, solver ID must not be set")
	}

	return admission.Allowed("")
}

func (v *PCValidator) HandleDelete(ctx context.Context, req admission.Request) admission.Response {
	// Here we could check if the peering candidate is reserved and if so, we need to check if the solver ID matches the one of the solver that is deleting the peering candidate
	// or if the solver ID is empty, we need to check if there is a Contract that is using this peering candidate
	// Maybe this is not the right logic but it need to be discussed and implemented
	return admission.Allowed("")
}

func (v *PCValidator) HandleUpdate(ctx context.Context, req admission.Request) admission.Response {
	pc, err := v.DecodePeeringCandidate(req.Object)
	if err != nil {
		klog.Errorf("Failed to decode peering candidate: %v", err)
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to decode peering candidate: %v", err))
	}

	pcOld, err := v.DecodePeeringCandidate(req.OldObject)
	if err != nil {
		klog.Errorf("Failed to decode peering old candidate: %v", err)
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to decode peering old candidate: %v", err))
	}

	// PC can be updated only if:
	// - both Reserved flag and SolverID are not set (i.e. it is not reserved)
	// - both Reserved flag and SolverID are set and you want to clear both in the same time

	if !pcOld.Spec.Reserved && pcOld.Spec.SolverID == "" {
		return admission.Allowed("")
	}

	if pcOld.Spec.Reserved && pcOld.Spec.SolverID != "" && !pc.Spec.Reserved && pc.Spec.SolverID == "" {
		return admission.Allowed("")
	}

	return admission.Denied("Peering candidate can be updated only if it is not reserved or if both Reserved flag and SolverID are set and you want to clear both in the same time")
}

/* func (v *PCValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
} */

func (v *PCValidator) DecodePeeringCandidate(obj runtime.RawExtension) (pc *advertisementv1alpha1.PeeringCandidate, err error) {
	pc = &advertisementv1alpha1.PeeringCandidate{}
	err = v.decoder.DecodeRaw(obj, pc)
	return
}
