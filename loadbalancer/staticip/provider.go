package staticip

import (
	"context"
	"fmt"
	"net"

	"golang.org/x/net/publicsuffix"

	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/loadbalancer"
)

type Provider struct {
	IPAddresses []net.IP
}

func NewProvider(config Config) (*Provider, error) {
	ips := make([]net.IP, len(config.IPAddresses))
	for i, s := range config.IPAddresses {
		ip := net.ParseIP(s)
		if ip == nil {
			return nil, fmt.Errorf("IP address '%s' is not valid", s)
		}
		ips[i] = ip
	}

	return &Provider{
		IPAddresses: ips,
	}, nil
}

var _ loadbalancer.Provider = &Provider{}

func (p *Provider) Type() string { return "static-ip" }

func (p *Provider) Provision(ctx context.Context, domain *domainv1beta1.CustomDomain) (*loadbalancer.ProvisionResult, error) {
	rootDomain, err := publicsuffix.EffectiveTLDPlusOne(domain.Name)
	if err != nil {
		return nil, err
	}

	dnsRecords := make([]loadbalancer.DNSRecord, len(p.IPAddresses))
	for i, ip := range p.IPAddresses {
		var recordType string
		if ip.To4() == nil {
			recordType = "AAAA"
		} else {
			recordType = "A"
		}

		name := domain.Name
		if name == rootDomain {
			name = "@"
		}

		dnsRecords[i] = loadbalancer.DNSRecord{
			Name:  name,
			Type:  recordType,
			Value: ip.String(),
		}
	}

	return &loadbalancer.ProvisionResult{
		DNSRecords: dnsRecords,
	}, nil
}

func (p *Provider) Release(ctx context.Context, domain *domainv1beta1.CustomDomain) (bool, error) {
	// Nothing to do.
	return true, nil
}
