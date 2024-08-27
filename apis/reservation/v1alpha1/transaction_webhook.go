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
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	advertisementv1alpha1 "github.com/fluidos-project/node/apis/advertisement/v1alpha1"
)

// log is for logging in this package.
var transactionlog = logf.Log.WithName("transaction-resource")

// Kubernetes client.
var k8sClientTransaction client.Client

// Context.
var ctxTransaction context.Context

// SetupWebhookWithManager sets up and registers the webhook with the manager.
func (r *Transaction) SetupWebhookWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()

	if err != nil {
		return err
	}

	// Get k8s client
	k8sClientTransaction = mgr.GetClient()

	// Get context
	ctxTransaction = context.Background()

	return nil
}

//nolint:lll // kubebuilder directives are too long, but they must be on the same line
//+kubebuilder:webhook:path=/mutate-reservation-fluidos-eu-v1alpha1-transaction,mutating=true,failurePolicy=fail,sideEffects=None,groups=reservation.fluidos.eu,resources=transactions,verbs=create;update,versions=v1alpha1,name=mtransaction.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Transaction{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *Transaction) Default() {
	transactionlog.Info("TRANSACTION DEFAULT WEBHOOK")
	transactionlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//nolint:lll // kubebuilder directives are too long, but they must be on the same line
//+kubebuilder:webhook:path=/validate-reservation-fluidos-eu-v1alpha1-transaction,mutating=false,failurePolicy=fail,sideEffects=None,groups=reservation.fluidos.eu,resources=transactions,verbs=create;update,versions=v1alpha1,name=vtransaction.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Transaction{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *Transaction) ValidateCreate() (admission.Warnings, error) {
	transactionlog.Info("TRANSACTION VALIDATE CREATE WEBHOOK")
	transactionlog.Info("validate create", "name", r.Name)

	if err := validateTransaction(r); err != nil {
		transactionlog.Error(err, "Error validating Transaction in update")
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *Transaction) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	transactionlog.Info("TRANSACTION VALIDATE UPDATE WEBHOOK")
	transactionlog.Info("validate update", "name", r.Name)

	transactionlog.Info("old", "name", old)

	if err := validateTransaction(r); err != nil {
		transactionlog.Error(err, "Error validating Transaction in update")
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *Transaction) ValidateDelete() (admission.Warnings, error) {
	transactionlog.Info("TRANSACTION VALIDATE DELETE WEBHOOK")
	transactionlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func validateTransaction(transaction *Transaction) error {
	// Retrieve all the PeeringCandidates
	peeringCandidates := &advertisementv1alpha1.PeeringCandidateList{}
	if err := k8sClientTransaction.List(ctxTransaction, peeringCandidates); err != nil {
		transactionlog.Error(err, "Error when listing PeeringCandidates")
		return err
	}

	// Get the PeeringCandidate which Flavor is the same as the Transaction
	var peeringCandidate *advertisementv1alpha1.PeeringCandidate
	for i := range peeringCandidates.Items {
		pc := &peeringCandidates.Items[i]
		if pc.Spec.Flavor.Name == transaction.Spec.FlavorID {
			peeringCandidate = pc
			break
		}
	}

	// Validate the Configuration
	return validateConfiguration(transaction.Spec.Configuration, &peeringCandidate.Spec.Flavor)
}
