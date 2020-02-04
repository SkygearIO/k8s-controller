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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/skygeario/k8s-controller/api"
)

// CustomDomainSpec defines the desired state of CustomDomain
type CustomDomainSpec struct {
	// LoadBalancerProvider is the load balancer provider for this domain.
	LoadBalancerProvider *string `json:"loadBalancerProvider,omitempty"`
	// VerificationKey is the domain verification token key.
	VerificationKey *string `json:"verificationKey,omitempty"`
	// Registrations are registrations from apps.
	Registrations []corev1.ObjectReference `json:"registrations,omitempty"`
}

// CustomDomainDNSRecord is a DNS record associated with the domain
type CustomDomainDNSRecord struct {
	// Name is name of DNS record
	Name string `json:"name"`
	// Type is type of DNS record
	Type string `json:"type"`
	// Value is value of DNS record
	Value string `json:"value"`
}

// CustomDomainConditionType is a valid CustomDomain condition type
type CustomDomainConditionType string

const (
	// DomainLoadBalancerProvisioned indicates the required domain resource is provisioned.
	DomainLoadBalancerProvisioned CustomDomainRegistrationConditionType = "LoadBalancerProvisioned"
)

// CustomDomainStatusLoadBalancer defines the status of the domain load balancer
type CustomDomainStatusLoadBalancer struct {
	// Provider is the provider of this load balancer
	Provider string `json:"provider"`
	// DNSRecords are DNS records that should be associated with the domain
	// +optional
	DNSRecords []CustomDomainDNSRecord `json:"dnsRecords,omitempty"`
}

// CustomDomainStatus defines the observed state of CustomDomain
type CustomDomainStatus struct {
	// Current state of custom domain.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []api.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// LoadBalancer is the status of the domain load balancer
	LoadBalancer *CustomDomainStatusLoadBalancer `json:"loadBalancer,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

// CustomDomain is the Schema for the customdomains API
type CustomDomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomDomainSpec   `json:"spec,omitempty"`
	Status CustomDomainStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CustomDomainList contains a list of CustomDomain
type CustomDomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomDomain `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomDomain{}, &CustomDomainList{})
}
