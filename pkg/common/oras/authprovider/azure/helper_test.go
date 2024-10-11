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

	"github.com/Azure/azure-sdk-for-go/sdk/containers/azcontainerregistry"
	"github.com/stretchr/testify/mock"
)

// MockAuthClient is a mock implementation of AuthClient.
type MockAuthClient struct {
	mock.Mock
}

// Mock method for ExchangeAADAccessTokenForACRRefreshToken
func (m *MockAuthClient) ExchangeAADAccessTokenForACRRefreshToken(ctx context.Context, grantType, service string, options *azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenOptions) (azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenResponse, error) {
	args := m.Called(ctx, grantType, service, options)
	return args.Get(0).(azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenResponse), args.Error(1)
}

// MockAuthClientFactory is a mock implementation of AuthClientFactory.
type MockAuthClientFactory struct {
	mock.Mock
}

// Mock method for CreateAuthClient
func (m *MockAuthClientFactory) CreateAuthClient(serverURL string, options *azcontainerregistry.AuthenticationClientOptions) (AuthClient, error) {
	args := m.Called(serverURL, options)
	return args.Get(0).(AuthClient), args.Error(1)
}

// MockRegistryHostGetter is a mock implementation of RegistryHostGetter.
type MockRegistryHostGetter struct {
	mock.Mock
}

// Mock method for GetRegistryHost
func (m *MockRegistryHostGetter) GetRegistryHost(artifact string) (string, error) {
	args := m.Called(artifact)
	return args.String(0), args.Error(1)
}

// // TestDefaultAuthClientFactoryImpl tests the default factory implementation.
// func TestDefaultAuthClientFactoryImpl(t *testing.T) {
// 	mockFactory := new(MockAuthClientFactory)
// 	mockAuthClient := new(MockAuthClient)

// 	serverURL := "https://example.azurecr.io"
// 	options := &azcontainerregistry.AuthenticationClientOptions{}

// 	// Set up expectations
// 	mockFactory.On("CreateAuthClient", serverURL, options).Return(mockAuthClient, nil)

// 	factory := &DefaultAuthClientFactoryImpl{}
// 	client, err := factory.CreateAuthClient(serverURL, options)

// 	// Verify expectations
// 	mockFactory.AssertCalled(t, "CreateAuthClient", serverURL, options)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, client)
// }

// // TestDefaultAuthClientFactory_Error tests error handling during client creation.
// func TestDefaultAuthClientFactory_Error(t *testing.T) {
// 	mockFactory := new(MockAuthClientFactory)

// 	serverURL := "https://example.azurecr.io"
// 	options := &azcontainerregistry.AuthenticationClientOptions{}
// 	expectedError := errors.New("failed to create client")

// 	// Set up expectations
// 	mockFactory.On("CreateAuthClient", serverURL, options).Return(nil, expectedError)

// 	factory := &DefaultAuthClientFactoryImpl{}
// 	client, err := factory.CreateAuthClient(serverURL, options)

// 	// Verify expectations
// 	mockFactory.AssertCalled(t, "CreateAuthClient", serverURL, options)
// 	assert.Error(t, err)
// 	assert.Nil(t, client)
// 	assert.Equal(t, expectedError, err)
// }

// // TestGetRegistryHost tests the GetRegistryHost function.
// func TestGetRegistryHost(t *testing.T) {
// 	mockGetter := new(MockRegistryHostGetter)

// 	artifact := "test/artifact"
// 	expectedHost := "example.azurecr.io"

// 	// Set up expectations
// 	mockGetter.On("GetRegistryHost", artifact).Return(expectedHost, nil)

// 	getter := &DefaultRegistryHostGetterImpl{}
// 	host, err := getter.GetRegistryHost(artifact)

// 	// Verify expectations
// 	mockGetter.AssertCalled(t, "GetRegistryHost", artifact)
// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedHost, host)
// }

// // TestGetRegistryHost_Error tests error handling in GetRegistryHost.
// func TestGetRegistryHost_Error(t *testing.T) {
// 	mockGetter := new(MockRegistryHostGetter)

// 	artifact := "test/artifact"
// 	expectedError := errors.New("failed to get registry host")

// 	// Set up expectations
// 	mockGetter.On("GetRegistryHost", artifact).Return("", expectedError)

// 	getter := &DefaultRegistryHostGetterImpl{}
// 	host, err := getter.GetRegistryHost(artifact)

// 	// Verify expectations
// 	mockGetter.AssertCalled(t, "GetRegistryHost", artifact)
// 	assert.Error(t, err)
// 	assert.Empty(t, host)
// 	assert.Equal(t, expectedError, err)
// }
