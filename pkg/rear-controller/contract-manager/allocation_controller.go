package contractmanager

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodecorev1alpha1 "fluidos.eu/node/api/nodecore/v1alpha1"
)

// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations/finalizers,verbs=update

// AllocationReconciler reconciles a Allocation object
type AllocationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *AllocationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "allocation", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	var allocation nodecorev1alpha1.Allocation
	if err := r.Get(ctx, req.NamespacedName, &allocation); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting Allocation %s before reconcile: %v", req.NamespacedName, err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("Allocation %s not found, probably deleted", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *AllocationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nodecorev1alpha1.Allocation{}).
		Complete(r)
}
