/*
Copyright The Ratify Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
IdentityTHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	provider "github.com/deislabs/ratify/pkg/referrerstore/oras/authprovider"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/services/preview/containerregistry/runtime/2019-08-15-preview/containerregistry"
)

type azureManagedIdentityProviderFactory struct{}
type azureManagedIdentityAuthProvider struct {
	identityToken azcore.AccessToken
	clientID      string
	tenantID      string
}

type azureManagedIdentityAuthProviderConf struct {
	Name     string `json:"name"`
	ClientID string `json:"clientID"`
}

var (
	// ErrorNoAuth indicates that no credentials are provided.
	ErrorNoAuth    = fmt.Errorf("no credentials provided for Azure cloud provider")
	ErrorNilEnv    = fmt.Errorf("env is nil pointer for GetServicePrincipalToken")
	ErrorNilConfig = fmt.Errorf("config is nil pointer for GetServicePrincipalToken")
)

const (
	azureManagedIdentityAuthProviderName string = "azureManagedIdentity"
)

// init calls Register for our Azure Workload Identity provider
func init() {
	provider.Register(azureManagedIdentityAuthProviderName, &azureManagedIdentityProviderFactory{})
}

// Create returns an azureManagedIdentityAuthProvider
func (s *azureManagedIdentityProviderFactory) Create(authProviderConfig provider.AuthProviderConfig) (provider.AuthProvider, error) {
	conf := azureManagedIdentityAuthProviderConf{}
	authProviderConfigBytes, err := json.Marshal(authProviderConfig)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(authProviderConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse azure managed identity auth provider configuration: %v", err)
	}

	tenant := os.Getenv("AZURE_TENANT_ID")
	if tenant == "" {
		return nil, fmt.Errorf("AZURE_TENANT_ID environment variable is empty")
	}
	client := os.Getenv("AZURE_CLIENT_ID")
	if client == "" {
		client = conf.ClientID
		if client == "" {
			return nil, fmt.Errorf("AZURE_CLIENT_ID environment variable is empty")
		}
	}
	if err != nil {
		return nil, err
	}
	// retrieve an AAD Access token
	token, err := getManagedIdentityToken(context.Background(), client)
	if err != nil {
		return nil, err
	}

	return &azureManagedIdentityAuthProvider{
		identityToken: token,
		clientID:      client,
		tenantID:      tenant,
	}, nil
}

// Enabled checks for non empty tenant ID and AAD access token
func (d *azureManagedIdentityAuthProvider) Enabled(ctx context.Context) bool {
	if d.clientID == "" {
		return false
	}

	if d.tenantID == "" {
		return false
	}

	if d.identityToken.Token == "" {
		return false
	}

	return true
}

// Provide returns the credentials for a specified artifact.
// Uses Managed Identity to retrieve an AAD access token which can be
// exchanged for a valid ACR refresh token for login.
func (d *azureManagedIdentityAuthProvider) Provide(ctx context.Context, artifact string) (provider.AuthConfig, error) {
	if !d.Enabled(ctx) {
		return provider.AuthConfig{}, fmt.Errorf("azure managed identity provider is not properly enabled")
	}
	// parse the artifact reference string to extract the registry host name
	artifactHostName, err := provider.GetRegistryHostName(artifact)
	if err != nil {
		return provider.AuthConfig{}, err
	}

	// need to refresh AAD token if it's expired
	if time.Now().Add(time.Minute * 5).After(d.identityToken.ExpiresOn) {
		newToken, err := getManagedIdentityToken(ctx, d.clientID)
		if err != nil {
			return provider.AuthConfig{}, errors.Wrap(err, "could not refresh azure managed identity token")
		}
		d.identityToken = newToken
		logrus.Info("sucessfully refreshed azure managed identity token")
	}
	// add protocol to generate complete URI
	serverUrl := "https://" + artifactHostName

	// create registry client and exchange AAD token for registry refresh token
	refreshTokenClient := containerregistry.NewRefreshTokensClient(serverUrl)
	rt, err := refreshTokenClient.GetFromExchange(ctx, "access_token", artifactHostName, d.tenantID, "", d.identityToken.Token)
	if err != nil {
		return provider.AuthConfig{}, fmt.Errorf("failed to get refresh token for container registry by azure managed identity token - %w", err)
	}

	authConfig := provider.AuthConfig{
		Username:  dockerTokenLoginUsernameGUID,
		Password:  *rt.RefreshToken,
		Provider:  d,
		ExpiresOn: d.identityToken.ExpiresOn,
	}

	return authConfig, nil
}

func getManagedIdentityToken(ctx context.Context, clientID string) (azcore.AccessToken, error) {
	id := azidentity.ClientID(clientID)
	opts := azidentity.ManagedIdentityCredentialOptions{ID: id}
	cred, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		return azcore.AccessToken{}, err
	}
	scopes := []string{AADResource}
	if cred != nil {
		return cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: scopes})
	}
	return azcore.AccessToken{}, ErrorNilConfig
}
