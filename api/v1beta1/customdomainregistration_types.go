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
)

// CustomDomainRegistrationSpec defines the desired state of CustomDomainRegistration
type CustomDomainRegistrationSpec struct {
	// DomainName is the custom domain name registered with the app.
	DomainName string `json:"domainName,omitempty"`
}

// CustomDomainRegistrationConditionType is a valid value for CustomDomainRegistrationCondition.Type
type CustomDomainRegistrationConditionType string

const (
	// Accepted indicates the registration is accepted.
	RegistrationAccepted CustomDomainRegistrationConditionType = "Accepted"
)

// CustomDomainRegistrationCondition contains details for the current condition of this registration.
type CustomDomainRegistrationCondition struct {
	// Type is the type of the condition.
	Type CustomDomainRegistrationConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status metav1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// CustomDomainRegistrationStatus defines the observed state of CustomDomainRegistration
type CustomDomainRegistrationStatus struct {
	// Current state of registration.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []CustomDomainRegistrationCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
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
