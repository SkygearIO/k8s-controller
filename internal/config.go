package internal

import "github.com/skygeario/k8s-controller/pkg/domain/loadbalancer/staticip"

type Config struct {
	StaticIP *staticip.Config
}
