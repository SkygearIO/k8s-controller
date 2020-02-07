/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers_test

import (
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/controllers"
	internaltest "github.com/skygeario/k8s-controller/internal/test"
	"github.com/skygeario/k8s-controller/pkg/domain/ingress/nginx"
	"github.com/skygeario/k8s-controller/pkg/domain/verification"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var mgrStop chan struct{}

var domainChecker = internaltest.NewDomainChecker()

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	controllers.VerificationCooldown = 3 * time.Second
	controllers.PollInterval = 1 * time.Second

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = domainv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())

	tlsProvider := internaltest.NewTLSProvider(mgr.GetClient())
	loadBalancer := internaltest.NewLoadBalancer()
	ingressProvider, err := nginx.NewProvider()
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.CustomDomainRegistrationReconciler{
		Client:                     mgr.GetClient(),
		Log:                        ctrl.Log.WithName("controllers").WithName("CustomDomainRegistration"),
		Scheme:                     mgr.GetScheme(),
		Now:                        metav1.Now,
		VerificationTokenGenerator: verification.GenerateDomainToken,
		DomainVerifier:             domainChecker.VerifyDomain,
		TLSProvider:                tlsProvider,
		IngressProvider:            ingressProvider,
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.CustomDomainReconciler{
		Client:                   mgr.GetClient(),
		Log:                      ctrl.Log.WithName("controllers").WithName("CustomDomain"),
		Scheme:                   mgr.GetScheme(),
		Now:                      metav1.Now,
		LoadBalancer:             loadBalancer,
		VerificationKeyGenerator: internaltest.DomainKeyGenerator,
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		if mgrStop != nil {
			close(mgrStop)
		}
		mgrStop = make(chan struct{})
		err = mgr.Start(mgrStop)
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = mgr.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	close(mgrStop)
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
