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

package utils

import (
	"encoding/json"
	"fmt"

	c "github.com/ratify-project/ratify/config"
	kmp "github.com/ratify-project/ratify/pkg/keymanagementprovider"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/config"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/factory"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/types"
)

// SpecToKeyManagementProvider creates KeyManagementProvider from  KeyManagementProviderSpec config
func SpecToKeyManagementProvider(raw []byte, keyManagamentSystemName string) (kmp.KeyManagementProvider, error) {
	kmProviderConfig, err := rawToKeyManagementProviderConfig(raw, keyManagamentSystemName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key management provider config: %w", err)
	}

	// TODO: add Version and Address to KeyManagementProviderSpec
	keyManagementProviderProvider, err := factory.CreateKeyManagementProviderFromConfig(kmProviderConfig, "0.1.0", c.GetDefaultPluginPath())
	if err != nil {
		return nil, fmt.Errorf("failed to create key management provider provider: %w", err)
	}

	return keyManagementProviderProvider, nil
}

// rawToKeyManagementProviderConfig converts raw json to KeyManagementProviderConfig
func rawToKeyManagementProviderConfig(raw []byte, keyManagamentSystemName string) (config.KeyManagementProviderConfig, error) {
	pluginConfig := config.KeyManagementProviderConfig{}

	if string(raw) == "" {
		return config.KeyManagementProviderConfig{}, fmt.Errorf("no key management provider parameters provided")
	}
	if err := json.Unmarshal(raw, &pluginConfig); err != nil {
		return config.KeyManagementProviderConfig{}, fmt.Errorf("unable to decode key management provider parameters.Raw: %s, err: %w", raw, err)
	}

	pluginConfig[types.Type] = keyManagamentSystemName

	return pluginConfig, nil
}
