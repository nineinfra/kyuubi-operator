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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var kyuubiclusterlog = logf.Log.WithName("kyuubicluster-resource")

func (r *KyuubiCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-kyuubi-nineinfra-tech-v1alpha1-kyuubicluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=kyuubi.nineinfra.tech,resources=kyuubiclusters,verbs=create;update,versions=v1alpha1,name=mkyuubicluster.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KyuubiCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KyuubiCluster) Default() {
	kyuubiclusterlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-kyuubi-nineinfra-tech-v1alpha1-kyuubicluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=kyuubi.nineinfra.tech,resources=kyuubiclusters,verbs=create;update,versions=v1alpha1,name=vkyuubicluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &KyuubiCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KyuubiCluster) ValidateCreate() (admission.Warnings, error) {
	kyuubiclusterlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KyuubiCluster) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	kyuubiclusterlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KyuubiCluster) ValidateDelete() (admission.Warnings, error) {
	kyuubiclusterlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
