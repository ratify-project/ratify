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

package azure

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/containers/azcontainerregistry"
	provider "github.com/ratify-project/ratify/pkg/common/oras/authprovider"
)

const GrantTypeAccessToken = "access_token"

// AuthClientFactory defines an interface for creating an authentication client.
type AuthClientFactory interface {
	CreateAuthClient(serverURL string, options *azcontainerregistry.AuthenticationClientOptions) (AuthClient, error)
}

// defaultAuthClientFactoryImpl is the default implementation of AuthClientFactory.
type defaultAuthClientFactoryImpl struct{}

// creates an AuthClient using the default factory implementation.
// Return an AuthClient and an error if the client creation fails.
func (f *defaultAuthClientFactoryImpl) CreateAuthClient(serverURL string, options *azcontainerregistry.AuthenticationClientOptions) (AuthClient, error) {
	return defaultAuthClientFactory(serverURL, options)
}

// Define a helper function that creates an instance of AuthenticationClientWrapper.
func defaultAuthClientFactory(serverURL string, options *azcontainerregistry.AuthenticationClientOptions) (AuthClient, error) {
	client, err := azcontainerregistry.NewAuthenticationClient(serverURL, options)
	if err != nil {
		return nil, err
	}
	return &AuthenticationClientWrapper{client: client}, nil
}

// Define the interface for azcontainerregistry.AuthenticationClient methods used
type AuthenticationClientInterface interface {
	ExchangeAADAccessTokenForACRRefreshToken(ctx context.Context, grantType azcontainerregistry.PostContentSchemaGrantType, service string, options *azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenOptions) (azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenResponse, error)
}

// Define the wrapper for AuthenticationClientInterface
type AuthenticationClientWrapper struct {
	client AuthenticationClientInterface
}

// A wrapper method that calls the underlying AuthenticationClientInterface's method.
// Exchanges an AAD access token for an ACR refresh token.
func (w *AuthenticationClientWrapper) ExchangeAADAccessTokenForACRRefreshToken(ctx context.Context, grantType azcontainerregistry.PostContentSchemaGrantType, service string, options *azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenOptions) (azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenResponse, error) {
	return w.client.ExchangeAADAccessTokenForACRRefreshToken(ctx, grantType, service, options)
}

// define the interface for authentication operations.
// It includes the method for exchanging an AAD access token for an ACR refresh token.
type AuthClient interface {
	ExchangeAADAccessTokenForACRRefreshToken(ctx context.Context, grantType azcontainerregistry.PostContentSchemaGrantType, service string, options *azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenOptions) (azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenResponse, error)
}

// RegistryHostGetter defines an interface for getting the registry host.
type RegistryHostGetter interface {
	GetRegistryHost(artifact string) (string, error)
}

// defaultRegistryHostGetterImpl is the default implementation of RegistryHostGetter.
type defaultRegistryHostGetterImpl struct{}

// Retrieves the registry host name for a given artifact.
// It utilizes the provider's GetRegistryHostName function to perform the lookup.
func (g *defaultRegistryHostGetterImpl) GetRegistryHost(artifact string) (string, error) {
	return provider.GetRegistryHostName(artifact)
}

// parseEndpoints checks if the endpoints are valid for auth provider. If no
// endpoints are provided, it defaults to the default ACR endpoint.
// A valid endpoint is either a fully qualified domain name or a wildcard domain
// name folloiwing RFC 1034.
// Valid examples:
// - *.example.com
// - example.com
//
// Invalid examples:
// - *
// - example.*
// - *example.com.*
// - *.
func parseEndpoints(endpoints []string) ([]string, error) {
	if len(endpoints) == 0 {
		return defaultACREndpoints, nil
	}
	for _, endpoint := range endpoints {
		switch strings.Count(endpoint, "*") {
		case 0:
			continue
		case 1:
			if !strings.HasPrefix(endpoint, "*.") {
				return nil, fmt.Errorf("invalid wildcard domain name: %s, it must start with '*.'", endpoint)
			}
			if len(endpoint) < 3 {
				return nil, fmt.Errorf("invalid wildcard domain name: %s, it must have at least one character after '*.'", endpoint)
			}
		default:
			return nil, fmt.Errorf("invalid wildcard domain name: %s, it must have at most one wildcard character", endpoint)
		}
	}
	return endpoints, nil
}

// validateHost checks if the host is matching endpoints supported by the auth
// provider.
func validateHost(host string, endpoints []string) error {
	for _, endpoint := range endpoints {
		if endpoint[0] == '*' {
			if _, zone, ok := strings.Cut(host, "."); ok && zone == endpoint[2:] {
				return nil
			}
		}
		if host == endpoint {
			return nil
		}
	}
	return fmt.Errorf("the artifact host %s is not in the scope of the store auth provider", host)
}
