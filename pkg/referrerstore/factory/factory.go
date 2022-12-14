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

package factory

import (
	"errors"
	"fmt"
	"os"
	"strings"

	pluginCommon "github.com/deislabs/ratify/pkg/common/plugin"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/deislabs/ratify/pkg/referrerstore/plugin"
	"github.com/deislabs/ratify/pkg/referrerstore/types"
	"github.com/sirupsen/logrus"
)

var builtInStores = make(map[string]StoreFactory)

// StoreFactory is an interface that defines methods to create a store
type StoreFactory interface {
	Create(version string, storesConfig config.StorePluginConfig) (referrerstore.ReferrerStore, error)
}

func Register(name string, factory StoreFactory) {
	if factory == nil {
		panic("store factory cannot be nil")
	}
	_, registered := builtInStores[name]
	if registered {
		panic(fmt.Sprintf("store factory named %s already registered", name))
	}

	builtInStores[name] = factory
}

func CreateStoreFromConfig(storeConfig config.StorePluginConfig, configVersion string, pluginBinDir []string) (referrerstore.ReferrerStore, error) {
	storeName, ok := storeConfig[types.Name]
	if !ok {
		return nil, fmt.Errorf("failed to find store name in the stores config with key %s", "name")
	}

	storeNameStr := fmt.Sprintf("%s", storeName)
	if strings.ContainsRune(storeNameStr, os.PathSeparator) {
		return nil, fmt.Errorf("invalid plugin name for a store: %s", storeName)
	}

	// if source is specified, download the plugin
	if source, ok := storeConfig[types.Source]; ok {
		sourceStr := fmt.Sprintf("%s", source)
		err := pluginCommon.DownloadPlugin(storeNameStr, sourceStr, pluginBinDir[0])
		if err != nil {
			return nil, fmt.Errorf("failed to download plugin: %v", err)
		}
		logrus.Infof("downloaded store plugin %s from %s to %s", storeNameStr, sourceStr, pluginBinDir[0])
	}

	storeFactory, ok := builtInStores[storeNameStr]
	if ok {
		return storeFactory.Create(configVersion, storeConfig)
	} else {
		return plugin.NewStore(configVersion, storeConfig, pluginBinDir)
	}
}

// CreateStoresFromConfig creates a stores from the provided configuration
func CreateStoresFromConfig(storesConfig config.StoresConfig, defaultPluginPath string) ([]referrerstore.ReferrerStore, error) {
	if storesConfig.Version == "" {
		storesConfig.Version = types.SpecVersion
	}

	err := validateStoresConfig(&storesConfig)
	if err != nil {
		return nil, err
	}

	if len(storesConfig.Stores) == 0 {
		return nil, errors.New("referrer store config should have at least one store")
	}

	var stores []referrerstore.ReferrerStore

	if len(storesConfig.PluginBinDirs) == 0 {
		storesConfig.PluginBinDirs = []string{defaultPluginPath}
	} else {
		storesConfig.PluginBinDirs = append(storesConfig.PluginBinDirs, defaultPluginPath)
	}

	for _, storeConfig := range storesConfig.Stores {
		store, err := CreateStoreFromConfig(storeConfig, storesConfig.Version, storesConfig.PluginBinDirs)
		if err != nil {
			return nil, err
		} else {
			stores = append(stores, store)
		}
	}

	return stores, nil
}

func validateStoresConfig(storesConfig *config.StoresConfig) error {
	// TODO check for existence of plugin dirs
	// TODO check if version is supported
	return nil

}
