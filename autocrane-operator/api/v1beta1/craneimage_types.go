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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CraneImageSpec defines the desired state of CraneImage.
type CraneImageSpec struct {
	// Source defines the source registry details.
	Source RegistryDetails `json:"source,omitempty"`

	// Destination defines the destination registry details.
	Destination RegistryDetails `json:"destination,omitempty"`

	// Image defines the image details.
	Image ImageDetails `json:"image,omitempty"`
}

// RegistryDetails defines the details of a container registry.
type RegistryDetails struct {
	// Registry is the URL of the container registry.
	Registry string `json:"registry,omitempty"`

	// CredentialsSecret is the name of the secret containing credentials for the registry.
	CredentialsSecret string `json:"credentialsSecret,omitempty"`

	// Prefix is the prefix for the image in the registry.
	Prefix string `json:"prefix,omitempty"`
}

// GetFullPrefix returns the full prefix for the registry, including the tailing slash.
func (rd *RegistryDetails) GetFullPrefix() string {
	if rd.Prefix == "" {
		return rd.Registry + "/"
	}
	return rd.Registry + "/" + rd.Prefix + "/"
}

func (rd *RegistryDetails) GetRepository(imageName string) string {
	if rd.Prefix == "" {
		return imageName
	}
	return rd.Prefix + "/" + imageName
}

func (rd *RegistryDetails) GetFullImageName(imageName string) string {
	return rd.Registry + "/" + rd.GetRepository(imageName)
}

// ImageDetails defines the details of the container image.
type ImageDetails struct {
	// Name is the name of the container image.
	Name string `json:"name,omitempty"`

	// Tag is the tag of the container image.
	Tag string `json:"tag,omitempty"`
}

// CraneImageStatus defines the observed state of CraneImage.
type CraneImageStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CraneImage is the Schema for the craneimages API.
type CraneImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CraneImageSpec   `json:"spec,omitempty"`
	Status CraneImageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CraneImageList contains a list of CraneImage.
type CraneImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CraneImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CraneImage{}, &CraneImageList{})
}
