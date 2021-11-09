package config

import (
	"github.com/deislabs/ratify/pkg/ocispecs"
	rc "github.com/deislabs/ratify/pkg/referrerstore/config"
)

type VerifierConfig map[string]interface{}

type PluginInputConfig struct {
	Config       VerifierConfig               `json:"config"`
	StoreConfig  rc.StoreConfig               `json:"storeConfig"`
	ReferencDesc ocispecs.ReferenceDescriptor `json:"referenceDesc"`
}

type VerifiersConfig struct {
	Version       string           `json:"version,omitempty"`
	PluginBinDirs []string         `json:"pluginBinDirs,omitempty"`
	Verifiers     []VerifierConfig `json:"plugins,omitempty"`
}
