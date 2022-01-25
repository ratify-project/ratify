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

package authprovider

import (
	"fmt"
	"os"
	"strings"
)

var builtInAuthProviders = make(map[string]AuthProviderFactory)

// AuthProviderFactory is an interface that defines methods to create an AuthProvider
type AuthProviderFactory interface {
	Create(authProviderConfig AuthProviderConfig) (AuthProvider, error)
}

// Register adds the factory to the built in providers map
func Register(name string, factory AuthProviderFactory) {
	if factory == nil {
		panic("auth provider factory cannot be nil")
	}
	_, registered := builtInAuthProviders[name]
	if registered {
		panic(fmt.Sprintf("auth provider factory named %s already registered", name))
	}

	builtInAuthProviders[name] = factory
}

// CreateAuthProvidersFromConfig creates the AuthProvider from the provided configuration.
// If the AuthProviderConfig isn't specified, use the default auth provider
func CreateAuthProviderFromConfig(authProviderConfig AuthProviderConfig) (AuthProvider, error) {
	// if auth provider not specified in config, return default provider
	if authProviderConfig == nil {
		return builtInAuthProviders[DefaultAuthProviderName].Create(authProviderConfig)
	}

	err := validateAuthProviderConfig(authProviderConfig)
	if err != nil {
		return nil, err
	}

	authProviderName, ok := authProviderConfig["name"]
	if !ok {
		return nil, fmt.Errorf("failed to find auth provider name in the auth providers config with key %s", "name")
	}

	providerNameStr := fmt.Sprintf("%s", authProviderName)
	if strings.ContainsRune(providerNameStr, os.PathSeparator) {
		return nil, fmt.Errorf("invalid auth provider name: %s", authProviderName)
	}

	authFactory, ok := builtInAuthProviders[providerNameStr]
	if ok {
		authProvider, err := authFactory.Create(authProviderConfig)
		if err != nil {
			return nil, err
		}

		return authProvider, nil
	}

	// return default docker config auth provider if no matching provider exists
	return builtInAuthProviders[DefaultAuthProviderName].Create(authProviderConfig)
}

// TODO: add validation
func validateAuthProviderConfig(authProviderConfig AuthProviderConfig) error {
	return nil
}
