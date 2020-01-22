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
)

// CustomDomainSpec defines the desired state of CustomDomain
type CustomDomainSpec struct {
	// LoadBalancerProvider is the load balancer provider for this domain.
	LoadBalancerProvider *string `json:"loadBalancerProvider,omitempty"`
	// Registrations are registrations from apps.
	Registrations []corev1.ObjectReference `json:"registrations,omitempty"`
}

type CustomDomainDNSRecord struct {
	// Name is name of DNS record
	Name string `json:"name"`
	// Type is type of DNS record
	Type string `json:"type"`
	// Value is value of DNS record
	Value string `json:"value"`
}

// CustomDomainStatusRegistration defines the status of a registration
type CustomDomainStatusRegistration struct {
	// SourceNamespace is originating namespace of the registration
	SourceNamespace string `json:"sourceNamespace"`
	// DNSRecords are DNS records that should be associated with the domain
	DNSRecords []CustomDomainDNSRecord `json:"dnsRecords"`
}

// CustomDomainStatus defines the observed state of CustomDomain
type CustomDomainStatus struct {
	// Registrations are statuses of the registrations
	Registrations []CustomDomainStatusRegistration `json:"registrations,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// CustomDomain is the Schema for the customdomains API
type CustomDomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomDomainSpec   `json:"spec,omitempty"`
	Status CustomDomainStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CustomDomainList contains a list of CustomDomain
type CustomDomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomDomain `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomDomain{}, &CustomDomainList{})
}
