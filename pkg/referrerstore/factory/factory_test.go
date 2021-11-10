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
	"context"
	"testing"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/deislabs/ratify/pkg/referrerstore/plugin"
	"github.com/opencontainers/go-digest"
)

type TestStore struct{}
type TestStoreFactory struct{}

func (s *TestStore) Name() string {
	return "test-store"
}

func (s *TestStore) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string) (referrerstore.ListReferrersResult, error) {
	return referrerstore.ListReferrersResult{}, nil
}

func (s *TestStore) GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	return nil, nil
}

func (s *TestStore) GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error) {
	return ocispecs.ReferenceManifest{}, nil
}

func (s *TestStore) GetConfig() *config.StoreConfig {
	return nil
}

func (f *TestStoreFactory) Create(version string, storesConfig config.StorePluginConfig) (referrerstore.ReferrerStore, error) {
	return &TestStore{}, nil
}

func TestCreateStoresFromConfig_BuiltInStores_ReturnsExpected(t *testing.T) {
	builtInStores = map[string]StoreFactory{
		"test-store": &TestStoreFactory{},
	}

	var storeConfig config.StorePluginConfig
	storeConfig = map[string]interface{}{
		"name": "test-store",
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

	if stores[0].Name() != "test-store" {
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
