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
	ratifyerrors "github.com/ratify-project/ratify/errors"
	"github.com/ratify-project/ratify/pkg/common/oras/authprovider"
)

// Verifies that Enabled checks if tenantID is empty or AAD token is empty
func TestAzureMSIEnabled_ExpectedResults(t *testing.T) {
	azAuthProvider := azureManagedIdentityAuthProvider{
		tenantID: "test_tenant",
		clientID: "test_client",
		identityToken: azcore.AccessToken{
			Token: "test_token",
		},
	}

	ctx := context.Background()

	if !azAuthProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned true but returned false")
	}

	azAuthProvider.tenantID = ""
	if azAuthProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned false but returned true for empty tenantID")
	}

	azAuthProvider.identityToken.Token = ""
	if azAuthProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned false but returned true for empty AAD access token")
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

	_, err = authprovider.CreateAuthProviderFromConfig(authProviderConfig)

	expectedErr := ratifyerrors.ErrorCodeAuthDenied.WithDetail("AZURE_TENANT_ID environment variable is empty")
	if err == nil || !errors.Is(err, expectedErr) {
		t.Fatalf("create auth provider should have failed: expected err %s, but got err %s", expectedErr, err)
	}

	err = os.Setenv("AZURE_TENANT_ID", "tenant id")
	if err != nil {
		t.Fatal("failed to set env variable AZURE_TENANT_ID")
	}

	_, err = authprovider.CreateAuthProviderFromConfig(authProviderConfig)

	expectedErr = ratifyerrors.ErrorCodeAuthDenied.WithDetail("AZURE_CLIENT_ID environment variable is empty")
	if err == nil || !errors.Is(err, expectedErr) {
		t.Fatalf("create auth provider should have failed: expected err %s, but got err %s", expectedErr, err)
	}
}
