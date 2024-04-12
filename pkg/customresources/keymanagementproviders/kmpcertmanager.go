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

	"github.com/deislabs/ratify/internal/constants"
	kmp "github.com/deislabs/ratify/pkg/keymanagementprovider"
)

// ActiveCertStores implements the KMPCertManager interface.
type ActiveCertStores struct {
	// scopedStores maps from scope to CertificateMap defined in /pkg/keymanagementprovider/keymanagementprovider.go
	// Example:
	// {
	//   "namespace1": kmp.CertificateMap{},
	//   "namespace2": kmp.CertificateMap{}
	// }
	scopedStores sync.Map
}

func NewActiveCertStores() KMPCertManager {
	return &ActiveCertStores{}
}

// GetCertStores fulfills the KMPCertManager interface.
// It returns the certificates for the given scope. If no certificates are found for the given scope, it returns cluster-wide certificates.
// TODO: Current implementation always fetches cluster-wide cert stores. Will support actual namespaced certStores in future.
func (c *ActiveCertStores) GetCertStores(_, storeName string) []*x509.Certificate {
	namespacedProvider, ok := c.scopedStores.Load(constants.EmptyNamespace)
	if !ok {
		return []*x509.Certificate{}
	}
	certMap := namespacedProvider.(*kmp.CertificateMap)
	return kmp.FlattenKMPMap(certMap.GetCertificatesFromMap(storeName))
}

// AddCerts fulfills the KMPCertManager interface.
// It adds the given certificates under the given scope.
// TODO: Current implementation always adds the given certificate to cluster-wide cert store. Will support actual namespaced certStores in future.
func (c *ActiveCertStores) AddCerts(_, storeName string, certs map[kmp.KMPMapKey][]*x509.Certificate) {
	scopedStore, _ := c.scopedStores.LoadOrStore(constants.EmptyNamespace, &kmp.CertificateMap{})
	scopedStore.(*kmp.CertificateMap).SetCertificatesInMap(storeName, certs)
}

// DeleteCerts fulfills the KMPCertManager interface.
// It deletes the store from the given scope.
// TODO: Current implementation always deletes the given certificate from cluster-wide cert store. Will support actual namespaced certStores in future.
func (c *ActiveCertStores) DeleteCerts(_, storeName string) {
	scopedKMPStore, ok := c.scopedStores.Load(constants.EmptyNamespace)
	if ok {
		scopedKMPStore.(*kmp.CertificateMap).DeleteCertificatesFromMap(storeName)
	}
}
