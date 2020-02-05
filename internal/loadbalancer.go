package internal

import (
	"context"
	"fmt"

	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/pkg/domain/loadbalancer"
	"github.com/skygeario/k8s-controller/pkg/domain/loadbalancer/staticip"
	"golang.org/x/net/publicsuffix"
)

type LoadBalancer struct {
	StaticIP *staticip.Provider
}

func NewLoadBalancer(config Config) (*LoadBalancer, error) {
	var err error
	var staticIP *staticip.Provider
	if config.StaticIP != nil {
		staticIP, err = staticip.NewProvider(*config.StaticIP)
		if err != nil {
			return nil, fmt.Errorf("cannot create static IP provider: %w", err)
		}
	}

	return &LoadBalancer{
		StaticIP: staticIP,
	}, nil
}

func (p *LoadBalancer) Provision(ctx context.Context, domain *domainv1beta1.CustomDomain) (string, *loadbalancer.ProvisionResult, error) {
	provider, err := p.selectProvider(domain)
	if err != nil {
		return "", nil, err
	}

	result, err := provider.Provision(ctx, domain)
	return provider.Type(), result, err
}

func (p *LoadBalancer) Release(ctx context.Context, domain *domainv1beta1.CustomDomain) (bool, error) {
	provider, err := p.selectProvider(domain)
	if err != nil {
		return false, err
	}
	return provider.Release(ctx, domain)
}

func (p *LoadBalancer) selectProvider(domain *domainv1beta1.CustomDomain) (loadbalancer.Provider, error) {
	if domain.Spec.LoadBalancerProvider != nil {
		return p.lookupProvider(*domain.Spec.LoadBalancerProvider)
	}

	rootDomain, err := publicsuffix.EffectiveTLDPlusOne(domain.Name)
	if err != nil {
		return nil, err
	}

	if domain.Name == rootDomain {
		// no CDN for root domain
		if p.StaticIP != nil {
			return p.StaticIP, nil
		}
	} else {
		// allow CDN for sub-domains
	}

	return nil, fmt.Errorf("no available load-balancer provider for the domain")
}

func (p *LoadBalancer) lookupProvider(providerType string) (loadbalancer.Provider, error) {
	providers := []loadbalancer.Provider{
		p.StaticIP,
	}
	for _, p := range providers {
		if p.Type() == providerType {
			return p, nil
		}
	}

	return nil, fmt.Errorf("load-balancer provider '%s' is unavailable", providerType)
}
