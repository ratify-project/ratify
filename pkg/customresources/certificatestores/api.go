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
	"context"
	"crypto/x509"
)

// CertStoreManager is an interface that defines the methods for managing certificate stores across different scopes.
type CertStoreManager interface {
	// GetCertsFromStore returns certificates from the given certificate store.
	GetCertsFromStore(ctx context.Context, storeName string) ([]*x509.Certificate, error)

	// AddStore adds the given certificate.
	AddStore(storeName string, cert []*x509.Certificate)

	// DeleteStore deletes the certificate from the given scope.
	DeleteStore(storeName string)
}
