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

package refresh

import (
	"context"
	"crypto"
	"crypto/x509"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	corecrl "github.com/notaryproject/notation-core-go/revocation/crl"
	re "github.com/ratify-project/ratify/errors"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/config"
	_ "github.com/ratify-project/ratify/pkg/keymanagementprovider/inline"
	mock "github.com/ratify-project/ratify/pkg/keymanagementprovider/mocks"
	nv "github.com/ratify-project/ratify/pkg/verifier/notation"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestKubeRefresher_Refresh(t *testing.T) {
	tests := []struct {
		name                    string
		providerRawParameters   []byte
		providerType            string
		providerRefreshInterval string
		GetCertsFunc            func(_ context.Context) (map[keymanagementprovider.KMPMapKey][]*x509.Certificate, keymanagementprovider.KeyManagementProviderStatus, error)
		GetKeysFunc             func(_ context.Context) (map[keymanagementprovider.KMPMapKey]crypto.PublicKey, keymanagementprovider.KeyManagementProviderStatus, error)
		IsRefreshableFunc       func() bool
		NewCRLHandler           nv.RevocationFactory
		expectedResult          ctrl.Result
		expectedError           bool
	}{
		{
			name:                  "Non-refreshable",
			providerRawParameters: []byte(`{"contentType": "certificate", "value": "-----BEGIN CERTIFICATE-----\nMIID2jCCAsKgAwIBAgIQXy2VqtlhSkiZKAGhsnkjbDANBgkqhkiG9w0BAQsFADBvMRswGQYDVQQD\nExJyYXRpZnkuZXhhbXBsZS5jb20xDzANBgNVBAsTBk15IE9yZzETMBEGA1UEChMKTXkgQ29tcGFu\neTEQMA4GA1UEBxMHUmVkbW9uZDELMAkGA1UECBMCV0ExCzAJBgNVBAYTAlVTMB4XDTIzMDIwMTIy\nNDUwMFoXDTI0MDIwMTIyNTUwMFowbzEbMBkGA1UEAxMScmF0aWZ5LmV4YW1wbGUuY29tMQ8wDQYD\nVQQLEwZNeSBPcmcxEzARBgNVBAoTCk15IENvbXBhbnkxEDAOBgNVBAcTB1JlZG1vbmQxCzAJBgNV\nBAgTAldBMQswCQYDVQQGEwJVUzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAL10bM81\npPAyuraORABsOGS8M76Bi7Guwa3JlM1g2D8CuzSfSTaaT6apy9GsccxUvXd5cmiP1ffna5z+EFmc\nizFQh2aq9kWKWXDvKFXzpQuhyqD1HeVlRlF+V0AfZPvGt3VwUUjNycoUU44ctCWmcUQP/KShZev3\n6SOsJ9q7KLjxxQLsUc4mg55eZUThu8mGB8jugtjsnLUYvIWfHhyjVpGrGVrdkDMoMn+u33scOmrt\nsBljvq9WVo4T/VrTDuiOYlAJFMUae2Ptvo0go8XTN3OjLblKeiK4C+jMn9Dk33oGIT9pmX0vrDJV\nX56w/2SejC1AxCPchHaMuhlwMpftBGkCAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeAMAkGA1UdEwQC\nMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAU0eaKkZj+MS9jCp9Dg1zdv3v/aKww\nHQYDVR0OBBYEFNHmipGY/jEvYwqfQ4Nc3b97/2isMA0GCSqGSIb3DQEBCwUAA4IBAQBNDcmSBizF\nmpJlD8EgNcUCy5tz7W3+AAhEbA3vsHP4D/UyV3UgcESx+L+Nye5uDYtTVm3lQejs3erN2BjW+ds+\nXFnpU/pVimd0aYv6mJfOieRILBF4XFomjhrJOLI55oVwLN/AgX6kuC3CJY2NMyJKlTao9oZgpHhs\nLlxB/r0n9JnUoN0Gq93oc1+OLFjPI7gNuPXYOP1N46oKgEmAEmNkP1etFrEjFRgsdIFHksrmlOlD\nIed9RcQ087VLjmuymLgqMTFX34Q3j7XgN2ENwBSnkHotE9CcuGRW+NuiOeJalL8DBmFXXWwHTKLQ\nPp5g6m1yZXylLJaFLKz7tdMmO355\n-----END CERTIFICATE-----\n"}`),
			providerType:          "inline",
			IsRefreshableFunc:     func() bool { return false },
			NewCRLHandler:         nv.CreateCRLHandlerFromConfig(),
			expectedResult:        ctrl.Result{},
			expectedError:         false,
		},
		{
			name:                    "Disabled",
			providerRawParameters:   []byte(`{"vaultURI": "https://yourkeyvault.vault.azure.net/", "certificates": [{"name": "cert1", "version": "1"}], "tenantID": "yourtenantID", "clientID": "yourclientID"}`),
			providerType:            "test-kmp",
			providerRefreshInterval: "",
			NewCRLHandler:           nv.CreateCRLHandlerFromConfig(),
			IsRefreshableFunc:       func() bool { return true },
			expectedResult:          ctrl.Result{},
			expectedError:           false,
		},
		{
			name:                    "Refreshable",
			providerRawParameters:   []byte(`{"vaultURI": "https://yourkeyvault.vault.azure.net/", "certificates": [{"name": "cert1", "version": "1"}], "tenantID": "yourtenantID", "clientID": "yourclientID"}`),
			providerType:            "test-kmp",
			providerRefreshInterval: "1m",
			NewCRLHandler:           nv.CreateCRLHandlerFromConfig(),
			IsRefreshableFunc:       func() bool { return true },
			expectedResult:          ctrl.Result{RequeueAfter: time.Minute},
			expectedError:           false,
		},
		{
			name:                    "Invalid Interval",
			providerRawParameters:   []byte(`{"vaultURI": "https://yourkeyvault.vault.azure.net/", "certificates": [{"name": "cert1", "version": "1"}], "tenantID": "yourtenantID", "clientID": "yourclientID"}`),
			providerType:            "test-kmp",
			providerRefreshInterval: "1mm",
			NewCRLHandler:           nv.CreateCRLHandlerFromConfig(),
			IsRefreshableFunc:       func() bool { return true },
			expectedResult:          ctrl.Result{},
			expectedError:           true,
		},
		{
			name: "Error Fetching Certificates",
			GetCertsFunc: func(_ context.Context) (map[keymanagementprovider.KMPMapKey][]*x509.Certificate, keymanagementprovider.KeyManagementProviderStatus, error) {
				// Example behavior: Return an error
				return nil, nil, errors.New("test error")
			},
			providerRawParameters: []byte(`{"vaultURI": "https://yourkeyvault.vault.azure.net/", "certificates": [{"name": "cert1", "version": "1"}], "tenantID": "yourtenantID", "clientID": "yourclientID"}`),
			providerType:          "test-kmp-error",
			IsRefreshableFunc:     func() bool { return true },
			NewCRLHandler:         nv.CreateCRLHandlerFromConfig(),
			expectedError:         true,
		},
		{
			name: "Error Fetching Keys",
			GetKeysFunc: func(_ context.Context) (map[keymanagementprovider.KMPMapKey]crypto.PublicKey, keymanagementprovider.KeyManagementProviderStatus, error) {
				// Example behavior: Return an error
				return nil, nil, errors.New("test error")
			},
			providerRawParameters: []byte(`{"vaultURI": "https://yourkeyvault.vault.azure.net/", "certificates": [{"name": "cert1", "version": "1"}], "tenantID": "yourtenantID", "clientID": "yourclientID"}`),
			providerType:          "test-kmp-error",
			IsRefreshableFunc:     func() bool { return true },
			NewCRLHandler:         nv.CreateCRLHandlerFromConfig(),
			expectedError:         true,
		},
		{
			name: "Error Caching with CRL Fetcher (non-blocking)",
			GetCertsFunc: func(_ context.Context) (map[keymanagementprovider.KMPMapKey][]*x509.Certificate, keymanagementprovider.KeyManagementProviderStatus, error) {
				return map[keymanagementprovider.KMPMapKey][]*x509.Certificate{
					{Name: "sample"}: {&x509.Certificate{}},
				}, keymanagementprovider.KeyManagementProviderStatus{}, nil
			},
			providerRawParameters:   []byte(`{"vaultURI": "https://yourkeyvault.vault.azure.net/", "certificates": [{"name": "cert1", "version": "1"}], "tenantID": "yourtenantID", "clientID": "yourclientID"}`),
			providerType:            "test-kmp",
			providerRefreshInterval: "1m",
			IsRefreshableFunc:       func() bool { return true },
			NewCRLHandler:           &MockCRLHandler{CacheDisabled: false, httpClient: &http.Client{}},
			expectedResult:          ctrl.Result{RequeueAfter: time.Minute},
			expectedError:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var factory mock.TestKeyManagementProviderFactory
			if tt.GetCertsFunc != nil {
				factory = mock.TestKeyManagementProviderFactory{
					GetCertsFunc:      tt.GetCertsFunc,
					IsRefreshableFunc: tt.IsRefreshableFunc,
				}
			} else if tt.GetKeysFunc != nil {
				factory = mock.TestKeyManagementProviderFactory{
					GetKeysFunc:       tt.GetKeysFunc,
					IsRefreshableFunc: tt.IsRefreshableFunc,
				}
			} else {
				factory = mock.TestKeyManagementProviderFactory{
					IsRefreshableFunc: tt.IsRefreshableFunc,
				}
			}

			provider, _ := factory.Create("", config.KeyManagementProviderConfig{}, "")

			kr := &KubeRefresher{
				Provider:                provider,
				ProviderType:            tt.providerType,
				ProviderRefreshInterval: tt.providerRefreshInterval,
				Resource:                "kmpname",
				CRLHandler:              tt.NewCRLHandler,
			}

			err := kr.Refresh(context.Background())
			result := kr.GetResult()
			if !reflect.DeepEqual(result, tt.expectedResult) {
				t.Fatalf("Expected nil but got %v with error %v", result, err)
			}
			if tt.expectedError && err == nil {
				t.Fatalf("Expected error but got nil")
			}
		})
	}
}

type MockCRLHandler struct {
	CacheDisabled bool
	httpClient    *http.Client
}

func (h *MockCRLHandler) NewFetcher() (corecrl.Fetcher, error) {
	return nil, re.ErrorCodeConfigInvalid.WithDetail("failed to create CRL fetcher")
}

func TestKubeRefresher_GetResult(t *testing.T) {
	kr := &KubeRefresher{
		Result:     ctrl.Result{RequeueAfter: time.Minute},
		CRLHandler: nv.CreateCRLHandlerFromConfig(),
	}

	result := kr.GetResult()
	expectedResult := ctrl.Result{RequeueAfter: time.Minute}

	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("Expected result %v, but got %v", expectedResult, result)
	}
}
func TestKubeRefresher_GetStatus(t *testing.T) {
	kr := &KubeRefresher{
		Status: keymanagementprovider.KeyManagementProviderStatus{
			"attribute1": "value1",
			"attribute2": "value2",
		},
		CRLHandler: nv.CreateCRLHandlerFromConfig(),
	}

	status := kr.GetStatus()
	expectedStatus := keymanagementprovider.KeyManagementProviderStatus{
		"attribute1": "value1",
		"attribute2": "value2",
	}

	if !reflect.DeepEqual(status, expectedStatus) {
		t.Fatalf("Expected status %v, but got %v", expectedStatus, status)
	}
}
func TestKubeRefresher_Create(t *testing.T) {
	tests := []struct {
		name                    string
		config                  RefresherConfig
		expectedProviderType    string
		expectedRefreshInterval string
		expectedResource        string
	}{
		{
			name: "Valid Config",
			config: RefresherConfig{
				Provider:                &mock.TestKeyManagementProvider{},
				ProviderType:            "test-kmp",
				ProviderRefreshInterval: "1m",
				Resource:                "test-resource",
			},
			expectedProviderType:    "test-kmp",
			expectedRefreshInterval: "1m",
			expectedResource:        "test-resource",
		},
		{
			name: "Empty Config",
			config: RefresherConfig{
				Provider:                nil,
				ProviderType:            "",
				ProviderRefreshInterval: "",
				Resource:                "",
			},
			expectedProviderType:    "",
			expectedRefreshInterval: "",
			expectedResource:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kr := &KubeRefresher{CRLHandler: nv.CreateCRLHandlerFromConfig()}
			refresher, err := kr.Create(tt.config)
			if err != nil {
				t.Fatalf("Expected no error, but got %v", err)
			}

			kubeRefresher, ok := refresher.(*KubeRefresher)
			if !ok {
				t.Fatalf("Expected KubeRefresher type, but got %T", refresher)
			}

			if kubeRefresher.ProviderType != tt.expectedProviderType {
				t.Fatalf("Expected ProviderType %v, but got %v", tt.expectedProviderType, kubeRefresher.ProviderType)
			}

			if kubeRefresher.ProviderRefreshInterval != tt.expectedRefreshInterval {
				t.Fatalf("Expected ProviderRefreshInterval %v, but got %v", tt.expectedRefreshInterval, kubeRefresher.ProviderRefreshInterval)
			}

			if kubeRefresher.Resource != tt.expectedResource {
				t.Fatalf("Expected Resource %v, but got %v", tt.expectedResource, kubeRefresher.Resource)
			}
		})
	}
}
