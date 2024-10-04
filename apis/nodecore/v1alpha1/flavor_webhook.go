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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var flavorlog = logf.Log.WithName("flavor-resource")

// SetupWebhookWithManager setups the webhooks for the Flavor resource with the manager.
func (r *Flavor) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//nolint:lll // kubebuilder directives are too long, but they must be on the same line
//+kubebuilder:webhook:path=/mutate-nodecore-fluidos-eu-v1alpha1-flavor,mutating=true,failurePolicy=fail,sideEffects=None,groups=nodecore.fluidos.eu,resources=flavors,verbs=create;update,versions=v1alpha1,name=mflavor.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Flavor{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *Flavor) Default() {
	flavorlog.Info("DEFAULT WEBHOOK")
	flavorlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//nolint:lll // kubebuilder directives are too long, but they must be on the same line
//+kubebuilder:webhook:path=/validate-nodecore-fluidos-eu-v1alpha1-flavor,mutating=false,failurePolicy=fail,sideEffects=None,groups=nodecore.fluidos.eu,resources=flavors,verbs=create;update,versions=v1alpha1,name=vflavor.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Flavor{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *Flavor) ValidateCreate() (admission.Warnings, error) {
	flavorlog.Info("VALIDATE CREATE WEBHOOK")
	flavorlog.Info("validate create", "name", r.Name)

	// Validate creation of Flavor checking FlavorType->TypeIdenfier matches the struct inside the FlavorType->TypeData
	typeIdenfier, _, err := ParseFlavorType(r)
	if err != nil {
		return nil, err
	}
	switch typeIdenfier {
	case TypeK8Slice:
		flavorlog.Info("FlavorTypeIdentifier is K8Slice")
	case TypeVM:
		flavorlog.Info("FlavorTypeIdentifier is VM")
	case TypeService:
		flavorlog.Info("FlavorTypeIdentifier is Service")
	default:
		flavorlog.Info("FlavorTypeIdentifier is not valid")
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *Flavor) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	flavorlog.Info("VALIDATE UPDATE WEBHOOK")
	flavorlog.Info("validate update", "name", r.Name)

	flavorlog.Info("old", "old", old)

	// Validate creation of Flavor checking FlavorType->TypeIdenfier matches the struct inside the FlavorType->TypeData
	typeIdenfier, _, err := ParseFlavorType(r)
	if err != nil {
		return nil, err
	}
	switch typeIdenfier {
	case TypeK8Slice:
		flavorlog.Info("FlavorTypeIdentifier is K8Slice")
	case TypeVM:
		flavorlog.Info("FlavorTypeIdentifier is VM")
	case TypeService:
		flavorlog.Info("FlavorTypeIdentifier is Service")
	default:
		flavorlog.Info("FlavorTypeIdentifier is not valid")
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *Flavor) ValidateDelete() (admission.Warnings, error) {
	flavorlog.Info("VALIDATE DELETE WEBHOOK")
	flavorlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
