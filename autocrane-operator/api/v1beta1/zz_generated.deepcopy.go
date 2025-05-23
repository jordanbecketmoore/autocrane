//go:build !ignore_autogenerated

/*
Copyright 2025.

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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CraneImage) DeepCopyInto(out *CraneImage) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CraneImage.
func (in *CraneImage) DeepCopy() *CraneImage {
	if in == nil {
		return nil
	}
	out := new(CraneImage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CraneImage) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CraneImageList) DeepCopyInto(out *CraneImageList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CraneImage, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CraneImageList.
func (in *CraneImageList) DeepCopy() *CraneImageList {
	if in == nil {
		return nil
	}
	out := new(CraneImageList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CraneImageList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CraneImagePolicy) DeepCopyInto(out *CraneImagePolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CraneImagePolicy.
func (in *CraneImagePolicy) DeepCopy() *CraneImagePolicy {
	if in == nil {
		return nil
	}
	out := new(CraneImagePolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CraneImagePolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CraneImagePolicyList) DeepCopyInto(out *CraneImagePolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CraneImagePolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CraneImagePolicyList.
func (in *CraneImagePolicyList) DeepCopy() *CraneImagePolicyList {
	if in == nil {
		return nil
	}
	out := new(CraneImagePolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CraneImagePolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CraneImagePolicySpec) DeepCopyInto(out *CraneImagePolicySpec) {
	*out = *in
	out.Source = in.Source
	out.Destination = in.Destination
	out.PassthroughCache = in.PassthroughCache
	out.ImagePolicy = in.ImagePolicy
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CraneImagePolicySpec.
func (in *CraneImagePolicySpec) DeepCopy() *CraneImagePolicySpec {
	if in == nil {
		return nil
	}
	out := new(CraneImagePolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CraneImagePolicyStatus) DeepCopyInto(out *CraneImagePolicyStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CraneImagePolicyStatus.
func (in *CraneImagePolicyStatus) DeepCopy() *CraneImagePolicyStatus {
	if in == nil {
		return nil
	}
	out := new(CraneImagePolicyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CraneImageSpec) DeepCopyInto(out *CraneImageSpec) {
	*out = *in
	out.Source = in.Source
	out.Destination = in.Destination
	out.PassthroughCache = in.PassthroughCache
	out.Image = in.Image
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CraneImageSpec.
func (in *CraneImageSpec) DeepCopy() *CraneImageSpec {
	if in == nil {
		return nil
	}
	out := new(CraneImageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CraneImageStatus) DeepCopyInto(out *CraneImageStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CraneImageStatus.
func (in *CraneImageStatus) DeepCopy() *CraneImageStatus {
	if in == nil {
		return nil
	}
	out := new(CraneImageStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImageDetails) DeepCopyInto(out *ImageDetails) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImageDetails.
func (in *ImageDetails) DeepCopy() *ImageDetails {
	if in == nil {
		return nil
	}
	out := new(ImageDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamePolicyDetails) DeepCopyInto(out *NamePolicyDetails) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamePolicyDetails.
func (in *NamePolicyDetails) DeepCopy() *NamePolicyDetails {
	if in == nil {
		return nil
	}
	out := new(NamePolicyDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyDetails) DeepCopyInto(out *PolicyDetails) {
	*out = *in
	out.Name = in.Name
	out.Tag = in.Tag
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyDetails.
func (in *PolicyDetails) DeepCopy() *PolicyDetails {
	if in == nil {
		return nil
	}
	out := new(PolicyDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RegistryDetails) DeepCopyInto(out *RegistryDetails) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RegistryDetails.
func (in *RegistryDetails) DeepCopy() *RegistryDetails {
	if in == nil {
		return nil
	}
	out := new(RegistryDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TagPolicyDetails) DeepCopyInto(out *TagPolicyDetails) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TagPolicyDetails.
func (in *TagPolicyDetails) DeepCopy() *TagPolicyDetails {
	if in == nil {
		return nil
	}
	out := new(TagPolicyDetails)
	in.DeepCopyInto(out)
	return out
}
