package verification

import (
	"fmt"

	"golang.org/x/net/publicsuffix"
)

func MakeDNSRecordName(domain string) (string, error) {
	rootDomain, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("_skygear.%s", rootDomain), nil
}
