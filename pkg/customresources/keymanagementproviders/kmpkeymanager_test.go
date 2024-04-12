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
	"testing"

	"github.com/deislabs/ratify/pkg/keymanagementprovider"
	"github.com/deislabs/ratify/pkg/utils"
)

func TestKeyStoresOperations(t *testing.T) {
	activeKeyStores := NewActiveKeyStores()

	keyStore1 := map[keymanagementprovider.KMPMapKey]crypto.PublicKey{
		{Name: "testName1", Version: "testVersion1"}: utils.CreateTestPublicKey(),
	}
	keyStore2 := map[keymanagementprovider.KMPMapKey]crypto.PublicKey{
		{Name: "testName2", Version: "testVersion2"}: utils.CreateTestPublicKey(),
	}

	if len(activeKeyStores.GetKeyStores(namespace1, name1)) != 0 {
		t.Errorf("Expected activeKeyStores to have 0 key store, but got %d", len(activeKeyStores.GetKeyStores(namespace1, name1)))
	}

	activeKeyStores.AddKeys(namespace1, name1, keyStore1)
	activeKeyStores.AddKeys(namespace2, name2, keyStore2)

	if len(activeKeyStores.GetKeyStores(namespace1, name1)) != 1 {
		t.Errorf("Expected activeKeyStores to have 1 key store, but got %d", len(activeKeyStores.GetKeyStores(namespace1, name1)))
	}
	if len(activeKeyStores.GetKeyStores(namespace2, name2)) != 1 {
		t.Errorf("Expected activeKeyStores to have 1 key store, but got %d", len(activeKeyStores.GetKeyStores(namespace2, name2)))
	}

	activeKeyStores.DeleteKeys(namespace1, name1)
	activeKeyStores.DeleteKeys(namespace2, name2)

	if len(activeKeyStores.GetKeyStores(namespace1, name1)) != 0 {
		t.Errorf("Expected activeKeyStores to have 0 key store, but got %d", len(activeKeyStores.GetKeyStores(namespace1, name1)))
	}
	if len(activeKeyStores.GetKeyStores(namespace2, name2)) != 0 {
		t.Errorf("Expected activeKeyStores to have 0 key store, but got %d", len(activeKeyStores.GetKeyStores(namespace2, name2)))
	}
}
