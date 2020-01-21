/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var customdomainregistrationlog = logf.Log.WithName("customdomainregistration-resource")

func (r *CustomDomainRegistration) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-domain-skygear-io-v1beta1-customdomainregistration,mutating=false,failurePolicy=fail,groups=domain.skygear.io,resources=customdomainregistrations,versions=v1beta1,name=vcustomdomainregistration.kb.io

var _ webhook.Validator = &CustomDomainRegistration{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CustomDomainRegistration) ValidateCreate() error {
	return r.validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CustomDomainRegistration) ValidateUpdate(old runtime.Object) error {
	return r.validate(old.(*CustomDomainRegistration))
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CustomDomainRegistration) ValidateDelete() error {
	return nil
}

func (r *CustomDomainRegistration) validate(old *CustomDomainRegistration) error {
	var errs field.ErrorList
	if old != nil && old.Name != r.Name {
		errs = append(errs, field.Invalid(field.NewPath("metadata", "name"), r.Name, "resource name cannot be changed"))
	}
	if r.Name != r.Spec.DomainName {
		errs = append(errs, field.Invalid(field.NewPath("spec", "domainName"), r.Spec.DomainName, "domainName must be same as resource name"))
	}

	if len(errs) != 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: GroupVersion.Group, Kind: "CustomDomainRegistration"},
			r.Name, errs)
	}
	return nil
}
