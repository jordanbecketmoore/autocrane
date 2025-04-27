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
	"crypto/sha256"
	"fmt"
	"regexp"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	imagev1beta1 "autocrane.io/api/v1beta1"
	semver "github.com/Masterminds/semver/v3"
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
	var err error

	// CraneImagePolicy function to return CraneImage pointer from input name and tag
	constructCraneImageForCraneImagePolicy := func(c *imagev1beta1.CraneImagePolicy, name string, tag string) (*imagev1beta1.CraneImage, error) {
		craneImage := &imagev1beta1.CraneImage{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
				Namespace:   c.Namespace,
			},
			Spec: imagev1beta1.CraneImageSpec{
				Source:           c.Spec.Source,
				Destination:      c.Spec.Destination,
				PassthroughCache: c.Spec.PassthroughCache,
				Image: imagev1beta1.ImageDetails{
					Name: name,
					Tag:  tag,
				},
			},
		}

		// Create name
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(name+":"+tag)))
		craneImage.ObjectMeta.Name = c.Name + "-" + hash[:7]

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
			log.V(1).Info("CraneImagePolicy resource not found. Ignoring since object must be deleted.")
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

	// Load child CraneImage objects
	var childCraneImages imagev1beta1.CraneImageList
	log.V(1).Info("Loading child CraneImage objects.")
	if err := r.List(ctx, &childCraneImages, client.InNamespace(req.Namespace), client.MatchingFields{craneImageOwnerKey: req.Name}); err != nil {
		log.Error(err, "Failed to list child CraneImage objects.")
		return result, err
	}
	// Make sure all child objects are up to date on uniform data
	for i, childCraneImage := range childCraneImages.Items {
		childCraneImageName := childCraneImage.Spec.Image.Name
		childCraneImageTag := childCraneImage.Spec.Image.Tag
		upToDatChildCraneImage, err := constructCraneImageForCraneImagePolicy(&craneImagePolicy, childCraneImageName, childCraneImageTag)
		if err != nil {
			log.Error(err, "Failed to create up-to-date child CraneImage object.")
			// Update status to reflect failure
			craneImagePolicy.Status.State = "Failed"
			craneImagePolicy.Status.Message = "Failed to create up-to-date child CraneImage object: " + err.Error()
			if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
		upToDatChildCraneImage.ObjectMeta = childCraneImage.ObjectMeta
		childCraneImages.Items[i] = *upToDatChildCraneImage
		if err := r.Update(ctx, &childCraneImages.Items[i]); err != nil {
			log.Error(err, "Failed to update child CraneImage object.")
			// Update status to reflect failure
			craneImagePolicy.Status.State = "Failed"
			craneImagePolicy.Status.Message = "Failed to update child CraneImage object: " + err.Error()
			if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
	}

	// Load source registry authenticator
	log.V(1).Info("Loading source registry authenticator.")
	var sourceAuth authn.Authenticator
	if craneImagePolicy.Spec.Source.CredentialsSecret != "" {
		log.V(1).Info("Using credentials secret for source registry.")
		// Fetch the secret
		var sourceRegistryCredentialsSecret corev1.Secret
		sourceSecretName := client.ObjectKey{
			Namespace: craneImagePolicy.Namespace,
			Name:      craneImagePolicy.Spec.Source.CredentialsSecret,
		}
		if err := r.Get(ctx, sourceSecretName, &sourceRegistryCredentialsSecret); err != nil {
			log.Error(err, "Failed to fetch credentials secret for source registry.")
			// Update status to reflect failure
			craneImagePolicy.Status.State = "Failed"
			craneImagePolicy.Status.Message = "Failed to fetch credentials secret: " + err.Error()
			if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
		log.V(1).Info("Successfully fetched credentials secret for source registry.")
		sourceAuth, err = secretToAuthenticator(&sourceRegistryCredentialsSecret, sourceRegistry, &log)
		if err != nil {
			log.Error(err, "Failed to create authenticator from credentials secret.")
			// Update status to reflect failure
			craneImagePolicy.Status.State = "Failed"
			craneImagePolicy.Status.Message = "Failed to create authenticator from credentials secret: " + err.Error()
			if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
		log.V(1).Info("Successfully created authenticator from destination credentials secret.")
	} else {
		log.V(1).Info("No credentials secret provided for source registry.")
		sourceAuth = authn.Anonymous
	}

	// Filter repositories based on policy details
	// First by image name, then by tags on those images

	// Define imageTagPairs object to hold image tag pairs to create CraneImage objects for
	policyImageTagPairs := make(imageTagPairs)

	// ################################### NAME FILTERING ##########################################

	// First, add exact image name if specified.
	imageNameExact := craneImagePolicy.Spec.ImagePolicy.Name.Exact
	// Check that imageExact is not empty and that it is in the source registry
	if imageNameExact != "" {
		log.Info("ImagePolicy contains exact image name.", "imageName", imageNameExact)
		if _, err := crane.Head(craneImagePolicy.Spec.Source.GetFullImageName(imageNameExact), crane.WithAuth(sourceAuth)); err == nil {
			log.V(1).Info("Adding exact image name to policyImageTagPairs.", "imageName", imageNameExact)
			policyImageTagPairs[imageNameExact] = []string{}
		} else {
			log.Info("Exact image name not found in source registry.", "imageName", imageNameExact)
			// Update status to reflect failure
			craneImagePolicy.Status.State = "Failed"
			craneImagePolicy.Status.Message = "Exact image name not found in source registry: " + imageNameExact
			if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
	}

	// If image field has non-exact name policy,
	// Catalog source registry to obtain a slice of repository names

	if craneImagePolicy.Spec.ImagePolicy.Name.Regex != "" {
		sourceRepositories := []string{}

		log.V(1).Info("Cataloging source registry.")
		sourceRepositories, err = crane.Catalog(sourceRegistry, crane.WithAuth(sourceAuth))
		if err != nil {
			log.Error(err, "Failed to catalog source registry.")
			// Update status to reflect failure
			craneImagePolicy.Status.State = "Failed"
			craneImagePolicy.Status.Message = "Failed to catalog source registry: " + err.Error()
			if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
		log.V(1).Info("Successfully cataloged source registry.", "repositories", sourceRepositories)
	}

	// ################################### TAG FILTERING ##########################################

	log.Info("Image pairs collected.", "imageNameTags", policyImageTagPairs)

	// Grab all existing image tag pairs from the source registry
	log.V(1).Info("Loading image tag pairs from source registry.")
	sourceImageTagPairs := make(imageTagPairs)

	for image, _ := range policyImageTagPairs {
		sourceImageTagPairs[image], err = crane.ListTags(craneImagePolicy.Spec.Source.GetFullImageName(image), crane.WithAuth(sourceAuth))
		if err != nil {
			log.Error(err, "Failed to list tags for image.", "imageName", image)
			// Update status to reflect failure
			craneImagePolicy.Status.State = "Failed"
			craneImagePolicy.Status.Message = "Failed to list tags for image: " + err.Error()
			if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
	}

	// First, add exact tags for images that match if specified
	imageTagExact := craneImagePolicy.Spec.ImagePolicy.Tag.Exact
	if imageTagExact != "" {
		for name, tags := range policyImageTagPairs {
			log.Info("Adding exact image tag to policyImageTagPairs.", "imageName", name, "imageTag", imageTagExact)
			tags = append(tags, imageTagExact)
			policyImageTagPairs[name] = tags
		}
	}

	imageTagRegexString := craneImagePolicy.Spec.ImagePolicy.Tag.Regex
	// TODO add logic to add regex '^' and '$' to regex string
	if imageTagRegexString != "" {
		log.Info("ImagePolicy contains regex image tag.", "regex", imageTagRegexString)
		for image, tags := range sourceImageTagPairs {
			matchingTags := []string{}
			for _, tag := range tags {
				match, err := regexp.Match(imageTagRegexString, []byte(tag))
				if err != nil {
					log.Error(err, "Error while attempting to match regexp.", "imageName", image, "regex", imageTagRegexString)
					// Update status to reflect failure
					craneImagePolicy.Status.State = "Failed"
					craneImagePolicy.Status.Message = "Failed to match regex: " + err.Error()
					if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
						log.Error(statusErr, "Failed to update CraneImage status.")
						return ctrl.Result{}, statusErr
					}
				}
				if match {
					log.Info("Adding regex image tag to policyImageTagPairs.", "imageName", image, "imageTag", tag, "regex", imageTagRegexString)
					matchingTags = append(matchingTags, tag)
				}
			}
			policyImageTagPairs[image] = append(policyImageTagPairs[image], matchingTags...)
		}

	}

	imageTagSemverString := craneImagePolicy.Spec.ImagePolicy.Tag.Semver
	if imageTagSemverString != "" {
		log.Info("ImagePolicy contains semver constraint.", "semver", imageTagSemverString)

		imageTagSemverConstraint, err := semver.NewConstraint(imageTagSemverString)
		if err != nil {
			log.Error(err, "Failed to parse semver constraint.", "semver", imageTagSemverString) // Update status to reflect failure
			craneImagePolicy.Status.State = "Failed"
			craneImagePolicy.Status.Message = "Failed to parse semver constraint: " + err.Error()
			if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}

		}
		for image, tags := range sourceImageTagPairs {
			matchingTags := []string{}
			for _, tag := range tags {
				tagVersion, err := semver.NewVersion(tag)
				// ignore tags that do not parse as semvers
				if err != nil {
					continue
				}
				if imageTagSemverConstraint.Check(tagVersion) {
					log.V(1).Info("Image tag matches semver constraint.", "imageName", image, "imageTag", tag, "semver", imageTagSemverString)
					matchingTags = append(matchingTags, tag)
				}

			}
			policyImageTagPairs[image] = matchingTags
		}
	}

	log.Info("Final image tag pairs.", "imageTagPairs", policyImageTagPairs)

	// ################################### CraneImage Provisioning ##########################################

	// Range over imageTagPairs and create CraneImage objects
	// for each image,tag pair if it doesn't already exist
	for name, tags := range policyImageTagPairs {
		tags = removeDuplicates(tags)
		for _, tag := range tags {
			craneImageExists := false
			for _, childCraneImage := range childCraneImages.Items {
				if childCraneImage.Spec.Image.Name == name && childCraneImage.Spec.Image.Tag == tag {
					log.V(1).Info("Child CraneImage object already exists.", "imageName", name, "imageTag", tag)
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
				log.V(1).Info("Successfully created new CraneImage object.", "imageName", name, "imageTag", tag, "craneImage", newCraneImage)

			}
		}
	}

	// Range over childCraneImages
	// and delete any that are not in imageTagPairs

	// Update status to reflect success
	childCraneImageCount := len(childCraneImages.Items)
	craneImagePolicy.Status.State = "Success"
	if childCraneImageCount == 0 {
		craneImagePolicy.Status.Message = "No child CraneImage objects found."
	} else if childCraneImageCount == 1 {
		craneImagePolicy.Status.Message = "1 child CraneImage object found."
	} else {
		craneImagePolicy.Status.Message = fmt.Sprintf("%d child CraneImage objects found.", childCraneImageCount)
	}
	if statusErr := r.Status().Update(ctx, &craneImagePolicy); statusErr != nil {
		log.Error(statusErr, "Failed to update CraneImage status.")
		return ctrl.Result{}, statusErr
	}
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

// Check if a string exists in a slice
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func removeDuplicates(strList []string) []string {
	list := []string{}
	for _, item := range strList {
		if !contains(list, item) {
			list = append(list, item)
		}
	}
	return list
}
