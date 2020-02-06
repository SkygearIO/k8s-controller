package tls

import (
	"context"

	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
)

type Provider interface {
	Type() string
	Provision(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (*ProvisionResult, error)
	Release(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (ok bool, err error)
}

type ProvisionResult struct {
	CertSecretName string
}
