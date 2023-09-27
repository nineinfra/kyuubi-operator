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

package controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kyuubiv1alpha1 "github.com/nineinfra/kyuubi-operator/api/v1alpha1"
)

// KyuubiClusterReconciler reconciles a KyuubiCluster object
type KyuubiClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kyuubi.nineinfra.tech,resources=kyuubiclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kyuubi.nineinfra.tech,resources=kyuubiclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kyuubi.nineinfra.tech,resources=kyuubiclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KyuubiCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *KyuubiClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var kyuubi kyuubiv1alpha1.KyuubiCluster
	err := r.Get(ctx, req.NamespacedName, &kyuubi)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Object not found, it could have been deleted")
		} else {
			logger.Info("Error occurred during fetching the object")
		}
		return ctrl.Result{}, err
	}
	requestArray := strings.Split(fmt.Sprint(req), "/")
	requestName := requestArray[1]

	if requestName == kyuubi.Name {
		logger.Info("create Or Update Clusters")
		err = r.createOrUpdateClusters(ctx, &kyuubi, logger)
		if err != nil {
			logger.Info("Error occurred during create Or Update Components")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateClusters(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, logger logr.Logger) error {
	err := r.createOrUpdateK8sResources(ctx, kyuubi, logger)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateK8sResources")
		return err
	}

	err = r.createOrUpdateKyuubiConfigmap(ctx, kyuubi, logger)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateKyuubiConfigmap")
		return err
	}

	err = r.createOrUpdateClusterRefsConfigmap(ctx, kyuubi, logger)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateClusterRefsConfigmap")
		return err
	}

	err = r.createOrUpdateKyuubi(ctx, kyuubi, logger)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateKyuubi")
		return err
	}

	existingService, err := r.createOrUpdateService(ctx, kyuubi, logger)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateService")
		return err
	}

	err = r.updateKyuubiClusterStatus(ctx, kyuubi, existingService, logger)
	if err != nil {
		logger.Error(err, "Error occurred during updateKyuubiClusterStatus")
		return err
	}
	return nil
}

func (r *KyuubiClusterReconciler) updateKyuubiClusterStatus(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, kyuubiService *corev1.Service, logger logr.Logger) error {
	exposedInfos := make([]kyuubiv1alpha1.ExposedInfo, 0)
	for k, v := range kyuubiService.Spec.Ports {
		var exposedInfo kyuubiv1alpha1.ExposedInfo
		exposedInfo.Name = kyuubi.Name + "-" + strconv.Itoa(k)
		exposedInfo.ExposedType = kyuubiv1alpha1.ExposedType(v.Name)
		v.DeepCopyInto(&exposedInfo.ServicePort)
		exposedInfos = append(exposedInfos, exposedInfo)
	}

	desiredKyuubiStatus := &kyuubiv1alpha1.KyuubiClusterStatus{
		ExposedInfos: exposedInfos,
	}

	if kyuubi.Status.ExposedInfos == nil || !reflect.DeepEqual(exposedInfos, kyuubi.Status.ExposedInfos) {
		if kyuubi.Status.ExposedInfos == nil {
			desiredKyuubiStatus.CreationTime = metav1.Now()
		} else {
			desiredKyuubiStatus.CreationTime = kyuubi.Status.CreationTime
		}
		desiredKyuubiStatus.UpdateTime = metav1.Now()
		desiredKyuubiStatus.DeepCopyInto(&kyuubi.Status)
		err := r.Status().Update(ctx, kyuubi)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *KyuubiClusterReconciler) constructServiceAccount(kyuubi *kyuubiv1alpha1.KyuubiCluster) (*corev1.ServiceAccount, error) {
	saDesired := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kyuubi.Name + "-kyuubi",
			Namespace: kyuubi.Namespace,
			Labels: map[string]string{
				"cluster": kyuubi.Name,
				"app":     "kyuubi",
			},
		},
	}

	if err := ctrl.SetControllerReference(kyuubi, saDesired, r.Scheme); err != nil {
		return saDesired, err
	}

	return saDesired, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateServiceAccount(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, logger logr.Logger) error {
	desiredKyuubiSa, _ := r.constructServiceAccount(kyuubi)

	existingKyuubeSa := &corev1.ServiceAccount{}

	err := r.Get(ctx, client.ObjectKeyFromObject(desiredKyuubiSa), existingKyuubeSa)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		if err := r.Create(ctx, desiredKyuubiSa); err != nil {
			return err
		}
	} else if !reflect.DeepEqual(existingKyuubeSa, desiredKyuubiSa) {
		logger.Info("updating kyuubi ServiceAccount")
	}
	return nil
}

func (r *KyuubiClusterReconciler) constructRole(kyuubi *kyuubiv1alpha1.KyuubiCluster) (*rbacv1.Role, error) {
	roleDesired := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kyuubi.Name + "-kyuubi",
			Namespace: kyuubi.Namespace,
			Labels: map[string]string{
				"cluster": kyuubi.Name,
				"app":     "kyuubi",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"pods",
					"configmaps",
					"services",
					"persistentvolumeclaims",
				},
				Verbs: []string{
					"create",
					"list",
					"delete",
					"watch",
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(kyuubi, roleDesired, r.Scheme); err != nil {
		return roleDesired, err
	}

	return roleDesired, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateRole(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, logger logr.Logger) error {
	desiredKyuubiRole, _ := r.constructRole(kyuubi)

	existingKyuubeRole := &rbacv1.Role{}

	err := r.Get(ctx, client.ObjectKeyFromObject(desiredKyuubiRole), existingKyuubeRole)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		if err := r.Create(ctx, desiredKyuubiRole); err != nil {
			return err
		}
	} else if !reflect.DeepEqual(existingKyuubeRole, desiredKyuubiRole) {
		logger.Info("updating kyuubi role")
	}
	return nil
}

func (r *KyuubiClusterReconciler) constructRoleBinding(kyuubi *kyuubiv1alpha1.KyuubiCluster) (*rbacv1.RoleBinding, error) {
	roleBindingDesired := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kyuubi.Name + "-kyuubi",
			Namespace: kyuubi.Namespace,
			Labels: map[string]string{
				"cluster": kyuubi.Name,
				"app":     "kyuubi",
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: kyuubi.Name + "-kyuubi",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     kyuubi.Name + "-kyuubi",
		},
	}

	if err := ctrl.SetControllerReference(kyuubi, roleBindingDesired, r.Scheme); err != nil {
		return roleBindingDesired, err
	}

	return roleBindingDesired, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateRoleBinding(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, logger logr.Logger) error {
	desiredKyuubiRoleBinding, _ := r.constructRoleBinding(kyuubi)

	existingKyuubiRoleBinding := &rbacv1.RoleBinding{}

	err := r.Get(ctx, client.ObjectKeyFromObject(desiredKyuubiRoleBinding), existingKyuubiRoleBinding)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		if err := r.Create(ctx, desiredKyuubiRoleBinding); err != nil {
			return err
		}
	} else if !reflect.DeepEqual(existingKyuubiRoleBinding, desiredKyuubiRoleBinding) {
		logger.Info("updating kyuubi rolebinding")
	}
	return nil
}

func (r *KyuubiClusterReconciler) createOrUpdateK8sResources(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, logger logr.Logger) error {
	err := r.createOrUpdateServiceAccount(ctx, kyuubi, logger)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateServiceAccount")
		return err
	}

	err = r.createOrUpdateRole(ctx, kyuubi, logger)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateRole")
		return err
	}

	err = r.createOrUpdateRoleBinding(ctx, kyuubi, logger)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateRoleBinding")
		return err
	}

	return nil
}

func (r *KyuubiClusterReconciler) constructDesiredKyuubiWorkload(kyuubi *kyuubiv1alpha1.KyuubiCluster) (*appsv1.StatefulSet, error) {
	stsDesired := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kyuubi.Name + "-kyuubi",
			Namespace: kyuubi.Namespace,
			Labels: map[string]string{
				"cluster": kyuubi.Name,
				"app":     "kyuubi",
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"cluster": kyuubi.Name,
					"app":     "kyuubi",
				},
			},
			ServiceName: kyuubi.Name + "-kyuubi",
			Replicas:    int32Ptr(kyuubi.Spec.KyuubiResource.Replicas),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"cluster": kyuubi.Name,
						"app":     "kyuubi",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  kyuubi.Name,
							Image: kyuubi.Spec.KyuubiImage.Repository + ":" + kyuubi.Spec.KyuubiImage.Tag,
							Ports: []corev1.ContainerPort{
								{
									Name:          "rest",
									ContainerPort: int32(10099),
								},
								{
									Name:          "thrift-binary",
									ContainerPort: int32(10009),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"/bin/bash",
											"-c",
											"bin/kyuubi status",
										},
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								TimeoutSeconds:      2,
								FailureThreshold:    10,
								SuccessThreshold:    1,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"/bin/bash",
											"-c",
											"$KYUUBI_HOME/bin/kyuubi status"},
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								TimeoutSeconds:      2,
								FailureThreshold:    10,
								SuccessThreshold:    1,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      kyuubi.Name + "-kyuubi",
									MountPath: "/opt/kyuubi/conf/kyuubi-defaults.conf",
									SubPath:   "kyuubi-defaults.conf",
								},
								{
									Name:      kyuubi.Name + "-spark",
									MountPath: "/opt/spark/conf/spark-defaults.conf",
									SubPath:   "spark-defaults.conf",
								},
								{
									Name:      kyuubi.Name + "-hdfssite",
									MountPath: "/opt/spark/conf/hdfs-site.xml",
									SubPath:   "hdfs-site.xml",
								},
								{
									Name:      kyuubi.Name + "-coresite",
									MountPath: "/opt/spark/conf/core-site.xml",
									SubPath:   "core-site.xml",
								},
								{
									Name:      kyuubi.Name + "-hivesite",
									MountPath: "/opt/spark/conf/hive-site.xml",
									SubPath:   "hive-site.xml",
								},
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyAlways,
					ServiceAccountName: kyuubi.Name + "-kyuubi",
					Volumes: []corev1.Volume{
						{
							Name: kyuubi.Name + "-kyuubi",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: kyuubi.Name + "-kyuubi",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "kyuubi-defaults.conf",
											Path: "kyuubi-defaults.conf",
										},
									},
								},
							},
						},
						{
							Name: kyuubi.Name + "-spark",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: kyuubi.Name + "-clusterrefs",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "spark-defaults.conf",
											Path: "spark-defaults.conf",
										},
									},
								},
							},
						},
						{
							Name: kyuubi.Name + "-hdfssite",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: kyuubi.Name + "-clusterrefs",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "hdfs-site.xml",
											Path: "hdfs-site.xml",
										},
									},
								},
							},
						},
						{
							Name: kyuubi.Name + "-coresite",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: kyuubi.Name + "-clusterrefs",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "core-site.xml",
											Path: "core-site.xml",
										},
									},
								},
							},
						},
						{
							Name: kyuubi.Name + "-hivesite",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: kyuubi.Name + "-clusterrefs",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "hive-site.xml",
											Path: "hive-site.xml",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(kyuubi, stsDesired, r.Scheme); err != nil {
		return stsDesired, err
	}
	return stsDesired, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateKyuubi(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, logger logr.Logger) error {

	desiredKyuubeWorkload, _ := r.constructDesiredKyuubiWorkload(kyuubi)

	existingKyuubeWorkload := &appsv1.StatefulSet{}

	err := r.Get(ctx, client.ObjectKeyFromObject(desiredKyuubeWorkload), existingKyuubeWorkload)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		if err := r.Create(ctx, desiredKyuubeWorkload); err != nil {
			return err
		}
	} else if !reflect.DeepEqual(existingKyuubeWorkload, desiredKyuubeWorkload) {
		logger.Info("updating kyuubi workload")
	}
	return nil
}

func (r *KyuubiClusterReconciler) desiredClusterRefsConfigMap(kyuubi *kyuubiv1alpha1.KyuubiCluster) (*corev1.ConfigMap, error) {
	cmDesired := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kyuubi.Name + "-clusterrefs",
			Namespace: kyuubi.Namespace,
			Labels: map[string]string{
				"cluster": kyuubi.Name,
				"app":     kyuubi.Name,
			},
		},
		Data: map[string]string{},
	}

	for _, cluster := range kyuubi.Spec.ClusterRefs {
		switch cluster.Type {
		case kyuubiv1alpha1.SparkClusterType:
			sparkConf := make(map[string]string)
			if cluster.Spark.SparkImage.Tag == "" {
				cluster.Spark.SparkImage.Tag = "latest"
			}
			sparkConf["spark.kubernetes.container.image"] = cluster.Spark.SparkImage.Repository + ":" + cluster.Spark.SparkImage.Tag
			sparkConf["spark.kubernetes.namespace"] = kyuubi.Namespace
			cmDesired.Data["spark-defaults.conf"] = map2String(sparkConf)
		case kyuubiv1alpha1.HdfsClusterType:
			cmDesired.Data["hdfs-site.xml"] = map2Xml(cluster.Hdfs.HdfsSite)
			cmDesired.Data["core-site.xml"] = map2Xml(cluster.Hdfs.CoreSite)
		case kyuubiv1alpha1.MetaStoreClusterType:
			cmDesired.Data["hive-site.xml"] = map2Xml(cluster.Metastore.HiveSite)
		}
	}

	if err := ctrl.SetControllerReference(kyuubi, cmDesired, r.Scheme); err != nil {
		return cmDesired, err
	}

	return cmDesired, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateClusterRefsConfigmap(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, logger logr.Logger) error {
	desiredConfigMap, _ := r.desiredClusterRefsConfigMap(kyuubi)

	existingConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKeyFromObject(desiredConfigMap), existingConfigMap)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Error occurred during Get configmap")
		return err
	}

	// Create or update the Service
	if errors.IsNotFound(err) {
		if err := r.Create(ctx, desiredConfigMap); err != nil {
			logger.Error(err, "Error occurred during Create configmap")
			return err
		}
		for {
			existingConfigMap := &corev1.ConfigMap{}
			err := r.Get(ctx, client.ObjectKeyFromObject(desiredConfigMap), existingConfigMap)
			if err != nil && !errors.IsNotFound(err) {
				return err
			}
			if errors.IsNotFound(err) {
				logger.Info("waiting to create cluster config ...")
				time.Sleep(100 * time.Millisecond)
			} else {
				break
			}
		}
	} else if !compareConf(desiredConfigMap.Data, existingConfigMap.Data) {
		logger.Info("updating clusterrefs configmap")
		existingConfigMap.Data = desiredConfigMap.Data
		if err := r.Update(ctx, existingConfigMap); err != nil {
			return err
		}
	}

	return nil
}

func (r *KyuubiClusterReconciler) constructKyuubiConfigMap(kyuubi *kyuubiv1alpha1.KyuubiCluster) (*corev1.ConfigMap, error) {

	cmDesired := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kyuubi.Name + "-kyuubi",
			Namespace: kyuubi.Namespace,
			Labels: map[string]string{
				"cluster": kyuubi.Name,
				"app":     "kyuubi",
			},
		},
		Data: map[string]string{
			"kyuubi-defaults.conf": map2String(kyuubi.Spec.KyuubiConf),
		},
	}

	if err := ctrl.SetControllerReference(kyuubi, cmDesired, r.Scheme); err != nil {
		return cmDesired, err
	}

	return cmDesired, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateKyuubiConfigmap(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, logger logr.Logger) error {
	desiredConfigMap, _ := r.constructKyuubiConfigMap(kyuubi)

	existingConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKeyFromObject(desiredConfigMap), existingConfigMap)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Error occurred during Get configmap")
		return err
	}

	// Create or update the Configmap
	if errors.IsNotFound(err) {
		if err := r.Create(ctx, desiredConfigMap); err != nil {
			logger.Error(err, "Error occurred during Create configmap")
			return err
		}
		for {
			existingConfigMap := &corev1.ConfigMap{}
			err := r.Get(ctx, client.ObjectKeyFromObject(desiredConfigMap), existingConfigMap)
			if err != nil && !errors.IsNotFound(err) {
				return err
			}
			if errors.IsNotFound(err) {
				logger.Info("waiting to create cluster config ...")
				time.Sleep(100 * time.Millisecond)
			} else {
				break
			}
		}
	} else if !compareConf(desiredConfigMap.Data, existingConfigMap.Data) {
		logger.Info("updating kyuubi configmap")
		existingConfigMap.Data = desiredConfigMap.Data
		if err := r.Update(ctx, existingConfigMap); err != nil {
			return err
		}
	}

	return nil
}

func (r *KyuubiClusterReconciler) contructDesiredService(kyuubi *kyuubiv1alpha1.KyuubiCluster) (*corev1.Service, error) {
	svcDesired := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kyuubi.Name + "-kyuubi",
			Namespace: kyuubi.Namespace,
			Labels: map[string]string{
				"cluster": kyuubi.Name,
				"app":     "kyuubi",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name: string(kyuubiv1alpha1.KyuubiRest),
					Port: 10099,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(10099),
					},
				},
				{
					Name: string(kyuubiv1alpha1.KyuubiThriftBinary),
					Port: 10009,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(10009),
					},
				},
			},
			Selector: map[string]string{
				"cluster": kyuubi.Name,
				"app":     "kyuubi",
			},
		},
	}

	if err := ctrl.SetControllerReference(kyuubi, svcDesired, r.Scheme); err != nil {
		return svcDesired, err
	}

	return svcDesired, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateService(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, logger logr.Logger) (*corev1.Service, error) {
	desiredService, _ := r.contructDesiredService(kyuubi)

	existingService := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKeyFromObject(desiredService), existingService)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	if errors.IsNotFound(err) {
		if err := r.Create(ctx, desiredService); err != nil {
			return desiredService, err
		}
	} else if !reflect.DeepEqual(desiredService, existingService) {
		logger.Info("updating service")
		desiredService.Spec.DeepCopyInto(&existingService.Spec)
		if err := r.Update(ctx, existingService); err != nil {
			return existingService, err
		}
	}
	return existingService, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KyuubiClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kyuubiv1alpha1.KyuubiCluster{}).
		Complete(r)
}
