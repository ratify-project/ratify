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

import "crypto/x509"

// CertStoreManager is an interface that defines the methods for managing certificate stores across different scopes.
type CertStoreManager interface {
	// GetCertStores returns certificates for the given scope.
	GetCertStores(scope string) map[string][]*x509.Certificate

	// AddStore adds the given certificate under the given scope.
	AddStore(scope, storeName string, cert []*x509.Certificate)

	// DeleteStore deletes the certificate from the given scope.
	DeleteStore(scope, storeName string)

	// IsEmpty returns true if there are no certificates.
	IsEmpty() bool
}
