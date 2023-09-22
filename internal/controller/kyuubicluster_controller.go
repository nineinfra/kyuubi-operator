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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
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

	// TODO(user): your logic here
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
		//logger.Info("create Or Update Components")
		err = r.createOrUpdateClusters(ctx, &kyuubi, logger)
		if err != nil {
			logger.Info("Error occurred during create Or Update Components")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateClusters(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, logger logr.Logger) error {
	err := r.createOrUpdateConfigmap(ctx, kyuubi, logger)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateConfigmap")
		return err
	}
	err = r.createOrUpdateKyuubi(ctx, kyuubi, logger)
	if err != nil {
		logger.Error(err, "Error occurred during createOrUpdateKyuubi")
		return err
	}
	return nil
}

func (r *KyuubiClusterReconciler) constructDesiredKyuubiWorkload(kyuubi *kyuubiv1alpha1.KyuubiCluster) (*appsv1.StatefulSet, error) {
	stsTemplate := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kyuubi.Name + "-kyuubi",
			Namespace: kyuubi.Namespace,
			Labels: map[string]string{
				"cluster": kyuubi.Name,
				"app":     "kyuubiOperator",
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"cluster": kyuubi.Name,
					"app":     "kyuubiOperator",
				},
			},
			ServiceName: kyuubi.Name + "-kyuubi",
			Replicas:    int32Ptr(kyuubi.Spec.KyuubiResource.Replicas),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"cluster": kyuubi.Name,
						"app":     "kyuubiOperator",
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
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"/bin/bash",
											"-c",
											"$KYUUBI_HOME/bin/kyuubi status"},
									},
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(kyuubi, stsTemplate, r.Scheme); err != nil {
		return stsTemplate, err
	}
	return stsTemplate, nil
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
		logger.Info("updating kyuubi")
	}
	return nil
}

func (r *KyuubiClusterReconciler) desiredClusterConfigMap(kyuubi *kyuubiv1alpha1.KyuubiCluster) (*corev1.ConfigMap, error) {

	cmTemplate := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kyuubi.Name + "-cluser-config",
			Namespace: kyuubi.Namespace,
			Labels: map[string]string{
				"app": kyuubi.Name,
			},
		},
		Data: map[string]string{
			"kyuubi-defaults.conf": map2String(kyuubi.Spec.KyuubiConf),
		},
	}

	if err := ctrl.SetControllerReference(kyuubi, cmTemplate, r.Scheme); err != nil {
		return nil, err
	}

	return cmTemplate, nil
}

func (r *KyuubiClusterReconciler) createOrUpdateConfigmap(ctx context.Context, kyuubi *kyuubiv1alpha1.KyuubiCluster, logger logr.Logger) error {
	desiredConfigMap, _ := r.desiredClusterConfigMap(kyuubi)

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
		logger.Info("updating configmap")
		existingConfigMap.Data = desiredConfigMap.Data
		if err := r.Update(ctx, existingConfigMap); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KyuubiClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kyuubiv1alpha1.KyuubiCluster{}).
		Complete(r)
}
