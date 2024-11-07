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

package azurekeyvault

// This class is based on implementation from  azure secret store csi provider
// Source: https://github.com/Azure/secrets-store-csi-driver-provider-azure/tree/release-1.4/pkg/provider
import (
	"context"
	"crypto"
	"errors"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azcertificates"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azkeys"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/azurekeyvault/types"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func SkipTestInitializeKVClient(t *testing.T) {
	testEnvs := []azure.Environment{
		azure.PublicCloud,
		azure.GermanCloud,
		azure.ChinaCloud,
		azure.USGovernmentCloud,
	}

	for i := range testEnvs {
		keyKVClient, secretKVClient, certificateKVClient, err := initializeKvClient(testEnvs[i].KeyVaultEndpoint, "", "", nil)
		assert.NoError(t, err)
		assert.NotNil(t, keyKVClient)
		assert.NotNil(t, secretKVClient)
		assert.NotNil(t, certificateKVClient)
	}
}

// TestCreate tests the Create function
func TestCreate(t *testing.T) {
	factory := &akvKMProviderFactory{}
	testCases := []struct {
		name      string
		config    config.KeyManagementProviderConfig
		expectErr bool
	}{
		{
			name: "valid config",
			config: config.KeyManagementProviderConfig{
				"inline":   "azurekeyvault",
				"vaultURI": "https://testkv.vault.azure.net/",
				"tenantID": "tid",
				"clientID": "clientid",
				"certificates": []map[string]interface{}{
					{
						"name": "cert1",
					},
				},
			},
			expectErr: false,
		},
		{
			name:      "keyvault uri not provided",
			config:    config.KeyManagementProviderConfig{},
			expectErr: true,
		},
		{
			name: "tenantID not provided",
			config: config.KeyManagementProviderConfig{
				"vaultUri": "https://testkv.vault.azure.net/",
			},
			expectErr: true,
		},
		{
			name: "clientID not provided",
			config: config.KeyManagementProviderConfig{
				"vaultUri": "https://testkv.vault.azure.net/",
				"tenantID": "tid",
			},
			expectErr: true,
		},
		{
			name: "invalid cloud name",
			config: config.KeyManagementProviderConfig{
				"vaultUri":  "https://testkv.vault.azure.net/",
				"tenantID":  "tid",
				"cloudName": "AzureCloud",
			},
			expectErr: true,
		},
		{
			name: "certificates & keys array not set",
			config: config.KeyManagementProviderConfig{
				"vaultUri":             "https://testkv.vault.azure.net/",
				"tenantID":             "tid",
				"useVMManagedIdentity": "true",
			},
			expectErr: true,
		},
		{
			name: "certificates empty",
			config: config.KeyManagementProviderConfig{
				"vaultUri":             "https://testkv.vault.azure.net/",
				"tenantID":             "tid",
				"useVMManagedIdentity": "true",
				"certificates":         []map[string]interface{}{},
			},
			expectErr: true,
		},
		{
			name: "invalid certificate name",
			config: config.KeyManagementProviderConfig{
				"vaultUri": "https://testkv.vault.azure.net/",
				"tenantID": "tid",
				"clientID": "clientid",
				"certificates": []map[string]interface{}{
					{
						"name":    "",
						"version": "version1",
					},
				},
			},
			expectErr: true,
		},
		{
			name: "invalid key name",
			config: config.KeyManagementProviderConfig{
				"vaultUri": "https://testkv.vault.azure.net/",
				"tenantID": "tid",
				"clientID": "clientid",
				"keys": []map[string]interface{}{
					{
						"name": "",
					},
				},
			},
			expectErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initKVClient = func(_, _, _ string, _ azcore.TokenCredential) (*azkeys.Client, *azsecrets.Client, *azcertificates.Client, error) {
				return &azkeys.Client{}, &azsecrets.Client{}, &azcertificates.Client{}, nil
			}
			_, err := factory.Create("v1", tc.config, "")
			if tc.expectErr != (err != nil) {
				t.Fatalf("error = %v, expectErr = %v", err, tc.expectErr)
			}
		})
	}
}

// TestGetCertificates tests the GetCertificates function
func TestGetCertificates(t *testing.T) {
	factory := &akvKMProviderFactory{}
	config := config.KeyManagementProviderConfig{
		"vaultUri": "https://testkv.vault.azure.net/",
		"tenantID": "tid",
		"clientID": "clientid",
		"certificates": []map[string]interface{}{
			{
				"name":    "cert1",
				"version": "",
			},
		},
	}

	provider, err := factory.Create("v1", config, "")
	if err != nil {
		t.Fatalf("expected no err but got error = %v", err)
	}

	certs, certStatus, err := provider.GetCertificates(context.Background())
	assert.NotNil(t, err)
	assert.Nil(t, certs)
	assert.Nil(t, certStatus)
}

// TestGetKeys tests the GetKeys function
func TestGetKeys(t *testing.T) {
	factory := &akvKMProviderFactory{}
	config := config.KeyManagementProviderConfig{
		"vaultUri": "https://testkv.vault.azure.net/",
		"tenantID": "tid",
		"clientID": "clientid",
		"keys": []map[string]interface{}{
			{
				"name": "key1",
			},
		},
	}

	initKVClient = func(_, _, _ string, _ azcore.TokenCredential) (*azkeys.Client, *azsecrets.Client, *azcertificates.Client, error) {
		return &azkeys.Client{}, &azsecrets.Client{}, &azcertificates.Client{}, nil
	}
	provider, err := factory.Create("v1", config, "")
	if err != nil {
		t.Fatalf("expected no err but got error = %v", err)
	}

	keys, keyStatus, err := provider.GetKeys(context.Background())
	assert.NotNil(t, err)
	assert.Nil(t, keys)
	assert.Nil(t, keyStatus)
}

func TestIsRefreshable(t *testing.T) {
	factory := &akvKMProviderFactory{}
	config := config.KeyManagementProviderConfig{
		"vaultUri": "https://testkv.vault.azure.net/",
		"tenantID": "tid",
		"clientID": "clientid",
		"certificates": []map[string]interface{}{
			{
				"name":    "cert1",
				"version": "",
			},
		},
	}

	provider, _ := factory.Create("v1", config, "")
	if provider.IsRefreshable() != true {
		t.Fatalf("expected true, got false")
	}
}

// TestGetStatusMap tests the getStatusMap function
func TestGetStatusMap(t *testing.T) {
	certsStatus := []map[string]string{}
	certsStatus = append(certsStatus, map[string]string{
		"CertName":    "Cert1",
		"CertVersion": "VersionABC",
	})
	certsStatus = append(certsStatus, map[string]string{
		"CertName":    "Cert2",
		"CertVersion": "VersionEDF",
	})

	actual := getStatusMap(certsStatus, types.CertificatesStatus)
	assert.NotNil(t, actual[types.CertificatesStatus])
}

// TestGetObjectVersion tests the getObjectVersion function
func TestGetObjectVersion(t *testing.T) {
	id := "https://kindkv.vault.azure.net/secrets/cert1/c55925c29c6743dcb9bb4bf091be03b0"
	expectedVersion := "c55925c29c6743dcb9bb4bf091be03b0"
	actual := getObjectVersion(id)
	assert.Equal(t, expectedVersion, actual)
}

// TestGetStatus tests the getStatusProperty function
func TestGetStatusProperty(t *testing.T) {
	timeNow := time.Now().String()
	certName := "certName"
	certVersion := "versionABC"
	isEnabled := true

	status := getStatusProperty(certName, certVersion, timeNow, isEnabled)
	assert.Equal(t, certName, status[types.StatusName])
	assert.Equal(t, timeNow, status[types.StatusLastRefreshed])
	assert.Equal(t, certVersion, status[types.StatusVersion])
}

// TestGetCertsFromSecretBundle tests the getCertsFromSecretBundle function
func TestGetCertsFromSecretBundle(t *testing.T) {
	cases := []struct {
		desc        string
		value       string
		contentType string
		id          azsecrets.ID
		expectedErr bool
	}{
		{
			desc:        "Pem Content Type",
			value:       "-----BEGIN CERTIFICATE-----\nMIIC8TCCAdmgAwIBAgIUaNrwbhs/I1ecqUYdzD2xuAVNdmowDQYJKoZIhvcNAQEL\nBQAwKjEPMA0GA1UECgwGUmF0aWZ5MRcwFQYDVQQDDA5SYXRpZnkgUm9vdCBDQTAe\nFw0yMzA2MjEwMTIyMzdaFw0yNDA2MjAwMTIyMzdaMBkxFzAVBgNVBAMMDnJhdGlm\neS5kZWZhdWx0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtskG1BUt\n4Fw2lbm53KbwZb1hnLmWdwRotZyznhhk/yrUDcq3uF6klwpk/E2IKfUKIo6doHSk\nXaEZXR68UtXygvA4wdg7xZ6kKpXy0gu+RxGE6CGtDHTyDDzITu+NBjo21ZSsyGpQ\nJeIKftUCHdwdygKf0CdJx8A29GBRpHGCmJadmt7tTzOnYjmbuPVLeqJo/Ex9qXcG\nZbxoxnxr5NCocFeKx+EbLo+k/KjdFB2PKnhgzxAaMMMP6eXPr8l5AlzkC83EmPvN\ntveuaBbamdlFkD+53TZeZlxt3GIdq93Iw/UpbQ/pvhbrztMT+UVEkm15sShfX8Xn\nL2st5A4n0V+66QIDAQABoyAwHjAMBgNVHRMBAf8EAjAAMA4GA1UdDwEB/wQEAwIH\ngDANBgkqhkiG9w0BAQsFAAOCAQEAGpOqozyfDSBjoTepsRroxxcZ4sq65gw45Bme\nm36BS6FG0WHIg3cMy6KIIBefTDSKrPkKNTtuF25AeGn9jM+26cnfDM78ZH0+Lnn7\n7hs0MA64WMPQaWs9/+89aM9NADV9vp2zdG4xMi6B7DruvKWyhJaNoRqK/qP6LdSQ\nw8M+21sAHvXgrRkQtJlVOzVhgwt36NOb1hzRlQiZB+nhv2Wbw7fbtAaADk3JAumf\nvM+YdPS1KfAFaYefm4yFd+9/C0KOkHico3LTbELO5hG0Mo/EYvtjM+Fljb42EweF\n3nAx1GSPe5Tn8p3h6RyJW5HIKozEKyfDuLS0ccB/nqT3oNjcTw==\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIDRTCCAi2gAwIBAgIUcC33VfaMhOnsl7avNTRVQozoVtUwDQYJKoZIhvcNAQEL\nBQAwKjEPMA0GA1UECgwGUmF0aWZ5MRcwFQYDVQQDDA5SYXRpZnkgUm9vdCBDQTAe\nFw0yMzA2MjEwMTIyMzZaFw0yMzA2MjIwMTIyMzZaMCoxDzANBgNVBAoMBlJhdGlm\neTEXMBUGA1UEAwwOUmF0aWZ5IFJvb3QgQ0EwggEiMA0GCSqGSIb3DQEBAQUAA4IB\nDwAwggEKAoIBAQDDFhDnyPrVDZaeRu6Tbg1a/iTwus+IuX+h8aKhKS1yHz4EF/Lz\nxCy7lNSQ9srGMMVumWuNom/ydIphff6PejZM1jFKPU6OQR/0JX5epcVIjbKa562T\nDguUxJ+h5V3EIyM4RqOWQ2g/xZo86x5TzyNJXiVdHHRvmDvUNwPpMeDjr/EHVAni\n5YQObxkJRiiZ7XOa5zz3YztVm8sSZAwPWroY1HIfvtP+KHpiNDIKSymmuJkH4SEr\nJn++iqN8na18a9DFBPTTrLPe3CxATGrMfosCMZ6LP3iFLLc/FaSpwcnugWdewsUK\nYs+sUY7jFWR7x7/1nyFWyRrQviM4f4TY+K7NAgMBAAGjYzBhMB0GA1UdDgQWBBQH\nYePW7QPP2p1utr3r6gqzEkKs+DAfBgNVHSMEGDAWgBQHYePW7QPP2p1utr3r6gqz\nEkKs+DAPBgNVHRMBAf8EBTADAQH/MA4GA1UdDwEB/wQEAwICBDANBgkqhkiG9w0B\nAQsFAAOCAQEAjKp4vx3bFaKVhAbQeTsDjWJgmXLK2vLgt74MiUwSF6t0wehlfszE\nIcJagGJsvs5wKFf91bnwiqwPjmpse/thPNBAxh1uEoh81tOklv0BN790vsVpq3t+\ncnUvWPiCZdRlAiGGFtRmKk3Keq4sM6UdiUki9s+wnxypHVb4wIpVxu5R271Lnp5I\n+rb2EQ48iblt4XZPczf/5QJdTgbItjBNbuO8WVPOqUIhCiFuAQziLtNUq3p81dHO\nQ2BPgmaitCpIUYHVYighLauBGCH8xOFzj4a4KbOxKdxyJTd0La/vRCKaUtJX67Lc\nfQYVR9HXQZ0YlmwPcmIG5v7wBfcW34NUvA==\n-----END CERTIFICATE-----\n",
			contentType: "application/x-pem-file",
			id:          "https://notarycerts.vault.azure.net/secrets/testCert6212/431ad135165741dcb95a46cf3e6686fb",
			expectedErr: false,
		},
		{
			desc:        "PKCS12 Content Type",
			value:       "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",
			contentType: "application/x-pkcs12",
			id:          "https://notarycerts.vault.azure.net/secrets/testCert6212/431ad135165741dcb95a46cf3e6686fb",
			expectedErr: false,
		},
		{
			desc:        "Invalid PKCS12 Content",
			value:       "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",
			contentType: "application/x-pkcs12",
			id:          "https://notarycerts.vault.azure.net/secrets/testCert6212/431ad135165741dcb95a46cf3e6686fb",
			expectedErr: true,
		},
		{
			desc:        "Secret Text File",
			value:       "text",
			contentType: "text",
			id:          "https://notarycerts.vault.azure.net/secrets/testCert6212/431ad135165741dcb95a46cf3e6686fb",
			expectedErr: true,
		},
		{
			desc:        "Test empty",
			value:       "",
			contentType: "",
			id:          "",
			expectedErr: true,
		},
	}

	for i, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			testdata := azsecrets.SecretBundle{
				Value:       &cases[i].value,
				ID:          &cases[i].id,
				ContentType: &cases[i].contentType,
			}

			certs, status, err := getCertsFromSecretBundle(context.Background(), testdata, "certName", true)
			if tc.expectedErr {
				assert.NotNil(t, err)
				assert.Nil(t, certs)
				assert.Nil(t, status)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetKeyFromKeyBundle(t *testing.T) {
	unsupportedType := azkeys.JSONWebKeyType("abc")
	cases := []struct {
		desc        string
		keyBundle   azkeys.KeyBundle
		expectedErr bool
		output      crypto.PublicKey
	}{
		{
			desc: "no key in key bundle",
			keyBundle: azkeys.KeyBundle{
				Key: nil,
			},
			expectedErr: true,
			output:      nil,
		},
		{
			desc: "invalid key in key bundle with nil Kty",
			keyBundle: azkeys.KeyBundle{
				Key: &azkeys.JSONWebKey{
					Kty: nil,
				},
			},
			expectedErr: true,
			output:      nil,
		},
		{
			desc: "key with unsupported Kty value",
			keyBundle: azkeys.KeyBundle{
				Key: &azkeys.JSONWebKey{
					Kty: &unsupportedType, // Unsupported key type
				},
			},
			expectedErr: true,
			output:      nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			key, err := getKeyFromKeyBundle(tc.keyBundle)
			if tc.expectedErr {
				assert.NotNil(t, err)
				assert.Nil(t, key)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, key)
			}
			if tc.output != nil {
				assert.Equal(t, tc.output, key)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	vaultURI := "https://test.vault.azure.net"
	tenantID := "testTenantID"
	clientID := "testClientID"
	validTestCerts := []types.KeyVaultValue{
		{
			Name:    "testCert",
			Version: "testVersion",
		},
	}
	validTestKeys := []types.KeyVaultValue{
		{
			Name:    "testKey",
			Version: "testVersion",
		},
	}

	cases := []struct {
		desc        string
		provider    akvKMProvider
		expectedErr bool
	}{
		{
			desc:        "Valid Provider",
			expectedErr: false,
			provider: akvKMProvider{
				vaultURI:     vaultURI,
				tenantID:     tenantID,
				clientID:     clientID,
				certificates: validTestCerts,
				keys:         validTestKeys,
			},
		},
		{
			desc:        "Missing Vault URI",
			expectedErr: true,
			provider: akvKMProvider{
				tenantID:     tenantID,
				clientID:     clientID,
				certificates: validTestCerts,
				keys:         validTestKeys,
			},
		},
		{
			desc:        "Missing Tenant ID",
			expectedErr: true,
			provider: akvKMProvider{
				vaultURI:     vaultURI,
				clientID:     clientID,
				certificates: validTestCerts,
				keys:         validTestKeys,
			},
		},
		{
			desc:        "Missing Client ID",
			expectedErr: true,
			provider: akvKMProvider{
				vaultURI:     vaultURI,
				tenantID:     tenantID,
				certificates: validTestCerts,
				keys:         validTestKeys,
			},
		},
		{
			desc:        "Missing Certificate Name",
			expectedErr: true,
			provider: akvKMProvider{
				vaultURI: vaultURI,
				tenantID: tenantID,
				clientID: clientID,
				keys:     validTestKeys,
				certificates: []types.KeyVaultValue{
					{
						Version: "testVersion",
					},
				},
			},
		},
		{
			desc:        "Missing Key Name",
			expectedErr: true,
			provider: akvKMProvider{
				vaultURI:     vaultURI,
				tenantID:     tenantID,
				clientID:     clientID,
				certificates: validTestCerts,
				keys: []types.KeyVaultValue{
					{
						Version: "testVersion",
					},
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.provider.validate()
			if tc.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

// Mock clients
type MockAzKeysClient struct {
	mock.Mock
}

type MockAzSecretsClient struct {
	mock.Mock
}

type MockAzCertificatesClient struct {
	mock.Mock
}

type MockWorkloadIdentityCredential struct {
	mock.Mock
}

// Mock functions
func (m *MockWorkloadIdentityCredential) NewWorkloadIdentityCredential(options *azidentity.WorkloadIdentityCredentialOptions) (*MockWorkloadIdentityCredential, error) {
	args := m.Called(options)
	return args.Get(0).(*MockWorkloadIdentityCredential), args.Error(1)
}

func (m *MockAzKeysClient) NewClient(endpoint string, credential *azidentity.WorkloadIdentityCredential, options *azkeys.ClientOptions) (*azkeys.Client, error) {
	args := m.Called(endpoint, credential, options)
	return args.Get(0).(*azkeys.Client), args.Error(1)
}

func (m *MockAzSecretsClient) NewClient(endpoint string, credential *azidentity.WorkloadIdentityCredential, options *azsecrets.ClientOptions) (*azsecrets.Client, error) {
	args := m.Called(endpoint, credential, options)
	return args.Get(0).(*azsecrets.Client), args.Error(1)
}

func (m *MockAzCertificatesClient) NewClient(endpoint string, credential *azidentity.WorkloadIdentityCredential, options *azcertificates.ClientOptions) (*azcertificates.Client, error) {
	args := m.Called(endpoint, credential, options)
	return args.Get(0).(*azcertificates.Client), args.Error(1)
}

func TestInitializeKvClient(t *testing.T) {
	mockCredential := new(MockWorkloadIdentityCredential)
	mockKeysClient := new(MockAzKeysClient)
	mockSecretsClient := new(MockAzSecretsClient)
	mockCertificatesClient := new(MockAzCertificatesClient)

	tests := []struct {
		name              string
		kvEndpoint        string
		userAgent         string
		tenantID          string
		clientID          string
		mockCredentialErr error
		mockKeysErr       error
		mockSecretsErr    error
		expectedErr       bool
	}{
		{
			name:        "Empty user agent",
			kvEndpoint:  "https://test.vault.azure.net",
			userAgent:   "",
			expectedErr: true,
		},
		{
			name:        "Auth failure",
			kvEndpoint:  "https://test.vault.azure.net",
			tenantID:    "testTenantID",
			clientID:    "testClientID",
			expectedErr: true,
		},
		{
			name:              "credential creation error",
			kvEndpoint:        "https://test-keyvault.vault.azure.net",
			tenantID:          "test-tenant-id",
			clientID:          "test-client-id",
			mockCredentialErr: errors.New("failed to create workload identity credential"),
			expectedErr:       true,
		},
		{
			name:        "azkeys client creation error",
			kvEndpoint:  "https://test-keyvault.vault.azure.net",
			tenantID:    "test-tenant-id",
			clientID:    "test-client-id",
			mockKeysErr: errors.New("failed to create azkeys client"),
			expectedErr: true,
		},
		{
			name:           "azsecrets client creation error",
			kvEndpoint:     "https://test-keyvault.vault.azure.net",
			tenantID:       "test-tenant-id",
			clientID:       "test-client-id",
			mockSecretsErr: errors.New("failed to create azsecrets client"),
			expectedErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mocks
			mockCredential.On("NewWorkloadIdentityCredential", mock.Anything).Return(mockCredential, tt.mockCredentialErr)
			mockKeysClient.On("NewClient", tt.kvEndpoint, mockCredential, mock.Anything).Return(mockKeysClient, tt.mockKeysErr)
			mockSecretsClient.On("NewClient", tt.kvEndpoint, mockCredential, mock.Anything).Return(mockSecretsClient, tt.mockSecretsErr)
			mockCertificatesClient.On("NewClient", tt.kvEndpoint, mockCredential, mock.Anything).Return(mockCertificatesClient, tt.mockSecretsErr)

			// Call function under test
			keysKVClient, secretsKVClient, certificatesKVClient, err := initializeKvClient(tt.kvEndpoint, tt.tenantID, tt.clientID, nil)

			// Validate expectations
			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, keysKVClient)
				assert.Nil(t, secretsKVClient)
				assert.Nil(t, certificatesKVClient)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, keysKVClient)
				assert.NotNil(t, secretsKVClient)
				assert.Nil(t, certificatesKVClient)
			}
		})
	}
}

// Test cases for keyType switch case handling
func TestGetKeyFromKeyBundlex(t *testing.T) {
	tests := []struct {
		name     string
		keyType  azkeys.JSONWebKeyType
		expected azkeys.JSONWebKeyType
		curve    azkeys.JSONWebKeyCurveName
		x        []byte
		y        []byte
		n        []byte
		e        []byte
	}{
		{
			name:     "Test ECHSM to EC",
			keyType:  azkeys.JSONWebKeyTypeECHSM,
			expected: azkeys.JSONWebKeyTypeEC,
			curve:    azkeys.JSONWebKeyCurveNameP256,                                                                                                                                                                         // Example curve name
			x:        []byte{0x6b, 0x17, 0xd1, 0xf2, 0xe1, 0x2c, 0x42, 0x47, 0xf8, 0xbc, 0xe6, 0xe5, 0x63, 0xa4, 0x40, 0xf2, 0x77, 0x03, 0x7d, 0x81, 0x2d, 0xeb, 0x33, 0xa0, 0xf4, 0xa1, 0x39, 0x45, 0xd8, 0x98, 0xc2, 0x96}, // Valid x-coordinate for P-256
			y:        []byte{0x4f, 0xe3, 0x42, 0xe2, 0xfe, 0x1a, 0x7f, 0x9b, 0x8e, 0xe7, 0xeb, 0x4a, 0x7c, 0x0f, 0x9e, 0x16, 0x2b, 0xce, 0x33, 0x57, 0x6b, 0x31, 0x5e, 0xce, 0xcb, 0xb6, 0x40, 0x68, 0x37, 0xbf, 0x51, 0xf5}, // Valid y-coordinate for P-256
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webKey := &azkeys.JSONWebKey{
				Kty: &tt.keyType,
			}
			if tt.keyType == azkeys.JSONWebKeyTypeECHSM {
				webKey.Crv = &tt.curve
				webKey.X = tt.x
				webKey.Y = tt.y
			}
			keyBundle := azkeys.KeyBundle{
				Key: webKey,
			}

			_, err := getKeyFromKeyBundle(keyBundle)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, *webKey.Kty)
		})
	}
}

func TestInitializeKvClient_Success(t *testing.T) {
	// Mock the context and input parameters
	keyVaultEndpoint := "https://myvault.vault.azure.net/"
	tenantID := "tenant-id"
	clientID := "client-id"

	// Create a mock credential provider
	mockCredential, err := azidentity.NewClientSecretCredential(tenantID, clientID, "fake-secret", nil)
	if err != nil {
		t.Fatalf("Failed to create mock credential: %v", err)
	}

	// Run the function with the mock credential
	keysKVClient, secretsKVClient, certificatesKVClient, err := initializeKvClient(keyVaultEndpoint, tenantID, clientID, mockCredential)

	// Assert the function succeeds without errors and clients are created
	assert.NotNil(t, keysKVClient)
	assert.NotNil(t, secretsKVClient)
	assert.NotNil(t, certificatesKVClient)
	assert.NoError(t, err)
}

func TestInitializeKvClient_FailureInAzKeysClient(t *testing.T) {
	// Mock the context and input parameters
	keyVaultEndpoint := "https://invalid-vault.vault.azure.net/"
	tenantID := "mock_tenant-id"
	clientID := "mock_client-id"

	// Run the function
	keysKVClient, secretsKVClient, certificatesKVClient, err := initializeKvClient(keyVaultEndpoint, tenantID, clientID, nil)

	// Assert that an error occurred and clients were not created
	assert.Nil(t, keysKVClient)
	assert.Nil(t, secretsKVClient)
	assert.Nil(t, certificatesKVClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create workload identity credential")
}

func TestInitializeKvClient_FailureInAzSecretsClient(t *testing.T) {
	// Mock the context and input parameters
	keyVaultEndpoint := "https://valid-vault.vault.azure.net/"
	tenantID := "tenant-id"
	clientID := "client-id"

	// Modify the azsecrets.NewClient function to simulate failure
	// Run the function
	keysKVClient, secretsKVClient, certificatesKVClient, err := initializeKvClient(keyVaultEndpoint, tenantID, clientID, nil)

	// Assert that an error occurred and clients were not created
	assert.Nil(t, keysKVClient)
	assert.Nil(t, secretsKVClient)
	assert.Nil(t, certificatesKVClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create workload identity credential")
}

func TestInitializeKvClient_FailureInAzCertificatesClient(t *testing.T) {
	// Mock the context and input parameters
	keyVaultEndpoint := "https://valid-vault.vault.azure.net/"
	tenantID := "tenant-id"
	clientID := "client-id"

	// Modify the azsecrets.NewClient function to simulate failure
	// Run the function
	keysKVClient, secretsKVClient, certificatesKVClient, err := initializeKvClient(keyVaultEndpoint, tenantID, clientID, nil)

	// Assert that an error occurred and clients were not created
	assert.Nil(t, keysKVClient)
	assert.Nil(t, secretsKVClient)
	assert.Nil(t, certificatesKVClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create workload identity credential")
}
