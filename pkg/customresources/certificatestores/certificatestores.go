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

// ActiveCertStores implements the CertStoreManager interface
type ActiveCertStores struct {
	// TODO: Implement concurrent safety using sync.Map
	// The structure of the map is as follows:
	// The first level maps from scope to certificate stores.
	// The second level maps from certificate store name to certificates.
	// The certificate store name is prefixed with the namespace.
	// Example:
	// {
	//   "namespace1": {
	//     "namespace1/store1": []*x509.Certificate,
	//     "namespace1/store2": []*x509.Certificate
	//   },
	//   "namespace2": {
	//     "namespace2/store1": []*x509.Certificate,
	//     "namespace2/store2": []*x509.Certificate
	//   }
	// }
	// Note: Scope is utilized for organizing and isolating cert stores. In a Kubernetes (K8s) environment, the scope can be either a namespace or an empty string ("") for cluster-wide cert stores.
	ScopedCertStores map[string]map[string][]*x509.Certificate
}

func NewActiveCertStores() CertStoreManager {
	return &ActiveCertStores{
		ScopedCertStores: make(map[string]map[string][]*x509.Certificate),
	}
}

// GetCertStores fulfills the CertStoreManager interface.
// It returns a list of cert stores for the given scope. If no cert stores are found for the given scope, it returns cluster-wide cert stores.
// TODO: Current implementation always fetches cluster-wide cert stores. Will support actual namespaced certStores in future.
func (c *ActiveCertStores) GetCertStores(_ string) map[string][]*x509.Certificate {
	return c.ScopedCertStores[constants.EmptyNamespace]
}

// AddStore fulfills the CertStoreManager interface.
// It adds the given certificate under the given scope.
// TODO: Current implementation always adds the given certificate to cluster-wide cert store. Will support actual namespaced certStores in future.
func (c *ActiveCertStores) AddStore(_, storeName string, certs []*x509.Certificate) {
	scope := constants.EmptyNamespace
	if c.ScopedCertStores[scope] == nil {
		c.ScopedCertStores[scope] = make(map[string][]*x509.Certificate)
	}
	c.ScopedCertStores[scope][storeName] = certs
}

// DeleteStore fulfills the CertStoreManager interface.
// It deletes the certificate from the given scope.
// TODO: Current implementation always deletes the cluster-wide cert store. Will support actual namespaced certStores in future.
func (c *ActiveCertStores) DeleteStore(_, storeName string) {
	if store, ok := c.ScopedCertStores[constants.EmptyNamespace]; ok {
		delete(store, storeName)
	}
}

// IsEmpty fulfills the CertStoreManager interface.
// It returns true if there are no certificates.
func (c *ActiveCertStores) IsEmpty() bool {
	count := 0
	for _, certStores := range c.ScopedCertStores {
		count += len(certStores)
	}
	return count == 0
}
