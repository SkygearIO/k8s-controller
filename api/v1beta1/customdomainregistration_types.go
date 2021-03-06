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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/skygeario/k8s-controller/api"
)

type CustomDomainConfig struct {
	// BackendServiceName is the name of backend Service.
	BackendServiceName string `json:"backendServiceName"`
	// BackendServicePort is the port of backend Service.
	BackendServicePort int `json:"backendServicePort"`
	// CertSecretName of the name of Secret storing custom TLS certificate
	CertSecretName *string `json:"certSecretName,omitempty"`
	// RedirectToURL is where to redirect the user
	RedirectToURL *string `json:"redirectToURL,omitempty"`
}

// CustomDomainRegistrationSpec defines the desired state of CustomDomainRegistration
type CustomDomainRegistrationSpec struct {
	// DomainName is the custom domain name registered with the app.
	DomainName string `json:"domainName"`
	// DomainConfig is the configuration of custom domain
	DomainConfig CustomDomainConfig `json:"domainConfig"`
	// VerifyAt is the time that next verification should be performed
	// +optional
	VerifyAt *metav1.Time `json:"verifyAt,omitempty"`
}

// CustomDomainRegistrationConditionType is a valid CustomDomainRegistration condition type
type CustomDomainRegistrationConditionType string

const (
	// RegistrationVerified indicates the registration is verified.
	RegistrationVerified CustomDomainRegistrationConditionType = "Verified"
	// RegistrationAccepted indicates the registration is accepted.
	RegistrationAccepted CustomDomainRegistrationConditionType = "Accepted"
	// RegistrationCertReady indicates TLS certificate for the registration is ready.
	RegistrationCertReady CustomDomainRegistrationConditionType = "CertReady"
	// RegistrationIngressReady indicates ingress for the registration is ready.
	RegistrationIngressReady CustomDomainRegistrationConditionType = "IngressReady"
)

// CustomDomainRegistrationStatus defines the observed state of CustomDomainRegistration
type CustomDomainRegistrationStatus struct {
	// Current state of registration.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []api.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// DNSRecords are DNS records that should be associated with the domain
	// +optional
	DNSRecords []CustomDomainDNSRecord `json:"dnsRecords,omitempty"`
	// LastVerificationTime is the time that last verification is performed
	// +optional
	LastVerificationTime *metav1.Time `json:"lastVerificationTime,omitempty"`
	// CertSecretName is the name of TLS certificate secret
	// +optional
	CertSecretName *string `json:"certSecretName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CustomDomainRegistration is the Schema for the customdomainregistrations API
type CustomDomainRegistration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomDomainRegistrationSpec   `json:"spec,omitempty"`
	Status CustomDomainRegistrationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CustomDomainRegistrationList contains a list of CustomDomainRegistration
type CustomDomainRegistrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomDomainRegistration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomDomainRegistration{}, &CustomDomainRegistrationList{})
}
