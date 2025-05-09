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
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ratify-project/ratify/v2/internal/executor"
)

const (
	configFileName = "config.json"
	configFileDir  = ".ratify"
)

var (
	initConfigDir         = new(sync.Once)
	configDir             string
	defaultConfigFilePath string
	homeDir               string
)

// Load reads the configuration file from the specified path and unmarshals it
// into an executor.Options struct.
func Load(configPath string) (*executor.Options, error) {
	body, err := os.ReadFile(getConfigurationFile(configPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	config := &executor.Options{}
	if err = json.Unmarshal(body, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return config, nil
}

func getConfigurationFile(configFilePath string) string {
	if configFilePath == "" {
		if configDir == "" {
			initConfigDir.Do(initDefaultPaths)
		}
		return defaultConfigFilePath
	}
	return configFilePath
}

func initDefaultPaths() {
	if configDir != "" {
		return
	}
	configDir = os.Getenv("RATIFY_CONFIG")
	if configDir == "" {
		configDir = filepath.Join(getHomeDir(), configFileDir)
	}
	defaultConfigFilePath = filepath.Join(configDir, configFileName)
}

func getHomeDir() string {
	if homeDir == "" {
		homeDir = get()
	}
	return homeDir
}
