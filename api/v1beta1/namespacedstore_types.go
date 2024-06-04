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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NamespacedStoreSpec defines the desired state of NamespacedStore
type NamespacedStoreSpec struct {
	// Important: Run "make install-crds" to regenerate code after modifying this file

	// Name of the store
	Name string `json:"name"`
	// Version of the store plugin. Optional
	Version string `json:"version,omitempty"`
	// Plugin path, optional
	Address string `json:"address,omitempty"`
	// OCI Artifact source to download the plugin from, optional
	Source *PluginSource `json:"source,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	// Parameters of the store
	Parameters runtime.RawExtension `json:"parameters,omitempty"`
}

// NamespacedStoreStatus defines the observed state of NamespacedStore
type NamespacedStoreStatus struct {
	// Important: Run "make install-crds" to regenerate code after modifying this file

	// Is successful in finding the plugin
	IsSuccess bool `json:"issuccess"`
	// Error message if operation was unsuccessful
	// +optional
	Error string `json:"error,omitempty"`
	// Truncated error message if the message is too long
	// +optional
	BriefError string `json:"brieferror,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Namespaced"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="IsSuccess",type=boolean,JSONPath=`.status.issuccess`
// +kubebuilder:printcolumn:name="Error",type=string,JSONPath=`.status.brieferror`
// NamespacedStore is the Schema for the namespacedstores API
type NamespacedStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespacedStoreSpec   `json:"spec,omitempty"`
	Status NamespacedStoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// NamespacedStoreList contains a list of NamespacedStore
type NamespacedStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NamespacedStore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NamespacedStore{}, &NamespacedStoreList{})
}
