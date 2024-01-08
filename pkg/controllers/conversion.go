package controllers

import (
	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

type storeSpec struct {
	// Name of the store
	name string `json:"name"`
	// Version of the store plugin. Optional
	version string `json:"version,omitempty"`
	// Plugin path, optional
	address string `json:"address,omitempty"`
	// OCI Artifact source to download the plugin from, optional
	source *configv1beta1.PluginSource `json:"source,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	// Parameters of the store
	parameters runtime.RawExtension `json:"parameters,omitempty"`
}

func convertStoreSpec(spec configv1beta1.StoreSpec) storeSpec {
	return storeSpec{
		name:       spec.Name,
		version:    spec.Version,
		address:    spec.Address,
		source:     spec.Source,
		parameters: spec.Parameters,
	}
}
