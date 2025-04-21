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

// CraneImagePolicySpec defines the desired state of CraneImagePolicy.
type CraneImagePolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Source RegistryDetails `json:"source,omitempty"`

	// Destination defines the destination registry details.
	Destination RegistryDetails `json:"destination,omitempty"`

	// PolicyDetails defines the policy details.
	ImagePolicy PolicyDetails `json:"imagePolicy,omitempty"`
}

// ImagePolicyDetails defines the details of the image policy.
type PolicyDetails struct {
	Name NamePolicyDetails `json:"image,omitempty"`
	Tag  TagPolicyDetails  `json:"tag,omitempty"`
}

type NamePolicyDetails struct {
	Regex string `json:"regex,omitempty"`
	Exact string `json:"exact,omitempty"`
}

type TagPolicyDetails struct {
	Regex string `json:"regex,omitempty"`
	Exact string `json:"exact,omitempty"`
}

// CraneImagePolicyStatus defines the observed state of CraneImagePolicy.
type CraneImagePolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CraneImagePolicy is the Schema for the craneimagepolicies API.
type CraneImagePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CraneImagePolicySpec   `json:"spec,omitempty"`
	Status CraneImagePolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CraneImagePolicyList contains a list of CraneImagePolicy.
type CraneImagePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CraneImagePolicy `json:"items"`
}

// CraneImagePolicy method to return CraneImage pointer from input name and tag
func (c *CraneImagePolicy) GetCraneImageFromPolicy(name string, tag string) *CraneImage {
	return &CraneImage{
		Spec: CraneImageSpec{
			Source: RegistryDetails{
				Registry:          c.Spec.Source.Registry,
				Prefix:            c.Spec.Source.Prefix,
				CredentialsSecret: c.Spec.Source.CredentialsSecret,
			},
			Destination: RegistryDetails{
				Registry:          c.Spec.Destination.Registry,
				Prefix:            c.Spec.Destination.Prefix,
				CredentialsSecret: c.Spec.Destination.CredentialsSecret,
			},
			Image: ImageDetails{
				Name: name,
				Tag:  tag,
			},
		},
	}
}

func init() {
	SchemeBuilder.Register(&CraneImagePolicy{}, &CraneImagePolicyList{})
}
