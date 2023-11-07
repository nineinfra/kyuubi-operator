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
	corev1 "k8s.io/api/core/v1"
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
	// The replicas of the kyuubi cluster workload.
	Replicas int32 `json:"replicas"`
}

type SparkCluster struct {
	// +optional
	// SparkMaster info. Used in the spark-submit operation
	SparkMaster string `json:"sparkMaster,omitempty"`
	// +optional
	// Spark image info.
	SparkImage ImageConfig `json:"sparkImage"`
	// +optional
	// Spark namespace.
	SparkNamespace string `json:"sparkNamespace"`
	// +optional
	// Spark defaults conf
	SparkDefaults map[string]string `json:"sparkDefaults"`
}

type MetastoreCluster struct {
	// +optional
	// Hive Metastore hive-site.xml.The type of the value must be string
	HiveSite map[string]string `json:"hiveSite,omitempty"`
}

type HdfsCluster struct {
	// +optional
	// HDFS core-site.xml.The type of the value must be string
	CoreSite map[string]string `json:"coreSite,omitempty"`
	// +optional
	// HDFS hdfs-site.xml.The type of the value must be string
	HdfsSite map[string]string `json:"hdfsSite,omitempty"`
}

type ClusterRef struct {
	// Name is the name of the referenced cluster
	Name string `json:"name"`
	// +kubebuilder:validation:Enum={spark,metastore,hdfs,flink}
	Type ClusterType `json:"type"`
	// ClusterInfo is the detail info of the cluster of the clustertype
	// +optional
	// Spark cluster infos referenced by the Kyuubi cluster
	Spark SparkCluster `json:"spark"`
	// +optional
	// Metastore cluster infos referenced by the Kyuubi cluster
	Metastore MetastoreCluster `json:"metastore"`
	// +optional
	// HDFS cluster infos referenced by the Kyuubi cluster
	Hdfs HdfsCluster `json:"hdfs"`
}

// KyuubiClusterSpec defines the desired state of KyuubiCluster
type KyuubiClusterSpec struct {
	//Kyuubi version
	KyuubiVersion string `json:"kyuubiVersion"`
	//Kyuubi image info
	KyuubiImage ImageConfig `json:"kyuubiImage"`
	//Kyuubi resouce configuration
	KyuubiResource ResourceConfig `json:"kyuubiResource"`
	//Kyuubi configurations.These cofigurations will be injected into the kyuubi-defatuls.conf.The type of the value must be string
	KyuubiConf  map[string]string `json:"kyuubiConf"`
	ClusterRefs []ClusterRef      `json:"clusterRefs"`
}

type ExposedType string

const (
	KyuubiRest         ExposedType = "rest"
	KyuubiThriftBinary ExposedType = "thrift-binary"
	KyuubiThriftHttp   ExposedType = "thrift-http"
	KyuubiMysql        ExposedType = "mysql"
)

type ExposedInfo struct {
	//Exposed name.
	Name string `json:"name"`
	//Exposed type. Support REST and THRIFT_BINARY
	ExposedType ExposedType `json:"exposedType"`
	//Exposed service name
	ServiceName string `json:"serviceName"`
	//Exposed service port info
	ServicePort corev1.ServicePort `json:"servicePort"`
}

// KyuubiClusterStatus defines the observed state of KyuubiCluster
type KyuubiClusterStatus struct {
	//Kyuubi exposedinfos
	ExposedInfos []ExposedInfo `json:"exposedInfos"`
	CreationTime metav1.Time   `json:"creationTime"`
	UpdateTime   metav1.Time   `json:"updateTime"`
}

// +genclient
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
