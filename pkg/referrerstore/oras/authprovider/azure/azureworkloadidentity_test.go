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
	"os"
	"testing"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/deislabs/ratify/pkg/referrerstore/oras/authprovider"
)

// Verifies that Enabled checks if tenantID is empty or AAD token is empty
func TestAzureWIEnabled_ExpectedResults(t *testing.T) {
	azAuthProvider := azureWIAuthProvider{
		tenantID: "test_tenant",
		aadToken: confidential.AuthResult{
			AccessToken: "test_token",
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

	azAuthProvider.aadToken.AccessToken = ""
	if azAuthProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned false but returned true for empty AAD access token")
	}
}

// Verifies that tenant id, client id, token file path, and authority host
// environment variables are properly set
func TestAzureWIValidation_EnvironmentVariables_ExpectedResults(t *testing.T) {
	authProviderConfig := map[string]interface{}{
		"name": "azureWorkloadIdentity",
	}

	err := os.Setenv("AZURE_TENANT_ID", "")
	if err != nil {
		t.Fatal("failed to set env variable AZURE_TENANT_ID")
	}

	_, err = authprovider.CreateAuthProviderFromConfig(authProviderConfig)

	expectedErr := fmt.Errorf("azure tenant id environment variable is empty")
	if err == nil || err.Error() != expectedErr.Error() {
		t.Fatalf("create auth provider should have failed: expected err %s, but got err %s", expectedErr, err)
	}

	err = os.Setenv("AZURE_TENANT_ID", "tenant id")
	if err != nil {
		t.Fatal("failed to set env variable AZURE_TENANT_ID")
	}

	_, err = authprovider.CreateAuthProviderFromConfig(authProviderConfig)

	expectedErr = fmt.Errorf("required environment variables not set, AZURE_CLIENT_ID: , AZURE_FEDERATED_TOKEN_FILE: , AZURE_AUTHORITY_HOST: ")
	if err == nil || err.Error() != expectedErr.Error() {
		t.Fatalf("create auth provider should have failed: expected err %s, but got err %s", expectedErr, err)
	}
}
