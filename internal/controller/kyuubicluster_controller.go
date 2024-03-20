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
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"os"
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
		if k8serrors.IsNotFound(err) {
			logger.Info("Object not found, it could have been deleted")
		} else {
			logger.Info("Error occurred during fetching the object")
		}
		return ctrl.Result{}, err
	}
	requestArray := strings.Split(fmt.Sprint(req), "/")
	requestName := requestArray[1]

	if requestName == kyuubi.Name {
		logger.Info("create or update clusters")
		err = r.createOrUpdateClusters(ctx, &kyuubi, logger)
		if err != nil {
			logger.Info("Error occurred during create or update clusters")
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

	existingService, err := r.createOrUpdateService(ctx, kyuubi)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateService")
		return err
	}

	err = r.updateKyuubiClusterStatus(ctx, kyuubi, existingService)
	if err != nil {
		logger.Error(err, "Error occurred during updateKyuubiClusterStatus")
		return err
	}
	return nil
}

func (r *KyuubiClusterReconciler) updateKyuubiClusterStatus(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, kyuubiService *corev1.Service) error {
	exposedInfos := make([]kyuubiv1alpha1.ExposedInfo, 0)
	for k, v := range kyuubiService.Spec.Ports {
		var exposedInfo kyuubiv1alpha1.ExposedInfo
		exposedInfo.ServiceName = kyuubiService.Name
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
			Name:      ClusterResourceName(kyuubi),
			Namespace: kyuubi.Namespace,
			Labels:    ClusterResourceLabels(kyuubi),
		},
	}

	if err := ctrl.SetControllerReference(kyuubi, saDesired, r.Scheme); err != nil {
		return saDesired, err
	}

	return saDesired, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateServiceAccount(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster) error {
	desiredKyuubiSa, _ := r.constructServiceAccount(kyuubi)

	existingKyuubeSa := &corev1.ServiceAccount{}

	err := r.Get(ctx, client.ObjectKeyFromObject(desiredKyuubiSa), existingKyuubeSa)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if k8serrors.IsNotFound(err) {
		if err := r.Create(ctx, desiredKyuubiSa); err != nil {
			return err
		}
	}
	return nil
}

func (r *KyuubiClusterReconciler) constructRole(kyuubi *kyuubiv1alpha1.KyuubiCluster) (*rbacv1.Role, error) {
	roleDesired := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterResourceName(kyuubi),
			Namespace: kyuubi.Namespace,
			Labels:    ClusterResourceLabels(kyuubi),
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
					"get",
					"create",
					"list",
					"delete",
					"watch",
					"deletecollection",
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(kyuubi, roleDesired, r.Scheme); err != nil {
		return roleDesired, err
	}

	return roleDesired, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateRole(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster) error {
	desiredKyuubiRole, _ := r.constructRole(kyuubi)

	existingKyuubeRole := &rbacv1.Role{}

	err := r.Get(ctx, client.ObjectKeyFromObject(desiredKyuubiRole), existingKyuubeRole)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if k8serrors.IsNotFound(err) {
		if err := r.Create(ctx, desiredKyuubiRole); err != nil {
			return err
		}
	}
	return nil
}

func (r *KyuubiClusterReconciler) constructRoleBinding(kyuubi *kyuubiv1alpha1.KyuubiCluster) (*rbacv1.RoleBinding, error) {
	roleBindingDesired := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterResourceName(kyuubi),
			Namespace: kyuubi.Namespace,
			Labels:    ClusterResourceLabels(kyuubi),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: ClusterResourceName(kyuubi),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     ClusterResourceName(kyuubi),
		},
	}

	if err := ctrl.SetControllerReference(kyuubi, roleBindingDesired, r.Scheme); err != nil {
		return roleBindingDesired, err
	}

	return roleBindingDesired, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateRoleBinding(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster) error {
	desiredKyuubiRoleBinding, _ := r.constructRoleBinding(kyuubi)

	existingKyuubiRoleBinding := &rbacv1.RoleBinding{}

	err := r.Get(ctx, client.ObjectKeyFromObject(desiredKyuubiRoleBinding), existingKyuubiRoleBinding)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if k8serrors.IsNotFound(err) {
		if err := r.Create(ctx, desiredKyuubiRoleBinding); err != nil {
			return err
		}
	}
	return nil
}

func (r *KyuubiClusterReconciler) createOrUpdateK8sResources(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, logger logr.Logger) error {
	err := r.createOrUpdateServiceAccount(ctx, kyuubi)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateServiceAccount")
		return err
	}

	err = r.createOrUpdateRole(ctx, kyuubi)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateRole")
		return err
	}

	err = r.createOrUpdateRoleBinding(ctx, kyuubi)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateRoleBinding")
		return err
	}

	return nil
}

func (r *KyuubiClusterReconciler) constructVolumeMounts(kyuubi *kyuubiv1alpha1.KyuubiCluster) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      ClusterResourceName(kyuubi),
			MountPath: fmt.Sprintf("%s/%s", DefaultKyuubiConfDir, DefaultKyuubiConfFile),
			SubPath:   DefaultKyuubiConfFile,
		},
	}
	for _, cluster := range kyuubi.Spec.ClusterRefs {
		switch cluster.Type {
		case kyuubiv1alpha1.SparkClusterType:
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      kyuubi.Name + "-spark",
				MountPath: fmt.Sprintf("%s/%s", DefaultSparkConfDir, DefaultSparkConfFile),
				SubPath:   DefaultSparkConfFile,
			})
		case kyuubiv1alpha1.HdfsClusterType:
			volumeMounts = append(volumeMounts,
				corev1.VolumeMount{
					Name:      kyuubi.Name + "-hdfssite",
					MountPath: fmt.Sprintf("%s/%s", DefaultSparkConfDir, DefaultHdfsSiteFile),
					SubPath:   DefaultHdfsSiteFile,
				},
				corev1.VolumeMount{
					Name:      kyuubi.Name + "-coresite",
					MountPath: fmt.Sprintf("%s/%s", DefaultSparkConfDir, DefaultCoreSiteFile),
					SubPath:   DefaultCoreSiteFile,
				})
		case kyuubiv1alpha1.MetaStoreClusterType:
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      kyuubi.Name + "-hivesite",
				MountPath: fmt.Sprintf("%s/%s", DefaultSparkConfDir, DefaultHiveSiteFile),
				SubPath:   DefaultHiveSiteFile,
			})
		}
	}
	return volumeMounts
}

func (r *KyuubiClusterReconciler) constructVolumes(kyuubi *kyuubiv1alpha1.KyuubiCluster) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: ClusterResourceName(kyuubi),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ClusterResourceName(kyuubi),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  DefaultKyuubiConfFile,
							Path: DefaultKyuubiConfFile,
						},
					},
				},
			},
		}}
	for _, cluster := range kyuubi.Spec.ClusterRefs {
		switch cluster.Type {
		case kyuubiv1alpha1.SparkClusterType:
			volumes = append(volumes, corev1.Volume{
				Name: kyuubi.Name + "-spark",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: NineResourceName(kyuubi, DefaultClusterRefsNameSuffix),
						},
						Items: []corev1.KeyToPath{
							{
								Key:  DefaultSparkConfFile,
								Path: DefaultSparkConfFile,
							},
						},
					},
				},
			})
		case kyuubiv1alpha1.HdfsClusterType:
			volumes = append(volumes,
				corev1.Volume{
					Name: kyuubi.Name + "-hdfssite",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: NineResourceName(kyuubi, DefaultClusterRefsNameSuffix),
							},
							Items: []corev1.KeyToPath{
								{
									Key:  DefaultHdfsSiteFile,
									Path: DefaultHdfsSiteFile,
								},
							},
						},
					},
				},
				corev1.Volume{
					Name: kyuubi.Name + "-coresite",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: NineResourceName(kyuubi, DefaultClusterRefsNameSuffix),
							},
							Items: []corev1.KeyToPath{
								{
									Key:  DefaultCoreSiteFile,
									Path: DefaultCoreSiteFile,
								},
							},
						},
					},
				},
			)
		case kyuubiv1alpha1.MetaStoreClusterType:
			volumes = append(volumes, corev1.Volume{
				Name: kyuubi.Name + "-hivesite",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: NineResourceName(kyuubi, DefaultClusterRefsNameSuffix),
						},
						Items: []corev1.KeyToPath{
							{
								Key:  DefaultHiveSiteFile,
								Path: DefaultHiveSiteFile,
							},
						},
					},
				},
			})
		}
	}
	return volumes
}

func (r *KyuubiClusterReconciler) constructDesiredKyuubiWorkload(kyuubi *kyuubiv1alpha1.KyuubiCluster) (*appsv1.StatefulSet, error) {
	stsDesired := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterResourceName(kyuubi),
			Namespace: kyuubi.Namespace,
			Labels:    ClusterResourceLabels(kyuubi),
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ClusterResourceLabels(kyuubi),
			},
			ServiceName: ClusterResourceName(kyuubi),
			Replicas:    int32Ptr(kyuubi.Spec.KyuubiResource.Replicas),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ClusterResourceLabels(kyuubi),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            kyuubi.Name,
							Image:           kyuubi.Spec.KyuubiImage.Repository + ":" + kyuubi.Spec.KyuubiImage.Tag,
							ImagePullPolicy: corev1.PullPolicy(kyuubi.Spec.KyuubiImage.PullPolicy),
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
							Env: DefaultDownwardAPI(),
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
							VolumeMounts: r.constructVolumeMounts(kyuubi),
						},
					},
					RestartPolicy:      corev1.RestartPolicyAlways,
					ServiceAccountName: ClusterResourceName(kyuubi),
					Volumes:            r.constructVolumes(kyuubi),
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
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if k8serrors.IsNotFound(err) {
		if err := r.Create(ctx, desiredKyuubeWorkload); err != nil {
			return err
		}
	} else if !reflect.DeepEqual(existingKyuubeWorkload.Spec, desiredKyuubeWorkload.Spec) {
		logger.Info("updating kyuubi workload")
		err = r.Update(ctx, desiredKyuubeWorkload)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *KyuubiClusterReconciler) desiredClusterRefsConfigMap(kyuubi *kyuubiv1alpha1.KyuubiCluster) (*corev1.ConfigMap, error) {
	cmDesired := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      NineResourceName(kyuubi, DefaultClusterRefsNameSuffix),
			Namespace: kyuubi.Namespace,
			Labels:    ClusterResourceLabels(kyuubi),
		},
		Data: map[string]string{},
	}
	var hdfsSite, coreSite map[string]string
	for _, cluster := range kyuubi.Spec.ClusterRefs {
		if cluster.Type == kyuubiv1alpha1.HdfsClusterType {
			hdfsSite = cluster.Hdfs.HdfsSite
			coreSite = cluster.Hdfs.CoreSite
		}
	}
	for _, cluster := range kyuubi.Spec.ClusterRefs {
		switch cluster.Type {
		case kyuubiv1alpha1.SparkClusterType:
			sparkConf := make(map[string]string)
			if cluster.Spark.SparkImage.Tag == "" {
				cluster.Spark.SparkImage.Tag = "latest"
			}
			sparkConf["spark.master"] = fmt.Sprintf("k8s://https://%s:%s", os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT"))
			sparkConf["spark.kubernetes.container.image"] = cluster.Spark.SparkImage.Repository + ":" + cluster.Spark.SparkImage.Tag
			sparkConf["spark.kubernetes.namespace"] = kyuubi.Namespace
			if cluster.Spark.SparkDefaults == nil {
				return nil, errors.New("spark default conf should not be nil")
			}
			for k, v := range cluster.Spark.SparkDefaults {
				sparkConf[k] = v
			}
			cmDesired.Data[DefaultSparkConfFile] = map2String(sparkConf)
		case kyuubiv1alpha1.HdfsClusterType:
			cmDesired.Data["hdfs-site.xml"] = map2Xml(cluster.Hdfs.HdfsSite)
			cmDesired.Data["core-site.xml"] = map2Xml(cluster.Hdfs.CoreSite)
		case kyuubiv1alpha1.MetaStoreClusterType:
			clusterConf := make(map[string]string, 0)
			if coreSite != nil {
				if _, ok := coreSite[FSDefaultFSConfKey]; ok {
					clusterConf[FSDefaultFSConfKey] = coreSite[FSDefaultFSConfKey]
				}
			}
			if hdfsSite != nil {
				if _, ok := hdfsSite[DFSNameSpacesConfKey]; ok {
					clusterConf[DFSNameSpacesConfKey] = hdfsSite[DFSNameSpacesConfKey]
					if _, ok := hdfsSite[fmt.Sprintf("dfs.ha.namenodes.%s", clusterConf[DFSNameSpacesConfKey])]; ok {
						clusterConf[fmt.Sprintf("dfs.ha.namenodes.%s", clusterConf[DFSNameSpacesConfKey])] = hdfsSite[fmt.Sprintf("dfs.ha.namenodes.%s", clusterConf[DFSNameSpacesConfKey])]
						clusterConf[fmt.Sprintf("dfs.client.failover.proxy.provider.%s", clusterConf[DFSNameSpacesConfKey])] = "org.apache.hadoop.hdfs.server.namenode.ha.ConfiguredFailoverProxyProvider"
						nameNodes := hdfsSite[fmt.Sprintf("dfs.ha.namenodes.%s", clusterConf[DFSNameSpacesConfKey])]
						nameNodeList := strings.Split(nameNodes, ",")
						for _, v := range nameNodeList {
							clusterConf[fmt.Sprintf("dfs.namenode.rpc-address.%s.%s", clusterConf[DFSNameSpacesConfKey], v)] = hdfsSite[fmt.Sprintf("dfs.namenode.rpc-address.%s.%s", clusterConf[DFSNameSpacesConfKey], v)]
						}
					} else {
						clusterConf[fmt.Sprintf("dfs.namenode.rpc-address.%s", clusterConf[DFSNameSpacesConfKey])] = hdfsSite[fmt.Sprintf("dfs.namenode.rpc-address.%s", clusterConf[DFSNameSpacesConfKey])]
					}
				}
			}

			for k, v := range cluster.Metastore.HiveSite {
				clusterConf[k] = v
			}
			cmDesired.Data["hive-site.xml"] = map2Xml(clusterConf)
		}
	}

	if err := ctrl.SetControllerReference(kyuubi, cmDesired, r.Scheme); err != nil {
		return cmDesired, err
	}

	return cmDesired, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateClusterRefsConfigmap(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, logger logr.Logger) error {
	desiredConfigMap, err := r.desiredClusterRefsConfigMap(kyuubi)
	if err != nil {
		return err
	}

	existingConfigMap := &corev1.ConfigMap{}
	err = r.Get(ctx, client.ObjectKeyFromObject(desiredConfigMap), existingConfigMap)
	if err != nil && !k8serrors.IsNotFound(err) {
		logger.Error(err, "Error occurred during Get configmap")
		return err
	}

	// Create or update the Service
	if k8serrors.IsNotFound(err) {
		if err := r.Create(ctx, desiredConfigMap); err != nil {
			logger.Error(err, "Error occurred during Create configmap")
			return err
		}
		for {
			existingConfigMap := &corev1.ConfigMap{}
			err := r.Get(ctx, client.ObjectKeyFromObject(desiredConfigMap), existingConfigMap)
			if err != nil && !k8serrors.IsNotFound(err) {
				return err
			}
			if k8serrors.IsNotFound(err) {
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
			Name:      ClusterResourceName(kyuubi),
			Namespace: kyuubi.Namespace,
			Labels:    ClusterResourceLabels(kyuubi),
		},
		Data: map[string]string{
			DefaultKyuubiConfFile: map2String(kyuubi.Spec.KyuubiConf),
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
	if err != nil && !k8serrors.IsNotFound(err) {
		logger.Error(err, "Error occurred during Get configmap")
		return err
	}

	// Create or update the Configmap
	if k8serrors.IsNotFound(err) {
		if err := r.Create(ctx, desiredConfigMap); err != nil {
			logger.Error(err, "Error occurred during Create configmap")
			return err
		}
		for {
			existingConfigMap := &corev1.ConfigMap{}
			err := r.Get(ctx, client.ObjectKeyFromObject(desiredConfigMap), existingConfigMap)
			if err != nil && !k8serrors.IsNotFound(err) {
				return err
			}
			if k8serrors.IsNotFound(err) {
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
			Name:      ClusterResourceName(kyuubi),
			Namespace: kyuubi.Namespace,
			Labels:    ClusterResourceLabels(kyuubi),
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
			Selector: ClusterResourceLabels(kyuubi),
		},
	}

	if err := ctrl.SetControllerReference(kyuubi, svcDesired, r.Scheme); err != nil {
		return svcDesired, err
	}

	return svcDesired, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateService(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster) (*corev1.Service, error) {
	desiredService, _ := r.contructDesiredService(kyuubi)

	existingService := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKeyFromObject(desiredService), existingService)
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, err
	}

	if k8serrors.IsNotFound(err) {
		if err := r.Create(ctx, desiredService); err != nil {
			return desiredService, err
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
