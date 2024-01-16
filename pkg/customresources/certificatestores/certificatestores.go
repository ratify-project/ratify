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

package certificatestores

import (
	"crypto/x509"

	"github.com/deislabs/ratify/internal/constants"
)

type ActiveCertStores struct {
	NamespacedCertStores map[string]map[string][]*x509.Certificate
}

func NewActiveCertStores() ActiveCertStores {
	return ActiveCertStores{
		NamespacedCertStores: make(map[string]map[string][]*x509.Certificate),
	}
}

// GetCertStores implements the CertificateStores interface.
// It returns a list of cert stores for the given scope. If no cert stores are found for the given scope, it returns cluster-wide cert stores.
func (c *ActiveCertStores) GetCertStores(scope string) map[string][]*x509.Certificate {
	if _, ok := c.NamespacedCertStores[scope]; ok {
		return c.NamespacedCertStores[scope]
	}
	return c.NamespacedCertStores[constants.EmptyNamespace]
}

func (c *ActiveCertStores) AddStore(scope, storeName string, certs []*x509.Certificate) {
	if c.NamespacedCertStores[scope] == nil {
		c.NamespacedCertStores[scope] = make(map[string][]*x509.Certificate)
	}
	c.NamespacedCertStores[scope][storeName] = certs
}

func (c *ActiveCertStores) DeleteStore(scope, storeName string) {
	if store, ok := c.NamespacedCertStores[scope]; ok {
		delete(store, storeName)
		if len(store) == 0 {
			delete(c.NamespacedCertStores, scope)
		}
	}
}

func (c *ActiveCertStores) IsEmpty() bool {
	return len(c.NamespacedCertStores) == 0
}
