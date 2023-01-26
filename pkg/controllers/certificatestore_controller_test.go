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
	"github.com/deislabs/ratify/pkg/certificateprovider/azurekeyvault/types"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestByteToCerts_ByteArrayToCertificates(t *testing.T) {
	certString1 := "-----BEGIN CERTIFICATE-----\nMIIDsDCCApigAwIBAgIQMdNmNTKwQ9aOe6iuMRokDzANBgkqhkiG9w0BAQsFADBa\nMQswCQYDVQQGEwJVUzELMAkGA1UECBMCV0ExEDAOBgNVBAcTB1NlYXR0bGUxDzAN\nBgNVBAoTBk5vdGFyeTEbMBkGA1UEAxMSd2FiYml0LW5ldHdvcmtzLmlvMB4XDTIy\nMTIxNDIxNTAzMVoXDTIzMTIxNDIyMDAzMVowWjELMAkGA1UEBhMCVVMxCzAJBgNV\nBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMQ8wDQYDVQQKEwZOb3RhcnkxGzAZBgNV\nBAMTEndhYmJpdC1uZXR3b3Jrcy5pbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC\nAQoCggEBAOP6AHCFz41kRqsAiv6guFtQVsrzMgzoCX7o9NtQ57rr8BESP1LTGRAO\n4bjyP0i+at5uwIm4tdz0gW+g0P+f9bmfiScYgOFuxAJxLkMkBWPN3dJ9ulP/OGgB\n6mSCsEGreB3uaGc5rMbWCRaux65bMPjEzx5ex0qRSsn+fFMTwINPQUJpXSvi/W2/\n1umEWE1x59x0vlkP2dN7CXtB5/Bh01QNNbMdKU9saYn0kaBrCYZLwr6AxFRzLqLM\nQggy/6bOp/+cTTVqTiChMcdyIX52GRr2lChRsB34dDPYxDeKSI5LoRy07bveLjex\n4wm9+vx/WOSS5z0QPvE/v8avuIkMXR0CAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeA\nMAkGA1UdEwQCMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAUwVvE\nvqQPxnE6j6pfX6jpSyv2dOAwHQYDVR0OBBYEFMFbxL6kD8ZxOo+qX1+o6Usr9nTg\nMA0GCSqGSIb3DQEBCwUAA4IBAQDE61FLbagvlCcXf0zcv+mUQ+0HvDVs7ofQe3Yw\naz7gAwxgTspr+jIFQWnPOOBupsyx/jucoz78ndbc5DGWPs2Qz/pIEGnLto2W/PYy\nas/9n8xHxembS4n/Mxxp60PF6ladi/nJAtDJds67sBeqLOfJzh6jV2uQvW7PXe1P\nOMSUHbBn8AfArZ/9njusiLs75+XcAgpnBFqKVv2Vd/INp2YQpVzusuiodeM8A9Qt\n/5xykjdCJw3ceZxD7dSkHgchKZPINFBYHt/EkN/d8mXFOKjGXZyntp4PO6PJ2HYN\nhMMDwdNu4mBmlMTdZMPEpIZIeW7G0P9KpCuvvD7po7NxdBgI\n-----END CERTIFICATE-----\n"
	certString2 := "-----BEGIN CERTIFICATE-----\nMIIDsDCCApigAwIBAgIQFJMQeqR8TRuHqNu+x0MuEDANBgkqhkiG9w0BAQsFADBa\nMQswCQYDVQQGEwJVUzELMAkGA1UECBMCV0ExEDAOBgNVBAcTB1NlYXR0bGUxDzAN\nBgNVBAoTBk5vdGFyeTEbMBkGA1UEAxMSd2FiYml0LW5ldHdvcmtzLmlvMB4XDTIz\nMDExMTE5MjAxMloXDTI0MDExMTE5MzAxMlowWjELMAkGA1UEBhMCVVMxCzAJBgNV\nBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMQ8wDQYDVQQKEwZOb3RhcnkxGzAZBgNV\nBAMTEndhYmJpdC1uZXR3b3Jrcy5pbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC\nAQoCggEBAMh7F6sZyeiQRva83SvQu0PbsyD4zkEeWAyj03n1dx91FEeEXItCr+Y1\nghQKgdBOY/wJQmSq/We1e+17NoNICrzy2Y1sOVMYR5sx8H/UxO3q8oS7bxctFy+e\nHs4BxlHIqeIiWnz9bFAJFqV6BkJDVjp9k5QfHlkqH08WBvm/D8YTpWzvEPn+71ZG\nN1RKqFUeeM949oGGnC63vVMRRYIx2LoJliNZXdj9qoOHZksDrX2jkgPykkOYcmfo\n9CH9v0JNX+0t0Enp0ruUFK1pSZW+TicI22GvENYHGZNZ0m+6oD5ePRZoYhWyAzgZ\nndHO5bYh3yC7DMc6ssOEJeNN0I2+iLUCAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeA\nMAkGA1UdEwQCMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAUYhhf\nPFgAqU8PF3ClvfKs67HmpWwwHQYDVR0OBBYEFGIYXzxYAKlPDxdwpb3yrOux5qVs\nMA0GCSqGSIb3DQEBCwUAA4IBAQCXu1w+6s2RO2/KPmC+29m9EjbDReI4bGlDGgiv\nwk1fmvPvDrqL4Ebpcrb1nstNlsxpKYQP+3Vi8gPiqNQ7JvPStd1NBu+ViCXdvOe5\nCtN7tBFTCBgdgXNZ9bvIM2dS+xW/ZAJdyHbV9Hn77+rs/uCDHtbaQMJ3N9LGW8GR\nGY+uYylrrCrjb9fzotMaONnF9c1GKiANskc9371wbaninpxcwMNA5j027XzfnMEW\nm807wjlNV3Kuf4fdDpzBLe940iplfTlQMylWMqgANpEw4EqHCrBJPQAHfQEpQlo+\n9H72lrqOiYNNwApfB9P+UqMDi1B7T2yzfvXcqQ75FpSRIxzK\n-----END CERTIFICATE-----\n"

	c1 := []byte(certString1)
	c2 := []byte(certString2)

	certificates := []types.Certificate{}
	certificates = append(certificates, types.Certificate{Content: c1})
	certificates = append(certificates, types.Certificate{Content: c2})

	r, err := byteToCerts(certificates)
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedLen := 2
	if len(r) != expectedLen {
		t.Fatalf("unexpected count of certificate, expected %+v, got %+v", expectedLen, len(r))
	}

	cert1 := r[0]
	serialNumber1 := cert1.SerialNumber.String()

	expectedserialNumber1 := "66229819451171253920043613209346319375"
	if serialNumber1 != expectedserialNumber1 {
		t.Fatalf("unexpected certificate, expected %+v, got %+v", expectedserialNumber1, serialNumber1)
	}

	cert2 := r[1]
	serialNumber2 := cert2.SerialNumber.String()

	expectedserialNumber2 := "27348161789198234828835474579392769552"
	if serialNumber2 != expectedserialNumber2 {
		t.Fatalf("unexpected certificate, expected %+v, got %+v", expectedserialNumber2, serialNumber2)
	}
}

func TestByteToCerts_FailedToDecode(t *testing.T) {
	certString := "foo"

	c1 := []byte(certString)

	certificates := []types.Certificate{}
	certificates = append(certificates, types.Certificate{Content: c1})

	_, err := byteToCerts(certificates)
	if err == nil {
		t.Fatalf("bytesToCerts should return an error")
	}

	expectedError := "Failed to decode certificate"
	if err.Error() != expectedError {
		t.Fatalf("unexpected error, expected %+v, got %+v", expectedError, err.Error())
	}
}

func TestByteToCerts_FailedX509ParseError(t *testing.T) {
	certString := "-----BEGIN CERTIFICATE-----\nMIIDsDCCApigAwIBAgIQFJMQeqR8TRuHqNu+x0MuEDANBgkqhkiG9w0BAQsFABAD\nMQswCQYDVQQGEwJVUzELMAkGA1UECBMCV0ExEDAOBgNVBAcTB1NlYXR0bGUxDzAN\nBgNVBAoTBk5vdGFyeTEbMBkGA1UEAxMSd2FiYml0LW5ldHdvcmtzLmlvMB4XDTIz\nMDExMTE5MjAxMloXDTI0MDExMTE5MzAxMlowWjELMAkGA1UEBhMCVVMxCzAJBgNV\nBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMQ8wDQYDVQQKEwZOb3RhcnkxGzAZBgNV\nBAMTEndhYmJpdC1uZXR3b3Jrcy5pbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC\nAQoCggEBAMh7F6sZyeiQRva83SvQu0PbsyD4zkEeWAyj03n1dx91FEeEXItCr+Y1\nghQKgdBOY/wJQmSq/We1e+17NoNICrzy2Y1sOVMYR5sx8H/UxO3q8oS7bxctFy+e\nHs4BxlHIqeIiWnz9bFAJFqV6BkJDVjp9k5QfHlkqH08WBvm/D8YTpWzvEPn+71ZG\nN1RKqFUeeM949oGGnC63vVMRRYIx2LoJliNZXdj9qoOHZksDrX2jkgPykkOYcmfo\n9CH9v0JNX+0t0Enp0ruUFK1pSZW+TicI22GvENYHGZNZ0m+6oD5ePRZoYhWyAzgZ\nndHO5bYh3yC7DMc6ssOEJeNN0I2+iLUCAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeA\nMAkGA1UdEwQCMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAUYhhf\nPFgAqU8PF3ClvfKs67HmpWwwHQYDVR0OBBYEFGIYXzxYAKlPDxdwpb3yrOux5qVs\nMA0GCSqGSIb3DQEBCwUAA4IBAQCXu1w+6s2RO2/KPmC+29m9EjbDReI4bGlDGgiv\nwk1fmvPvDrqL4Ebpcrb1nstNlsxpKYQP+3Vi8gPiqNQ7JvPStd1NBu+ViCXdvOe5\nCtN7tBFTCBgdgXNZ9bvIM2dS+xW/ZAJdyHbV9Hn77+rs/uCDHtbaQMJ3N9LGW8GR\nGY+uYylrrCrjb9fzotMaONnF9c1GKiANskc9371wbaninpxcwMNA5j027XzfnMEW\nm807wjlNV3Kuf4fdDpzBLe940iplfTlQMylWMqgANpEw4EqHCrBJPQAHfQEpQlo+\n9H72lrqOiYNNwApfB9P+UqMDi1B7T2yzfvXcqQ75FpSRIBAD\n-----END CERTIFICATE-----\n"

	c1 := []byte(certString)

	certificates := []types.Certificate{}
	certificates = append(certificates, types.Certificate{Content: c1})

	_, err := byteToCerts(certificates)
	if err == nil {
		t.Fatalf("bytesToCerts should return an error")
	}

	expectedError := "x509: malformed issuer"
	if err.Error() != expectedError {
		t.Fatalf("unexpected error, expected %+v, got %+v", expectedError, err.Error())
	}
}

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
