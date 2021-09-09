package config

type StorePluginConfig map[string]interface{}

type StoresConfig struct {
	Version       string              `json:"version,omitempty"`
	PluginBinDirs []string            `json:"pluginBinDirs,omitempty"`
	Stores        []StorePluginConfig `json:"plugins,omitempty"`
}

type StoreConfig struct {
	Version       string            `json:"version"`
	PluginBinDirs []string          `json:"pluginBinDirs"`
	Store         StorePluginConfig `json:"store"`
}
