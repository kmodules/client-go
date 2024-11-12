//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright AppsCode Inc. and Contributors

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPIClusterInfo) DeepCopyInto(out *CAPIClusterInfo) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPIClusterInfo.
func (in *CAPIClusterInfo) DeepCopy() *CAPIClusterInfo {
	if in == nil {
		return nil
	}
	out := new(CAPIClusterInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertificatePrivateKey) DeepCopyInto(out *CertificatePrivateKey) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertificatePrivateKey.
func (in *CertificatePrivateKey) DeepCopy() *CertificatePrivateKey {
	if in == nil {
		return nil
	}
	out := new(CertificatePrivateKey)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertificateSpec) DeepCopyInto(out *CertificateSpec) {
	*out = *in
	if in.IssuerRef != nil {
		in, out := &in.IssuerRef, &out.IssuerRef
		*out = new(corev1.TypedLocalObjectReference)
		(*in).DeepCopyInto(*out)
	}
	if in.Subject != nil {
		in, out := &in.Subject, &out.Subject
		*out = new(X509Subject)
		(*in).DeepCopyInto(*out)
	}
	if in.Duration != nil {
		in, out := &in.Duration, &out.Duration
		*out = new(metav1.Duration)
		**out = **in
	}
	if in.RenewBefore != nil {
		in, out := &in.RenewBefore, &out.RenewBefore
		*out = new(metav1.Duration)
		**out = **in
	}
	if in.DNSNames != nil {
		in, out := &in.DNSNames, &out.DNSNames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.IPAddresses != nil {
		in, out := &in.IPAddresses, &out.IPAddresses
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.URIs != nil {
		in, out := &in.URIs, &out.URIs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.EmailAddresses != nil {
		in, out := &in.EmailAddresses, &out.EmailAddresses
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PrivateKey != nil {
		in, out := &in.PrivateKey, &out.PrivateKey
		*out = new(CertificatePrivateKey)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertificateSpec.
func (in *CertificateSpec) DeepCopy() *CertificateSpec {
	if in == nil {
		return nil
	}
	out := new(CertificateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterClaimFeatures) DeepCopyInto(out *ClusterClaimFeatures) {
	*out = *in
	if in.EnabledFeatures != nil {
		in, out := &in.EnabledFeatures, &out.EnabledFeatures
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.NotManagedFeatures != nil {
		in, out := &in.NotManagedFeatures, &out.NotManagedFeatures
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.DisabledFeatures != nil {
		in, out := &in.DisabledFeatures, &out.DisabledFeatures
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterClaimFeatures.
func (in *ClusterClaimFeatures) DeepCopy() *ClusterClaimFeatures {
	if in == nil {
		return nil
	}
	out := new(ClusterClaimFeatures)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterClaimInfo) DeepCopyInto(out *ClusterClaimInfo) {
	*out = *in
	in.ClusterMetadata.DeepCopyInto(&out.ClusterMetadata)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterClaimInfo.
func (in *ClusterClaimInfo) DeepCopy() *ClusterClaimInfo {
	if in == nil {
		return nil
	}
	out := new(ClusterClaimInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterInfo) DeepCopyInto(out *ClusterInfo) {
	*out = *in
	if in.ClusterManagers != nil {
		in, out := &in.ClusterManagers, &out.ClusterManagers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.CAPI != nil {
		in, out := &in.CAPI, &out.CAPI
		*out = new(CAPIClusterInfo)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterInfo.
func (in *ClusterInfo) DeepCopy() *ClusterInfo {
	if in == nil {
		return nil
	}
	out := new(ClusterInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterMetadata) DeepCopyInto(out *ClusterMetadata) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterMetadata.
func (in *ClusterMetadata) DeepCopy() *ClusterMetadata {
	if in == nil {
		return nil
	}
	out := new(ClusterMetadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Condition) DeepCopyInto(out *Condition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Condition.
func (in *Condition) DeepCopy() *Condition {
	if in == nil {
		return nil
	}
	out := new(Condition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Conditions) DeepCopyInto(out *Conditions) {
	{
		in := &in
		*out = make(Conditions, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Conditions.
func (in Conditions) DeepCopy() Conditions {
	if in == nil {
		return nil
	}
	out := new(Conditions)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HealthCheckSpec) DeepCopyInto(out *HealthCheckSpec) {
	*out = *in
	in.ReadonlyHealthCheckSpec.DeepCopyInto(&out.ReadonlyHealthCheckSpec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HealthCheckSpec.
func (in *HealthCheckSpec) DeepCopy() *HealthCheckSpec {
	if in == nil {
		return nil
	}
	out := new(HealthCheckSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImageInfo) DeepCopyInto(out *ImageInfo) {
	*out = *in
	if in.Lineages != nil {
		in, out := &in.Lineages, &out.Lineages
		*out = make([]Lineage, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.PullCredentials != nil {
		in, out := &in.PullCredentials, &out.PullCredentials
		*out = new(PullCredentials)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImageInfo.
func (in *ImageInfo) DeepCopy() *ImageInfo {
	if in == nil {
		return nil
	}
	out := new(ImageInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Lineage) DeepCopyInto(out *Lineage) {
	*out = *in
	if in.Chain != nil {
		in, out := &in.Chain, &out.Chain
		*out = make([]ObjectInfo, len(*in))
		copy(*out, *in)
	}
	if in.Containers != nil {
		in, out := &in.Containers, &out.Containers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Lineage.
func (in *Lineage) DeepCopy() *Lineage {
	if in == nil {
		return nil
	}
	out := new(Lineage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectID) DeepCopyInto(out *ObjectID) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectID.
func (in *ObjectID) DeepCopy() *ObjectID {
	if in == nil {
		return nil
	}
	out := new(ObjectID)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectInfo) DeepCopyInto(out *ObjectInfo) {
	*out = *in
	out.Resource = in.Resource
	out.Ref = in.Ref
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectInfo.
func (in *ObjectInfo) DeepCopy() *ObjectInfo {
	if in == nil {
		return nil
	}
	out := new(ObjectInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectReference) DeepCopyInto(out *ObjectReference) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectReference.
func (in *ObjectReference) DeepCopy() *ObjectReference {
	if in == nil {
		return nil
	}
	out := new(ObjectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PullCredentials) DeepCopyInto(out *PullCredentials) {
	*out = *in
	if in.SecretRefs != nil {
		in, out := &in.SecretRefs, &out.SecretRefs
		*out = make([]corev1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PullCredentials.
func (in *PullCredentials) DeepCopy() *PullCredentials {
	if in == nil {
		return nil
	}
	out := new(PullCredentials)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReadonlyHealthCheckSpec) DeepCopyInto(out *ReadonlyHealthCheckSpec) {
	*out = *in
	if in.PeriodSeconds != nil {
		in, out := &in.PeriodSeconds, &out.PeriodSeconds
		*out = new(int32)
		**out = **in
	}
	if in.TimeoutSeconds != nil {
		in, out := &in.TimeoutSeconds, &out.TimeoutSeconds
		*out = new(int32)
		**out = **in
	}
	if in.FailureThreshold != nil {
		in, out := &in.FailureThreshold, &out.FailureThreshold
		*out = new(int32)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReadonlyHealthCheckSpec.
func (in *ReadonlyHealthCheckSpec) DeepCopy() *ReadonlyHealthCheckSpec {
	if in == nil {
		return nil
	}
	out := new(ReadonlyHealthCheckSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceID) DeepCopyInto(out *ResourceID) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceID.
func (in *ResourceID) DeepCopy() *ResourceID {
	if in == nil {
		return nil
	}
	out := new(ResourceID)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSConfig) DeepCopyInto(out *TLSConfig) {
	*out = *in
	if in.IssuerRef != nil {
		in, out := &in.IssuerRef, &out.IssuerRef
		*out = new(corev1.TypedLocalObjectReference)
		(*in).DeepCopyInto(*out)
	}
	if in.Certificates != nil {
		in, out := &in.Certificates, &out.Certificates
		*out = make([]CertificateSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSConfig.
func (in *TLSConfig) DeepCopy() *TLSConfig {
	if in == nil {
		return nil
	}
	out := new(TLSConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TimeOfDay.
func (in *TimeOfDay) DeepCopy() *TimeOfDay {
	if in == nil {
		return nil
	}
	out := new(TimeOfDay)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TypedObjectReference) DeepCopyInto(out *TypedObjectReference) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TypedObjectReference.
func (in *TypedObjectReference) DeepCopy() *TypedObjectReference {
	if in == nil {
		return nil
	}
	out := new(TypedObjectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *X509Subject) DeepCopyInto(out *X509Subject) {
	*out = *in
	if in.Organizations != nil {
		in, out := &in.Organizations, &out.Organizations
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Countries != nil {
		in, out := &in.Countries, &out.Countries
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.OrganizationalUnits != nil {
		in, out := &in.OrganizationalUnits, &out.OrganizationalUnits
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Localities != nil {
		in, out := &in.Localities, &out.Localities
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Provinces != nil {
		in, out := &in.Provinces, &out.Provinces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.StreetAddresses != nil {
		in, out := &in.StreetAddresses, &out.StreetAddresses
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PostalCodes != nil {
		in, out := &in.PostalCodes, &out.PostalCodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new X509Subject.
func (in *X509Subject) DeepCopy() *X509Subject {
	if in == nil {
		return nil
	}
	out := new(X509Subject)
	in.DeepCopyInto(out)
	return out
}
