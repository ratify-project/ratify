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

package utils

import (
	"context"
	"crypto/x509"
	"testing"

	"github.com/deislabs/ratify/pkg/certificateprovider"
	"github.com/deislabs/ratify/pkg/certificateprovider/inline"
	"github.com/deislabs/ratify/pkg/controllers"

	ctxUtils "github.com/deislabs/ratify/internal/context"
	cs "github.com/deislabs/ratify/pkg/customresources/certificatestores"
)

func TestGetCertStoreConfig_ValidConfig(t *testing.T) {
	var parametersString = "{\"certificates\":\"array:\\n  - |\\n    certificateName: wabbit-networks-io\\n    certificateVersion: 97a1545d893344079ce57699c8810590\\n\",\"clientID\":\"sampleClientID\",\"keyvaultName\":\"sampleKeyVault\",\"tenantID\":\"sampleTenantID\"}"
	var certStoreParameters = []byte(parametersString)

	result, err := GetCertStoreConfig(certStoreParameters)
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

	_, err := GetCertStoreConfig(certStoreParameters)
	if err == nil {
		t.Fatalf("Expected error")
	}

	expectedError := "received empty parameters"
	if err.Error() != expectedError {
		t.Fatalf("Unexpected error, expected %+v, got %+v", expectedError, err.Error())
	}
}

func TestGetCertStoreConfig_InvalidParams(t *testing.T) {
	var parametersString = "test\n"
	var certStoreParameters = []byte(parametersString)

	_, err := GetCertStoreConfig(certStoreParameters)
	if err == nil {
		t.Fatalf("Expected error")
	}

	expectedError := "invalid character 'e' in literal true (expecting 'r')"
	if err.Error() != expectedError {
		t.Fatalf("Unexpected error, expected %+v, got %+v", expectedError, err.Error())
	}
}

func TestGetCertificateProvider(t *testing.T) {
	providers := map[string]certificateprovider.CertificateProvider{}
	providers["inline"] = inline.Create()
	result, _ := GetCertificateProvider(providers, "inline")

	if result == nil {
		t.Fatalf("Expected GetCertificateProvider() to return inline cert provider")
	}

	_, err := GetCertificateProvider(providers, "azurekv")
	if err == nil {
		t.Fatalf("Getting unregistered provider should returns an error")
	}
}

func TestGetCertificatesMap(t *testing.T) {
	controllers.CertificatesMap = cs.NewActiveCertStores()
	controllers.CertificatesMap.AddStore("default", "default/certStore", []*x509.Certificate{})
	ctx := ctxUtils.SetContextWithNamespace(context.Background(), "default")

	if certs := GetCertificatesMap(ctx); len(certs) != 1 {
		t.Fatalf("Expected 1 certificate store, got %d", len(certs))
	}
}
