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
var customdomainlog = logf.Log.WithName("customdomain-resource")

func (r *CustomDomain) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-domain-skygear-io-v1beta1-customdomain,mutating=false,failurePolicy=fail,groups=domain.skygear.io,resources=customdomains,versions=v1beta1,name=vcustomdomain.kb.io

var _ webhook.Validator = &CustomDomain{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CustomDomain) ValidateCreate() error {
	return r.validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CustomDomain) ValidateUpdate(old runtime.Object) error {
	return r.validate(old.(*CustomDomain))
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CustomDomain) ValidateDelete() error {
	return nil
}

func (r *CustomDomain) validate(old *CustomDomain) error {
	var errs field.ErrorList
	if old != nil && old.Name != r.Name {
		errs = append(errs, field.Invalid(field.NewPath("metadata", "name"), r.Name, "resource name cannot be changed"))
	}
	for i, reg := range r.Spec.Registrations {
		if reg.GroupVersionKind() != GroupVersion.WithKind("CustomDomainRegistration") {
			errs = append(errs, field.Invalid(field.NewPath("spec", "registrations").Index(i), r.Name, "only CustomDomainRegistration is supported"))
		}
	}
	if old != nil &&
		old.Spec.LoadBalancerProvider != nil &&
		(r.Spec.LoadBalancerProvider == nil || *old.Spec.LoadBalancerProvider != *r.Spec.LoadBalancerProvider) {
		errs = append(errs, field.Invalid(field.NewPath("spec", "loadBalancerProvider"), r.Name, "load balancer provider cannot be changed"))
	}

	if len(errs) != 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: GroupVersion.Group, Kind: "CustomDomain"},
			r.Name, errs)
	}
	return nil
}
