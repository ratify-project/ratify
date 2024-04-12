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

package keymanagementproviders

import (
	"crypto"
	"sync"

	"github.com/deislabs/ratify/internal/constants"
	kmp "github.com/deislabs/ratify/pkg/keymanagementprovider"
)

// ActiveKeyStores implements the KMPKeyManager interface.
type ActiveKeyStores struct {
	// scopedStores maps from scope to KeyMap defined in /pkg/keymanagementprovider/keymanagementprovider.go
	// Example:
	// {
	//	 "namespace1": kmp.KeyMap{},
	//   "namespace2": kmp.KeyMap{}
	// }
	scopedStores sync.Map
}

func NewActiveKeyStores() KMPKeyManager {
	return &ActiveKeyStores{}
}

// GetKeyStores fulfills the KMPKeyManager interface.
// It returns the keys for the given scope. If no keys are found for the given scope, it returns cluster-wide keys.
// TODO: Current implementation always fetches cluster-wide key stores. Will support actual namespaced keyStores in future.
func (k *ActiveKeyStores) GetKeyStores(_, storeName string) map[kmp.KMPMapKey]crypto.PublicKey {
	namespacedProvider, ok := k.scopedStores.Load(constants.EmptyNamespace)
	if !ok {
		return map[kmp.KMPMapKey]crypto.PublicKey{}
	}
	keyMap := namespacedProvider.(*kmp.KeyMap)
	return keyMap.GetKeysFromMap(storeName)
}

// AddKeys fulfills the KMPKeyManager interface.
// It adds the given keys under the given scope.
// TODO: Current implementation always adds cluster-wide key stores. Will support actual namespaced keyStores in future.
func (k *ActiveKeyStores) AddKeys(_, storeName string, keys map[kmp.KMPMapKey]crypto.PublicKey) {
	scopedStore, _ := k.scopedStores.LoadOrStore(constants.EmptyNamespace, &kmp.KeyMap{})
	scopedStore.(*kmp.KeyMap).SetKeysInMap(storeName, keys)
}

// DeleteKeys fulfills the KMPKeyManager interface.
// It deletes the keys for the given scope.
// TODO: Current implementation always deletes cluster-wide key stores. Will support actual namespaced keyStores in future.
func (k *ActiveKeyStores) DeleteKeys(_, storeName string) {
	scopedKMPStore, ok := k.scopedStores.Load(constants.EmptyNamespace)
	if ok {
		scopedKMPStore.(*kmp.KeyMap).DeleteKeysFromMap(storeName)
	}
}
