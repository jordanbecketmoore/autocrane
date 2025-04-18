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

	"github.com/google/go-containerregistry/pkg/crane"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	imagev1beta1 "autocrane.io/api/v1beta1"
)

// CraneImageReconciler reconciles a CraneImage object
type CraneImageReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Clock
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// Clock knows how to get the current time.
// It can be used to fake out timing for testing.
type Clock interface {
	Now() time.Time
}

// +kubebuilder:rbac:groups=image.autocrane.io,resources=craneimages,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=image.autocrane.io,resources=craneimages/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=image.autocrane.io,resources=craneimages/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CraneImage object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *CraneImageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Define result with requeue after 1 minute
	result := ctrl.Result{
		RequeueAfter: time.Minute,
	}
	// Fetch the CraneImage instance
	var craneImage imagev1beta1.CraneImage
	if err := r.Get(ctx, req.NamespacedName, &craneImage); err != nil {
		if errors.IsNotFound(err) {
			log.Info("CraneImage resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get CraneImage resource.")
		return result, err
	}

	sourceRegistry := craneImage.Spec.Source.Registry
	destinationRegistry := craneImage.Spec.Destination.Registry
	imageName := craneImage.Spec.Image.Name
	imageTag := craneImage.Spec.Image.Tag

	sourceImage := sourceRegistry + "/" + imageName + ":" + imageTag
	destinationImage := destinationRegistry + "/" + imageName + ":" + imageTag

	log = log.WithValues("sourceRegistry", sourceRegistry,
		"destinationRegistry", destinationRegistry,
		"imageName", imageName,
		"imageTag", imageTag)

	log.Info("Reconciling CraneImage")

	// Check if the image exists in the destination registry
	log.Info("Checking if image exists in destination registry")

	if _, err := crane.Head(destinationImage); err != nil {

		log.Info("Image not found in destination registry. Copying from source.")

		// Pull the image from the source registry
		log.Info("Pulling image from source registry.")
		image, err := crane.Pull(sourceImage)
		if err != nil {
			log.Error(err, "Failed to pull image from source registry.")

			// Update status to reflect failure
			craneImage.Status.State = "Failed"
			craneImage.Status.Message = err.Error()
			if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
			}
			return result, err
		}

		// Push the image to the destination registry
		log.Info("Pushing image to destination registry.")
		if err := crane.Push(image, destinationImage); err != nil {
			log.Error(err, "Failed to push image to destination registry.")

			// Update status to reflect failure
			craneImage.Status.State = "Failed"
			craneImage.Status.Message = err.Error()
			if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
			}
			return result, err
		}

		log.Info("Image successfully pushed to destination registry.")

		// Update status to reflect success
		craneImage.Status.State = "Succeeded"
		craneImage.Status.Message = "Image successfully copied to destination registry."
		if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
			log.Error(statusErr, "Failed to update CraneImage status.")
			return result, statusErr
		}
	} else {
		log.Info("Image already exists in destination registry.")

		// Update status to reflect no action needed
		craneImage.Status.State = "Succeeded"
		craneImage.Status.Message = "Image already exists in destination registry."
		if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
			log.Error(statusErr, "Failed to update CraneImage status.")
			return result, statusErr
		}
	}

	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CraneImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// set up a real clock, since we're not in a test
	if r.Clock == nil {
		r.Clock = realClock{}
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&imagev1beta1.CraneImage{}).
		Named("craneimage").
		Complete(r)
}
