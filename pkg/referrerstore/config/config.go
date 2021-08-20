package config

type StoreConfig map[string]interface{}

type StoresConfig struct {
	Version       string        `json:"version,omitempty"`
	PluginBinDirs []string      `json:"pluginBinDirs,omitempty`
	Stores        []StoreConfig `json:"plugins,omitempty"`
}
