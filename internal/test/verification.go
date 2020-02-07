package test

import (
	"context"
	"fmt"

	"github.com/skygeario/k8s-controller/pkg/domain/verification"
)

func DomainKeyGenerator() string {
	return "domain-verification-key"
}

type DomainChecker struct {
	Records map[string][]string
}

func NewDomainChecker() *DomainChecker {
	return &DomainChecker{
		Records: map[string][]string{},
	}
}

func (c *DomainChecker) Reset() {
	c.Records = map[string][]string{}
}

func (c *DomainChecker) VerifyDomain(ctx context.Context, domain string, token string) error {
	name, err := verification.MakeDNSRecordName(domain)
	if err != nil {
		return err
	}

	for _, rtoken := range c.Records[name] {
		if rtoken == token {
			return nil
		}
	}
	return fmt.Errorf("verification DNS record not found")
}
