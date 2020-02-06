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

package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"

	// +kubebuilder:scaffold:imports

	domainv1beta1 "github.com/skygeario/k8s-controller/api/v1beta1"
	"github.com/skygeario/k8s-controller/controllers"
	"github.com/skygeario/k8s-controller/internal"
	"github.com/skygeario/k8s-controller/pkg/domain/verification"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = cm.AddToScheme(scheme)

	_ = domainv1beta1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var enableWebhooks bool
	var configFile string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enableWebhooks, "enable-webhooks", true, "Enable CRD webhooks.")
	flag.StringVar(&configFile, "config-file", "", "Path to configuration JSON file.")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	configJSON, err := ioutil.ReadFile(configFile)
	if err != nil {
		setupLog.Error(err, "unable read configuration")
		os.Exit(1)
	}
	var config internal.Config
	if err := json.Unmarshal(configJSON, &config); err != nil {
		setupLog.Error(err, "unable parse configuration")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	loadBalancer, err := internal.NewLoadBalancer(config)
	if err != nil {
		setupLog.Error(err, "unable create load balancer")
		os.Exit(1)
	}

	tlsProvider, err := internal.NewTLSProvider(mgr.GetClient(), config)
	if err != nil {
		setupLog.Error(err, "unable create TLS provider")
		os.Exit(1)
	}

	ingressProvider, err := internal.NewIngressProvider(config)
	if err != nil {
		setupLog.Error(err, "unable create ingress provider")
		os.Exit(1)
	}

	if enableWebhooks {
		if err = (&domainv1beta1.CustomDomainRegistration{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "CustomDomainRegistration")
			os.Exit(1)
		}
		if err = (&domainv1beta1.CustomDomain{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "CustomDomain")
			os.Exit(1)
		}
	}
	if err = (&controllers.CustomDomainRegistrationReconciler{
		Client:                     mgr.GetClient(),
		Log:                        ctrl.Log.WithName("controllers").WithName("CustomDomainRegistration"),
		Scheme:                     mgr.GetScheme(),
		Now:                        metav1.Now,
		VerificationTokenGenerator: verification.GenerateDomainToken,
		DomainVerifier:             verification.VerifyDomain,
		TLSProvider:                tlsProvider,
		IngressProvider:            ingressProvider,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CustomDomainRegistration")
		os.Exit(1)
	}
	if err = (&controllers.CustomDomainReconciler{
		Client:                   mgr.GetClient(),
		Log:                      ctrl.Log.WithName("controllers").WithName("CustomDomain"),
		Scheme:                   mgr.GetScheme(),
		LoadBalancer:             loadBalancer,
		VerificationKeyGenerator: verification.GenerateDomainKey,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CustomDomain")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
