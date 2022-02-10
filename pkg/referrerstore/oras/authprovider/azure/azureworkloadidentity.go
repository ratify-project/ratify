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
	"strings"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	provider "github.com/deislabs/ratify/pkg/referrerstore/oras/authprovider"
	"github.com/pkg/errors"
)

type AzureWIProviderFactory struct{}
type azureWIAuthProvider struct{}

type azureWIAuthProviderConf struct {
	Name string `json:"name"`
}

const (
	azureWIAuthProviderName      string = "azure-wi"
	dockerTokenLoginUsernameGUID string = "00000000-0000-0000-0000-000000000000"
	AADResource                  string = "https://containerregistry.azure.net"
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

	return &azureWIAuthProvider{}, nil
}

// Enabled always returns true since there are no fields to verify
func (d *azureWIAuthProvider) Enabled() bool {
	return true
}

// Provide returns the credentials for a specificed artifact.
// Uses Azure Workload Identity to retrieve an AAD access token which can be
// exchanged for a valid ACR refresh token for login.
func (d *azureWIAuthProvider) Provide(artifact string) (provider.AuthConfig, error) {
	tenantID := os.Getenv("AZURE_TENANT_ID")

	// parse the artifact reference string to extract the registry host name
	artifactHostName, err := provider.GetRegistryHostName(artifact)
	if err != nil {
		return provider.AuthConfig{}, err
	}

	// retrieve an AAD Access token
	aadToken, err := getAADAccessToken(tenantID, AADResource)
	if err != nil {
		return provider.AuthConfig{}, err
	}

	// send a challenge to the login server
	directive, err := receiveChallengeFromLoginServer(artifactHostName, "https")
	if err != nil {
		return provider.AuthConfig{}, err
	}

	// use challenge directive and AAD token to exchange for a registry token
	refreshToken, err := performTokenExchange(artifactHostName, directive, tenantID, aadToken)
	if err != nil {
		return provider.AuthConfig{}, err
	}

	authConfig := provider.AuthConfig{
		Username: dockerTokenLoginUsernameGUID,
		Password: refreshToken,
		Provider: d,
	}
	return authConfig, nil
}

// Source: https://github.com/Azure/azure-workload-identity/blob/d126293e3c7c669378b225ad1b1f29cf6af4e56d/examples/msal-go/token_credential.go#L25
func getAADAccessToken(tenantID, resource string) (string, error) {
	// Azure AD Workload Identity webhook will inject the following env vars:
	// 	AZURE_CLIENT_ID with the clientID set in the service account annotation
	// 	AZURE_TENANT_ID with the tenantID set in the service account annotation. If not defined, then
	// 	the tenantID provided via azure-wi-webhook-config for the webhook will be used.
	// 	AZURE_FEDERATED_TOKEN_FILE is the service account token path
	// 	AZURE_AUTHORITY_HOST is the AAD authority hostname
	clientID := os.Getenv("AZURE_CLIENT_ID")
	tokenFilePath := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
	authorityHost := os.Getenv("AZURE_AUTHORITY_HOST")

	// read the service account token from the filesystem
	signedAssertion, err := readJWTFromFS(tokenFilePath)
	if err != nil {
		return "", errors.Wrap(err, "failed to read service account token")
	}
	cred, err := confidential.NewCredFromAssertion(signedAssertion)
	if err != nil {
		return "", errors.Wrap(err, "failed to create confidential creds")
	}

	// create the confidential client to request an AAD token
	confidentialClientApp, err := confidential.New(
		clientID,
		cred,
		confidential.WithAuthority(fmt.Sprintf("%s%s/oauth2/token", authorityHost, tenantID)))
	if err != nil {
		return "", errors.Wrap(err, "failed to create confidential client app")
	}

	// .default needs to be added to the scope
	if !strings.HasSuffix(resource, ".default") {
		resource += "/.default"
	}

	result, err := confidentialClientApp.AcquireTokenByCredential(context.Background(), []string{resource})
	if err != nil {
		return "", errors.Wrap(err, "failed to acquire token")
	}

	return result.AccessToken, nil
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
