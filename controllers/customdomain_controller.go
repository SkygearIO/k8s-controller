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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/util/slice"
)

// CustomDomainReconciler reconciles a CustomDomain object
type CustomDomainReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=domain.skygear.io,resources=customdomains,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=domain.skygear.io,resources=customdomains/status,verbs=get;update;patch

func (r *CustomDomainReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("customdomain", req.NamespacedName)

	var domain domainv1beta1.CustomDomain
	if err := r.Get(ctx, req.NamespacedName, &domain); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.validateRegistrations(ctx, &domain); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CustomDomainReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&domainv1beta1.CustomDomain{}).
		Complete(r)
}

func (r *CustomDomainReconciler) validateRegistrations(ctx context.Context, d *domainv1beta1.CustomDomain) error {
	n := 0
	for _, ref := range d.Spec.Registrations {
		var reg domainv1beta1.CustomDomainRegistration
		err := r.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, &reg)
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return err
		}

		d.Spec.Registrations[n] = ref
		n++

		if !slice.ContainsOwnerReference(reg.OwnerReferences, d) {
			if err := ctrl.SetControllerReference(d, &reg, r.Scheme); err != nil {
				return err
			}
			if err := r.Update(ctx, &reg); err != nil {
				return err
			}
		}
	}
	if n != len(d.Spec.Registrations) {
		d.Spec.Registrations = d.Spec.Registrations[:n]
		if err := r.Update(ctx, d); err != nil {
			return err
		}
	}
	return nil
}
