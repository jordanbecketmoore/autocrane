/*
Copyright 2025.

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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	imagev1beta1 "autocrane.io/api/v1beta1"
)

// CraneImagePolicyReconciler reconciles a CraneImagePolicy object
type CraneImagePolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Clock
}

// +kubebuilder:rbac:groups=image.autocrane.io,resources=craneimagepolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=image.autocrane.io,resources=craneimagepolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=image.autocrane.io,resources=craneimagepolicies/finalizers,verbs=update
// +kubebuilder:rbac:groups=image.autocrane.io,resources=craneimages,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=image.autocrane.io,resources=craneimages/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CraneImagePolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *CraneImagePolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	result := ctrl.Result{
		RequeueAfter: 5 * time.Minute,
	}

	var craneImagePolicy imagev1beta1.CraneImagePolicy
	if err := r.Get(ctx, req.NamespacedName, &craneImagePolicy); err != nil {
		if errors.IsNotFound(err) {
			log.Info("CraneImagePolicy resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get CraneImagePolicy resource.")
		return result, err
	}
	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CraneImagePolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&imagev1beta1.CraneImagePolicy{}).
		Named("craneimagepolicy").
		Complete(r)
}
