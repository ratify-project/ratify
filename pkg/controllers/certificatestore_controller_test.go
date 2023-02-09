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

package controllers

import (
	"testing"

	configv1alpha1 "github.com/deislabs/ratify/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetCertStoreConfig_ValidConfig(t *testing.T) {
	var parametersString = "{\"certificates\":\"array:\\n  - |\\n    certificateName: wabbit-networks-io\\n    certificateVersion: 97a1545d893344079ce57699c8810590\\n\",\"clientID\":\"sampleClientID\",\"keyvaultName\":\"sampleKeyVault\",\"tenantID\":\"sampleTenantID\"}"
	var certStoreParameters = []byte(parametersString)

	spec := configv1alpha1.CertificateStoreSpec{
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

	spec := configv1alpha1.CertificateStoreSpec{
		Provider: "azurekeyvault",
		Parameters: runtime.RawExtension{
			Raw: certStoreParameters,
		},
	}

	_, err := getCertStoreConfig(spec)
	if err == nil {
		t.Fatalf("Expected error")
	}

	expectedError := "Received empty parameters"
	if err.Error() != expectedError {
		t.Fatalf("Unexpected error, expected %+v, got %+v", expectedError, err.Error())
	}
}
