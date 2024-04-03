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

package referrerstores

import (
	"fmt"

	"github.com/deislabs/ratify/pkg/referrerstore"
)

type ActiveStores struct {
	NamespacedStores map[string]map[string]referrerstore.ReferrerStore
}

func NewActiveStores() ActiveStores {
	return ActiveStores{
		NamespacedStores: make(map[string]map[string]referrerstore.ReferrerStore),
	}
}

func NewActiveStoresWithoutNames(stores []referrerstore.ReferrerStore) ActiveStores {
	activeStores := make(map[string]map[string]referrerstore.ReferrerStore)
	activeStores[""] = make(map[string]referrerstore.ReferrerStore)

	for index, store := range stores {
		activeStores[""][fmt.Sprintf("%d", index)] = store
	}

	return ActiveStores{
		NamespacedStores: activeStores,
	}
}

// GetStores implements the Stores interface.
// It returns all the stores in the ActiveStores for the given scope. If no stores are found for the given scope, it returns cluster-wide stores.
func (s *ActiveStores) GetStores(scope string) []referrerstore.ReferrerStore {
	stores := []referrerstore.ReferrerStore{}
	for _, namespacedStores := range s.NamespacedStores {
		for _, store := range namespacedStores {
			stores = append(stores, store)
		}
	}
	return stores
}

func (s *ActiveStores) AddStore(scope, storeName string, store referrerstore.ReferrerStore) {
	if _, ok := s.NamespacedStores[scope]; !ok {
		s.NamespacedStores[scope] = make(map[string]referrerstore.ReferrerStore)
	}
	s.NamespacedStores[scope][storeName] = store
}

func (s *ActiveStores) DeleteStore(scope, storeName string) {
	if stores, ok := s.NamespacedStores[scope]; ok {
		delete(stores, storeName)
	}
}

func (s *ActiveStores) IsEmpty() bool {
	return s.GetStoreCount() == 0
}

func (s *ActiveStores) GetStoreCount() int {
	count := 0
	for _, stores := range s.NamespacedStores {
		count += len(stores)
	}
	return count
}
