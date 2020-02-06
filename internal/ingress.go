package internal

import (
	"fmt"

	"github.com/skygeario/k8s-controller/pkg/domain/ingress"
	"github.com/skygeario/k8s-controller/pkg/domain/ingress/nginx"
)

func NewIngressProvider(config Config) (ingress.Provider, error) {
	p, err := nginx.NewProvider()
	if err != nil {
		return nil, fmt.Errorf("cannot create nginx ingress provider: %w", err)
	}

	return p, nil
}
