package internal

import "github.com/skygeario/k8s-controller/loadbalancer/staticip"

type Config struct {
	StaticIP *staticip.Config
}
