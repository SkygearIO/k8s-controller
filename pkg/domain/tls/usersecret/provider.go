package usersecret

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/pkg/domain/tls"
)

var scheme = runtime.NewScheme()

func init() {
	_ = domainv1beta1.AddToScheme(scheme)
}

type Provider struct {
}

func NewProvider() (*Provider, error) {
	return &Provider{}, nil
}

var _ tls.Provider = &Provider{}

func (p *Provider) Provision(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (*tls.ProvisionResult, error) {
	if reg.Spec.DomainConfig.CertSecretName == nil {
		return nil, fmt.Errorf("certificate secret name not provided")
	}

	return &tls.ProvisionResult{
		CertSecretName: *reg.Spec.DomainConfig.CertSecretName,
	}, nil
}

func (p *Provider) Release(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (bool, error) {
	// Nothing to do here.
	return true, nil
}
