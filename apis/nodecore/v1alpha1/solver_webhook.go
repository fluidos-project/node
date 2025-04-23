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

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var solverlog = logf.Log.WithName("solver-resource")

// SetupWebhookWithManager sets up and registers the webhook with the manager.
func (r *Solver) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&Solver{}).
		WithDefaulter(&Solver{}).
		WithValidator(&Solver{}).
		Complete()
}

//nolint:lll // kubebuilder directives are too long, but they must be on the same line
//+kubebuilder:webhook:path=/mutate-nodecore-fluidos-eu-v1alpha1-solver,mutating=true,failurePolicy=fail,sideEffects=None,groups=nodecore.fluidos.eu,resources=solvers,verbs=create;update,versions=v1alpha1,name=msolver.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &Solver{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *Solver) Default(ctx context.Context, obj runtime.Object) error {
	_ = ctx
	solver := obj.(*Solver)
	solverlog.Info("DEFAULT WEBHOOK")
	solverlog.Info("default", "name", solver.Name)

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//nolint:lll // kubebuilder directives are too long, but they must be on the same line
//+kubebuilder:webhook:path=/validate-nodecore-fluidos-eu-v1alpha1-solver,mutating=false,failurePolicy=fail,sideEffects=None,groups=nodecore.fluidos.eu,resources=solvers,verbs=create;update,versions=v1alpha1,name=vsolver.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &Solver{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *Solver) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	_ = ctx
	solver := obj.(*Solver)
	solverlog.Info("VALIDATE CREATE WEBHOOK")
	solverlog.Info("validate create", "name", solver.Name)

	if err := validateSelector(solver.Spec.Selector); err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *Solver) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	_ = ctx
	solver := newObj.(*Solver)
	solverlog.Info("VALIDATE UPDATE WEBHOOK")
	solverlog.Info("validate update", "name", solver.Name)

	solverlog.Info("old", "old", oldObj)

	if err := validateSelector(solver.Spec.Selector); err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *Solver) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	_ = ctx
	solver := obj.(*Solver)
	solverlog.Info("VALIDATE DELETE WEBHOOK")
	solverlog.Info("validate delete", "name", solver.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func validateSelector(selector *Selector) error {
	if selector == nil {
		return nil
	}
	// Validate creation of Solver checking SolverType->typeIdentifier matches the struct inside the SolverType->TypeData
	typeIdentifier, _, err := ParseSolverSelector(selector)
	if err != nil {
		return err
	}
	switch typeIdentifier {
	case TypeK8Slice:
		solverlog.Info("Selector Flavor Type is K8Slice")
	case TypeVM:
		solverlog.Info("Selector Flavor Type is VM")
	case TypeService:
		solverlog.Info("Selector Flavor Type is Service")
	default:
		solverlog.Info("Selector Flavor Type is not valid")
	}

	return nil
}
