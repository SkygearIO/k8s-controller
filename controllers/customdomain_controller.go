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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/skygeario/k8s-controller/api"
	domain "github.com/skygeario/k8s-controller/api"
	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/loadbalancer"
	"github.com/skygeario/k8s-controller/util/condition"
	"github.com/skygeario/k8s-controller/util/finalizer"
	"github.com/skygeario/k8s-controller/util/slice"
)

type LoadBalancer interface {
	Provision(ctx context.Context, domain *domainv1beta1.CustomDomain) (providerType string, result *loadbalancer.ProvisionResult, err error)
	Release(ctx context.Context, domain *domainv1beta1.CustomDomain) (ok bool, err error)
}

// CustomDomainReconciler reconciles a CustomDomain object
type CustomDomainReconciler struct {
	client.Client
	Log                      logr.Logger
	Scheme                   *runtime.Scheme
	LoadBalancer             LoadBalancer
	VerificationKeyGenerator func() string
}

// +kubebuilder:rbac:groups=domain.skygear.io,resources=customdomains,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=domain.skygear.io,resources=customdomains/status,verbs=get;update;patch

func (r *CustomDomainReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("customdomain", req.NamespacedName)

	var d domainv1beta1.CustomDomain
	if err := r.Get(ctx, req.NamespacedName, &d); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.validateRegistrations(ctx, &d); err != nil {
		return ctrl.Result{}, err
	}

	var conditions []api.Condition
	doFinalize := false
	if d.DeletionTimestamp == nil {
		finalizerAdded, err := finalizer.Ensure(r, ctx, &d, domain.DomainFinalizer)
		if err != nil {
			return ctrl.Result{}, err
		}
		if finalizerAdded {
			return ctrl.Result{Requeue: true}, nil
		}

		if len(d.Spec.Registrations) == 0 {
			// Release custom domain when no registrations
			err = r.Delete(ctx, &d)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}

		provisioned, err := r.provisionLoadBalancer(ctx, &d)
		if err != nil {
			conditions = append(conditions, api.Condition{
				Type:    string(domainv1beta1.DomainLoadBalancerProvisioned),
				Status:  metav1.ConditionUnknown,
				Message: err.Error(),
			})
		} else {
			conditions = append(conditions, api.Condition{
				Type:   string(domainv1beta1.DomainLoadBalancerProvisioned),
				Status: condition.ToStatus(provisioned),
			})
		}

		err = r.processRegistrations(ctx, &d)
		if err != nil {
			return ctrl.Result{}, err
		}

	} else {
		doFinalize = true

		released, err := r.releaseLoadBalancer(ctx, &d)
		if err != nil {
			doFinalize = false
			conditions = append(conditions, api.Condition{
				Type:    string(domainv1beta1.DomainLoadBalancerProvisioned),
				Status:  metav1.ConditionUnknown,
				Message: err.Error(),
			})
		} else {
			doFinalize = doFinalize && released
			conditions = append(conditions, api.Condition{
				Type:   string(domainv1beta1.DomainLoadBalancerProvisioned),
				Status: condition.ToStatus(!released),
			})
		}
	}

	if err := r.Status().Update(ctx, &d); err != nil {
		return ctrl.Result{}, err
	}

	if doFinalize {
		err := finalizer.Remove(r, ctx, &d, domain.DomainFinalizer)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CustomDomainReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&domainv1beta1.CustomDomain{}).
		Owns(&domainv1beta1.CustomDomainRegistration{}).
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
			patch := client.MergeFrom(reg.DeepCopy())
			if err := ctrl.SetControllerReference(d, &reg, r.Scheme); err != nil {
				return err
			}
			if err := r.Patch(ctx, &reg, patch); err != nil {
				return err
			}
		}
	}
	if n != len(d.Spec.Registrations) {
		patch := client.MergeFrom(d.DeepCopy())
		d.Spec.Registrations = d.Spec.Registrations[:n]
		if err := r.Patch(ctx, d, patch); err != nil {
			return err
		}
	}
	return nil
}

func (r *CustomDomainReconciler) provisionLoadBalancer(ctx context.Context, d *domainv1beta1.CustomDomain) (bool, error) {
	providerType, result, err := r.LoadBalancer.Provision(ctx, d)
	if err != nil {
		return false, err
	}
	if d.Spec.LoadBalancerProvider == nil {
		patch := client.MergeFrom(d.DeepCopy())
		d.Spec.LoadBalancerProvider = &providerType
		if err := r.Patch(ctx, d, patch); err != nil {
			return false, err
		}
	}

	dnsRecords := make([]domainv1beta1.CustomDomainDNSRecord, len(result.DNSRecords))
	for i, r := range result.DNSRecords {
		dnsRecords[i] = domainv1beta1.CustomDomainDNSRecord{
			Name:  r.Name,
			Type:  r.Type,
			Value: r.Value,
		}
	}

	d.Status.LoadBalancer = &domainv1beta1.CustomDomainStatusLoadBalancer{
		Provider:   providerType,
		DNSRecords: dnsRecords,
	}

	return result != nil, nil
}

func (r *CustomDomainReconciler) releaseLoadBalancer(ctx context.Context, d *domainv1beta1.CustomDomain) (bool, error) {
	return r.LoadBalancer.Release(ctx, d)
}

func (r *CustomDomainReconciler) processRegistrations(ctx context.Context, d *domainv1beta1.CustomDomain) error {
	if d.Spec.VerificationKey == nil {
		patch := client.MergeFrom(d.DeepCopy())
		d.Spec.VerificationKey = pointer.StringPtr(r.VerificationKeyGenerator())
		if err := r.Patch(ctx, d, patch); err != nil {
			return err
		}
	}

	if d.Spec.OwnerApp == nil {
		appToAccept := ""
		for _, ref := range d.Spec.Registrations {
			var reg domainv1beta1.CustomDomainRegistration
			if err := r.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, &reg); err != nil {
				return err
			}

			cond := condition.Lookup(reg.Status.Conditions, string(domainv1beta1.RegistrationVerified))
			if cond != nil && cond.Status == metav1.ConditionTrue {
				appToAccept = reg.Namespace
				break
			}
		}

		if appToAccept != "" {
			patch := client.MergeFrom(d.DeepCopy())
			d.Spec.OwnerApp = &appToAccept
			if err := r.Patch(ctx, d, patch); err != nil {
				return err
			}
		}
	} else {
		ownerOk := false
		for _, ref := range d.Spec.Registrations {
			if ref.Namespace != *d.Spec.OwnerApp {
				continue
			}
			var reg domainv1beta1.CustomDomainRegistration
			if err := r.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, &reg); err != nil {
				if !apierrors.IsNotFound(err) {
					return err
				}
				break
			}

			cond := condition.Lookup(reg.Status.Conditions, string(domainv1beta1.RegistrationVerified))
			if cond == nil || cond.Status != metav1.ConditionTrue {
				break
			}

			ownerOk = true
			break
		}

		if !ownerOk {
			patch := client.MergeFrom(d.DeepCopy())
			d.Spec.OwnerApp = nil
			if err := r.Patch(ctx, d, patch); err != nil {
				return err
			}
		}
	}

	return nil
}
