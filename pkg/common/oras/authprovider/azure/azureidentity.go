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
	"encoding/json"
	"fmt"
	"os"
	"time"

	re "github.com/deislabs/ratify/errors"
	"github.com/deislabs/ratify/internal/logger"
	provider "github.com/deislabs/ratify/pkg/common/oras/authprovider"

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
		return nil, re.ErrorCodeConfigInvalid.WithError(err).WithComponentType(re.AuthProvider)
	}

	if err := json.Unmarshal(authProviderConfigBytes, &conf); err != nil {
		return nil, re.ErrorCodeConfigInvalid.NewError(re.AuthProvider, "", re.AzureManagedIdentityLink, err, "failed to parse azure managed identity auth provider configuration.", re.HideStackTrace)
	}

	tenant := os.Getenv("AZURE_TENANT_ID")
	if tenant == "" {
		return nil, re.ErrorCodeEnvNotSet.WithDetail("AZURE_TENANT_ID environment variable is empty").WithComponentType(re.AuthProvider)
	}
	client := os.Getenv("AZURE_CLIENT_ID")
	if client == "" {
		client = conf.ClientID
		if client == "" {
			return nil, re.ErrorCodeEnvNotSet.WithDetail("AZURE_CLIENT_ID environment variable is empty").WithComponentType(re.AuthProvider)
		}
	}
	if err != nil {
		return nil, err
	}
	// retrieve an AAD Access token
	token, err := getManagedIdentityToken(context.Background(), client)
	if err != nil {
		return nil, re.ErrorCodeAuthDenied.NewError(re.AuthProvider, "", re.AzureManagedIdentityLink, err, "", re.HideStackTrace)
	}

	return &azureManagedIdentityAuthProvider{
		identityToken: token,
		clientID:      client,
		tenantID:      tenant,
	}, nil
}

// Enabled checks for non empty tenant ID and AAD access token
func (d *azureManagedIdentityAuthProvider) Enabled(_ context.Context) bool {
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
			return provider.AuthConfig{}, re.ErrorCodeAuthDenied.NewError(re.AuthProvider, "", re.AzureManagedIdentityLink, err, "could not refresh azure managed identity token", re.HideStackTrace)
		}
		d.identityToken = newToken
		logger.GetLogger(ctx, logOpt).Info("successfully refreshed azure managed identity token")
	}
	// add protocol to generate complete URI
	serverURL := "https://" + artifactHostName

	// create registry client and exchange AAD token for registry refresh token
	refreshTokenClient := containerregistry.NewRefreshTokensClient(serverURL)
	rt, err := refreshTokenClient.GetFromExchange(ctx, "access_token", artifactHostName, d.tenantID, "", d.identityToken.Token)
	if err != nil {
		return provider.AuthConfig{}, re.ErrorCodeAuthDenied.NewError(re.AuthProvider, "", re.AzureManagedIdentityLink, err, "failed to get refresh token for container registry by azure managed identity token", re.HideStackTrace)
	}

	expiresOn := getACRExpiryIfEarlier(d.identityToken.ExpiresOn)

	authConfig := provider.AuthConfig{
		Username:  dockerTokenLoginUsernameGUID,
		Password:  *rt.RefreshToken,
		Provider:  d,
		ExpiresOn: expiresOn,
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
	return azcore.AccessToken{}, re.ErrorCodeConfigInvalid.WithComponentType(re.AuthProvider).WithDetail("config is nil pointer for GetServicePrincipalToken")
}
