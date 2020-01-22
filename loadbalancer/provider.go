package loadbalancer

import (
	"context"

	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
)

type Provider interface {
	Type() string
	Provision(ctx context.Context, domain *domainv1beta1.CustomDomain) (*ProvisionResult, error)
	Release(ctx context.Context, domain *domainv1beta1.CustomDomain) (ok bool, err error)
}

type ProvisionResult struct {
	ProviderType string
	DNSRecords   []DNSRecord
}

type DNSRecord struct {
	Name  string
	Type  string
	Value string
}
