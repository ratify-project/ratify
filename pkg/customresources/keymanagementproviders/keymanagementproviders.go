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
	"crypto/x509"
	"sync"

	kmp "github.com/deislabs/ratify/pkg/keymanagementprovider"
)

type ActiveKMPs struct {
	NamespacedKMPs sync.Map
	//NamespacedKMPs map[string]map[kmp.KMPMapKey][]*x509.Certificate
}

func NewActiveKMPs() KeyManagementProviders {
	return &ActiveKMPs{}
}

func (c *ActiveKMPs) GetCertStores(scope, certStore string) []*x509.Certificate {
	namespacedProvider, ok := c.NamespacedKMPs.Load(scope)
	if !ok {
		return []*x509.Certificate{}
	}
	provider := namespacedProvider.(*sync.Map)
	return kmp.FlattenKMPMap(kmp.GetCertificatesFromMap(provider, certStore))
}

func (c *ActiveKMPs) AddStore(scope, certName string, certs map[kmp.KMPMapKey][]*x509.Certificate) {
	namespacedKMPStore, _ := c.NamespacedKMPs.LoadOrStore(scope, &sync.Map{})
	kmp.SetCertificatesInMap(namespacedKMPStore.(*sync.Map), certName, certs)
}

func (c *ActiveKMPs) DeleteStore(scope, certName string) {
	namespacedKMPStore, ok := c.NamespacedKMPs.Load(scope)
	if ok {
		kmp.DeleteCertificatesFromMap(namespacedKMPStore.(*sync.Map), certName)
	}
}

func (c *ActiveKMPs) IsEmpty() bool {
	isEmpty := true
	c.NamespacedKMPs.Range(func(key, value interface{}) bool {
		store := key.(*sync.Map)
		isEmptyInnerMap := true
		store.Range(func(_, _ interface{}) bool {
			isEmptyInnerMap = false
			return false
		})
		if !isEmptyInnerMap {
			isEmpty = false
			return false
		}
		return true
	})
	return isEmpty
}
