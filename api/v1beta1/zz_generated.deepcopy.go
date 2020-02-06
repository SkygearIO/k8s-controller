// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	"github.com/skygeario/k8s-controller/api"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomDomain) DeepCopyInto(out *CustomDomain) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomDomain.
func (in *CustomDomain) DeepCopy() *CustomDomain {
	if in == nil {
		return nil
	}
	out := new(CustomDomain)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomDomain) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomDomainDNSRecord) DeepCopyInto(out *CustomDomainDNSRecord) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomDomainDNSRecord.
func (in *CustomDomainDNSRecord) DeepCopy() *CustomDomainDNSRecord {
	if in == nil {
		return nil
	}
	out := new(CustomDomainDNSRecord)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomDomainList) DeepCopyInto(out *CustomDomainList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CustomDomain, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomDomainList.
func (in *CustomDomainList) DeepCopy() *CustomDomainList {
	if in == nil {
		return nil
	}
	out := new(CustomDomainList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomDomainList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomDomainRegistration) DeepCopyInto(out *CustomDomainRegistration) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomDomainRegistration.
func (in *CustomDomainRegistration) DeepCopy() *CustomDomainRegistration {
	if in == nil {
		return nil
	}
	out := new(CustomDomainRegistration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomDomainRegistration) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomDomainRegistrationList) DeepCopyInto(out *CustomDomainRegistrationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CustomDomainRegistration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomDomainRegistrationList.
func (in *CustomDomainRegistrationList) DeepCopy() *CustomDomainRegistrationList {
	if in == nil {
		return nil
	}
	out := new(CustomDomainRegistrationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomDomainRegistrationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomDomainRegistrationSpec) DeepCopyInto(out *CustomDomainRegistrationSpec) {
	*out = *in
	if in.VerifyAt != nil {
		in, out := &in.VerifyAt, &out.VerifyAt
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomDomainRegistrationSpec.
func (in *CustomDomainRegistrationSpec) DeepCopy() *CustomDomainRegistrationSpec {
	if in == nil {
		return nil
	}
	out := new(CustomDomainRegistrationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomDomainRegistrationStatus) DeepCopyInto(out *CustomDomainRegistrationStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]api.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.DNSRecords != nil {
		in, out := &in.DNSRecords, &out.DNSRecords
		*out = make([]CustomDomainDNSRecord, len(*in))
		copy(*out, *in)
	}
	if in.LastVerificationTime != nil {
		in, out := &in.LastVerificationTime, &out.LastVerificationTime
		*out = (*in).DeepCopy()
	}
	if in.CertSecretName != nil {
		in, out := &in.CertSecretName, &out.CertSecretName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomDomainRegistrationStatus.
func (in *CustomDomainRegistrationStatus) DeepCopy() *CustomDomainRegistrationStatus {
	if in == nil {
		return nil
	}
	out := new(CustomDomainRegistrationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomDomainSpec) DeepCopyInto(out *CustomDomainSpec) {
	*out = *in
	if in.LoadBalancerProvider != nil {
		in, out := &in.LoadBalancerProvider, &out.LoadBalancerProvider
		*out = new(string)
		**out = **in
	}
	if in.VerificationKey != nil {
		in, out := &in.VerificationKey, &out.VerificationKey
		*out = new(string)
		**out = **in
	}
	if in.Registrations != nil {
		in, out := &in.Registrations, &out.Registrations
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.OwnerApp != nil {
		in, out := &in.OwnerApp, &out.OwnerApp
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomDomainSpec.
func (in *CustomDomainSpec) DeepCopy() *CustomDomainSpec {
	if in == nil {
		return nil
	}
	out := new(CustomDomainSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomDomainStatus) DeepCopyInto(out *CustomDomainStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]api.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.LoadBalancer != nil {
		in, out := &in.LoadBalancer, &out.LoadBalancer
		*out = new(CustomDomainStatusLoadBalancer)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomDomainStatus.
func (in *CustomDomainStatus) DeepCopy() *CustomDomainStatus {
	if in == nil {
		return nil
	}
	out := new(CustomDomainStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomDomainStatusLoadBalancer) DeepCopyInto(out *CustomDomainStatusLoadBalancer) {
	*out = *in
	if in.DNSRecords != nil {
		in, out := &in.DNSRecords, &out.DNSRecords
		*out = make([]CustomDomainDNSRecord, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomDomainStatusLoadBalancer.
func (in *CustomDomainStatusLoadBalancer) DeepCopy() *CustomDomainStatusLoadBalancer {
	if in == nil {
		return nil
	}
	out := new(CustomDomainStatusLoadBalancer)
	in.DeepCopyInto(out)
	return out
}
