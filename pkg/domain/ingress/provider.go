package ingress

import (
	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
)

type Provider interface {
	MakeIngress(reg *domainv1beta1.CustomDomainRegistration) (*networkingv1beta1.Ingress, error)
}
