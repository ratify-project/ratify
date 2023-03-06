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

package certificateprovider

import (
	"fmt"
)

var certificateProviders = make(map[string]CertificateProvider)

// AuthProviderFactory is an interface that defines methods to create an AuthProvider
type CertProviderFactory interface {
	Create() (CertificateProvider, error)
}

// Register adds the factory to the built in providers map
func Register(name string, factory CertProviderFactory) {
	if factory == nil {
		panic("auth provider factory cannot be nil")
	}
	_, registered := certificateProviders[name]
	if registered {
		panic(fmt.Sprintf("cert provider factory named %s already registered", name))
	}

	provider, err := factory.Create()
	if err != nil {
		panic(fmt.Sprintf("cert provider factory creation failed %s already registered", err))
	}
	certificateProviders[name] = provider
}

// returns the internal certificate provider map
func GetCertificateProviders() map[string]CertificateProvider {
	return certificateProviders
}
