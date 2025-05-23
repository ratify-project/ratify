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
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/notaryproject/ratify/v2/internal/executor"
	"github.com/notaryproject/ratify/v2/internal/store"
	"github.com/notaryproject/ratify/v2/internal/verifier/factory"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tempDir := t.TempDir()
	validConfig := `{"verifiers":[],"stores":{}}`

	validConfigPath := filepath.Join(tempDir, "valid_config.json")
	err := os.WriteFile(validConfigPath, []byte(validConfig), 0600)
	assert.NoError(t, err)

	t.Run("valid config file", func(t *testing.T) {
		config, err := Load(validConfigPath)
		assert.NoError(t, err)
		assert.Equal(t, &executor.Options{
			Verifiers: []factory.NewVerifierOptions{},
			Stores:    store.PatternOptions{},
		}, config)
	})

	t.Run("non-existent config file", func(t *testing.T) {
		_, err := Load(filepath.Join(tempDir, "nonexistent.json"))
		assert.Error(t, err)
	})

	t.Run("invalid json format", func(t *testing.T) {
		invalidConfigPath := filepath.Join(tempDir, "invalid_config.json")
		err := os.WriteFile(invalidConfigPath, []byte(`{"Field1": "value1", "Field2":}`), 0600)
		assert.NoError(t, err)

		_, err = Load(invalidConfigPath)
		assert.Error(t, err)
	})

	t.Run("empty config path uses default", func(t *testing.T) {
		initConfigDir = new(sync.Once)
		configDir = tempDir
		defaultConfigFilePath = filepath.Join(configDir, configFileName)
		err := os.WriteFile(defaultConfigFilePath, []byte(validConfig), 0600)
		assert.NoError(t, err)

		config, err := Load("")
		assert.NoError(t, err)
		assert.Equal(t, &executor.Options{
			Verifiers: []factory.NewVerifierOptions{},
			Stores:    store.PatternOptions{},
		}, config)
	})

	t.Run("empty config path and no default", func(t *testing.T) {
		initConfigDir = new(sync.Once)
		configDir = ""
		defaultConfigFilePath = filepath.Join(configDir, configFileName)

		config, err := Load("")
		assert.Error(t, err)
		assert.Nil(t, config)
	})
}
