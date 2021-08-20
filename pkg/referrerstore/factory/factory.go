package factory

import (
	"errors"

	"github.com/notaryproject/hora/pkg/referrerstore"
	"github.com/notaryproject/hora/pkg/referrerstore/config"
	"github.com/notaryproject/hora/pkg/referrerstore/plugin"
	"github.com/notaryproject/hora/pkg/referrerstore/types"
)

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
		store, err := plugin.NewStore(storesConfig.Version, storeConfig, append(storesConfig.PluginBinDirs, defaultPluginPath))

		if err != nil {
			return nil, err
		}

		stores = append(stores, store)
	}

	return stores, nil
}

func validateStoresConfig(storesConfig *config.StoresConfig) error {
	// TODO check for existence of plugin dirs
	// TODO check if version is supported
	return nil

}
