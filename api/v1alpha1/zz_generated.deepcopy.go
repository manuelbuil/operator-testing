//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2023.

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

package v1alpha1

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Egressgwk3s) DeepCopyInto(out *Egressgwk3s) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Egressgwk3s.
func (in *Egressgwk3s) DeepCopy() *Egressgwk3s {
	if in == nil {
		return nil
	}
	out := new(Egressgwk3s)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Egressgwk3s) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Egressgwk3sList) DeepCopyInto(out *Egressgwk3sList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Egressgwk3s, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Egressgwk3sList.
func (in *Egressgwk3sList) DeepCopy() *Egressgwk3sList {
	if in == nil {
		return nil
	}
	out := new(Egressgwk3sList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Egressgwk3sList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Egressgwk3sSpec) DeepCopyInto(out *Egressgwk3sSpec) {
	*out = *in
	if in.SourcePods != nil {
		in, out := &in.SourcePods, &out.SourcePods
		*out = make([]SourcePodsSelector, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.DestinationCIDRs != nil {
		in, out := &in.DestinationCIDRs, &out.DestinationCIDRs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Egressgwk3sSpec.
func (in *Egressgwk3sSpec) DeepCopy() *Egressgwk3sSpec {
	if in == nil {
		return nil
	}
	out := new(Egressgwk3sSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Egressgwk3sStatus) DeepCopyInto(out *Egressgwk3sStatus) {
	*out = *in
	if in.Pods != nil {
		in, out := &in.Pods, &out.Pods
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.NodeIP != nil {
		in, out := &in.NodeIP, &out.NodeIP
		*out = make([]v1.NodeAddress, len(*in))
		copy(*out, *in)
	}
	if in.DestinationCIDRs != nil {
		in, out := &in.DestinationCIDRs, &out.DestinationCIDRs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Egressgwk3sStatus.
func (in *Egressgwk3sStatus) DeepCopy() *Egressgwk3sStatus {
	if in == nil {
		return nil
	}
	out := new(Egressgwk3sStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPBlock) DeepCopyInto(out *IPBlock) {
	*out = *in
	if in.Except != nil {
		in, out := &in.Except, &out.Except
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPBlock.
func (in *IPBlock) DeepCopy() *IPBlock {
	if in == nil {
		return nil
	}
	out := new(IPBlock)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SourcePodsSelector) DeepCopyInto(out *SourcePodsSelector) {
	*out = *in
	in.NamespaceSelector.DeepCopyInto(&out.NamespaceSelector)
	in.PodSelector.DeepCopyInto(&out.PodSelector)
	if in.IpBlock != nil {
		in, out := &in.IpBlock, &out.IpBlock
		*out = new(IPBlock)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SourcePodsSelector.
func (in *SourcePodsSelector) DeepCopy() *SourcePodsSelector {
	if in == nil {
		return nil
	}
	out := new(SourcePodsSelector)
	in.DeepCopyInto(out)
	return out
}
