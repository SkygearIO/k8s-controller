package nginx

import (
	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/skygeario/k8s-controller/pkg/domain/ingress"
)

var scheme = runtime.NewScheme()

func init() {
	_ = domainv1beta1.AddToScheme(scheme)
}

type Provider struct {
}

func NewProvider() (*Provider, error) {
	return &Provider{}, nil
}

var _ ingress.Provider = &Provider{}

func (p *Provider) Type() string { return "nginx" }

func (p *Provider) MakeIngress(reg *domainv1beta1.CustomDomainRegistration) (*networkingv1beta1.Ingress, error) {
	ingress := networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      reg.Spec.DomainName,
			Namespace: reg.Namespace,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "nginx",
			},
		},
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{
				networkingv1beta1.IngressRule{
					Host: reg.Spec.DomainName,
					IngressRuleValue: networkingv1beta1.IngressRuleValue{
						HTTP: &networkingv1beta1.HTTPIngressRuleValue{
							Paths: []networkingv1beta1.HTTPIngressPath{
								networkingv1beta1.HTTPIngressPath{
									Path: "/",
									Backend: networkingv1beta1.IngressBackend{
										ServiceName: reg.Spec.BackendServiceName,
										ServicePort: intstr.FromInt(reg.Spec.BackendServicePort),
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1beta1.IngressTLS{
				networkingv1beta1.IngressTLS{
					Hosts:      []string{reg.Spec.DomainName},
					SecretName: "",
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(reg, &ingress, scheme); err != nil {
		return nil, err
	}

	if reg.Status.CertSecretName != nil {
		ingress.Spec.TLS[0].SecretName = *reg.Status.CertSecretName
	}

	return &ingress, nil
}
