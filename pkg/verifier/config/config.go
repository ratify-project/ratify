package config

import (
	"github.com/deislabs/hora/pkg/ocispecs"
	rc "github.com/deislabs/hora/pkg/referrerstore/config"
)

type VerifierConfig map[string]interface{}

type PluginInputConfig struct {
	Config       VerifierConfig               `yaml:"config"`
	StoreConfig  rc.StoreConfig               `yaml:"storeConfig"`
	ReferencDesc ocispecs.ReferenceDescriptor `yaml:"referenceDesc"`
}

type VerifiersConfig struct {
	Version       string           `yaml:"version,omitempty"`
	PluginBinDirs []string         `yaml:"pluginBinDirs,omitempty"`
	Verifiers     []VerifierConfig `yaml:"plugins,omitempty"`
}
