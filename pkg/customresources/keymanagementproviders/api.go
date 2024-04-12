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
	"crypto/x509"

	kmp "github.com/deislabs/ratify/pkg/keymanagementprovider"
)

// KMPCertManager is an interface that defines the methods for managing certificate stores across different scopes.
type KMPCertManager interface {
	// GetCertStores returns certificates for the given scope.
	GetCertStores(scope, storeName string) []*x509.Certificate

	// AddCerts adds the given certificate under the given scope.
	AddCerts(scope, storeName string, certs map[kmp.KMPMapKey][]*x509.Certificate)

	// DeleteCerts deletes the store from the given scope.
	DeleteCerts(scope, storeName string)
}

// KMPKeyManager is an interface that defines the methods for managing key stores across different scopes.
type KMPKeyManager interface {
	// GetKeyStores returns keys for the given scope.
	GetKeyStores(scope, storeName string) map[kmp.KMPMapKey]crypto.PublicKey

	// AddKeys adds the given keys under the given scope.
	AddKeys(scope, storeName string, keys map[kmp.KMPMapKey]crypto.PublicKey)

	// DeleteKeys deletes the store from the given scope.
	DeleteKeys(scope, storeName string)
}
