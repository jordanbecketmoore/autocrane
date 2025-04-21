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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// A type for holding desired image,tag pairs
type imageTagPairs map[string][]string

// CraneImagePolicyReconciler reconciles a CraneImagePolicy object
type CraneImagePolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Clock
}

var (
	craneImageOwnerKey = ".metadata.controller"
)

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
	// TODO uncomment this line
	// var err error

	// CraneImagePolicy function to return CraneImage pointer from input name and tag
	constructCraneImageForCraneImagePolicy := func(c *imagev1beta1.CraneImagePolicy, name string, tag string) (*imagev1beta1.CraneImage, error) {
		craneImage := &imagev1beta1.CraneImage{
			Spec: imagev1beta1.CraneImageSpec{
				Source: imagev1beta1.RegistryDetails{
					Registry:          c.Spec.Source.Registry,
					Prefix:            c.Spec.Source.Prefix,
					CredentialsSecret: c.Spec.Source.CredentialsSecret,
				},
				Destination: imagev1beta1.RegistryDetails{
					Registry:          c.Spec.Destination.Registry,
					Prefix:            c.Spec.Destination.Prefix,
					CredentialsSecret: c.Spec.Destination.CredentialsSecret,
				},
				Image: imagev1beta1.ImageDetails{
					Name: name,
					Tag:  tag,
				},
			},
		}
		if err := ctrl.SetControllerReference(c, craneImage, r.Scheme); err != nil {
			return nil, err
		}

		return craneImage, nil
	}

	// default result return value
	result := ctrl.Result{
		RequeueAfter: 5 * time.Minute,
	}

	// Fetch the CraneImagePolicy object
	var craneImagePolicy imagev1beta1.CraneImagePolicy
	if err := r.Get(ctx, req.NamespacedName, &craneImagePolicy); err != nil {
		if errors.IsNotFound(err) {
			log.Info("CraneImagePolicy resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get CraneImagePolicy resource.")
		return result, err
	}

	sourceRegistry := craneImagePolicy.Spec.Source.Registry
	destinationRegistry := craneImagePolicy.Spec.Destination.Registry
	sourcePrefix := craneImagePolicy.Spec.Source.Prefix
	destinationPrefix := craneImagePolicy.Spec.Destination.Prefix

	log = log.WithValues(
		"sourceRegistry", sourceRegistry,
		"destinationRegistry", destinationRegistry,
		"sourcePrefix", sourcePrefix,
		"destinationPrefix", destinationPrefix,
	)

	// Load source registry authenticator
	log.Info("Loading source registry authenticator.")
	// TODO uncomment vvv
	// var sourceAuth authn.Authenticator
	// if craneImagePolicy.Spec.Source.CredentialsSecret != "" {
	// 	log.Info("Using credentials secret for source registry.")
	// 	// Fetch the secret
	// 	var sourceRegistryCredentialsSecret corev1.Secret
	// 	sourceSecretName := client.ObjectKey{
	// 		Namespace: craneImagePolicy.Namespace,
	// 		Name:      craneImagePolicy.Spec.Source.CredentialsSecret,
	// 	}
	// 	if err := r.Get(ctx, sourceSecretName, &sourceRegistryCredentialsSecret); err != nil {
	// 		log.Error(err, "Failed to fetch credentials secret for source registry.")
	// 		// Update status to reflect failure
	// 		craneImagePolicy.Status.State = "Failed"
	// 		craneImagePolicy.Status.Message = "Failed to fetch credentials secret: " + err.Error()
	// 		if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
	// 			log.Error(statusErr, "Failed to update CraneImage status.")
	// 			return ctrl.Result{}, statusErr
	// 		}
	// 		return result, nil
	// 	}
	// 	log.Info("Successfully fetched credentials secret for source registry.")
	// 	sourceAuth, err = secretToAuthenticator(&sourceRegistryCredentialsSecret, sourceRegistry, &log)
	// 	if err != nil {
	// 		log.Error(err, "Failed to create authenticator from credentials secret.")
	// 		// Update status to reflect failure
	// 		craneImagePolicy.Status.State = "Failed"
	// 		craneImagePolicy.Status.Message = "Failed to create authenticator from credentials secret: " + err.Error()
	// 		if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
	// 			log.Error(statusErr, "Failed to update CraneImage status.")
	// 			return ctrl.Result{}, statusErr
	// 		}
	// 		return result, nil
	// 	}
	// 	log.Info("Successfully created authenticator from destination credentials secret.")
	// } else {
	// 	log.Info("No credentials secret provided for source registry.")
	// 	sourceAuth = authn.Anonymous
	// }

	// If image field has non-exact name policy,
	// Catalog source registry to obtain a slice of repository names

	// TODO uncomment block vvv
	// if craneImagePolicy.Spec.ImagePolicy.Name.Regex != "" {
	// 	log.Info("Cataloging source registry.")
	// 	sourceRepositories, err := crane.Catalog(sourceRegistry, crane.WithAuth(sourceAuth))
	// 	if err != nil {
	// 		log.Error(err, "Failed to catalog source registry.")
	// 		// Update status to reflect failure
	// 		craneImagePolicy.Status.State = "Failed"
	// 		craneImagePolicy.Status.Message = "Failed to catalog source registry: " + err.Error()
	// 		if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
	// 			log.Error(statusErr, "Failed to update CraneImage status.")
	// 			return ctrl.Result{}, statusErr
	// 		}
	// 		return result, nil
	// 	}
	// 	log.Info("Successfully cataloged source registry.", "repositories", sourceRepositories)
	// } else {
	// 	sourceRepositories := []string{}
	// }

	// Filter repositories based on policy details
	// First by image name, then by tags on those images

	// Define imageTagPairs object to hold image tag pairs to create CraneImage objects for
	imageTagPairs := make(imageTagPairs)

	// ################################### NAME FILTERING ##########################################

	// First, add exact image name if specified.
	imageExact := craneImagePolicy.Spec.ImagePolicy.Name.Exact
	if imageExact != "" {
		log.Info("Adding exact image name to imageTagPairs.", "imageName", imageExact)
		imageTagPairs[imageExact] = []string{}
	}

	// TODO: Add regex support for image name
	// Check all repositories for regex match
	// then add to imageTagPairs keys

	// ################################### TAG FILTERING ##########################################

	// First, add exact tags for images that match if specified
	imageTagExact := craneImagePolicy.Spec.ImagePolicy.Tag.Exact
	if imageTagExact != "" {
		for name, tags := range imageTagPairs {
			log.Info("Adding exact image tag to imageTagPairs.", "imageName", name, "imageTag", imageTagExact)
			tags = append(tags, imageTagExact)
			imageTagPairs[name] = tags
		}
	}

	// TODO: Add regex support for image tags
	// Check all repositories for regex-matching tags
	// then add to imageTagPair values

	// ################################### CraneImage Provisioning ##########################################

	// Load child CraneImage objects
	var childCraneImages imagev1beta1.CraneImageList
	log.Info("Loading child CraneImage objects.")
	if err := r.List(ctx, &childCraneImages, client.InNamespace(req.Namespace), client.MatchingFields{craneImageOwnerKey: req.Name}); err != nil {
		log.Error(err, "Failed to list child CraneImage objects.")
		return result, err
	}

	// Range over imageTagPairs and create CraneImage objects
	// for each image,tag pair if it doesn't already exist
	for name, tags := range imageTagPairs {
		for _, tag := range tags {
			craneImageExists := false
			for _, childCraneImage := range childCraneImages.Items {
				if childCraneImage.Spec.Image.Name == name && childCraneImage.Spec.Image.Tag == tag {
					log.Info("Child CraneImage object already exists.", "imageName", name, "imageTag", tag)
					// Break back to tags loop
					craneImageExists = true
					break
				}
			}

			if !craneImageExists {
				// Create a new CraneImage object (pointer)
				newCraneImage, err := constructCraneImageForCraneImagePolicy(&craneImagePolicy, name, tag)
				if err != nil {
					log.Error(err, "Failed to create new CraneImage object.")
					// Update status to reflect failure
					craneImagePolicy.Status.State = "Failed"
					craneImagePolicy.Status.Message = "Failed to create new CraneImage object: " + err.Error()
					if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
						log.Error(statusErr, "Failed to update CraneImage status.")
						return ctrl.Result{}, statusErr
					}
					return result, nil
				}
				if err := r.Create(ctx, newCraneImage); err != nil {
					log.Error(err, "Failed to create new CraneImage object.", "imageName", name, "imageTag", tag, "craneImage", newCraneImage)
					// Update status to reflect failure
					craneImagePolicy.Status.State = "Failed"
					craneImagePolicy.Status.Message = "Failed to create new CraneImage object: " + err.Error()
					if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
						log.Error(statusErr, "Failed to update CraneImage status.")
						return ctrl.Result{}, statusErr
					}
					return result, nil
				}
				log.Info("Successfully created new CraneImage object.", "imageName", name, "imageTag", tag, "craneImage", newCraneImage)

			}
		}
	}

	// Range over childCraneImages
	// and delete any that are not in imageTagPairs

	// Final return if haven't already
	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CraneImagePolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Clock == nil {
		r.Clock = realClock{}
	}

	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&imagev1beta1.CraneImage{},
		craneImageOwnerKey,
		func(rawObj client.Object) []string {

			craneImage := rawObj.(*imagev1beta1.CraneImage)
			owner := metav1.GetControllerOf(craneImage)
			if owner == nil {
				return nil
			}
			if owner.APIVersion != imagev1beta1.GroupVersion.String() || owner.Kind != "CraneImagePolicy" {
				return nil
			}
			return []string{owner.Name}
		}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&imagev1beta1.CraneImagePolicy{}).
		Named("craneimagepolicy").
		Complete(r)
}
