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
	"errors"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	ratifyerrors "github.com/deislabs/ratify/errors"
	provider "github.com/deislabs/ratify/pkg/common/oras/authprovider"
)

// func TestAzureMSICreate_ExpectedResults(t *testing.T) {
// 	var testProviderFactory azureManagedIdentityProviderFactory
// 	tests := []struct {
// 		name       string
// 		configMap  map[string]interface{}
// 		isNegative bool
// 		expect     error
// 	}{
// 		{
// 			name: "input type for unmarshal is unsupported",
// 			configMap: map[string]interface{}{
// 				"Name":     "test_name",
// 				"ClientID": "test_clientID",
// 			},
// 			isNegative: true,
// 			expect:     re.ErrorCodeConfigInvalid,
// 		},
// 	}
// 	for _, testCase := range tests {
// 		_, err := testProviderFactory.Create(provider.AuthProviderConfig(testCase.configMap))
// 		if testCase.isNegative != (err != nil) {
// 			t.Errorf("Expected %v in case %v, but got %v", testCase.expect, testCase.name, err)
// 		}
// 	}
// }

// Verifiers that Enable checks if tenantID is empty or AAD token is empty
func TestAzureMSIEnabled_ExpectedResults(t *testing.T) {
	tests := []struct {
		name           string
		azAuthProvider azureManagedIdentityAuthProvider
		expect         bool
	}{
		{
			name: "complete config",
			azAuthProvider: azureManagedIdentityAuthProvider{
				tenantID: "test_tenant",
				clientID: "test_client",
				identityToken: azcore.AccessToken{
					Token: "test_token",
				},
			},
			expect: true,
		},
		{
			name: "config miss tenantID",
			azAuthProvider: azureManagedIdentityAuthProvider{
				tenantID: "",
				clientID: "test_client",
				identityToken: azcore.AccessToken{
					Token: "test_token",
				},
			},
			expect: false,
		},
		{
			name: "config miss clientID",
			azAuthProvider: azureManagedIdentityAuthProvider{
				tenantID: "test_tenant",
				clientID: "",
				identityToken: azcore.AccessToken{
					Token: "test_token",
				},
			},
			expect: false,
		},
		{
			name: "config miss Token",
			azAuthProvider: azureManagedIdentityAuthProvider{
				tenantID: "test_tenant",
				clientID: "test_client",
				identityToken: azcore.AccessToken{
					Token: "",
				},
			},
			expect: false,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.azAuthProvider.Enabled(ctx)
			if got != tt.expect {
				t.Fatalf("Expect: %v, got %v for %v", tt.expect, got, tt.name)
			}
		})
	}

}

// Verifies that tenant id, client id, token file path, and authority host
// environment variables are properly set
func TestAzureMSIValidation_EnvironmentVariables_ExpectedResults(t *testing.T) {
	authProviderConfig := map[string]interface{}{
		"name": "azureManagedIdentity",
	}

	err := os.Setenv("AZURE_TENANT_ID", "")
	if err != nil {
		t.Fatal("failed to set env variable AZURE_TENANT_ID")
	}

	err = os.Setenv("AZURE_CLIENT_ID", "")
	if err != nil {
		t.Fatal("failed to set env variable AZURE_CLIENT_ID")
	}

	_, err = provider.CreateAuthProviderFromConfig(authProviderConfig)

	expectedErr := ratifyerrors.ErrorCodeAuthDenied.WithDetail("AZURE_TENANT_ID environment variable is empty")
	if err == nil || !errors.Is(err, expectedErr) {
		t.Fatalf("create auth provider should have failed: expected err %s, but got err %s", expectedErr, err)
	}

	err = os.Setenv("AZURE_TENANT_ID", "tenant id")
	if err != nil {
		t.Fatal("failed to set env variable AZURE_TENANT_ID")
	}

	_, err = provider.CreateAuthProviderFromConfig(authProviderConfig)

	expectedErr = ratifyerrors.ErrorCodeAuthDenied.WithDetail("AZURE_CLIENT_ID environment variable is empty")
	if err == nil || !errors.Is(err, expectedErr) {
		t.Fatalf("create auth provider should have failed: expected err %s, but got err %s", expectedErr, err)
	}
}
