package config

type StorePluginConfig map[string]interface{}

type StoresConfig struct {
	Version       string              `yaml:"version,omitempty"`
	PluginBinDirs []string            `yaml:"pluginBinDirs,omitempty"`
	Stores        []StorePluginConfig `yaml:"plugins,omitempty"`
}

type StoreConfig struct {
	Version       string            `yaml:"version"`
	PluginBinDirs []string          `yaml:"pluginBinDirs"`
	Store         StorePluginConfig `yaml:"store"`
}
