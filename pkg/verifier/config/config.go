package config

type VerifierConfig map[string]interface{}

type PluginInputConfig struct {
	Config VerifierConfig `json:"config"`
	Blob   []byte         `json:"blob"`
}

type VerifiersConfig struct {
	Version       string           `json:"version,omitempty"`
	PluginBinDirs []string         `json:"pluginBinDirs,omitempty`
	Verifiers     []VerifierConfig `json:"plugins,omitempty"`
}
