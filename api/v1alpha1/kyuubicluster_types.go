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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DeployMode describes the type of deployment of a Spark application.
type DeployMode string

// Different types of deployments.
const (
	ClusterMode         DeployMode = "cluster"
	ClientMode          DeployMode = "client"
	InClusterClientMode DeployMode = "in-cluster-client"
)

type ClusterType string

// Different types of clusters.
const (
	SparkClusterType     ClusterType = "spark"
	MetaStoreClusterType ClusterType = "metastore"
	HdfsClusterType      ClusterType = "hdfs"
	FlinkClusterType     ClusterType = "flink"
)

type ImageConfig struct {
	Repository string `json:"repository"`
	// Image tag. Usually the vesion of the kyuubi, default: `latest`.
	// +optional
	Tag string `json:"tag,omitempty"`
	// Image pull policy. One of `Always, Never, IfNotPresent`, default: `Always`.
	// +kubebuilder:default:=Always
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	// +optional
	PullPolicy string `json:"pullPolicy,omitempty"`
	// Secrets for image pull.
	// +optional
	PullSecrets string `json:"pullSecret,omitempty"`
}

type ResourceConfig struct {
	Replicas int32 `json:"replicas"`
}

type SparkCluster struct {
	SparkMaster    string      `json:"sparkMaster,omitempty"`
	SparkImage     ImageConfig `json:"sparkImage"`
	SparkNamespace string      `json:"sparkNamespace"`
}

type MetastoreCluster struct {
	// +optional
	HiveSite map[string]string `json:"hiveSizte,omitempty"`
}

type HdfsCluster struct {
	// +optional
	CoreSite map[string]string `json:"coreSite,omitempty"`
	// +optional
	HdfsSite map[string]string `json:"hdfsSite,omitempty"`
}

type ClusterRef struct {
	// Name is the name of the referenced cluster
	Name string `json:"name"`
	// +kubebuilder:validation:Enum={spark,metastore,hdfs,flink}
	ClusterType ClusterType `json:"clusterType"`
	// ClusterInfo is the detail info of the cluster of the clustertype
	// +optional
	SparkCluster SparkCluster `json:"sparkCluster"`
	// +optional
	MetastoreCluster MetastoreCluster `json:"metastoreCluster"`
	// +optional
	HdfsCluster HdfsCluster `json:"hdfsCluster"`
}

// KyuubiClusterSpec defines the desired state of KyuubiCluster
type KyuubiClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	KyuubiVersion string `json:"kyuubiVersion"`
	// OperatorImage  ImageConfig       `json:"operatorImage"`
	KyuubiImage    ImageConfig       `json:"kyuubiImage"`
	KyuubiResource ResourceConfig    `json:"kyuubiResource"`
	KyuubiConf     map[string]string `json:"kyuubiConf"`
	ClusterRefs    []ClusterRef      `json:"clusterRefs"`
}

type ExposedType string

const (
	KyuubiRest         ExposedType = "REST"
	KyuubiThriftBinary ExposedType = "THRIFT_BINARY"
	KyuubiThriftHttp   ExposedType = "THRIFT_HTTP"
	KyuubiMysql        ExposedType = "MYSQL"
)

type K8sServiceType string

const (
	ClusterIP K8sServiceType = "ClusterIP"
	NodePort  K8sServiceType = "NodePort"
)

type K8sService struct {
	Name string `json:"name"`
	// +kubebuilder:validation:Enum={spark,metastore,hdfs,flink}
	Type K8sServiceType `json:"type"`
	Port int            `json:"port"`
}
type ExposedInfo struct {
	Name        string      `json:"name"`
	ExposedType ExposedType `json:"exposedType"`
	Port        int         `json:"port"`
	Service     K8sService  `json:"service"`
}

// KyuubiClusterStatus defines the observed state of KyuubiCluster
type KyuubiClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ExposedInfos []ExposedInfo `json:"exposedInfos"`
	CreationTime metav1.Time   `json:"creationTime"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KyuubiCluster is the Schema for the kyuubiclusters API
type KyuubiCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KyuubiClusterSpec   `json:"spec,omitempty"`
	Status KyuubiClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KyuubiClusterList contains a list of KyuubiCluster
type KyuubiClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KyuubiCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KyuubiCluster{}, &KyuubiClusterList{})
}
