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

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/deislabs/ratify/pkg/homedir"
	pcConfig "github.com/deislabs/ratify/pkg/policyprovider/config"
	rsConfig "github.com/deislabs/ratify/pkg/referrerstore/config"
	vfConfig "github.com/deislabs/ratify/pkg/verifier/config"
)

const (
	ConfigFileName = "config.json"
	configFileDir  = ".ratify"
	PluginsFolder  = "plugins"
)

type Config struct {
	StoresConfig    rsConfig.StoresConfig    `json:"stores,omitempty"`
	PoliciesConfig  pcConfig.PoliciesConfig  `json:"policies,omitempty"`
	VerifiersConfig vfConfig.VerifiersConfig `json:"verifiers,omitempty"`
}

var (
	initConfigDir         = new(sync.Once)
	homeDir               string
	configDir             string
	defaultConfigFilePath string
	defaultPluginsPath    string
)

func InitDefaultPaths() {
	if configDir != "" {
		return
	}
	configDir = os.Getenv("RATIFY_CONFIG")
	if configDir == "" {
		configDir = filepath.Join(getHomeDir(), configFileDir)

	}
	defaultPluginsPath = filepath.Join(configDir, PluginsFolder)
	defaultConfigFilePath = filepath.Join(configDir, ConfigFileName)
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
			initConfigDir.Do(InitDefaultPaths)
		}

		configFilePath = defaultConfigFilePath
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
	if defaultPluginsPath == "" {
		initConfigDir.Do(InitDefaultPaths)
	}
	return defaultPluginsPath
}
