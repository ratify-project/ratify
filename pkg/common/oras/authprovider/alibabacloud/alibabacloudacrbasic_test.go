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

package alibabacloud

import (
	"context"
	"os"
	"testing"

	cr20181201 "github.com/alibabacloud-go/cr-20181201/v2/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	// #nosec G101 (Ref: https://github.com/securego/gosec/issues/295)
	testPassword              = "eyJwYXlsb2FkIjoiOThPNTFqemhaUmZWVG"
	testUserMeta              = `{"instanceId":"cri-xxxxxxx","time":"1730185474000","type":"user","userId":"123456"}`
	testHost                  = "test-registry.cn-hangzhou.cr.aliyuncs.com"
	testInvalidHost           = "http://test-registry.cn-hangzhou.cr.aliyuncs.com"
	testArtifact              = testHost + "/foo:latest"
	testArtifactWithoutRegion = "test-registry-vpc.cr.aliyuncs.com/foo/test:latest"
	testInstanceID            = "cri-testing"
)

// Mock types for external dependencies
type MockAlibabaCloudAcrAuthTokenGetter struct {
	mock.Mock
}

// Mock AlibabaCloudAcrAuthTokenGetter.GetAcrAuthToken
func (m *MockAlibabaCloudAcrAuthTokenGetter) GetAcrAuthToken(artifact string, acrInstanceConfig map[string]string) (*cr20181201.GetAuthorizationTokenResponseBody, error) {
	args := m.Called(artifact, acrInstanceConfig)
	return args.Get(0).(*cr20181201.GetAuthorizationTokenResponseBody), args.Error(1)
}

// Verifies that alibabacloudAcrBasicAuthProvider is properly constructed by mocking AcrAuthToken
func mockAuthProvider() alibabacloudAcrBasicAuthProvider {
	// Setup
	return alibabacloudAcrBasicAuthProvider{
		providerName: alibabacloudAcrAuthProviderName,
	}
}

func TestAlibabaCloudAcrBasicAuthProvider_Create(t *testing.T) {
	authProviderConfig := map[string]interface{}{
		"name":              "alibabacloudAcrBasic",
		"defaultInstanceId": testInstanceID,
	}

	factory := AlibabaCloudAcrBasicProviderFactory{}
	_, err := factory.Create(authProviderConfig)

	if err != nil {
		t.Fatalf("create failed %v", err)
	}
}

func TestAlibabaCloudAcrBasicAuthProvider_CreateWithoutInstanceId(t *testing.T) {
	authProviderConfig := map[string]interface{}{
		"name": "alibabacloudAcrBasic",
	}
	factory := AlibabaCloudAcrBasicProviderFactory{}
	_, err := factory.Create(authProviderConfig)
	assert.Error(t, err)
}

func TestAlibabaCloudAcrBasicAuthProvider_CreateWithInvalidProviderConfig(t *testing.T) {
	authProviderConfig := map[string]interface{}{
		"name":    000,
		"invalid": "test",
	}
	factory := AlibabaCloudAcrBasicProviderFactory{}
	_, err := factory.Create(authProviderConfig)
	assert.Error(t, err)
}

func TestAlibabaCloudAcrBasicAuthProvider_CreateWithMultiProviderConfig(t *testing.T) {
	ic1 := acrInstancesConfig{
		InstanceID:   "cri-testing1",
		InstanceName: "test-instance-1",
	}
	ic2 := acrInstancesConfig{
		InstanceID:   "cri-testing2",
		InstanceName: "test-instance-2",
	}
	authProviderConfig := map[string]interface{}{
		"name":               "alibabacloudAcrBasic",
		"acrInstancesConfig": []acrInstancesConfig{ic1, ic2},
	}
	factory := AlibabaCloudAcrBasicProviderFactory{}
	_, err := factory.Create(authProviderConfig)
	assert.NoError(t, err)
}

func TestAlibabaCloudAcrBasicAuthProvider_Enabled(t *testing.T) {
	authProvider := mockAuthProvider()

	ctx := context.Background()

	if !authProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned true but returned false")
	}

	authProvider.providerName = ""
	if authProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned false but returned true")
	}
}

func TestAlibabaCloudAcrBasicAuthProvider_ProvidesNotEnabled(t *testing.T) {
	authProvider := alibabacloudAcrBasicAuthProvider{
		providerName: "",
	}
	_, err := authProvider.Provide(context.TODO(), testInvalidHost)
	assert.Error(t, err)
}

func TestAlibabaCloudAcrBasicAuthProvider_GetAuthTokenWithoutRRSAEnv(t *testing.T) {
	acrInstancesConfig := make(map[string]string, 0)
	_, err := getAcrAuthToken(testArtifactWithoutRegion, acrInstancesConfig)
	assert.Error(t, err)

	expectedMessage := "required environment variables not set, ALIBABA_CLOUD_ROLE_ARN: , ALIBABA_CLOUD_OIDC_PROVIDER_ARN: , ALIBABA_CLOUD_OIDC_TOKEN_FILE: "
	if err.Error() != expectedMessage {
		t.Fatalf("expected message: %s, instead got error: %s", expectedMessage, err.Error())
	}
}

func TestAlibabaCloudAcrBasicAuthProvider_GetAuthTokenWithInvalidHost(t *testing.T) {
	os.Setenv("ALIBABA_CLOUD_OIDC_PROVIDER_ARN", "placeholder")
	os.Setenv("ALIBABA_CLOUD_ROLE_ARN", "placeholder")
	os.Setenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE", "placeholder")
	acrInstancesConfig := map[string]string{
		"test": testInstanceID,
	}
	_, err := getAcrAuthToken(testInvalidHost, acrInstancesConfig)
	assert.Error(t, err)
}

func TestAlibabaCloudAcrBasicAuthProvider_GetAuthTokenWithoutRegion(t *testing.T) {
	os.Setenv("ALIBABA_CLOUD_OIDC_PROVIDER_ARN", "placeholder")
	os.Setenv("ALIBABA_CLOUD_ROLE_ARN", "placeholder")
	os.Setenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE", "placeholder")
	acrInstancesConfig := make(map[string]string, 0)
	_, err := getAcrAuthToken(testArtifactWithoutRegion, acrInstancesConfig)
	assert.Error(t, err)

	expectedMessage := "ALIBABACLOUD_IMAGE_INVALID: Invalid Alibaba Cloud Registry image format"
	if err.Error() != expectedMessage {
		t.Fatalf("expected message: %s, instead got error: %s", expectedMessage, err.Error())
	}
}

func TestAlibabaCloudAcrBasicAuthProvider_GetAuthTokenWithInvalidInstanceId(t *testing.T) {
	os.Setenv("ALIBABA_CLOUD_OIDC_PROVIDER_ARN", "placeholder")
	os.Setenv("ALIBABA_CLOUD_ROLE_ARN", "placeholder")
	os.Setenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE", "placeholder")
	acrInstancesConfig := map[string]string{
		"testInvalid": testInstanceID,
	}
	_, err := getAcrAuthToken(testHost, acrInstancesConfig)
	assert.Error(t, err)
}

func TestAlibabaCloudAcrBasicAuthProvider_GetAuthTokenWithInvalidAcrClient(t *testing.T) {
	os.Setenv("ALIBABA_CLOUD_OIDC_PROVIDER_ARN", "placeholder")
	os.Setenv("ALIBABA_CLOUD_ROLE_ARN", "placeholder")
	os.Setenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE", "placeholder")
	acrInstancesConfig := map[string]string{
		"test": testInstanceID,
	}
	_, err := getAcrAuthToken(testHost, acrInstancesConfig)
	assert.Error(t, err)
}

func TestAlibabaCloudAcrBasicAuthProvider_Provide_TokenRefreshSuccess(t *testing.T) {
	mockAlibabaCloudAcrAuthTokenGetter := new(MockAlibabaCloudAcrAuthTokenGetter)
	mockACRToken := cr20181201.GetAuthorizationTokenResponseBody{}
	mockAlibabaCloudAcrAuthTokenGetter.On("GetAcrAuthToken", "mockartifact", mock.Anything).Return(&mockACRToken, nil)

	authProvider := alibabacloudAcrBasicAuthProvider{
		providerName:    alibabacloudAcrAuthProviderName,
		getAcrAuthToken: mockAlibabaCloudAcrAuthTokenGetter,
	}

	ctx := context.Background()
	_, err := authProvider.Provide(ctx, "mockartifact")

	// Validate success and token refresh
	assert.NoError(t, err)
}
