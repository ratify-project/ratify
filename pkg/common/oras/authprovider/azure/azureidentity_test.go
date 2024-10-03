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
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/containers/azcontainerregistry"
	ratifyerrors "github.com/ratify-project/ratify/errors"
	"github.com/ratify-project/ratify/pkg/common/oras/authprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGetManagedIdentityToken struct {
	mock.Mock
}

func (m *MockGetManagedIdentityToken) GetManagedIdentityToken(ctx context.Context, clientID string) (azcore.AccessToken, error) {
	args := m.Called(ctx, clientID)
	return args.Get(0).(azcore.AccessToken), args.Error(1)
}

// Verifies that Enabled checks if tenantID is empty or AAD token is empty
func TestAzureMSIEnabled_ExpectedResults(t *testing.T) {
	azAuthProvider := MIAuthProvider{
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

func TestMIProvide_Success(t *testing.T) {
	const registryHost = "myregistry.azurecr.io"
	mockClient := new(MockAuthClient)
	expectedRefreshToken := "mocked_refresh_token"
	mockClient.On("ExchangeAADAccessTokenForACRRefreshToken", mock.Anything, "access_token", registryHost, mock.Anything).
		Return(azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenResponse{
			ACRRefreshToken: azcontainerregistry.ACRRefreshToken{RefreshToken: &expectedRefreshToken},
		}, nil)

	provider := &MIAuthProvider{
		identityToken: azcore.AccessToken{
			Token:     "mockToken",
			ExpiresOn: time.Now().Add(time.Hour),
		},
		tenantID: "mockTenantID",
		clientID: "mockClientID",
		authClientFactory: func(_ string, _ *azcontainerregistry.AuthenticationClientOptions) (AuthClient, error) {
			return mockClient, nil
		},
		getRegistryHost: func(_ string) (string, error) {
			return registryHost, nil
		},
		getManagedIdentityToken: func(_ context.Context, _ string) (azcore.AccessToken, error) {
			return azcore.AccessToken{
				Token:     "mockToken",
				ExpiresOn: time.Now().Add(time.Hour),
			}, nil
		},
	}

	authConfig, err := provider.Provide(context.Background(), "artifact")

	assert.NoError(t, err)
	// Assert that getManagedIdentityToken was not called
	mockClient.AssertNotCalled(t, "getManagedIdentityToken", mock.Anything, mock.Anything)
	// Assert that the returned refresh token matches the expected one
	assert.Equal(t, expectedRefreshToken, authConfig.Password)
}

func TestMIProvide_RefreshAAD(t *testing.T) {
	const registryHost = "myregistry.azurecr.io"
	// Arrange
	mockClient := new(MockAuthClient)

	// Create a mock function for getManagedIdentityToken
	mockGetManagedIdentityToken := new(MockGetManagedIdentityToken)

	provider := &MIAuthProvider{
		identityToken: azcore.AccessToken{
			Token:     "mockToken",
			ExpiresOn: time.Now(), // Expired token to force a refresh
		},
		tenantID: "mockTenantID",
		clientID: "mockClientID",
		authClientFactory: func(_ string, _ *azcontainerregistry.AuthenticationClientOptions) (AuthClient, error) {
			return mockClient, nil
		},
		getRegistryHost: func(_ string) (string, error) {
			return registryHost, nil
		},
		getManagedIdentityToken: mockGetManagedIdentityToken.GetManagedIdentityToken, // Use the mock
	}

	mockClient.On("ExchangeAADAccessTokenForACRRefreshToken", mock.Anything, "access_token", registryHost, mock.Anything).
		Return(azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenResponse{
			ACRRefreshToken: azcontainerregistry.ACRRefreshToken{RefreshToken: new(string)},
		}, nil)

	// Set up the expectation for the mocked method
	mockGetManagedIdentityToken.On("GetManagedIdentityToken", mock.Anything, "mockClientID").
		Return(azcore.AccessToken{
			Token:     "newMockToken",
			ExpiresOn: time.Now().Add(time.Hour),
		}, nil)

	ctx := context.TODO()
	artifact := "testArtifact"

	// Act
	_, err := provider.Provide(ctx, artifact)

	// Assert
	assert.NoError(t, err)
	mockGetManagedIdentityToken.AssertCalled(t, "GetManagedIdentityToken", mock.Anything, "mockClientID") // Assert that getManagedIdentityToken was called
}

func TestMIProvide_Failure_InvalidHostName(t *testing.T) {
	provider := &MIAuthProvider{
		tenantID: "test_tenant",
		clientID: "test_client",
		identityToken: azcore.AccessToken{
			Token: "test_token",
		},
		getRegistryHost: func(_ string) (string, error) {
			return "", errors.New("invalid hostname")
		},
	}

	_, err := provider.Provide(context.Background(), "artifact")
	assert.Error(t, err)
}
