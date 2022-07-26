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
	"context"
	"testing"
)

type TestAuthProviderFactory struct{}

func (f *TestAuthProviderFactory) Create(authProviderConfig AuthProviderConfig) (AuthProvider, error) {
	return &TestAuthProvider{}, nil
}

// Checks the correct registered auth provider is invoked based on config
func TestCreateAuthProvidersFromConfig_BuiltInAuthProviders_ReturnsExpected(t *testing.T) {
	builtInAuthProviders = map[string]AuthProviderFactory{
		"testAuthProvider": &TestAuthProviderFactory{},
	}

	authProviderConfig := map[string]interface{}{
		"name": "testAuthProvider",
	}

	authProvider, err := CreateAuthProviderFromConfig(authProviderConfig)

	if err != nil {
		t.Fatalf("create auth provider failed with err %v", err)
	}

	authConfig, err := authProvider.Provide(context.Background(), "test-artifact")
	if err != nil {
		t.Fatalf("provide failed to get credential with err %v", err)
	}

	if authConfig.Username != "test" || authConfig.Password != "testpw" {
		t.Fatalf("provide failed to return correct credentials")
	}
}

// Checks the auth provider creation fails if auth provider specified does not exist
func TestCreateAuthProvidersFromConfig_NonexistentAuthProviders_ReturnsExpected(t *testing.T) {
	builtInAuthProviders = map[string]AuthProviderFactory{
		"dockerConfig": &defaultProviderFactory{},
	}

	authProviderConfig := map[string]interface{}{
		"name": "test-non-existent",
	}

	_, err := CreateAuthProviderFromConfig(authProviderConfig)

	if err == nil {
		t.Fatalf("create auth provider should have failed for non existent provider")
	}
}

// Checks the default auth provider is returned when no auth provider is
// specified in config
func TestCreateAuthProvidersFromConfig_NullAuthProviders_ReturnsExpected(t *testing.T) {
	builtInAuthProviders = map[string]AuthProviderFactory{
		"dockerConfig": &defaultProviderFactory{},
	}

	authProvider, err := CreateAuthProviderFromConfig(nil)

	if err != nil {
		t.Fatalf("create auth provider failed with err %v", err)
	}

	providerType, isType := authProvider.(*defaultAuthProvider)
	if !isType {
		t.Fatalf("default auth provider not returned. instead provider of type %v returned", providerType)
	}
}
