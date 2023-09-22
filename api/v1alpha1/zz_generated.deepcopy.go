//go:build !ignore_autogenerated

/*
Copyright 2023 nineinfra.

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
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterRef) DeepCopyInto(out *ClusterRef) {
	*out = *in
	out.SparkClusterInfo = in.SparkClusterInfo
	out.MetastoreClusterInfo = in.MetastoreClusterInfo
	out.HdfsClusterInfo = in.HdfsClusterInfo
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterRef.
func (in *ClusterRef) DeepCopy() *ClusterRef {
	if in == nil {
		return nil
	}
	out := new(ClusterRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonConfig) DeepCopyInto(out *CommonConfig) {
	*out = *in
	out.OperatorImageConfig = in.OperatorImageConfig
	out.KyuubiImageConfig = in.KyuubiImageConfig
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonConfig.
func (in *CommonConfig) DeepCopy() *CommonConfig {
	if in == nil {
		return nil
	}
	out := new(CommonConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExposedInfo) DeepCopyInto(out *ExposedInfo) {
	*out = *in
	out.Service = in.Service
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExposedInfo.
func (in *ExposedInfo) DeepCopy() *ExposedInfo {
	if in == nil {
		return nil
	}
	out := new(ExposedInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HdfsClusterInfo) DeepCopyInto(out *HdfsClusterInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HdfsClusterInfo.
func (in *HdfsClusterInfo) DeepCopy() *HdfsClusterInfo {
	if in == nil {
		return nil
	}
	out := new(HdfsClusterInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImageConfig) DeepCopyInto(out *ImageConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImageConfig.
func (in *ImageConfig) DeepCopy() *ImageConfig {
	if in == nil {
		return nil
	}
	out := new(ImageConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *K8sService) DeepCopyInto(out *K8sService) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new K8sService.
func (in *K8sService) DeepCopy() *K8sService {
	if in == nil {
		return nil
	}
	out := new(K8sService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KyuubiCluster) DeepCopyInto(out *KyuubiCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KyuubiCluster.
func (in *KyuubiCluster) DeepCopy() *KyuubiCluster {
	if in == nil {
		return nil
	}
	out := new(KyuubiCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KyuubiCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KyuubiClusterList) DeepCopyInto(out *KyuubiClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KyuubiCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KyuubiClusterList.
func (in *KyuubiClusterList) DeepCopy() *KyuubiClusterList {
	if in == nil {
		return nil
	}
	out := new(KyuubiClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KyuubiClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KyuubiClusterSpec) DeepCopyInto(out *KyuubiClusterSpec) {
	*out = *in
	out.CommonConfig = in.CommonConfig
	if in.ClusterRefs != nil {
		in, out := &in.ClusterRefs, &out.ClusterRefs
		*out = make([]ClusterRef, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KyuubiClusterSpec.
func (in *KyuubiClusterSpec) DeepCopy() *KyuubiClusterSpec {
	if in == nil {
		return nil
	}
	out := new(KyuubiClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KyuubiClusterStatus) DeepCopyInto(out *KyuubiClusterStatus) {
	*out = *in
	if in.ExposedInfos != nil {
		in, out := &in.ExposedInfos, &out.ExposedInfos
		*out = make([]ExposedInfo, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KyuubiClusterStatus.
func (in *KyuubiClusterStatus) DeepCopy() *KyuubiClusterStatus {
	if in == nil {
		return nil
	}
	out := new(KyuubiClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetastoreClusterInfo) DeepCopyInto(out *MetastoreClusterInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetastoreClusterInfo.
func (in *MetastoreClusterInfo) DeepCopy() *MetastoreClusterInfo {
	if in == nil {
		return nil
	}
	out := new(MetastoreClusterInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SparkClusterInfo) DeepCopyInto(out *SparkClusterInfo) {
	*out = *in
	out.SparkImageConfig = in.SparkImageConfig
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SparkClusterInfo.
func (in *SparkClusterInfo) DeepCopy() *SparkClusterInfo {
	if in == nil {
		return nil
	}
	out := new(SparkClusterInfo)
	in.DeepCopyInto(out)
	return out
}
