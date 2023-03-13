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

package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"strings"
	"net/url"
)

// Source: https://github.com/Azure/azure-workload-identity/blob/d126293e3c7c669378b225ad1b1f29cf6af4e56d/examples/msal-go/token_credential.go#L25
func GetAADAccessToken(ctx context.Context, tenantID, clientID, scope string) (confidential.AuthResult, error) {
	// Azure AD Workload Identity webhook will inject the following env vars:
	// 	AZURE_FEDERATED_TOKEN_FILE is the service account token path
	// 	AZURE_AUTHORITY_HOST is the AAD authority hostname

	tokenFilePath := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
	authority := os.Getenv("AZURE_AUTHORITY_HOST")

	if tokenFilePath == "" || authority == "" {
		return confidential.AuthResult{}, fmt.Errorf("required environment variables not set, AZURE_FEDERATED_TOKEN_FILE: %s, AZURE_AUTHORITY_HOST: %s", tokenFilePath, authority)
	}

	cred := confidential.NewCredFromAssertionCallback(func(context.Context, confidential.AssertionRequestOptions) (string, error) {
		// read the service account token from the filesystem
		return readJWTFromFS(tokenFilePath)
	})

	// create the confidential client to request an AAD token
	confidentialClientApp, err := confidential.New(
		clientID,
		cred,
		confidential.WithAuthority(fmt.Sprintf("%s%s/oauth2/token", authority, tenantID)))
	if err != nil {
		return confidential.AuthResult{}, fmt.Errorf("failed to create confidential client app: %w", err)
	}

	result, err := confidentialClientApp.AcquireTokenByCredential(ctx, []string{scope})
	if err != nil {
		return confidential.AuthResult{}, fmt.Errorf("failed to acquire AAD token: %w", err)
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

// RegionFromImage parses region from image url
func RegionFromImage(image string) (string, error) {
	registry, err := RegistryFromImage(image)
	if err != nil {
		return "", err
	}

	region := RegionFromRegistry(registry)

	return region, nil
}

// RegionFromRegistry parses AWS region ID from registry url
func RegionFromRegistry(registry string) string {
	a := strings.Split(registry, ".")
	if len(a) >= 6 {
		return a[3]
	}
	return ""
}

// RegistryFromImage parses registry host from image url
func RegistryFromImage(image string) (string, error) {
	if strings.Contains(image, "https://") {
		u, err := url.Parse(image)
		if err != nil {
			return "", err
		}
		return u.Host, nil
	}

	return image[:strings.IndexByte(image, '/')], nil
}
