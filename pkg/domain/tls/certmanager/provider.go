package certmanager

import (
	"context"

	cmutil "github.com/jetstack/cert-manager/pkg/api/util"
	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/pkg/domain/tls"
)

var scheme = runtime.NewScheme()

func init() {
	_ = domainv1beta1.AddToScheme(scheme)
}

type Provider struct {
	KubeClient        client.Client
	ClusterIssuerName string
}

func NewProvider(client client.Client, config Config) (*Provider, error) {
	return &Provider{
		KubeClient:        client,
		ClusterIssuerName: config.ClusterIssuerName,
	}, nil
}

var _ tls.Provider = &Provider{}

func (p *Provider) Type() string { return "certmanager" }

func (p *Provider) Provision(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (*tls.ProvisionResult, error) {
	var cert cm.Certificate
	err := p.KubeClient.Get(ctx, types.NamespacedName{Namespace: reg.Namespace, Name: reg.Name}, &cert)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
	}

	if apierrors.IsNotFound(err) {
		cert.Namespace = reg.Namespace
		cert.Name = reg.Name
		cert.Spec.IssuerRef = cmmeta.ObjectReference{
			Kind: "ClusterIssuer",
			Name: p.ClusterIssuerName,
		}
		cert.Spec.SecretName = reg.Name + "-tls"
		cert.Spec.DNSNames = []string{reg.Spec.DomainName}
		if err := ctrl.SetControllerReference(reg, &cert, scheme); err != nil {
			return nil, err
		}
		if err := p.KubeClient.Create(ctx, &cert); err != nil {
			return nil, err
		}
	}

	if !cmutil.CertificateHasCondition(&cert, cm.CertificateCondition{Type: cm.CertificateConditionReady, Status: cmmeta.ConditionTrue}) {
		return nil, nil
	}

	return &tls.ProvisionResult{
		CertSecretName: cert.Spec.SecretName,
	}, nil
}

func (p *Provider) Release(ctx context.Context, reg *domainv1beta1.CustomDomainRegistration) (bool, error) {
	var cert cm.Certificate
	err := p.KubeClient.Get(ctx, types.NamespacedName{Namespace: reg.Namespace, Name: reg.Name}, &cert)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}

	if !metav1.IsControlledBy(&cert, reg) {
		return true, nil
	}

	if err := p.KubeClient.Delete(ctx, &cert); err != nil {
		return false, err
	}
	return true, nil
}
