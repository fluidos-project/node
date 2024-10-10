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
var serviceblueprintlog = logf.Log.WithName("serviceblueprint-resource")

// SetupWebhookWithManager setups the webhooks for the ServiceBlueprint resource with the manager.
func (r *ServiceBlueprint) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//nolint:lll // kubebuilder directives are too long, but they must be on the same line
//+kubebuilder:webhook:path=/mutate-nodecore-fluidos-eu-v1alpha1-serviceblueprint,mutating=true,failurePolicy=fail,sideEffects=None,groups=nodecore.fluidos.eu,resources=serviceblueprints,verbs=create;update,versions=v1alpha1,name=mserviceblueprint.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ServiceBlueprint{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *ServiceBlueprint) Default() {
	serviceblueprintlog.Info("DEFAULT WEBHOOK")
	serviceblueprintlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//nolint:lll // kubebuilder directives are too long, but they must be on the same line
//+kubebuilder:webhook:path=/validate-nodecore-fluidos-eu-v1alpha1-serviceblueprint,mutating=false,failurePolicy=fail,sideEffects=None,groups=nodecore.fluidos.eu,resources=serviceblueprints,verbs=create;update,versions=v1alpha1,name=vserviceblueprint.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ServiceBlueprint{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *ServiceBlueprint) ValidateCreate() (admission.Warnings, error) {
	serviceblueprintlog.Info("VALIDATE CREATE WEBHOOK")
	serviceblueprintlog.Info("validate create", "name", r.Name)

	// Validate ServiceBlueprint templates
	manifests, err := ValidateAndExtractManifests(r.Spec.Templates)
	if err != nil {
		return nil, err
	}
	for _, manifest := range manifests {
		serviceblueprintlog.Info("manifest", "manifest", manifest)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *ServiceBlueprint) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	serviceblueprintlog.Info("VALIDATE UPDATE WEBHOOK")
	serviceblueprintlog.Info("validate update", "name", r.Name)

	serviceblueprintlog.Info("old", "old", old)

	// Validate ServiceBlueprint templates
	manifests, err := ValidateAndExtractManifests(r.Spec.Templates)
	if err != nil {
		return nil, err
	}
	for _, manifest := range manifests {
		serviceblueprintlog.Info("manifest", "manifest", manifest)
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *ServiceBlueprint) ValidateDelete() (admission.Warnings, error) {
	serviceblueprintlog.Info("VALIDATE DELETE WEBHOOK")
	serviceblueprintlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
