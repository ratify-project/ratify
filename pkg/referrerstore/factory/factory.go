package factory

import (
	"errors"
	"fmt"

	"github.com/deislabs/hora/pkg/referrerstore"
	"github.com/deislabs/hora/pkg/referrerstore/config"
	"github.com/deislabs/hora/pkg/referrerstore/plugin"
	"github.com/deislabs/hora/pkg/referrerstore/types"
)

var builtInStores = make(map[string]StoreFactory)

type StoreFactory interface {
	Create(version string, storesConfig config.StorePluginConfig) (referrerstore.ReferrerStore, error)
}

func Register(name string, factory StoreFactory) {
	if factory == nil {
		panic("Store factor cannot be nil")
	}
	_, registered := builtInStores[name]
	if registered {
		panic(fmt.Sprintf("store factory named %s already registered", name))
	}

	builtInStores[name] = factory
}

func CreateStoresFromConfig(storesConfig config.StoresConfig, defaultPluginPath string) ([]referrerstore.ReferrerStore, error) {
	if storesConfig.Version == "" {
		storesConfig.Version = types.SpecVersion
	}

	err := validateStoresConfig(&storesConfig)
	if err != nil {
		return nil, err
	}

	if len(storesConfig.Stores) == 0 {
		return nil, errors.New("referrer store config should have atleast one store")
	}

	var stores []referrerstore.ReferrerStore

	if len(storesConfig.PluginBinDirs) == 0 {
		storesConfig.PluginBinDirs = []string{defaultPluginPath}
	}
	for _, storeConfig := range storesConfig.Stores {
		storeName, ok := storeConfig[types.Name]
		if !ok {
			return nil, fmt.Errorf("failed to find store name in the stores config with key %s", "name")
		}

		storeFactory, ok := builtInStores[fmt.Sprintf("%s", storeName)]
		if ok {
			store, err := storeFactory.Create(storesConfig.Version, storeConfig)

			if err != nil {
				return nil, err
			}

			stores = append(stores, store)
		} else {
			store, err := plugin.NewStore(storesConfig.Version, storeConfig, append(storesConfig.PluginBinDirs, defaultPluginPath))

			if err != nil {
				return nil, err
			}

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
