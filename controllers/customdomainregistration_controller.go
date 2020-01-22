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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/skygeario/k8s-controller/api"
	domain "github.com/skygeario/k8s-controller/api"
	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/util/condition"
	"github.com/skygeario/k8s-controller/util/finalizer"
	"github.com/skygeario/k8s-controller/util/slice"
)

// CustomDomainRegistrationReconciler reconciles a CustomDomainRegistration object
type CustomDomainRegistrationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=domain.skygear.io,resources=customdomainregistrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=domain.skygear.io,resources=customdomainregistrations/status,verbs=get;update;patch

func (r *CustomDomainRegistrationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("customdomainregistration", req.NamespacedName)

	var reg domainv1beta1.CustomDomainRegistration
	if err := r.Get(ctx, req.NamespacedName, &reg); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var conditions []api.Condition
	doFinalize := false
	if reg.DeletionTimestamp == nil {
		finalizerAdded, err := finalizer.Ensure(r, ctx, &reg, domain.DomainFinalizer)
		if err != nil {
			return ctrl.Result{}, err
		}
		if finalizerAdded {
			return ctrl.Result{Requeue: true}, nil
		}

		registered, err := r.registerDomain(ctx, &reg)
		if err != nil {
			conditions = append(conditions, api.Condition{
				Type:    string(domainv1beta1.RegistrationAccepted),
				Status:  metav1.ConditionUnknown,
				Message: err.Error(),
			})
		} else {
			conditions = append(conditions, api.Condition{
				Type:   string(domainv1beta1.RegistrationAccepted),
				Status: condition.ToStatus(registered),
			})
		}

	} else {
		doFinalize = true

		registered, err := r.unregisterDomain(ctx, &reg)
		if err != nil {
			doFinalize = false
			conditions = append(conditions, api.Condition{
				Type:    string(domainv1beta1.RegistrationAccepted),
				Status:  metav1.ConditionUnknown,
				Message: err.Error(),
			})
		} else {
			doFinalize = doFinalize && !registered
			conditions = append(conditions, api.Condition{
				Type:   string(domainv1beta1.RegistrationAccepted),
				Status: condition.ToStatus(registered),
			})
		}
	}

	condition.MergeFrom(conditions, reg.Status.Conditions)
	reg.Status.Conditions = conditions
	if err := r.Status().Update(ctx, &reg); err != nil {
		return ctrl.Result{}, err
	}

	if doFinalize {
		err := finalizer.Remove(r, ctx, &reg, domain.DomainFinalizer)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CustomDomainRegistrationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&domainv1beta1.CustomDomainRegistration{}).
		Complete(r)
}

func (r *CustomDomainRegistrationReconciler) registerDomain(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (registered bool, err error) {
	var domain domainv1beta1.CustomDomain
	err = r.Get(ctx, types.NamespacedName{Name: reg.Name}, &domain)
	if err != nil && !apierrors.IsNotFound(err) {
		return false, err
	}

	regRef := corev1.ObjectReference{
		APIVersion: reg.APIVersion,
		Kind:       reg.Kind,
		Name:       reg.Name,
		Namespace:  reg.Namespace,
		UID:        reg.UID,
	}
	if apierrors.IsNotFound(err) {
		domain = domainv1beta1.CustomDomain{
			ObjectMeta: metav1.ObjectMeta{
				Name: reg.Name,
			},
			Spec: domainv1beta1.CustomDomainSpec{
				Registrations: []corev1.ObjectReference{regRef},
			},
		}
		if err := r.Create(ctx, &domain); err != nil {
			return false, err
		}
	} else {
		if !slice.ContainsObjectReference(domain.Spec.Registrations, reg) {
			domain.Spec.Registrations = append(domain.Spec.Registrations, regRef)
			if err := r.Update(ctx, &domain); err != nil {
				return false, err
			}
		}
	}

	registered = slice.ContainsObjectReference(domain.Spec.Registrations, reg)
	return registered, nil
}

func (r *CustomDomainRegistrationReconciler) unregisterDomain(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (registered bool, err error) {
	var domain domainv1beta1.CustomDomain
	err = r.Get(ctx, types.NamespacedName{Name: reg.Name}, &domain)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if slice.ContainsObjectReference(domain.Spec.Registrations, reg) {
		domain.Spec.Registrations = slice.RemoveObjectReference(domain.Spec.Registrations, reg)
		if err := r.Update(ctx, &domain); err != nil {
			return false, err
		}
	}

	registered = slice.ContainsObjectReference(domain.Spec.Registrations, reg)
	return registered, nil
}
