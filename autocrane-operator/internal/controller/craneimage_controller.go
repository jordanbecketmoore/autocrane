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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	imagev1beta1 "autocrane.io/api/v1beta1"
	"github.com/docker/cli/cli/config/configfile"
	corev1 "k8s.io/api/core/v1"
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
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get,list

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
	var image v1.Image
	var err error

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

	// Standardize dockerhub registry addresses to docker.io
	if sourceRegistry == "index.docker.io" {
		sourceRegistry = "docker.io"
	}

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

	if _, err := crane.Head(destinationImage); err == nil {
		log.Info("Image already exists in destination registry.")

		// Update status to reflect no action needed
		craneImage.Status.State = "Succeeded"
		craneImage.Status.Message = "Image already exists in destination registry."
		if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
			log.Error(statusErr, "Failed to update CraneImage status.")
			return result, statusErr
		}
		return result, nil
	}

	log.Info("Image not found in destination registry. Copying from source.")

	// ################################### PULL ##########################################
	// Pull the image from the source registry
	log.Info("Pulling image from source registry.")

	if craneImage.Spec.Source.CredentialsSecret != "" {
		log.Info("Using credentials secret for source registry.")

		// Fetch the secret
		var secret corev1.Secret
		secretName := client.ObjectKey{
			Namespace: craneImage.Namespace,
			Name:      craneImage.Spec.Source.CredentialsSecret,
		}
		if err := r.Get(ctx, secretName, &secret); err != nil {
			log.Error(err, "Failed to fetch credentials secret for source registry.")

			// Update status to reflect failure
			craneImage.Status.State = "Failed"
			craneImage.Status.Message = "Failed to fetch credentials secret: " + err.Error()
			if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
			}
			return result, err
		}

		log.Info("Successfully fetched credentials secret for source registry.")

		// Check if the secret type is DockerConfigJson
		if secret.Type == corev1.SecretTypeDockerConfigJson {
			// Get encoded Docker config JSON
			dockerConfigJSON := secret.Data[corev1.DockerConfigJsonKey]

			// Base64 decode dockerConfigJSON
			decodedDockerConfigJSON, err := base64.StdEncoding.DecodeString(string(dockerConfigJSON))
			if err != nil {
				log.Error(err, "Failed to decode Docker config JSON.")

				// Update status to reflect failure
				craneImage.Status.State = "Failed"
				craneImage.Status.Message = "Failed to decode Docker config JSON: " + err.Error()
				if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
					log.Error(statusErr, "Failed to update CraneImage status.")
				}
				return result, err
			}
			// Use the decoded Docker config JSON to authenticate with the source registry
			log.Info("Unmarshalling Docker config JSON.")
			var dockerConfig configfile.ConfigFile
			if err := json.Unmarshal(decodedDockerConfigJSON, &dockerConfig); err != nil {
				// Update status to reflect failure
				craneImage.Status.State = "Failed"
				craneImage.Status.Message = "Failed to unmarshal Docker config JSON: " + err.Error()
				if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
					log.Error(statusErr, "Failed to update CraneImage status.")
				}
				return result, err
			}
			log.Info("Successfully unmarshalled Docker config JSON.")

			// Generate crane authenticator from Docker config JSON
			auth, err := configFileToAuthenticator(dockerConfig, sourceRegistry)
			if err != nil {
				log.Error(err, "Failed to create authenticator for source registry.")

				// Update status to reflect failure
				craneImage.Status.State = "Failed"
				craneImage.Status.Message = err.Error()
				if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
					log.Error(statusErr, "Failed to update CraneImage status.")
				}
				return result, err
			}

			image, err = crane.Pull(sourceImage, crane.WithAuth(auth))
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
			log.Info("Successfully pulled image from source registry.")
		} // TODO: block must either return or define image to continue

	} else {
		image, err = crane.Pull(sourceImage)
		if err != nil {
			log.Error(err, "Failed to pull image from source registry.")

			// Update status to reflect failure
			craneImage.Status.State = "Failed"
			craneImage.Status.Message = err.Error()
			if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
			}
		}
	}
	// ################################### PUSH ##########################################
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

func configFileToAuthenticator(configFile configfile.ConfigFile, registry string) (authn.Authenticator, error) {
	if registry == "docker.io" {
		registry = "index.docker.io"
	}
	authConfig, err := configFile.GetAuthConfig(registry)
	if err != nil {
		return nil, err
	}
	// Create auth from username and password
	if authConfig.Username != "" && authConfig.Password != "" {
		return authn.FromConfig(authn.AuthConfig{
			Username: authConfig.Username,
			Password: authConfig.Password,
		}), nil
	}
	// Create auth from token
	if authConfig.Auth != "" {
		return authn.FromConfig(authn.AuthConfig{
			Auth: authConfig.Auth,
		}), nil
	}

	return nil, fmt.Errorf("Unable to create authenticator for registry: %x", registry)
}
