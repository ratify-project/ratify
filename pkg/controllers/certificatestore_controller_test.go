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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/pkg/certificateprovider"
	"github.com/deislabs/ratify/pkg/certificateprovider/inline"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	certStr  = "-----BEGIN CERTIFICATE-----\nMIIDsDCCApigAwIBAgIQMdNmNTKwQ9aOe6iuMRokDzANBgkqhkiG9w0BAQsFADBa\nMQswCQYDVQQGEwJVUzELMAkGA1UECBMCV0ExEDAOBgNVBAcTB1NlYXR0bGUxDzAN\nBgNVBAoTBk5vdGFyeTEbMBkGA1UEAxMSd2FiYml0LW5ldHdvcmtzLmlvMB4XDTIy\nMTIxNDIxNTAzMVoXDTIzMTIxNDIyMDAzMVowWjELMAkGA1UEBhMCVVMxCzAJBgNV\nBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMQ8wDQYDVQQKEwZOb3RhcnkxGzAZBgNV\nBAMTEndhYmJpdC1uZXR3b3Jrcy5pbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC\nAQoCggEBAOP6AHCFz41kRqsAiv6guFtQVsrzMgzoCX7o9NtQ57rr8BESP1LTGRAO\n4bjyP0i+at5uwIm4tdz0gW+g0P+f9bmfiScYgOFuxAJxLkMkBWPN3dJ9ulP/OGgB\n6mSCsEGreB3uaGc5rMbWCRaux65bMPjEzx5ex0qRSsn+fFMTwINPQUJpXSvi/W2/\n1umEWE1x59x0vlkP2dN7CXtB5/Bh01QNNbMdKU9saYn0kaBrCYZLwr6AxFRzLqLM\nQggy/6bOp/+cTTVqTiChMcdyIX52GRr2lChRsB34dDPYxDeKSI5LoRy07bveLjex\n4wm9+vx/WOSS5z0QPvE/v8avuIkMXR0CAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeA\nMAkGA1UdEwQCMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAUwVvE\nvqQPxnE6j6pfX6jpSyv2dOAwHQYDVR0OBBYEFMFbxL6kD8ZxOo+qX1+o6Usr9nTg\nMA0GCSqGSIb3DQEBCwUAA4IBAQDE61FLbagvlCcXf0zcv+mUQ+0HvDVs7ofQe3Yw\naz7gAwxgTspr+jIFQWnPOOBupsyx/jucoz78ndbc5DGWPs2Qz/pIEGnLto2W/PYy\nas/9n8xHxembS4n/Mxxp60PF6ladi/nJAtDJds67sBeqLOfJzh6jV2uQvW7PXe1P\nOMSUHbBn8AfArZ/9njusiLs75+XcAgpnBFqKVv2Vd/INp2YQpVzusuiodeM8A9Qt\n/5xykjdCJw3ceZxD7dSkHgchKZPINFBYHt/EkN/d8mXFOKjGXZyntp4PO6PJ2HYN\nhMMDwdNu4mBmlMTdZMPEpIZIeW7G0P9KpCuvvD7po7NxdBgI\n-----END CERTIFICATE-----\n"
	certStr2 = "-----BEGIN CERTIFICATE-----\nMIIDsDCCApigAwIBAgIQFJMQeqR8TRuHqNu+x0MuEDANBgkqhkiG9w0BAQsFADBa\nMQswCQYDVQQGEwJVUzELMAkGA1UECBMCV0ExEDAOBgNVBAcTB1NlYXR0bGUxDzAN\nBgNVBAoTBk5vdGFyeTEbMBkGA1UEAxMSd2FiYml0LW5ldHdvcmtzLmlvMB4XDTIz\nMDExMTE5MjAxMloXDTI0MDExMTE5MzAxMlowWjELMAkGA1UEBhMCVVMxCzAJBgNV\nBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMQ8wDQYDVQQKEwZOb3RhcnkxGzAZBgNV\nBAMTEndhYmJpdC1uZXR3b3Jrcy5pbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC\nAQoCggEBAMh7F6sZyeiQRva83SvQu0PbsyD4zkEeWAyj03n1dx91FEeEXItCr+Y1\nghQKgdBOY/wJQmSq/We1e+17NoNICrzy2Y1sOVMYR5sx8H/UxO3q8oS7bxctFy+e\nHs4BxlHIqeIiWnz9bFAJFqV6BkJDVjp9k5QfHlkqH08WBvm/D8YTpWzvEPn+71ZG\nN1RKqFUeeM949oGGnC63vVMRRYIx2LoJliNZXdj9qoOHZksDrX2jkgPykkOYcmfo\n9CH9v0JNX+0t0Enp0ruUFK1pSZW+TicI22GvENYHGZNZ0m+6oD5ePRZoYhWyAzgZ\nndHO5bYh3yC7DMc6ssOEJeNN0I2+iLUCAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeA\nMAkGA1UdEwQCMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAUYhhf\nPFgAqU8PF3ClvfKs67HmpWwwHQYDVR0OBBYEFGIYXzxYAKlPDxdwpb3yrOux5qVs\nMA0GCSqGSIb3DQEBCwUAA4IBAQCXu1w+6s2RO2/KPmC+29m9EjbDReI4bGlDGgiv\nwk1fmvPvDrqL4Ebpcrb1nstNlsxpKYQP+3Vi8gPiqNQ7JvPStd1NBu+ViCXdvOe5\nCtN7tBFTCBgdgXNZ9bvIM2dS+xW/ZAJdyHbV9Hn77+rs/uCDHtbaQMJ3N9LGW8GR\nGY+uYylrrCrjb9fzotMaONnF9c1GKiANskc9371wbaninpxcwMNA5j027XzfnMEW\nm807wjlNV3Kuf4fdDpzBLe940iplfTlQMylWMqgANpEw4EqHCrBJPQAHfQEpQlo+\n9H72lrqOiYNNwApfB9P+UqMDi1B7T2yzfvXcqQ75FpSRIxzK\n-----END CERTIFICATE-----\n"
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

func TestUpdateCertificatesMap(t *testing.T) {
	kv1Cert := getCert(certStr)
	kv2Cert := getCert(certStr2)
	certificates := []*x509.Certificate{kv1Cert}
	certificatesMap["certs-ca"] = []*x509.Certificate{kv2Cert}

	// test with empty map
	updateCertificatesMap("certs-sa", certificates)

	if len(certificatesMap) == 0 {
		t.Fatalf("Properties should not be empty")
	}

	// test with non-empty map
	updateCertificatesMap("certs-ca", certificates)

	if len(certificatesMap) == 0 {
		t.Fatalf("Properties should not be empty")
	}
}

// convert string to a x509 certificate
func getCert(certString string) *x509.Certificate {
	block, _ := pem.Decode([]byte(certString))
	if block == nil {
		panic("failed to parse certificate PEM")
	}

	test, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}

	return test
}
