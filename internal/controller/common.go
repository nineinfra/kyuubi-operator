package controller

import (
	kyuubiv1alpha1 "github.com/nineinfra/kyuubi-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func DefaultDownwardAPI() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: "NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name: "POD_UID",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.uid",
				},
			},
		},
		{
			Name: "HOST_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.hostIP",
				},
			},
		},
	}
}

func ClusterResourceName(cluster *kyuubiv1alpha1.KyuubiCluster, suffixs ...string) string {
	return cluster.Name + DefaultNameSuffix + strings.Join(suffixs, "-")
}

func ClusterResourceLabels(cluster *kyuubiv1alpha1.KyuubiCluster) map[string]string {
	return map[string]string{
		"cluster": cluster.Name,
		"app":     DefaultClusterSign,
	}
}
