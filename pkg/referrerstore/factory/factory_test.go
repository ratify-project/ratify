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
	"testing"

	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/deislabs/ratify/pkg/referrerstore/mocks"
	"github.com/deislabs/ratify/pkg/referrerstore/plugin"
)

type TestStoreFactory struct{}

func (f *TestStoreFactory) Create(version string, storesConfig config.StorePluginConfig) (referrerstore.ReferrerStore, error) {
	return &mocks.TestStore{}, nil
}

func TestCreateStoresFromConfig_BuiltInStores_ReturnsExpected(t *testing.T) {
	builtInStores = map[string]StoreFactory{
		"testStore": &TestStoreFactory{},
	}

	var storeConfig config.StorePluginConfig
	storeConfig = map[string]interface{}{
		"name": "testStore",
	}
	storesConfig := config.StoresConfig{
		Stores: []config.StorePluginConfig{storeConfig},
	}

	stores, err := CreateStoresFromConfig(storesConfig, "")

	if err != nil {
		t.Fatalf("create stores failed with err %v", err)
	}

	if len(stores) != 1 {
		t.Fatalf("expected to have %d stores, actual count %d", 1, len(stores))
	}

	if stores[0].Name() != "testStore" {
		t.Fatalf("expected to create test store")
	}

	if _, ok := stores[0].(*plugin.StorePlugin); ok {
		t.Fatalf("type assertion failed expected a built in store")
	}
}

func TestCreateStoresFromConfig_PluginStores_ReturnsExpected(t *testing.T) {
	var storeConfig config.StorePluginConfig
	storeConfig = map[string]interface{}{
		"name": "plugin-store",
	}
	storesConfig := config.StoresConfig{
		Stores: []config.StorePluginConfig{storeConfig},
	}

	stores, err := CreateStoresFromConfig(storesConfig, "")

	if err != nil {
		t.Fatalf("create stores failed with err %v", err)
	}

	if len(stores) != 1 {
		t.Fatalf("expected to have %d stores, actual count %d", 1, len(stores))
	}

	if stores[0].Name() != "plugin-store" {
		t.Fatalf("expected to create plugin store")
	}

	if _, ok := stores[0].(*plugin.StorePlugin); !ok {
		t.Fatalf("type assertion failed expected a plugin store")
	}
}
