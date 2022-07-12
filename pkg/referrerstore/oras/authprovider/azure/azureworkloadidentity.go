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

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	provider "github.com/deislabs/ratify/pkg/referrerstore/oras/authprovider"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/preview/containerregistry/runtime/2019-08-15-preview/containerregistry"
)

type AzureWIProviderFactory struct{}
type azureWIAuthProvider struct {
	aadToken confidential.AuthResult
	tenantID string
}

type azureWIAuthProviderConf struct {
	Name string `json:"name"`
}

const (
	azureWIAuthProviderName      string = "azureWorkloadIdentity"
	dockerTokenLoginUsernameGUID string = "00000000-0000-0000-0000-000000000000"
	AADResource                  string = "https://containerregistry.azure.net/.default"
)

// init calls Register for our Azure Workload Identity provider
func init() {
	provider.Register(azureWIAuthProviderName, &AzureWIProviderFactory{})
}

// Create returns an AzureWIAuthProvider
func (s *AzureWIProviderFactory) Create(authProviderConfig provider.AuthProviderConfig) (provider.AuthProvider, error) {
	conf := azureWIAuthProviderConf{}
	authProviderConfigBytes, err := json.Marshal(authProviderConfig)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(authProviderConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse auth provider configuration: %v", err)
	}

	tenant := os.Getenv("AZURE_TENANT_ID")
	if tenant == "" {
		return nil, fmt.Errorf("azure tenant id environment variable is empty")
	}
	// retrieve an AAD Access token
	token, err := getAADAccessToken(context.Background(), tenant)
	if err != nil {
		return nil, err
	}

	return &azureWIAuthProvider{
		aadToken: token,
		tenantID: tenant,
	}, nil
}

// Enabled checks for non empty tenant ID and AAD access token
func (d *azureWIAuthProvider) Enabled(ctx context.Context) bool {
	if d.tenantID == "" {
		return false
	}

	if d.aadToken.AccessToken == "" {
		return false
	}

	return true
}

// Provide returns the credentials for a specified artifact.
// Uses Azure Workload Identity to retrieve an AAD access token which can be
// exchanged for a valid ACR refresh token for login.
func (d *azureWIAuthProvider) Provide(ctx context.Context, artifact string) (provider.AuthConfig, error) {
	if !d.Enabled(ctx) {
		return provider.AuthConfig{}, fmt.Errorf("azure workload identity auth provider is not properly enabled")
	}
	// parse the artifact reference string to extract the registry host name
	artifactHostName, err := provider.GetRegistryHostName(artifact)
	if err != nil {
		return provider.AuthConfig{}, err
	}

	// need to refresh AAD token if it's expired
	if time.Now().After(d.aadToken.ExpiresOn) {
		newToken, err := getAADAccessToken(ctx, d.tenantID)
		if err != nil {
			return provider.AuthConfig{}, errors.Wrap(err, "could not refresh AAD token")
		}
		d.aadToken = newToken
		logrus.Info("sucessfully refreshed AAD token")
	}

	// add protocol to generate complete URI
	serverUrl := "https://" + artifactHostName

	// create registry client and exchange AAD token for registry refresh token
	refreshTokenClient := containerregistry.NewRefreshTokensClient(serverUrl)
	rt, err := refreshTokenClient.GetFromExchange(context.Background(), "access_token", artifactHostName, d.tenantID, "", d.aadToken.AccessToken)
	if err != nil {
		return provider.AuthConfig{}, fmt.Errorf("failed to get refresh token for container registry - %w", err)
	}

	authConfig := provider.AuthConfig{
		Username:  dockerTokenLoginUsernameGUID,
		Password:  *rt.RefreshToken,
		Provider:  d,
		ExpiresOn: d.aadToken.ExpiresOn,
	}

	return authConfig, nil
}

// Source: https://github.com/Azure/azure-workload-identity/blob/d126293e3c7c669378b225ad1b1f29cf6af4e56d/examples/msal-go/token_credential.go#L25
func getAADAccessToken(ctx context.Context, tenantID string) (confidential.AuthResult, error) {
	// Azure AD Workload Identity webhook will inject the following env vars:
	// 	AZURE_CLIENT_ID with the clientID set in the service account annotation
	// 	AZURE_TENANT_ID with the tenantID set in the service account annotation. If not defined, then
	// 	the tenantID provided via azure-wi-webhook-config for the webhook will be used.
	// 	AZURE_FEDERATED_TOKEN_FILE is the service account token path
	// 	AZURE_AUTHORITY_HOST is the AAD authority hostname
	clientID := os.Getenv("AZURE_CLIENT_ID")
	tokenFilePath := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
	authority := os.Getenv("AZURE_AUTHORITY_HOST")
	if clientID == "" || tokenFilePath == "" || authority == "" {
		return confidential.AuthResult{}, fmt.Errorf("required environment variables not set, AZURE_CLIENT_ID: %s, AZURE_FEDERATED_TOKEN_FILE: %s, AZURE_AUTHORITY_HOST: %s", clientID, tokenFilePath, authority)
	}

	// read the service account token from the filesystem
	signedAssertion, err := readJWTFromFS(tokenFilePath)
	if err != nil {
		return confidential.AuthResult{}, errors.Wrap(err, "failed to read service account token")
	}
	cred, err := confidential.NewCredFromAssertion(signedAssertion)
	if err != nil {
		return confidential.AuthResult{}, errors.Wrap(err, "failed to create confidential creds")
	}

	// create the confidential client to request an AAD token
	confidentialClientApp, err := confidential.New(
		clientID,
		cred,
		confidential.WithAuthority(fmt.Sprintf("%s%s/oauth2/token", authority, tenantID)))
	if err != nil {
		return confidential.AuthResult{}, errors.Wrap(err, "failed to create confidential client app")
	}

	result, err := confidentialClientApp.AcquireTokenByCredential(ctx, []string{AADResource})
	if err != nil {
		return confidential.AuthResult{}, errors.Wrap(err, "failed to acquire AAD token")
	}

	return result, nil
}

// readJWTFromFS reads the jwt from file system
// Source: https://github.com/Azure/azure-workload-identity/blob/d126293e3c7c669378b225ad1b1f29cf6af4e56d/examples/msal-go/token_credential.go#L88
func readJWTFromFS(tokenFilePath string) (string, error) {
	token, err := os.ReadFile(tokenFilePath)
	if err != nil {
		return "", err
	}
	return string(token), nil
}
