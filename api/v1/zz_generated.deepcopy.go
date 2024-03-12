//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2021.

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

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Bridge) DeepCopyInto(out *Bridge) {
	*out = *in
	if in.OVS != nil {
		in, out := &in.OVS, &out.OVS
		*out = new(OVSConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Bridge.
func (in *Bridge) DeepCopy() *Bridge {
	if in == nil {
		return nil
	}
	out := new(Bridge)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Bridges) DeepCopyInto(out *Bridges) {
	*out = *in
	if in.OVS != nil {
		in, out := &in.OVS, &out.OVS
		*out = make([]OVSConfigExt, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Bridges.
func (in *Bridges) DeepCopy() *Bridges {
	if in == nil {
		return nil
	}
	out := new(Bridges)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in ByPriority) DeepCopyInto(out *ByPriority) {
	{
		in := &in
		*out = make(ByPriority, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ByPriority.
func (in ByPriority) DeepCopy() ByPriority {
	if in == nil {
		return nil
	}
	out := new(ByPriority)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Interface) DeepCopyInto(out *Interface) {
	*out = *in
	if in.VfGroups != nil {
		in, out := &in.VfGroups, &out.VfGroups
		*out = make([]VfGroup, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Interface.
func (in *Interface) DeepCopy() *Interface {
	if in == nil {
		return nil
	}
	out := new(Interface)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InterfaceExt) DeepCopyInto(out *InterfaceExt) {
	*out = *in
	if in.VFs != nil {
		in, out := &in.VFs, &out.VFs
		*out = make([]VirtualFunction, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InterfaceExt.
func (in *InterfaceExt) DeepCopy() *InterfaceExt {
	if in == nil {
		return nil
	}
	out := new(InterfaceExt)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in InterfaceExts) DeepCopyInto(out *InterfaceExts) {
	{
		in := &in
		*out = make(InterfaceExts, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InterfaceExts.
func (in InterfaceExts) DeepCopy() InterfaceExts {
	if in == nil {
		return nil
	}
	out := new(InterfaceExts)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Interfaces) DeepCopyInto(out *Interfaces) {
	{
		in := &in
		*out = make(Interfaces, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Interfaces.
func (in Interfaces) DeepCopy() Interfaces {
	if in == nil {
		return nil
	}
	out := new(Interfaces)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OVSBridgeConfig) DeepCopyInto(out *OVSBridgeConfig) {
	*out = *in
	if in.ExternalIDs != nil {
		in, out := &in.ExternalIDs, &out.ExternalIDs
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.OtherConfig != nil {
		in, out := &in.OtherConfig, &out.OtherConfig
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OVSBridgeConfig.
func (in *OVSBridgeConfig) DeepCopy() *OVSBridgeConfig {
	if in == nil {
		return nil
	}
	out := new(OVSBridgeConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OVSConfig) DeepCopyInto(out *OVSConfig) {
	*out = *in
	in.Bridge.DeepCopyInto(&out.Bridge)
	in.Uplink.DeepCopyInto(&out.Uplink)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OVSConfig.
func (in *OVSConfig) DeepCopy() *OVSConfig {
	if in == nil {
		return nil
	}
	out := new(OVSConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OVSConfigExt) DeepCopyInto(out *OVSConfigExt) {
	*out = *in
	in.Bridge.DeepCopyInto(&out.Bridge)
	if in.Uplinks != nil {
		in, out := &in.Uplinks, &out.Uplinks
		*out = make([]OVSUplinkConfigExt, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OVSConfigExt.
func (in *OVSConfigExt) DeepCopy() *OVSConfigExt {
	if in == nil {
		return nil
	}
	out := new(OVSConfigExt)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OVSInterfaceConfig) DeepCopyInto(out *OVSInterfaceConfig) {
	*out = *in
	if in.Options != nil {
		in, out := &in.Options, &out.Options
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ExternalIDs != nil {
		in, out := &in.ExternalIDs, &out.ExternalIDs
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.OtherConfig != nil {
		in, out := &in.OtherConfig, &out.OtherConfig
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OVSInterfaceConfig.
func (in *OVSInterfaceConfig) DeepCopy() *OVSInterfaceConfig {
	if in == nil {
		return nil
	}
	out := new(OVSInterfaceConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OVSNetwork) DeepCopyInto(out *OVSNetwork) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OVSNetwork.
func (in *OVSNetwork) DeepCopy() *OVSNetwork {
	if in == nil {
		return nil
	}
	out := new(OVSNetwork)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OVSNetwork) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OVSNetworkList) DeepCopyInto(out *OVSNetworkList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OVSNetwork, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OVSNetworkList.
func (in *OVSNetworkList) DeepCopy() *OVSNetworkList {
	if in == nil {
		return nil
	}
	out := new(OVSNetworkList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OVSNetworkList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OVSNetworkSpec) DeepCopyInto(out *OVSNetworkSpec) {
	*out = *in
	if in.Trunk != nil {
		in, out := &in.Trunk, &out.Trunk
		*out = make([]*TrunkConfig, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(TrunkConfig)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OVSNetworkSpec.
func (in *OVSNetworkSpec) DeepCopy() *OVSNetworkSpec {
	if in == nil {
		return nil
	}
	out := new(OVSNetworkSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OVSNetworkStatus) DeepCopyInto(out *OVSNetworkStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OVSNetworkStatus.
func (in *OVSNetworkStatus) DeepCopy() *OVSNetworkStatus {
	if in == nil {
		return nil
	}
	out := new(OVSNetworkStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OVSUplinkConfig) DeepCopyInto(out *OVSUplinkConfig) {
	*out = *in
	in.Interface.DeepCopyInto(&out.Interface)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OVSUplinkConfig.
func (in *OVSUplinkConfig) DeepCopy() *OVSUplinkConfig {
	if in == nil {
		return nil
	}
	out := new(OVSUplinkConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OVSUplinkConfigExt) DeepCopyInto(out *OVSUplinkConfigExt) {
	*out = *in
	in.Interface.DeepCopyInto(&out.Interface)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OVSUplinkConfigExt.
func (in *OVSUplinkConfigExt) DeepCopy() *OVSUplinkConfigExt {
	if in == nil {
		return nil
	}
	out := new(OVSUplinkConfigExt)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OvsHardwareOffloadConfig) DeepCopyInto(out *OvsHardwareOffloadConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OvsHardwareOffloadConfig.
func (in *OvsHardwareOffloadConfig) DeepCopy() *OvsHardwareOffloadConfig {
	if in == nil {
		return nil
	}
	out := new(OvsHardwareOffloadConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in PluginNameSlice) DeepCopyInto(out *PluginNameSlice) {
	{
		in := &in
		*out = make(PluginNameSlice, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PluginNameSlice.
func (in PluginNameSlice) DeepCopy() PluginNameSlice {
	if in == nil {
		return nil
	}
	out := new(PluginNameSlice)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovIBNetwork) DeepCopyInto(out *SriovIBNetwork) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovIBNetwork.
func (in *SriovIBNetwork) DeepCopy() *SriovIBNetwork {
	if in == nil {
		return nil
	}
	out := new(SriovIBNetwork)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SriovIBNetwork) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovIBNetworkList) DeepCopyInto(out *SriovIBNetworkList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SriovIBNetwork, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovIBNetworkList.
func (in *SriovIBNetworkList) DeepCopy() *SriovIBNetworkList {
	if in == nil {
		return nil
	}
	out := new(SriovIBNetworkList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SriovIBNetworkList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovIBNetworkSpec) DeepCopyInto(out *SriovIBNetworkSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovIBNetworkSpec.
func (in *SriovIBNetworkSpec) DeepCopy() *SriovIBNetworkSpec {
	if in == nil {
		return nil
	}
	out := new(SriovIBNetworkSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovIBNetworkStatus) DeepCopyInto(out *SriovIBNetworkStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovIBNetworkStatus.
func (in *SriovIBNetworkStatus) DeepCopy() *SriovIBNetworkStatus {
	if in == nil {
		return nil
	}
	out := new(SriovIBNetworkStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetwork) DeepCopyInto(out *SriovNetwork) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetwork.
func (in *SriovNetwork) DeepCopy() *SriovNetwork {
	if in == nil {
		return nil
	}
	out := new(SriovNetwork)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SriovNetwork) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkList) DeepCopyInto(out *SriovNetworkList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SriovNetwork, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkList.
func (in *SriovNetworkList) DeepCopy() *SriovNetworkList {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SriovNetworkList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkNicSelector) DeepCopyInto(out *SriovNetworkNicSelector) {
	*out = *in
	if in.RootDevices != nil {
		in, out := &in.RootDevices, &out.RootDevices
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PfNames != nil {
		in, out := &in.PfNames, &out.PfNames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkNicSelector.
func (in *SriovNetworkNicSelector) DeepCopy() *SriovNetworkNicSelector {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkNicSelector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkNodePolicy) DeepCopyInto(out *SriovNetworkNodePolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkNodePolicy.
func (in *SriovNetworkNodePolicy) DeepCopy() *SriovNetworkNodePolicy {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkNodePolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SriovNetworkNodePolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkNodePolicyList) DeepCopyInto(out *SriovNetworkNodePolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SriovNetworkNodePolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkNodePolicyList.
func (in *SriovNetworkNodePolicyList) DeepCopy() *SriovNetworkNodePolicyList {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkNodePolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SriovNetworkNodePolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkNodePolicySpec) DeepCopyInto(out *SriovNetworkNodePolicySpec) {
	*out = *in
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.NicSelector.DeepCopyInto(&out.NicSelector)
	in.Bridge.DeepCopyInto(&out.Bridge)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkNodePolicySpec.
func (in *SriovNetworkNodePolicySpec) DeepCopy() *SriovNetworkNodePolicySpec {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkNodePolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkNodePolicyStatus) DeepCopyInto(out *SriovNetworkNodePolicyStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkNodePolicyStatus.
func (in *SriovNetworkNodePolicyStatus) DeepCopy() *SriovNetworkNodePolicyStatus {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkNodePolicyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkNodeState) DeepCopyInto(out *SriovNetworkNodeState) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkNodeState.
func (in *SriovNetworkNodeState) DeepCopy() *SriovNetworkNodeState {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkNodeState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SriovNetworkNodeState) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkNodeStateList) DeepCopyInto(out *SriovNetworkNodeStateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SriovNetworkNodeState, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkNodeStateList.
func (in *SriovNetworkNodeStateList) DeepCopy() *SriovNetworkNodeStateList {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkNodeStateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SriovNetworkNodeStateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkNodeStateSpec) DeepCopyInto(out *SriovNetworkNodeStateSpec) {
	*out = *in
	if in.Interfaces != nil {
		in, out := &in.Interfaces, &out.Interfaces
		*out = make(Interfaces, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Bridges.DeepCopyInto(&out.Bridges)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkNodeStateSpec.
func (in *SriovNetworkNodeStateSpec) DeepCopy() *SriovNetworkNodeStateSpec {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkNodeStateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkNodeStateStatus) DeepCopyInto(out *SriovNetworkNodeStateStatus) {
	*out = *in
	if in.Interfaces != nil {
		in, out := &in.Interfaces, &out.Interfaces
		*out = make(InterfaceExts, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Bridges.DeepCopyInto(&out.Bridges)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkNodeStateStatus.
func (in *SriovNetworkNodeStateStatus) DeepCopy() *SriovNetworkNodeStateStatus {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkNodeStateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkPoolConfig) DeepCopyInto(out *SriovNetworkPoolConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkPoolConfig.
func (in *SriovNetworkPoolConfig) DeepCopy() *SriovNetworkPoolConfig {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkPoolConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SriovNetworkPoolConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkPoolConfigList) DeepCopyInto(out *SriovNetworkPoolConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SriovNetworkPoolConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkPoolConfigList.
func (in *SriovNetworkPoolConfigList) DeepCopy() *SriovNetworkPoolConfigList {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkPoolConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SriovNetworkPoolConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkPoolConfigSpec) DeepCopyInto(out *SriovNetworkPoolConfigSpec) {
	*out = *in
	out.OvsHardwareOffloadConfig = in.OvsHardwareOffloadConfig
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkPoolConfigSpec.
func (in *SriovNetworkPoolConfigSpec) DeepCopy() *SriovNetworkPoolConfigSpec {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkPoolConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkPoolConfigStatus) DeepCopyInto(out *SriovNetworkPoolConfigStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkPoolConfigStatus.
func (in *SriovNetworkPoolConfigStatus) DeepCopy() *SriovNetworkPoolConfigStatus {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkPoolConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkSpec) DeepCopyInto(out *SriovNetworkSpec) {
	*out = *in
	if in.MinTxRate != nil {
		in, out := &in.MinTxRate, &out.MinTxRate
		*out = new(int)
		**out = **in
	}
	if in.MaxTxRate != nil {
		in, out := &in.MaxTxRate, &out.MaxTxRate
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkSpec.
func (in *SriovNetworkSpec) DeepCopy() *SriovNetworkSpec {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovNetworkStatus) DeepCopyInto(out *SriovNetworkStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovNetworkStatus.
func (in *SriovNetworkStatus) DeepCopy() *SriovNetworkStatus {
	if in == nil {
		return nil
	}
	out := new(SriovNetworkStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovOperatorConfig) DeepCopyInto(out *SriovOperatorConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovOperatorConfig.
func (in *SriovOperatorConfig) DeepCopy() *SriovOperatorConfig {
	if in == nil {
		return nil
	}
	out := new(SriovOperatorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SriovOperatorConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovOperatorConfigList) DeepCopyInto(out *SriovOperatorConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SriovOperatorConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovOperatorConfigList.
func (in *SriovOperatorConfigList) DeepCopy() *SriovOperatorConfigList {
	if in == nil {
		return nil
	}
	out := new(SriovOperatorConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SriovOperatorConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovOperatorConfigSpec) DeepCopyInto(out *SriovOperatorConfigSpec) {
	*out = *in
	if in.ConfigDaemonNodeSelector != nil {
		in, out := &in.ConfigDaemonNodeSelector, &out.ConfigDaemonNodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.DisablePlugins != nil {
		in, out := &in.DisablePlugins, &out.DisablePlugins
		*out = make(PluginNameSlice, len(*in))
		copy(*out, *in)
	}
	if in.FeatureGates != nil {
		in, out := &in.FeatureGates, &out.FeatureGates
		*out = make(map[string]bool, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovOperatorConfigSpec.
func (in *SriovOperatorConfigSpec) DeepCopy() *SriovOperatorConfigSpec {
	if in == nil {
		return nil
	}
	out := new(SriovOperatorConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SriovOperatorConfigStatus) DeepCopyInto(out *SriovOperatorConfigStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SriovOperatorConfigStatus.
func (in *SriovOperatorConfigStatus) DeepCopy() *SriovOperatorConfigStatus {
	if in == nil {
		return nil
	}
	out := new(SriovOperatorConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrunkConfig) DeepCopyInto(out *TrunkConfig) {
	*out = *in
	if in.MinID != nil {
		in, out := &in.MinID, &out.MinID
		*out = new(uint)
		**out = **in
	}
	if in.MaxID != nil {
		in, out := &in.MaxID, &out.MaxID
		*out = new(uint)
		**out = **in
	}
	if in.ID != nil {
		in, out := &in.ID, &out.ID
		*out = new(uint)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrunkConfig.
func (in *TrunkConfig) DeepCopy() *TrunkConfig {
	if in == nil {
		return nil
	}
	out := new(TrunkConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VfGroup) DeepCopyInto(out *VfGroup) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VfGroup.
func (in *VfGroup) DeepCopy() *VfGroup {
	if in == nil {
		return nil
	}
	out := new(VfGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualFunction) DeepCopyInto(out *VirtualFunction) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualFunction.
func (in *VirtualFunction) DeepCopy() *VirtualFunction {
	if in == nil {
		return nil
	}
	out := new(VirtualFunction)
	in.DeepCopyInto(out)
	return out
}
