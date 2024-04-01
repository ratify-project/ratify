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

// ClusterKeyManagementProviderSpec defines the desired state of ClusterKeyManagementProvider
type ClusterKeyManagementProviderSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name of the key management provider
	Type string `json:"type,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	// Parameters of the key management provider
	Parameters runtime.RawExtension `json:"parameters,omitempty"`
}

// ClusterKeyManagementProviderStatus defines the observed state of ClusterKeyManagementProvider
type ClusterKeyManagementProviderStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Is successful in loading certificate/key files
	IsSuccess bool `json:"issuccess"`
	// Error message if operation was unsuccessful
	// +optional
	Error string `json:"error,omitempty"`
	// Truncated error message if the message is too long
	// +optional
	BriefError string `json:"brieferror,omitempty"`
	// The time stamp of last successful certificate/key fetch operation. If operation failed, last fetched time shows the time of error
	// +optional
	LastFetchedTime *metav1.Time `json:"lastfetchedtime,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// provider specific properties of the each individual certificate/key
	// +optional
	Properties runtime.RawExtension `json:"properties,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="IsSuccess",type=boolean,JSONPath=`.status.issuccess`
// +kubebuilder:printcolumn:name="Error",type=string,JSONPath=`.status.brieferror`
// +kubebuilder:printcolumn:name="LastFetchedTime",type=date,JSONPath=`.status.lastfetchedtime`
// ClusterKeyManagementProvider is the Schema for the clusterkeymanagementproviders API
type ClusterKeyManagementProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterKeyManagementProviderSpec   `json:"spec,omitempty"`
	Status ClusterKeyManagementProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterKeyManagementProviderList contains a list of ClusterKeyManagementProvider
type ClusterKeyManagementProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterKeyManagementProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterKeyManagementProvider{}, &ClusterKeyManagementProviderList{})
}
