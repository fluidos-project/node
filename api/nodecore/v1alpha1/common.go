package v1alpha1

const (
	//PhaseReady   Phase = "Ready"
	PhaseSolved  Phase = "Solved"
	PhaseFailed  Phase = "Failed"
	PhaseRunning Phase = "Running"
	PhaseIdle    Phase = "Idle"
	PhaseTimeout Phase = "Timeout"
)

// GenericRef represents a reference to a generic Kubernetes resource,
// and it is composed of the resource name and (optionally) its namespace.
type GenericRef struct {
	// The name of the resource to be referenced.
	Name string `json:"name"`

	// The namespace containing the resource to be referenced. It should be left
	// empty in case of cluster-wide resources.
	Namespace string `json:"namespace,omitempty"`
}

type NodeIdentity struct {
	Domain string `json:"domain"`
	NodeID string `json:"nodeID"`
	IP     string `json:"ip"`
}
