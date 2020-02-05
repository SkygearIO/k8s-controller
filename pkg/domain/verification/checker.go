package verification

import (
	"context"
	"fmt"
	"net"
)

var resolver = &net.Resolver{}

func VerifyDomain(ctx context.Context, domain string, token string) error {
	recordName, err := MakeDNSRecordName(domain)
	if err != nil {
		return fmt.Errorf("cannot lookup verification DNS record: %w", err)
	}

	records, err := resolver.LookupTXT(ctx, recordName)
	if err != nil {
		return fmt.Errorf("cannot lookup verification DNS record: %w", err)
	}

	for _, value := range records {
		if value == token {
			return nil
		}
	}
	return fmt.Errorf("verification DNS record not found")
}
