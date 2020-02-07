package test

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/pkg/domain/tls"
)

type TLSProvider struct {
	KubeClient        client.Client
	Now               func() time.Time
	ProvisionTime     time.Duration
	ProvisionRequests map[types.NamespacedName]time.Time
}

func NewTLSProvider(client client.Client) *TLSProvider {
	return &TLSProvider{
		KubeClient:        client,
		Now:               time.Now,
		ProvisionTime:     time.Second * 1,
		ProvisionRequests: map[types.NamespacedName]time.Time{},
	}
}

func (p *TLSProvider) Provision(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (*tls.ProvisionResult, error) {
	n := types.NamespacedName{Namespace: reg.Namespace, Name: reg.Name}
	reqTime, ok := p.ProvisionRequests[n]
	if !ok {
		reqTime = p.Now()
		p.ProvisionRequests[n] = reqTime
	}

	if p.Now().Before(reqTime.Add(p.ProvisionTime)) {
		return nil, nil
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: reg.Namespace,
			Name:      reg.Name + "-tls",
		},
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte{0},
			corev1.TLSPrivateKeyKey: []byte{0},
		},
	}
	if err := p.KubeClient.Create(ctx, secret); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return nil, err
		}
	}
	return &tls.ProvisionResult{CertSecretName: secret.Name}, nil
}

func (p *TLSProvider) Release(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (ok bool, err error) {
	secret := &corev1.Secret{}
	if err := p.KubeClient.Get(ctx, types.NamespacedName{
		Namespace: reg.Namespace,
		Name:      reg.Name + "-tls",
	}, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}

	if err := p.KubeClient.Delete(ctx, secret); err != nil {
		return false, err
	}
	return true, nil
}
