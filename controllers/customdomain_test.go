package controllers_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/controllers"
	"github.com/skygeario/k8s-controller/pkg/util/condition"
)

var _ = Describe("Custom Domain", func() {
	const timeout = time.Second * 10
	const interval = time.Millisecond * 100

	Context("Custom domain verification", func() {
		It("Should create multiple registration of same unverified domain", func() {
			ctx := context.Background()
			createReg := func(namespace, domain string, backendName string, backendPort int) {
				r := &domainv1beta1.CustomDomainRegistration{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
						Name:      domain,
					},
					Spec: domainv1beta1.CustomDomainRegistrationSpec{
						DomainName:         domain,
						BackendServiceName: backendName,
						BackendServicePort: backendPort,
					},
				}
				Expect(k8sClient.Create(ctx, r)).Should(Succeed())
			}
			createReg("app1", "my-app.test", "app", 80)
			createReg("app2", "my-app.test", "my-app", 8001)
			createReg("app2", "sub.my-app.test", "sub-app", 8080)
		})
		It("Should have correct unverified status and DNS records", func() {
			ctx := context.Background()
			test := func(namespace, domain string) []domainv1beta1.CustomDomainDNSRecord {
				var records []domainv1beta1.CustomDomainDNSRecord
				Eventually(func() error {
					domainReg := &domainv1beta1.CustomDomainRegistration{}
					Expect(k8sClient.Get(ctx,
						types.NamespacedName{Namespace: namespace, Name: domain},
						domainReg)).To(Succeed())

					verified := condition.Lookup(domainReg.Status.Conditions, string(domainv1beta1.RegistrationVerified))
					if verified == nil || verified.Status != metav1.ConditionFalse {
						return fmt.Errorf("unexpected verified status: %#v", verified)
					}

					records = domainReg.Status.DNSRecords
					if len(records) < 2 {
						return fmt.Errorf("unexpected DNS record count: %#v", records)
					}
					return nil
				}, timeout, interval).Should(Succeed())
				return records
			}

			reg1Records := test("app1", "my-app.test")
			reg2Records := test("app2", "my-app.test")
			reg3Records := test("app2", "sub.my-app.test")

			Expect(reg1Records).To(Equal([]domainv1beta1.CustomDomainDNSRecord{
				{Name: "my-app.test", Type: "A", Value: "127.0.0.1"},
				{Name: "_skygear.my-app.test", Type: "TXT", Value: "c4fe13c3968005a8d8fddd37fd2738450b131c6881a501e62d8393660664330d"},
			}))
			Expect(reg2Records).To(Equal([]domainv1beta1.CustomDomainDNSRecord{
				{Name: "my-app.test", Type: "A", Value: "127.0.0.1"},
				{Name: "_skygear.my-app.test", Type: "TXT", Value: "bf46fcae092bcfdbbfb6900e0c343c4447cc284a98e0e3cf49df0470e90085ab"},
			}))
			Expect(reg3Records).To(Equal([]domainv1beta1.CustomDomainDNSRecord{
				{Name: "sub.my-app.test", Type: "A", Value: "127.0.0.1"},
				{Name: "_skygear.my-app.test", Type: "TXT", Value: "bf46fcae092bcfdbbfb6900e0c343c4447cc284a98e0e3cf49df0470e90085ab"},
			}))
		})
		It("Should verify the domains using DNS records", func() {
			ctx := context.Background()
			statusOf := func(namespace, domain string) error {
				n := types.NamespacedName{Namespace: namespace, Name: domain}
				domainReg := &domainv1beta1.CustomDomainRegistration{}
				Expect(k8sClient.Get(ctx, n, domainReg)).To(Succeed())

				status := metav1.ConditionUnknown
				message := ""
				verified := condition.Lookup(domainReg.Status.Conditions, string(domainv1beta1.RegistrationVerified))
				if verified != nil {
					status = verified.Status
					message = verified.Message
				}

				switch status {
				case metav1.ConditionTrue:
					return nil
				case metav1.ConditionFalse:
					return fmt.Errorf("%s", message)
				}
				return fmt.Errorf("unknown status")
			}
			verify := func(namespace, domain string) error {
				n := types.NamespacedName{Namespace: namespace, Name: domain}
				domainReg := &domainv1beta1.CustomDomainRegistration{}

				var verifyAt metav1.Time
				Eventually(func() error {
					if err := k8sClient.Get(ctx, n, domainReg); err != nil {
						return err
					}

					verifyAt = metav1.Unix(metav1.Now().Unix(), 0)
					if domainReg.Status.LastVerificationTime != nil &&
						!verifyAt.After(domainReg.Status.LastVerificationTime.Time) {
						verifyAt = metav1.Time{Time: domainReg.Status.LastVerificationTime.Add(1 + time.Second)}
					}
					domainReg.Spec.VerifyAt = &verifyAt
					if err := k8sClient.Update(ctx, domainReg); err != nil {
						return err
					}
					return nil
				}, timeout, interval).Should(Succeed())

				Eventually(func() error {
					Expect(k8sClient.Get(ctx, n, domainReg)).To(Succeed())

					if domainReg.Status.LastVerificationTime == nil ||
						!domainReg.Status.LastVerificationTime.After(verifyAt.Time) {
						return fmt.Errorf("verification not yet performed: %#v", domainReg.Status)
					}
					return nil
				}, timeout, interval).Should(Succeed())

				return statusOf(namespace, domain)
			}

			domainChecker.Records["_skygear.my-app.test"] = []string{
				"bf46fcae092bcfdbbfb6900e0c343c4447cc284a98e0e3cf49df0470e90085ab",
			}
			Expect(verify("app1", "my-app.test")).To(MatchError("verification DNS record not found"))
			Expect(verify("app2", "my-app.test")).To(Succeed())
			Expect(verify("app2", "sub.my-app.test")).To(Succeed())

			n := types.NamespacedName{Namespace: "app1", Name: "my-app.test"}
			domainRegOld := &domainv1beta1.CustomDomainRegistration{}
			Expect(k8sClient.Get(ctx, n, domainRegOld)).To(Succeed())

			domainChecker.Records["_skygear.my-app.test"] = []string{
				"c4fe13c3968005a8d8fddd37fd2738450b131c6881a501e62d8393660664330d",
			}
			Expect(verify("app1", "my-app.test")).To(Succeed())
			Expect(verify("app2", "my-app.test")).To(MatchError("verification DNS record not found"))
			Expect(verify("app2", "sub.my-app.test")).To(MatchError("verification DNS record not found"))

			// Should have cooldown applied
			domainRegNew := &domainv1beta1.CustomDomainRegistration{}
			Expect(k8sClient.Get(ctx, n, domainRegNew)).To(Succeed())
			interval := domainRegNew.Status.LastVerificationTime.Sub(domainRegOld.Status.LastVerificationTime.Time)
			Expect(interval).To(BeNumerically(">=", controllers.VerificationCooldown))

			domainChecker.Records["_skygear.my-app.test"] = []string{
				"bf46fcae092bcfdbbfb6900e0c343c4447cc284a98e0e3cf49df0470e90085ab",
				"c4fe13c3968005a8d8fddd37fd2738450b131c6881a501e62d8393660664330d",
			}
			Expect(verify("app1", "my-app.test")).To(Succeed())
			Expect(verify("app2", "my-app.test")).To(Succeed())
			Expect(verify("app2", "sub.my-app.test")).To(Succeed())
		})
		It("Should accept first verified domain", func() {
			ctx := context.Background()
			accepted := func(namespace, domain string) func() bool {
				return func() bool {
					n := types.NamespacedName{Namespace: namespace, Name: domain}
					domainReg := &domainv1beta1.CustomDomainRegistration{}
					Expect(k8sClient.Get(ctx, n, domainReg)).To(Succeed())

					accepted := condition.Lookup(domainReg.Status.Conditions, string(domainv1beta1.RegistrationAccepted))
					return accepted != nil && accepted.Status == metav1.ConditionTrue
				}
			}

			Eventually(accepted("app1", "my-app.test"), timeout, interval).Should(BeTrue())
			Eventually(accepted("app2", "my-app.test"), timeout, interval).Should(BeFalse())
			Eventually(accepted("app2", "sub.my-app.test"), timeout, interval).Should(BeTrue())
		})
		It("Should configure Ingress and certificate Secret", func() {
			ctx := context.Background()
			ready := func(namespace, domain string) func() bool {
				return func() bool {
					n := types.NamespacedName{Namespace: namespace, Name: domain}
					domainReg := &domainv1beta1.CustomDomainRegistration{}
					Expect(k8sClient.Get(ctx, n, domainReg)).To(Succeed())

					ingressReady := condition.Lookup(domainReg.Status.Conditions, string(domainv1beta1.RegistrationIngressReady))
					if ingressReady == nil || ingressReady.Status != metav1.ConditionTrue {
						return false
					}

					certReady := condition.Lookup(domainReg.Status.Conditions, string(domainv1beta1.RegistrationCertReady))
					if certReady == nil || certReady.Status != metav1.ConditionTrue {
						return false
					}

					return true
				}
			}
			Eventually(ready("app1", "my-app.test"), timeout, interval).Should(BeTrue())
			Eventually(ready("app2", "my-app.test"), timeout, interval).Should(BeFalse())
			Eventually(ready("app2", "sub.my-app.test"), timeout, interval).Should(BeTrue())

			secretName := func(namespace, domain string) string {
				n := types.NamespacedName{Namespace: namespace, Name: domain}
				domainReg := &domainv1beta1.CustomDomainRegistration{}
				Expect(k8sClient.Get(ctx, n, domainReg)).To(Succeed())

				if domainReg.Status.CertSecretName == nil {
					return ""
				}
				return *domainReg.Status.CertSecretName
			}

			Expect(secretName("app1", "my-app.test")).Should(Equal("my-app.test-tls"))
			Expect(secretName("app2", "sub.my-app.test")).Should(Equal("sub.my-app.test-tls"))

			type Ingress struct {
				Annotations map[string]string
				Spec        networkingv1beta1.IngressSpec
			}
			ingress := func(namespace, domain string) Ingress {
				n := types.NamespacedName{Namespace: namespace, Name: domain}
				ingress := &networkingv1beta1.Ingress{}
				Expect(k8sClient.Get(ctx, n, ingress)).To(Succeed())

				return Ingress{
					Annotations: ingress.Annotations,
					Spec:        ingress.Spec,
				}
			}

			Expect(ingress("app1", "my-app.test")).Should(Equal(Ingress{
				map[string]string{
					"kubernetes.io/ingress.class": "nginx",
				},
				networkingv1beta1.IngressSpec{
					Rules: []networkingv1beta1.IngressRule{{
						Host: "my-app.test",
						IngressRuleValue: networkingv1beta1.IngressRuleValue{
							HTTP: &networkingv1beta1.HTTPIngressRuleValue{
								Paths: []networkingv1beta1.HTTPIngressPath{{
									Path: "/",
									Backend: networkingv1beta1.IngressBackend{
										ServiceName: "app",
										ServicePort: intstr.FromInt(80),
									},
								}},
							},
						},
					}},
					TLS: []networkingv1beta1.IngressTLS{{
						Hosts:      []string{"my-app.test"},
						SecretName: "my-app.test-tls",
					}},
				},
			}))
			Expect(ingress("app2", "sub.my-app.test")).Should(Equal(Ingress{
				map[string]string{
					"kubernetes.io/ingress.class": "nginx",
				},
				networkingv1beta1.IngressSpec{
					Rules: []networkingv1beta1.IngressRule{{
						Host: "sub.my-app.test",
						IngressRuleValue: networkingv1beta1.IngressRuleValue{
							HTTP: &networkingv1beta1.HTTPIngressRuleValue{
								Paths: []networkingv1beta1.HTTPIngressPath{{
									Path: "/",
									Backend: networkingv1beta1.IngressBackend{
										ServiceName: "sub-app",
										ServicePort: intstr.FromInt(8080),
									},
								}},
							},
						},
					}},
					TLS: []networkingv1beta1.IngressTLS{{
						Hosts:      []string{"sub.my-app.test"},
						SecretName: "sub.my-app.test-tls",
					}},
				},
			}))
		})
		It("Should cleanup associated resources", func() {
			ctx := context.Background()
			domainReg := &domainv1beta1.CustomDomainRegistration{}
			Expect(k8sClient.Get(
				ctx,
				types.NamespacedName{Namespace: "app1", Name: "my-app.test"},
				domainReg)).To(Succeed())
			Expect(k8sClient.Delete(ctx, domainReg)).To(Succeed())
			Expect(k8sClient.Get(
				ctx,
				types.NamespacedName{Namespace: "app2", Name: "my-app.test"},
				domainReg)).To(Succeed())
			Expect(k8sClient.Delete(ctx, domainReg)).To(Succeed())
			Expect(k8sClient.Get(
				ctx,
				types.NamespacedName{Namespace: "app2", Name: "sub.my-app.test"},
				domainReg)).To(Succeed())
			Expect(k8sClient.Delete(ctx, domainReg)).To(Succeed())

			regList := &domainv1beta1.CustomDomainRegistrationList{}
			Eventually(func() error {
				if err := k8sClient.List(ctx, regList); err != nil {
					return err
				}
				if len(regList.Items) > 0 {
					return fmt.Errorf("CustomDomainRegistration is not cleaned up: %#v", regList)
				}
				return nil
			}, timeout, interval).Should(Succeed())

			domainList := &domainv1beta1.CustomDomainList{}
			Eventually(func() error {
				if err := k8sClient.List(ctx, domainList); err != nil {
					return err
				}
				if len(domainList.Items) > 0 {
					return fmt.Errorf("CustomDomain is not cleaned up: %#v", domainList)
				}
				return nil
			}, timeout, interval).Should(Succeed())

			// Secret and Ingress is deleted by GC in real env, not tested here.
		})

		It("Should not register to a terminating domain", func() {
			ctx := context.Background()
			d := &domainv1beta1.CustomDomain{}
			r := &domainv1beta1.CustomDomainRegistration{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "app1",
					Name:      "my-app.test",
				},
				Spec: domainv1beta1.CustomDomainRegistrationSpec{
					DomainName:         "my-app.test",
					BackendServiceName: "app",
					BackendServicePort: 80,
				},
			}
			Expect(k8sClient.Create(ctx, r.DeepCopy())).Should(Succeed())
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: r.Spec.DomainName}, d)
			}, timeout, interval).Should(Succeed())
			// CustomDomain & CustomDomainRegistration is created
			domainUID := d.UID

			Expect(k8sClient.Delete(ctx, r.DeepCopy())).Should(Succeed())
			Eventually(func() bool {
				rr := &domainv1beta1.CustomDomainRegistration{}
				err := k8sClient.Get(ctx, types.NamespacedName{Namespace: r.Namespace, Name: r.Name}, rr)
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
			// CustomDomainRegistration is deleted

			Eventually(func() bool {
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: r.Spec.DomainName}, d)).To(Succeed())
				return d.DeletionTimestamp != nil
			}, timeout, interval).Should(BeTrue())
			// CustomDomain is terminating

			Expect(k8sClient.Create(ctx, r.DeepCopy())).Should(Succeed())
			// Recreating CustomDomainRegistration
			Eventually(func() error {
				rr := &domainv1beta1.CustomDomainRegistration{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: r.Namespace, Name: r.Name}, rr); err != nil {
					return err
				}
				owner := metav1.GetControllerOf(rr)
				// Should not owned by the terminating domain
				if owner == nil || owner.UID == domainUID {
					return fmt.Errorf("unexpected owner: %#v", rr.OwnerReferences)
				}
				return nil
			}, timeout, interval).Should(Succeed())
		})
	})
})
