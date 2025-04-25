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
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
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
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

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

	sourceImage := craneImage.Spec.Source.GetFullPrefix() + imageName + ":" + imageTag
	destinationImage := craneImage.Spec.Destination.GetFullPrefix() + imageName + ":" + imageTag

	log = log.WithValues("sourceRegistry", sourceRegistry,
		"destinationRegistry", destinationRegistry,
		"sourcePrefix", craneImage.Spec.Source.Prefix,
		"destinationPrefix", craneImage.Spec.Destination.Prefix,
		"imageName", imageName,
		"imageTag", imageTag)

	log.Info("Reconciling CraneImage")

	// Check if the image exists in the source registry and fetch its digest

	// Load source registry authenticator
	log.Info("Loading source registry authenticator.")
	var sourceAuth authn.Authenticator
	if craneImage.Spec.Source.CredentialsSecret != "" {
		log.Info("Using credentials secret for source registry.")
		// Fetch the secret
		var sourceRegistryCredentialsSecret corev1.Secret
		sourceSecretName := client.ObjectKey{
			Namespace: craneImage.Namespace,
			Name:      craneImage.Spec.Source.CredentialsSecret,
		}
		if err := r.Get(ctx, sourceSecretName, &sourceRegistryCredentialsSecret); err != nil {
			log.Error(err, "Failed to fetch credentials secret for source registry.")
			// Update status to reflect failure
			craneImage.Status.State = "Failed"
			craneImage.Status.Message = "Failed to fetch credentials secret: " + err.Error()
			if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
		log.Info("Successfully fetched credentials secret for source registry.")
		sourceAuth, err = secretToAuthenticator(&sourceRegistryCredentialsSecret, sourceRegistry, &log)
		if err != nil {
			log.Error(err, "Failed to create authenticator from credentials secret.")
			// Update status to reflect failure
			craneImage.Status.State = "Failed"
			craneImage.Status.Message = "Failed to create authenticator from credentials secret: " + err.Error()
			if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
		log.Info("Successfully created authenticator from destination credentials secret.")
	} else {
		log.Info("No credentials secret provided for source registry.")
		sourceAuth = authn.Anonymous
	}

	// Check source image registry digest
	sourceDigest, err := crane.Digest(sourceImage, crane.WithAuth(sourceAuth))
	if err != nil {
		log.Error(err, "Failed to get image digest from source registry.")
		// Update status to reflect failure
		craneImage.Status.State = "Failed"
		craneImage.Status.Message = "Failed to get image digest from source registry: " + err.Error()
		if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
			log.Error(statusErr, "Failed to update CraneImage status.")
			return ctrl.Result{}, statusErr
		}
		return result, nil
	}
	log.Info("Successfully fetched image digest from source registry.")

	// Load destination registry authenticator
	log.Info("Loading destination registry authenticator.")
	var destinationAuth authn.Authenticator
	if craneImage.Spec.Destination.CredentialsSecret != "" {
		log.Info("Using credentials secret for destination registry.")
		// Fetch the secret
		var destinationRegistryCredentialsSecret corev1.Secret
		destinationSecretName := client.ObjectKey{
			Namespace: craneImage.Namespace,
			Name:      craneImage.Spec.Destination.CredentialsSecret,
		}
		if err := r.Get(ctx, destinationSecretName, &destinationRegistryCredentialsSecret); err != nil {
			log.Error(err, "Failed to fetch credentials secret for destination registry.")
			// Update status to reflect failure
			craneImage.Status.State = "Failed"
			craneImage.Status.Message = "Failed to fetch credentials secret: " + err.Error()
			if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
		log.Info("Successfully fetched credentials secret for destination registry.")
		destinationAuth, err = secretToAuthenticator(&destinationRegistryCredentialsSecret, destinationRegistry, &log)
		if err != nil {
			log.Error(err, "Failed to create authenticator from credentials secret.")
			// Update status to reflect failure
			craneImage.Status.State = "Failed"
			craneImage.Status.Message = "Failed to create authenticator from credentials secret: " + err.Error()
			if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
		log.Info("Successfully created authenticator from destination credentials secret.")
	} else {
		log.Info("No credentials secret provided for destination registry.")
		destinationAuth = authn.Anonymous
	}
	// Check if the image exists in the destination registry
	log.Info("Checking if image exists in destination registry")
	destinationDigest, err := crane.Digest(destinationImage, crane.WithAuth(destinationAuth))
	// Check if something's there
	if err == nil {
		log.Info("An image was found in the destination registry.")
		// Check if image has the same digest
		// If it does, no need to copy
		if destinationDigest == sourceDigest {
			log.Info("Image found in destination registry. No action needed.")
			// Update status to reflect success
			craneImage.Status.State = "Succeeded"
			craneImage.Status.Message = "Image already exists in destination registry."
			if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
		log.Info("Image found in destination registry, but with different digest.")
	}
	// Here, error != nil or destinationDigest != sourceDigest, so we must copy
	log.Info("Image not found in destination registry. Copying from source.")

	// ################################### PULL ##########################################

	// If no passthrough cache, pull the image from the source registry
	if craneImage.Spec.PassthroughCache.Registry == "" {
		log.Info("Pulling image from source registry.")

		image, err = crane.Pull(sourceImage, crane.WithAuth(sourceAuth))
		if err != nil {
			log.Error(err, "Failed to pull image from source registry.")

			// Update status to reflect failure
			craneImage.Status.State = "Failed"
			craneImage.Status.Message = err.Error()
			if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
		log.Info("Successfully pulled image from source registry.")
	} else {
		// If passthrough cache, pull the image from the passthrough cache
		log.Info("Pulling image from passthrough cache registry.")
		passthroughCacheImage := craneImage.Spec.PassthroughCache.GetFullImageName(imageName) + ":" + imageTag
		passthroughCacheRegistry := craneImage.Spec.PassthroughCache.Registry
		// TODO add authenticator for passthrough cache
		// Load source registry authenticator
		log.Info("Loading source registry authenticator.")
		var passthroughCacheAuth authn.Authenticator
		if craneImage.Spec.PassthroughCache.CredentialsSecret != "" {
			log.Info("Using credentials secret for source registry.")
			// Fetch the secret
			var passthroughCacheRegistryCredentialsSecret corev1.Secret
			passthroughCacheSecretName := client.ObjectKey{
				Namespace: craneImage.Namespace,
				Name:      craneImage.Spec.PassthroughCache.CredentialsSecret,
			}
			if err := r.Get(ctx, passthroughCacheSecretName, &passthroughCacheRegistryCredentialsSecret); err != nil {
				log.Error(err, "Failed to fetch credentials secret for passthrough cache registry.")
				// Update status to reflect failure
				craneImage.Status.State = "Failed"
				craneImage.Status.Message = "Failed to fetch credentials secret: " + err.Error()
				if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
					log.Error(statusErr, "Failed to update CraneImage status.")
					return ctrl.Result{}, statusErr
				}
				return result, nil
			}
			log.Info("Successfully fetched credentials secret for source registry.")
			passthroughCacheAuth, err = secretToAuthenticator(&passthroughCacheRegistryCredentialsSecret, passthroughCacheRegistry, &log)
			if err != nil {
				log.Error(err, "Failed to create authenticator from credentials secret.")
				// Update status to reflect failure
				craneImage.Status.State = "Failed"
				craneImage.Status.Message = "Failed to create authenticator from credentials secret: " + err.Error()
				if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
					log.Error(statusErr, "Failed to update CraneImage status.")
					return ctrl.Result{}, statusErr
				}
				return result, nil
			}
			log.Info("Successfully created authenticator from passthrough cache credentials secret.")
		} else {
			log.Info("No credentials secret provided for passthrough cache registry.")
			passthroughCacheAuth = authn.Anonymous
		}
		image, err = crane.Pull(passthroughCacheImage, crane.WithAuth(passthroughCacheAuth))
		if err != nil {
			log.Error(err, "Failed to pull image from passthrough cache registry.")

			// Update status to reflect failure
			craneImage.Status.State = "Failed"
			craneImage.Status.Message = err.Error()
			if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
				log.Error(statusErr, "Failed to update CraneImage status.")
				return ctrl.Result{}, statusErr
			}
			return result, nil
		}
		log.Info("Successfully pulled image from passthrough cache registry.")
	}

	// ################################### PUSH ##########################################
	// Push the image to the destination registry
	log.Info("Pushing image to destination registry.")
	err = crane.Push(image, destinationImage, crane.WithAuth(destinationAuth))
	if err != nil {
		log.Error(err, "Failed to push image to destination registry.")

		// Update status to reflect failure
		craneImage.Status.State = "Failed"
		craneImage.Status.Message = err.Error()
		if statusErr := r.Status().Update(ctx, &craneImage); statusErr != nil {
			log.Error(statusErr, "Failed to update CraneImage status.")
			return ctrl.Result{}, statusErr
		}
		return result, nil
	}
	log.Info("Successfully pushed image to destination registry.")

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

func configFileToAuthenticator(configFile configfile.ConfigFile, registry string, log *logr.Logger) (authn.Authenticator, error) {
	if registry == "docker.io" {
		registry = "index.docker.io"
	}
	authConfig, err := configFile.GetAuthConfig(registry)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Loaded authConfig for %s", registry))
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

	return nil, fmt.Errorf("unable to create authenticator for registry: %s", registry)
}

func secretToAuthenticator(secret *corev1.Secret, registry string, log *logr.Logger) (authn.Authenticator, error) {
	// Check if the secret type is DockerConfigJson type or has non-empty DockerConfigJson key
	if (secret.Type == corev1.SecretTypeDockerConfigJson) || (secret.Data[corev1.DockerConfigJsonKey] != nil) {
		log.Info("Using Docker config JSON secret")
		// Get encoded Docker config JSON
		dockerConfigJSON := secret.Data[corev1.DockerConfigJsonKey]

		// Decode the Docker config JSON
		var dockerConfig configfile.ConfigFile
		if err := json.Unmarshal(dockerConfigJSON, &dockerConfig); err != nil {
			return nil, err
		}

		return configFileToAuthenticator(dockerConfig, registry, log)

	}
	// Check if the secret type is BasicAuth or has non-empty username and password
	if (secret.Type == corev1.SecretTypeBasicAuth) || (secret.Data[corev1.BasicAuthUsernameKey] != nil && secret.Data[corev1.BasicAuthPasswordKey] != nil) {
		// Get the username and password from the secret
		username := string(secret.Data[corev1.BasicAuthUsernameKey])
		password := string(secret.Data[corev1.BasicAuthPasswordKey])
		// TODO REMOVE
		log.Info("Using Basic Auth secret")
		return authn.FromConfig(authn.AuthConfig{
			Username: username,
			Password: password,
		}), nil
	}

	return nil, fmt.Errorf("secret could not be converted to authenticator")
}
