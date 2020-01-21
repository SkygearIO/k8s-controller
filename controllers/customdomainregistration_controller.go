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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	domain "github.com/skygeario/k8s-controller/api"
	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
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

	if reg.DeletionTimestamp == nil {
		finalizerAdded, err := r.ensureFinalizer(ctx, &reg)
		if err != nil {
			return ctrl.Result{}, err
		}
		if finalizerAdded {
			return ctrl.Result{Requeue: true}, nil
		}

	} else {
		err := r.removeFinalizer(ctx, &reg)
		return ctrl.Result{}, err
	}

	if err := r.Status().Update(ctx, &reg); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CustomDomainRegistrationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&domainv1beta1.CustomDomainRegistration{}).
		Complete(r)
}

func (r *CustomDomainRegistrationReconciler) ensureFinalizer(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (added bool, err error) {
	if slice.ContainsString(reg.Finalizers, domain.DomainFinalizer) {
		return false, nil
	}
	reg.Finalizers = append(reg.Finalizers, domain.DomainFinalizer)
	added = true
	err = r.Update(ctx, reg)
	return
}

func (r *CustomDomainRegistrationReconciler) removeFinalizer(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) error {
	reg.Finalizers = slice.RemoveString(reg.Finalizers, domain.DomainFinalizer)
	return r.Update(ctx, reg)
}
