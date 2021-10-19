package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/deislabs/hora/pkg/homedir"
	pcConfig "github.com/deislabs/hora/pkg/policyprovider/config"
	rsConfig "github.com/deislabs/hora/pkg/referrerstore/config"
	vfConfig "github.com/deislabs/hora/pkg/verifier/config"
)

const (
	ConfigFileName = "config.json"
	configFileDir  = ".hora"
	PluginsFolder  = "plugins"
)

type Config struct {
	StoresConfig    rsConfig.StoresConfig    `json:"stores,omitempty"`
	PoliciesConfig  pcConfig.PoliciesConfig  `json:"policies,omitempty"`
	VerifiersConfig vfConfig.VerifiersConfig `json:"verifiers,omitempty"`
}

var (
	initConfigDir = new(sync.Once)
	homeDir       string
	configDir     string
	pluginPath    string
)

func setConfigDir() {
	if configDir != "" {
		return
	}
	configDir = os.Getenv("HORA_CONFIG")
	if configDir == "" {
		configDir = filepath.Join(getHomeDir(), configFileDir)

	}
	pluginPath = filepath.Join(configDir, PluginsFolder)
}

func Dir() string {
	initConfigDir.Do(setConfigDir)
	return configDir
}

func getHomeDir() string {
	if homeDir == "" {
		homeDir = homedir.Get()
	}
	return homeDir
}

func Load(configFilePath string) (Config, error) {

	config := Config{}
	if configFilePath == "" {

		if configDir == "" {
			configDir = Dir()
		}

		configFilePath = filepath.Join(configDir, ConfigFileName)
		// FIX the race here
		pluginPath = filepath.Join(configDir, PluginsFolder)
	}

	file, err := os.OpenFile(configFilePath, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return config, fmt.Errorf("could not find config file at path %s", configFilePath)
		}
		return config, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&config); err != nil && !errors.Is(err, io.EOF) {
		return config, err
	}

	return config, nil

}

func GetDefaultPluginPath() string {
	return pluginPath
}
