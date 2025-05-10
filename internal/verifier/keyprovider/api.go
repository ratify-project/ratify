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

package keyprovider

import (
	"context"
	"crypto/x509"
	"fmt"
)

// KeyProvider defines methods to fetch crypto material for signature
// verification.
type KeyProvider interface {
	GetCertificates(ctx context.Context) ([]*x509.Certificate, error)
}

type keyProviderFactory func(options any) (KeyProvider, error)

var keyProviderFactories = make(map[string]keyProviderFactory)

// RegisterKeyProvider registers a key provider factory with the given name.
func RegisterKeyProvider(name string, factory keyProviderFactory) {
	keyProviderFactories[name] = factory
}

// CreateKeyProvider creates a new key provider instance.
func CreateKeyProvider(name string, options any) (KeyProvider, error) {
	factory, exists := keyProviderFactories[name]
	if !exists {
		return nil, fmt.Errorf("key provider %s not registered", name)
	}
	return factory(options)
}
