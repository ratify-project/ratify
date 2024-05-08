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

package namespaceresource

import (
	"fmt"
	"testing"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/pkg/certificateprovider"
	"github.com/deislabs/ratify/pkg/certificateprovider/inline"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetCertStoreConfig_ValidConfig(t *testing.T) {
	var parametersString = "{\"certificates\":\"array:\\n  - |\\n    certificateName: wabbit-networks-io\\n    certificateVersion: 97a1545d893344079ce57699c8810590\\n\",\"clientID\":\"sampleClientID\",\"keyvaultName\":\"sampleKeyVault\",\"tenantID\":\"sampleTenantID\"}"
	var certStoreParameters = []byte(parametersString)

	spec := configv1beta1.CertificateStoreSpec{
		Provider: "azurekeyvault",
		Parameters: runtime.RawExtension{
			Raw: certStoreParameters,
		},
	}

	result, err := getCertStoreConfig(spec)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if len(result) != 4 ||
		result["clientID"] != "sampleClientID" ||
		result["tenantID"] != "sampleTenantID" ||
		result["keyvaultName"] != "sampleKeyVault" ||
		result["certificates"] != "array:\n  - |\n    certificateName: wabbit-networks-io\n    certificateVersion: 97a1545d893344079ce57699c8810590\n" {
		t.Fatalf("unexpected value")
	}
}

func TestGetCertStoreConfig_EmptyStringError(t *testing.T) {
	var parametersString = ""
	var certStoreParameters = []byte(parametersString)

	spec := configv1beta1.CertificateStoreSpec{
		Provider: "azurekeyvault",
		Parameters: runtime.RawExtension{
			Raw: certStoreParameters,
		},
	}

	_, err := getCertStoreConfig(spec)
	if err == nil {
		t.Fatalf("Expected error")
	}

	expectedError := "received empty parameters"
	if err.Error() != expectedError {
		t.Fatalf("Unexpected error, expected %+v, got %+v", expectedError, err.Error())
	}
}

func TestUpdateErrorStatus(t *testing.T) {
	var parametersString = "{\"certs\":{\"name\":\"certName\"}}"
	var certStatus = []byte(parametersString)

	status := configv1beta1.CertificateStoreStatus{
		IsSuccess: true,
		Properties: runtime.RawExtension{
			Raw: certStatus,
		},
	}
	certStore := configv1beta1.CertificateStore{
		Status: status,
	}
	expectedErr := "it's a long error from unit test"
	lastFetchedTime := metav1.Now()
	updateErrorStatus(&certStore, expectedErr, &lastFetchedTime)

	if certStore.Status.IsSuccess != false {
		t.Fatalf("Unexpected error, expected isSuccess to be false , actual %+v", certStore.Status.IsSuccess)
	}

	if certStore.Status.Error != expectedErr {
		t.Fatalf("Unexpected error string, expected %+v, got %+v", expectedErr, certStore.Status.Error)
	}
	expectedBriedErr := fmt.Sprintf("%s...", expectedErr[:30])
	if certStore.Status.BriefError != expectedBriedErr {
		t.Fatalf("Unexpected error string, expected %+v, got %+v", expectedBriedErr, certStore.Status.Error)
	}

	//make sure properties of last cached cert was not overridden
	if len(certStore.Status.Properties.Raw) == 0 {
		t.Fatalf("Unexpected properties,  expected %+v, got %+v", parametersString, string(certStore.Status.Properties.Raw))
	}
}

func TestUpdateSuccessStatus(t *testing.T) {
	certStatus := certificateprovider.CertificatesStatus{}
	properties := map[string]string{}
	properties["CertName"] = "wabbit"
	properties["Version"] = "ABC"

	certStatus["Certificates"] = properties

	lastFetchedTime := metav1.Now()

	status := configv1beta1.CertificateStoreStatus{
		IsSuccess: false,
		Error:     "error from last operation",
	}
	certStore := configv1beta1.CertificateStore{
		Status: status,
	}

	updateSuccessStatus(&certStore, &lastFetchedTime, certStatus)

	if certStore.Status.IsSuccess != true {
		t.Fatalf("Expected isSuccess to be true , actual %+v", certStore.Status.IsSuccess)
	}

	if certStore.Status.Error != "" {
		t.Fatalf("Unexpected error string, actual %+v", certStore.Status.Error)
	}

	//make sure properties of last cached cert was updated
	if len(certStore.Status.Properties.Raw) == 0 {
		t.Fatalf("Properties should not be empty")
	}
}

func TestUpdateSuccessStatus_emptyProperties(t *testing.T) {
	lastFetchedTime := metav1.Now()
	status := configv1beta1.CertificateStoreStatus{
		IsSuccess: false,
		Error:     "error from last operation",
	}
	certStore := configv1beta1.CertificateStore{
		Status: status,
	}

	updateSuccessStatus(&certStore, &lastFetchedTime, nil)

	if certStore.Status.IsSuccess != true {
		t.Fatalf("Expected isSuccess to be true , actual %+v", certStore.Status.IsSuccess)
	}

	if certStore.Status.Error != "" {
		t.Fatalf("Unexpected error string, actual %+v", certStore.Status.Error)
	}

	//make sure properties of last cached cert was updated
	if len(certStore.Status.Properties.Raw) != 0 {
		t.Fatalf("Properties should be empty")
	}
}

func TestGetCertificateProvider(t *testing.T) {
	providers := map[string]certificateprovider.CertificateProvider{}
	providers["inline"] = inline.Create()
	result, _ := getCertificateProvider(providers, "inline")

	if result == nil {
		t.Fatalf("Expected getCertificateProvider() to return inline cert provider")
	}

	_, err := getCertificateProvider(providers, "azurekv")
	if err == nil {
		t.Fatalf("Getting unregistered provider should returns an error")
	}
}
