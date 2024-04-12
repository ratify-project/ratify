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
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"reflect"
	"testing"
	"time"
)

func TestCreateTestCert(t *testing.T) {
	cert := CreateTestCert()

	if cert == nil {
		t.Fatal("Expected a non-nil certificate, got nil")
	}

	// Check certificate fields
	expectedSerialNumber := big.NewInt(1)
	if cert.SerialNumber.Cmp(expectedSerialNumber) != 0 {
		t.Fatalf("Expected serial number %v, got %v", expectedSerialNumber, cert.SerialNumber)
	}

	expectedOrganization := []string{"My Organization"}
	if !reflect.DeepEqual(cert.Subject.Organization, expectedOrganization) {
		t.Fatalf("Expected organization %v, got %v", expectedOrganization, cert.Subject.Organization)
	}

	expectedCountry := []string{"Country"}
	if !reflect.DeepEqual(cert.Subject.Country, expectedCountry) {
		t.Fatalf("Expected country %v, got %v", expectedCountry, cert.Subject.Country)
	}

	expectedProvince := []string{"Province"}
	if !reflect.DeepEqual(cert.Subject.Province, expectedProvince) {
		t.Fatalf("Expected province %v, got %v", expectedProvince, cert.Subject.Province)
	}

	// Check NotBefore and NotAfter dates
	now := time.Now()
	if cert.NotBefore.After(now) {
		t.Fatalf("NotBefore is after current time: %v", cert.NotBefore)
	}

	// Check KeyUsage
	expectedKeyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	if cert.KeyUsage != expectedKeyUsage {
		t.Fatalf("Expected KeyUsage %v, got %v", expectedKeyUsage, cert.KeyUsage)
	}

	// Check ExtKeyUsage
	expectedExtKeyUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	if !reflect.DeepEqual(cert.ExtKeyUsage, expectedExtKeyUsage) {
		t.Fatalf("Expected ExtKeyUsage %v, got %v", expectedExtKeyUsage, cert.ExtKeyUsage)
	}

	// Check BasicConstraintsValid
	if !cert.BasicConstraintsValid {
		t.Fatal("Expected BasicConstraintsValid to be true, got false")
	}

	// Check PublicKey
	if cert.PublicKey.(*rsa.PublicKey).N.Cmp(cert.PublicKey.(*rsa.PublicKey).N) != 0 {
		t.Fatal("Public key mismatch")
	}
}

func TestCreateTestPublicKey(t *testing.T) {
	publicKey := CreateTestPublicKey()

	if publicKey == nil {
		t.Fatal("Expected a non-nil public key, got nil")
	}

	// Check the type of the public key
	_, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		t.Fatal("Expected *rsa.PublicKey, got", reflect.TypeOf(publicKey))
	}

	// Marshal the public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey.(*rsa.PublicKey))
	if err != nil {
		t.Fatal("Error marshaling public key:", err)
	}

	// Create a PEM block for the public key
	expectedPublicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	// Encode the PEM block
	expectedPublicKeyPEMEncoded := pem.EncodeToMemory(expectedPublicKeyPEM)
	if expectedPublicKeyPEMEncoded == nil {
		t.Fatal("Error encoding PEM block")
	}

	// Decode the public key from the function's output
	block, _ := pem.Decode(expectedPublicKeyPEMEncoded)
	if block == nil {
		t.Fatal("Error decoding PEM block")
	}

	// Parse the public key
	expectedPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatal("Error parsing public key:", err)
	}

	// Check if the parsed public key matches the expected public key
	if !reflect.DeepEqual(publicKey, expectedPublicKey) {
		t.Fatal("Parsed public key does not match the expected public key")
	}
}
