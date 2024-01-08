/*
Copyright The Ratify Authors.

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

// +kubebuilder:skip
package unversioned

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// ClusterVerifierSpec defines the desired state of ClusterVerifier
type ClusterVerifierSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Name of the cluster verifier
	Name string `json:"name,omitempty"`

	// Version of the cluster verifier plugin. Optional
	Version string `json:"version,omitempty"`

	// The type of artifact this cluster verifier handles
	ArtifactTypes string `json:"artifactTypes,omitempty"`

	// # Optional. URL/file path
	Address string `json:"address,omitempty"`

	// OCI Artifact source to download the plugin from, optional
	Source *PluginSource `json:"source,omitempty"`

	// Parameters for this cluster verifier
	Parameters runtime.RawExtension `json:"parameters,omitempty"`
}

// ClusterVerifierStatus defines the observed state of ClusterVerifier
type ClusterVerifierStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// ClusterVerifier is the Schema for the clusterverifiers API
type ClusterVerifier struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterVerifierSpec   `json:"spec,omitempty"`
	Status ClusterVerifierStatus `json:"status,omitempty"`
}

// ClusterVerifierList contains a list of ClusterVerifier
type ClusterVerifierList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterVerifier `json:"items"`
}
