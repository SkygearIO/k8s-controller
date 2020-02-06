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
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/skygeario/k8s-controller/api"
	domain "github.com/skygeario/k8s-controller/api"
	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/pkg/domain/ingress"
	"github.com/skygeario/k8s-controller/pkg/domain/tls"
	"github.com/skygeario/k8s-controller/pkg/domain/verification"
	"github.com/skygeario/k8s-controller/pkg/util/condition"
	"github.com/skygeario/k8s-controller/pkg/util/deadline"
	"github.com/skygeario/k8s-controller/pkg/util/finalizer"
	"github.com/skygeario/k8s-controller/pkg/util/slice"
)

const (
	VerificationCooldownSeconds int = 60
	VerificationTimeoutSeconds  int = 5
)

type TLSProvider interface {
	Provision(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (result *tls.ProvisionResult, err error)
	Release(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (ok bool, err error)
}

type IngressProvider interface {
	MakeIngress(reg *domainv1beta1.CustomDomainRegistration) (*networkingv1beta1.Ingress, error)
}

// CustomDomainRegistrationReconciler reconciles a CustomDomainRegistration object
type CustomDomainRegistrationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	Now                        func() metav1.Time
	VerificationTokenGenerator func(key, nonce string) string
	DomainVerifier             func(ctx context.Context, domain, token string) error
	TLSProvider                TLSProvider
	IngressProvider            ingress.Provider
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
	var requeueDeadline deadline.Deadline
	if reg.DeletionTimestamp == nil {
		finalizerAdded, err := finalizer.Ensure(r, ctx, &reg, domain.DomainFinalizer)
		if err != nil {
			return ctrl.Result{}, err
		}
		if finalizerAdded {
			return ctrl.Result{Requeue: true}, nil
		}

		err = r.registerDomain(ctx, &reg)
		if err != nil {
			return ctrl.Result{}, err
		}

		requeueTime, verified, err := r.verifyDomainIfNeeded(ctx, &reg)
		if err != nil {
			conditions = append(conditions, api.Condition{
				Type:    string(domainv1beta1.RegistrationVerified),
				Status:  condition.ToStatus(verified),
				Message: err.Error(),
			})
		} else {
			conditions = append(conditions, api.Condition{
				Type:   string(domainv1beta1.RegistrationVerified),
				Status: condition.ToStatus(verified),
			})
		}
		if requeueTime != nil {
			requeueDeadline.Set(*requeueTime)
		}

		accepted, err := r.checkAcceptance(ctx, &reg)
		if err != nil {
			conditions = append(conditions, api.Condition{
				Type:    string(domainv1beta1.RegistrationAccepted),
				Status:  metav1.ConditionUnknown,
				Message: err.Error(),
			})
		} else {
			conditions = append(conditions, api.Condition{
				Type:   string(domainv1beta1.RegistrationAccepted),
				Status: condition.ToStatus(accepted),
			})
		}

		var certSecretName *string
		if accepted {
			tlsResult, err := r.TLSProvider.Provision(ctx, &reg)
			if err != nil {
				conditions = append(conditions, api.Condition{
					Type:    string(domainv1beta1.RegistrationCertReady),
					Status:  metav1.ConditionUnknown,
					Message: err.Error(),
				})
			} else {
				conditions = append(conditions, api.Condition{
					Type:   string(domainv1beta1.RegistrationCertReady),
					Status: condition.ToStatus(tlsResult != nil),
				})
			}
			if tlsResult == nil {
				requeueDeadline.Set(r.Now().Add(10 * time.Second))
			} else {
				certSecretName = &tlsResult.CertSecretName
			}
		} else {
			released, err := r.TLSProvider.Release(ctx, &reg)
			if err != nil {
				conditions = append(conditions, api.Condition{
					Type:    string(domainv1beta1.RegistrationCertReady),
					Status:  metav1.ConditionUnknown,
					Message: err.Error(),
				})
			} else {
				conditions = append(conditions, api.Condition{
					Type:   string(domainv1beta1.RegistrationCertReady),
					Status: condition.ToStatus(!released),
				})
			}
			if !released {
				requeueDeadline.Set(r.Now().Add(10 * time.Second))
			}
		}
		reg.Status.CertSecretName = certSecretName

		if accepted {
			ok, err := r.updateIngress(ctx, &reg)
			if err != nil {
				conditions = append(conditions, api.Condition{
					Type:    string(domainv1beta1.RegistrationIngressReady),
					Status:  metav1.ConditionUnknown,
					Message: err.Error(),
				})
			} else {
				conditions = append(conditions, api.Condition{
					Type:   string(domainv1beta1.RegistrationIngressReady),
					Status: condition.ToStatus(ok),
				})
			}
		} else {
			ok, err := r.deleteIngress(ctx, &reg)
			if err != nil {
				conditions = append(conditions, api.Condition{
					Type:    string(domainv1beta1.RegistrationIngressReady),
					Status:  metav1.ConditionUnknown,
					Message: err.Error(),
				})
			} else {
				conditions = append(conditions, api.Condition{
					Type:   string(domainv1beta1.RegistrationIngressReady),
					Status: condition.ToStatus(!ok),
				})
			}
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

	return ctrl.Result{RequeueAfter: requeueDeadline.Duration(r.Now().Time)}, nil
}

func (r *CustomDomainRegistrationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&domainv1beta1.CustomDomainRegistration{}).
		Watches(
			&source.Kind{Type: &domainv1beta1.CustomDomain{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(func(o handler.MapObject) []ctrl.Request {
					d := o.Object.(*domainv1beta1.CustomDomain)
					reqs := make([]ctrl.Request, len(d.Spec.Registrations))
					for i, reg := range d.Spec.Registrations {
						reqs[i] = ctrl.Request{NamespacedName: types.NamespacedName{Namespace: reg.Namespace, Name: reg.Name}}
					}
					return reqs
				}),
			},
		).
		Complete(r)
}

func (r *CustomDomainRegistrationReconciler) registerDomain(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) error {
	var domain domainv1beta1.CustomDomain
	err := r.Get(ctx, types.NamespacedName{Name: reg.Name}, &domain)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
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
			return err
		}
	} else {
		if !slice.ContainsObjectReference(domain.Spec.Registrations, reg) {
			patch := client.MergeFrom(domain.DeepCopy())
			domain.Spec.Registrations = append(domain.Spec.Registrations, regRef)
			if err := r.Patch(ctx, &domain, patch); err != nil {
				return err
			}
		}
	}

	return nil
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
		patch := client.MergeFrom(domain.DeepCopy())
		domain.Spec.Registrations = slice.RemoveObjectReference(domain.Spec.Registrations, reg)
		if err := r.Patch(ctx, &domain, patch); err != nil {
			return false, err
		}
	}

	registered = slice.ContainsObjectReference(domain.Spec.Registrations, reg)
	return registered, nil
}

func (r *CustomDomainRegistrationReconciler) verifyDomainIfNeeded(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (requeueTime *time.Time, verified bool, err error) {
	var domain domainv1beta1.CustomDomain
	err = r.Get(ctx, types.NamespacedName{Name: reg.Name}, &domain)
	if err != nil {
		return nil, false, err
	}

	if domain.Spec.VerificationKey == nil || domain.Status.LoadBalancer == nil {
		return nil, false, nil
	}

	token := r.VerificationTokenGenerator(*domain.Spec.VerificationKey, string(reg.UID))
	dnsRecordName, err := verification.MakeDNSRecordName(domain.Name)
	if err != nil {
		return nil, false, err
	}
	records := append(
		domain.Status.LoadBalancer.DNSRecords,
		domainv1beta1.CustomDomainDNSRecord{Name: dnsRecordName, Type: "TXT", Value: token},
	)
	reg.Status.DNSRecords = records

	currentVerified := false
	for _, cond := range reg.Status.Conditions {
		if cond.Type == string(domainv1beta1.RegistrationVerified) {
			currentVerified = cond.Status == metav1.ConditionTrue
			break
		}
	}

	now := r.Now()
	if reg.Spec.VerifyAt == nil ||
		(reg.Status.LastVerificationTime != nil && reg.Status.LastVerificationTime.After(reg.Spec.VerifyAt.Time)) {
		return nil, currentVerified, nil
	}
	verifyTime := reg.Spec.VerifyAt.Time
	if reg.Status.LastVerificationTime != nil &&
		verifyTime.Before(reg.Status.LastVerificationTime.Add(time.Duration(VerificationCooldownSeconds)*time.Second)) {
		verifyTime = reg.Status.LastVerificationTime.Add(time.Duration(VerificationCooldownSeconds) * time.Second)
	}
	if !now.After(verifyTime) {
		return &verifyTime, currentVerified, nil
	}

	err = func() error {
		verifyCtx, cancel := context.WithTimeout(ctx, time.Duration(VerificationTimeoutSeconds)*time.Second)
		defer cancel()
		return r.DomainVerifier(verifyCtx, domain.Name, token)
	}()

	// TODO(domain): re-verify periodically

	reg.Status.LastVerificationTime = &now
	return nil, err == nil, err
}

func (r *CustomDomainRegistrationReconciler) checkAcceptance(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (accepted bool, err error) {
	var domain domainv1beta1.CustomDomain
	err = r.Get(ctx, types.NamespacedName{Name: reg.Name}, &domain)
	if err != nil {
		return false, err
	}

	accepted = domain.Spec.OwnerApp != nil && *domain.Spec.OwnerApp == reg.Namespace
	return accepted, nil
}

func (r *CustomDomainRegistrationReconciler) updateIngress(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (bool, error) {
	ingress, err := r.IngressProvider.MakeIngress(reg)
	if err != nil {
		return false, err
	}

	existingIngress := &networkingv1beta1.Ingress{}
	if err = r.Get(ctx, types.NamespacedName{Namespace: ingress.Namespace, Name: ingress.Name}, existingIngress); err != nil {
		if !apierrors.IsNotFound(err) {
			return false, err
		}
	}

	if apierrors.IsNotFound(err) {
		if err = r.Create(ctx, ingress); err != nil {
			return false, err
		}
	} else {
		existingIngress = existingIngress.DeepCopy()
		existingIngress.Labels = ingress.Labels
		existingIngress.Annotations = ingress.Annotations
		existingIngress.Spec = ingress.Spec
		if err = r.Update(ctx, existingIngress); err != nil {
			return false, err
		}
	}

	return true, nil
}

func (r *CustomDomainRegistrationReconciler) deleteIngress(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (bool, error) {
	ingress, err := r.IngressProvider.MakeIngress(reg)
	if err != nil {
		return false, err
	}

	existingIngress := &networkingv1beta1.Ingress{}
	if err = r.Get(ctx, types.NamespacedName{Namespace: ingress.Namespace, Name: ingress.Name}, existingIngress); err != nil {
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}

	if err = r.Delete(ctx, existingIngress); err != nil {
		return false, err
	}

	return true, nil
}
