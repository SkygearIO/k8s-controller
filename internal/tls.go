package internal

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/pkg/domain/tls"
	"github.com/skygeario/k8s-controller/pkg/domain/tls/certmanager"
	"github.com/skygeario/k8s-controller/pkg/domain/tls/usersecret"
)

const (
	tlsCertManager string = "cert-manager"
	tlsUserSecret  string = "user-secret"
)

type TLSProvider struct {
	CertManager *certmanager.Provider
	UserSecret  *usersecret.Provider
}

func NewTLSProvider(client client.Client, config Config) (*TLSProvider, error) {
	var err error

	var certManager *certmanager.Provider
	if config.CertManager != nil {
		certManager, err = certmanager.NewProvider(client, *config.CertManager)
		if err != nil {
			return nil, fmt.Errorf("cannot create cert-manager provider: %w", err)
		}
	} else {
		return nil, fmt.Errorf("cert-manager config is missing")
	}

	var userSecret *usersecret.Provider
	userSecret, err = usersecret.NewProvider()
	if err != nil {
		return nil, fmt.Errorf("cannot create user secret certificate provider: %w", err)
	}

	return &TLSProvider{
		CertManager: certManager,
		UserSecret:  userSecret,
	}, nil
}

func (p *TLSProvider) Provision(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (*tls.ProvisionResult, error) {
	providerType, provider, err := p.selectProvider(reg)
	if err != nil {
		return nil, err
	}

	// release provisioned resources from other providers
	for t, p := range p.allProviders() {
		if t == providerType {
			continue
		}

		released, err := p.Release(ctx, reg)
		if err != nil {
			return nil, err
		}
		if !released {
			return nil, nil
		}
	}

	result, err := provider.Provision(ctx, reg)
	return result, err
}

func (p *TLSProvider) Release(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (bool, error) {
	// release provisioned resources from all providers
	for _, p := range p.allProviders() {
		released, err := p.Release(ctx, reg)
		if err != nil {
			return false, err
		}
		if !released {
			return false, nil
		}
	}
	return true, nil
}

func (p *TLSProvider) allProviders() map[string]tls.Provider {
	return map[string]tls.Provider{
		tlsCertManager: p.CertManager,
		tlsUserSecret:  p.UserSecret,
	}
}

func (p *TLSProvider) selectProvider(reg *domainv1beta1.CustomDomainRegistration) (string, tls.Provider, error) {
	if reg.Spec.DomainConfig.CertSecretName != nil {
		return tlsUserSecret, p.UserSecret, nil
	}
	return tlsCertManager, p.CertManager, nil
}
