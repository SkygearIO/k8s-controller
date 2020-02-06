package internal

import (
	"github.com/skygeario/k8s-controller/pkg/domain/loadbalancer/staticip"
	"github.com/skygeario/k8s-controller/pkg/domain/tls/certmanager"
)

type Config struct {
	StaticIP    *staticip.Config
	CertManager *certmanager.Config
}
