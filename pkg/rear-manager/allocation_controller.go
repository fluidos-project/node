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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
)

// +kubebuilder:rbac:groups=nodecore.github.com/fluidos-project/,resources=allocations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nodecore.github.com/fluidos-project/,resources=allocations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nodecore.github.com/fluidos-project/,resources=allocations/finalizers,verbs=update

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
