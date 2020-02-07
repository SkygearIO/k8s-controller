package test

import (
	"context"
	"fmt"
	"time"

	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/pkg/domain/loadbalancer"
)

type LoadBalancer struct {
	Now               func() time.Time
	ProvisionTime     time.Duration
	ProvisionRequests map[string]time.Time
	ReleaseTime       time.Duration
	ReleaseRequests   map[string]time.Time
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		Now:               time.Now,
		ProvisionTime:     time.Second * 1,
		ProvisionRequests: map[string]time.Time{},
		ReleaseTime:       time.Second * 1,
		ReleaseRequests:   map[string]time.Time{},
	}
}

func (p *LoadBalancer) Provision(ctx context.Context, domain *domainv1beta1.CustomDomain) (string, *loadbalancer.ProvisionResult, error) {
	reqTime, ok := p.ProvisionRequests[domain.Name]
	if !ok {
		reqTime = p.Now()
		p.ProvisionRequests[domain.Name] = reqTime
	}

	if p.Now().Before(reqTime.Add(p.ProvisionTime)) {
		return "test", nil, nil
	}

	return "test", &loadbalancer.ProvisionResult{DNSRecords: []loadbalancer.DNSRecord{
		{Name: domain.Name, Type: "A", Value: "127.0.0.1"},
	}}, nil
}

func (p *LoadBalancer) Release(ctx context.Context, domain *domainv1beta1.CustomDomain) (ok bool, err error) {
	if _, ok := p.ProvisionRequests[domain.Name]; !ok {
		return true, nil
	}

	reqTime, ok := p.ReleaseRequests[domain.Name]
	if !ok {
		reqTime = p.Now()
		p.ReleaseRequests[domain.Name] = reqTime
	}

	if p.Now().Before(reqTime.Add(p.ReleaseTime)) {
		return false, nil
	}

	delete(p.ReleaseRequests, domain.Name)
	delete(p.ProvisionRequests, domain.Name)
	return true, nil
}
