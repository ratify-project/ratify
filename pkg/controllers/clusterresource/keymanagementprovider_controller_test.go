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

package clusterresource

import (
	"context"
	"fmt"
	"testing"

	configv1beta1 "github.com/ratify-project/ratify/api/v1beta1"
	"github.com/ratify-project/ratify/internal/constants"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	test "github.com/ratify-project/ratify/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	kmpName = "kmpName"
)

// TestUpdateErrorStatus tests the updateErrorStatus method
func TestKMProviderUpdateErrorStatus(t *testing.T) {
	var parametersString = "{\"certs\":{\"name\":\"certName\"}}"
	var kmProviderStatus = []byte(parametersString)

	status := configv1beta1.KeyManagementProviderStatus{
		IsSuccess: true,
		Properties: runtime.RawExtension{
			Raw: kmProviderStatus,
		},
	}
	keyManagementProvider := configv1beta1.KeyManagementProvider{
		Status: status,
	}
	expectedErr := "it's a long error from unit test"
	lastFetchedTime := metav1.Now()
	updateKMProviderErrorStatus(&keyManagementProvider, expectedErr, &lastFetchedTime)

	if keyManagementProvider.Status.IsSuccess != false {
		t.Fatalf("Unexpected error, expected isSuccess to be false , actual %+v", keyManagementProvider.Status.IsSuccess)
	}

	if keyManagementProvider.Status.Error != expectedErr {
		t.Fatalf("Unexpected error string, expected %+v, got %+v", expectedErr, keyManagementProvider.Status.Error)
	}
	expectedBriedErr := fmt.Sprintf("%s...", expectedErr[:30])
	if keyManagementProvider.Status.BriefError != expectedBriedErr {
		t.Fatalf("Unexpected error string, expected %+v, got %+v", expectedBriedErr, keyManagementProvider.Status.Error)
	}

	//make sure properties of last cached cert was not overridden
	if len(keyManagementProvider.Status.Properties.Raw) == 0 {
		t.Fatalf("Unexpected properties,  expected %+v, got %+v", parametersString, string(keyManagementProvider.Status.Properties.Raw))
	}
}

// TestKMProviderUpdateSuccessStatus tests the updateSuccessStatus method
func TestKMProviderUpdateSuccessStatus(t *testing.T) {
	kmProviderStatus := keymanagementprovider.KeyManagementProviderStatus{}
	properties := map[string]string{}
	properties["Name"] = "wabbit"
	properties["Version"] = "ABC"

	kmProviderStatus["Certificates"] = properties

	lastFetchedTime := metav1.Now()

	status := configv1beta1.KeyManagementProviderStatus{
		IsSuccess: false,
		Error:     "error from last operation",
	}
	keyManagementProvider := configv1beta1.KeyManagementProvider{
		Status: status,
	}

	updateKMProviderSuccessStatus(&keyManagementProvider, &lastFetchedTime, kmProviderStatus)

	if keyManagementProvider.Status.IsSuccess != true {
		t.Fatalf("Expected isSuccess to be true , actual %+v", keyManagementProvider.Status.IsSuccess)
	}

	if keyManagementProvider.Status.Error != "" {
		t.Fatalf("Unexpected error string, actual %+v", keyManagementProvider.Status.Error)
	}

	//make sure properties of last cached cert was updated
	if len(keyManagementProvider.Status.Properties.Raw) == 0 {
		t.Fatalf("Properties should not be empty")
	}
}

// TestKMProviderUpdateSuccessStatus tests the updateSuccessStatus method with empty properties
func TestKMProviderUpdateSuccessStatus_emptyProperties(t *testing.T) {
	lastFetchedTime := metav1.Now()
	status := configv1beta1.KeyManagementProviderStatus{
		IsSuccess: false,
		Error:     "error from last operation",
	}
	keyManagementProvider := configv1beta1.KeyManagementProvider{
		Status: status,
	}

	updateKMProviderSuccessStatus(&keyManagementProvider, &lastFetchedTime, nil)

	if keyManagementProvider.Status.IsSuccess != true {
		t.Fatalf("Expected isSuccess to be true , actual %+v", keyManagementProvider.Status.IsSuccess)
	}

	if keyManagementProvider.Status.Error != "" {
		t.Fatalf("Unexpected error string, actual %+v", keyManagementProvider.Status.Error)
	}

	//make sure properties of last cached cert was updated
	if len(keyManagementProvider.Status.Properties.Raw) != 0 {
		t.Fatalf("Properties should be empty")
	}
}

func TestWriteKMProviderStatus(t *testing.T) {
	logger := logrus.WithContext(context.Background())
	lastFetchedTime := metav1.Now()
	testCases := []struct {
		name       string
		isSuccess  bool
		kmProvider *configv1beta1.KeyManagementProvider
		errString  string
		reconciler client.StatusClient
	}{
		{
			name:       "success status",
			isSuccess:  true,
			errString:  "",
			kmProvider: &configv1beta1.KeyManagementProvider{},
			reconciler: &test.MockStatusClient{},
		},
		{
			name:       "error status",
			isSuccess:  false,
			kmProvider: &configv1beta1.KeyManagementProvider{},
			errString:  "a long error string that exceeds the max length of 30 characters",
			reconciler: &test.MockStatusClient{},
		},
		{
			name:       "status update failed",
			isSuccess:  true,
			kmProvider: &configv1beta1.KeyManagementProvider{},
			reconciler: &test.MockStatusClient{
				UpdateFailed: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			writeKMProviderStatus(context.Background(), tc.reconciler, tc.kmProvider, logger, tc.isSuccess, tc.errString, lastFetchedTime, nil)

			if tc.kmProvider.Status.IsSuccess != tc.isSuccess {
				t.Fatalf("Expected isSuccess to be %+v , actual %+v", tc.isSuccess, tc.kmProvider.Status.IsSuccess)
			}

			if tc.kmProvider.Status.Error != tc.errString {
				t.Fatalf("Expected Error to be %+v , actual %+v", tc.errString, tc.kmProvider.Status.Error)
			}
		})
	}
}

func TestKMPReconcile(t *testing.T) {
	tests := []struct {
		name             string
		description      string
		provider         *configv1beta1.KeyManagementProvider
		req              *reconcile.Request
		expectedErr      bool
		expectedKMPCount int
	}{
		{
			name:        "nonexistent KMP",
			description: "Reconciling a non-existent KMP CR, it should be deleted from maps",
			req: &reconcile.Request{
				NamespacedName: types.NamespacedName{Name: "nonexistent"},
			},
			provider: &configv1beta1.KeyManagementProvider{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.EmptyNamespace,
					Name:      kmpName,
				},
				Spec: configv1beta1.KeyManagementProviderSpec{
					Type: "inline",
				},
			},
			expectedErr:      false,
			expectedKMPCount: 0,
		},
		{
			name:        "invalid params",
			description: "Received invalid parameters of the KMP Spec, it should fail the reconcile and return an error",
			provider: &configv1beta1.KeyManagementProvider{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.EmptyNamespace,
					Name:      kmpName,
				},
				Spec: configv1beta1.KeyManagementProviderSpec{
					Type: "inline",
				},
			},
			expectedErr:      true,
			expectedKMPCount: 0,
		},
		{
			name:        "valid params",
			description: "Received a valid KMP manifest, it should be added to the cert map",
			provider: &configv1beta1.KeyManagementProvider{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.EmptyNamespace,
					Name:      kmpName,
				},
				Spec: configv1beta1.KeyManagementProviderSpec{
					Type: "inline",
					Parameters: runtime.RawExtension{
						Raw: []byte(`{"type": "inline", "contentType": "certificate", "value": "-----BEGIN CERTIFICATE-----\nMIID2jCCAsKgAwIBAgIQXy2VqtlhSkiZKAGhsnkjbDANBgkqhkiG9w0BAQsFADBvMRswGQYDVQQD\nExJyYXRpZnkuZXhhbXBsZS5jb20xDzANBgNVBAsTBk15IE9yZzETMBEGA1UEChMKTXkgQ29tcGFu\neTEQMA4GA1UEBxMHUmVkbW9uZDELMAkGA1UECBMCV0ExCzAJBgNVBAYTAlVTMB4XDTIzMDIwMTIy\nNDUwMFoXDTI0MDIwMTIyNTUwMFowbzEbMBkGA1UEAxMScmF0aWZ5LmV4YW1wbGUuY29tMQ8wDQYD\nVQQLEwZNeSBPcmcxEzARBgNVBAoTCk15IENvbXBhbnkxEDAOBgNVBAcTB1JlZG1vbmQxCzAJBgNV\nBAgTAldBMQswCQYDVQQGEwJVUzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAL10bM81\npPAyuraORABsOGS8M76Bi7Guwa3JlM1g2D8CuzSfSTaaT6apy9GsccxUvXd5cmiP1ffna5z+EFmc\nizFQh2aq9kWKWXDvKFXzpQuhyqD1HeVlRlF+V0AfZPvGt3VwUUjNycoUU44ctCWmcUQP/KShZev3\n6SOsJ9q7KLjxxQLsUc4mg55eZUThu8mGB8jugtjsnLUYvIWfHhyjVpGrGVrdkDMoMn+u33scOmrt\nsBljvq9WVo4T/VrTDuiOYlAJFMUae2Ptvo0go8XTN3OjLblKeiK4C+jMn9Dk33oGIT9pmX0vrDJV\nX56w/2SejC1AxCPchHaMuhlwMpftBGkCAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeAMAkGA1UdEwQC\nMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAU0eaKkZj+MS9jCp9Dg1zdv3v/aKww\nHQYDVR0OBBYEFNHmipGY/jEvYwqfQ4Nc3b97/2isMA0GCSqGSIb3DQEBCwUAA4IBAQBNDcmSBizF\nmpJlD8EgNcUCy5tz7W3+AAhEbA3vsHP4D/UyV3UgcESx+L+Nye5uDYtTVm3lQejs3erN2BjW+ds+\nXFnpU/pVimd0aYv6mJfOieRILBF4XFomjhrJOLI55oVwLN/AgX6kuC3CJY2NMyJKlTao9oZgpHhs\nLlxB/r0n9JnUoN0Gq93oc1+OLFjPI7gNuPXYOP1N46oKgEmAEmNkP1etFrEjFRgsdIFHksrmlOlD\nIed9RcQ087VLjmuymLgqMTFX34Q3j7XgN2ENwBSnkHotE9CcuGRW+NuiOeJalL8DBmFXXWwHTKLQ\nPp5g6m1yZXylLJaFLKz7tdMmO355\n-----END CERTIFICATE-----\n"}`),
					},
				},
			},
			expectedErr:      false,
			expectedKMPCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetKMP()
			scheme, _ := test.CreateScheme()
			client := fake.NewClientBuilder().WithScheme(scheme)
			client.WithObjects(tt.provider)
			r := &KeyManagementProviderReconciler{
				Scheme: scheme,
				Client: client.Build(),
			}
			var req reconcile.Request
			if tt.req != nil {
				req = *tt.req
			} else {
				req = reconcile.Request{
					NamespacedName: test.KeyFor(tt.provider),
				}
			}

			_, err := r.Reconcile(context.Background(), req)
			if tt.expectedErr != (err != nil) {
				t.Fatalf("Reconcile() expected error %v, actual %v", tt.expectedErr, err)
			}
			certs, _ := keymanagementprovider.GetCertificatesFromMap(context.Background(), kmpName)
			if len(certs) != tt.expectedKMPCount {
				t.Fatalf("Cert map expected size %v, actual %v", tt.expectedKMPCount, len(certs))
			}
		})
	}
}

func resetKMP() {
	keymanagementprovider.DeleteCertificatesFromMap(storeName)
}
