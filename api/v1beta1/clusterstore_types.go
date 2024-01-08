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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// ClusterStoreSpec defines the desired state of ClusterStore
type ClusterStoreSpec struct {
	// Important: Run "make install-crds" to regenerate code after modifying this file

	// Name of the cluster store
	Name string `json:"name"`
	// Version of the cluster store plugin. Optional
	Version string `json:"version,omitempty"`
	// Plugin path, optional
	Address string `json:"address,omitempty"`
	// OCI Artifact source to download the plugin from, optional
	Source *PluginSource `json:"source,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	// Parameters of the cluster store
	Parameters runtime.RawExtension `json:"parameters,omitempty"`
}

// ClusterStoreStatus defines the observed state of ClusterStore
type ClusterStoreStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:storageversion
// ClusterStore is the Schema for the clusterstores API
type ClusterStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterStoreSpec   `json:"spec,omitempty"`
	Status ClusterStoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// ClusterStoreList contains a list of ClusterStore
type ClusterStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterStore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterStore{}, &ClusterStoreList{})
}
