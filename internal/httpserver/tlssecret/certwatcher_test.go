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

package tlssecret

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestLoadingTLSCerts_EmptyPaths(t *testing.T) {
	if _, err := NewWatcher("", "", ""); err == nil {
		t.Fatalf("expected error but got none")
	}
}

func TestLoadingTLSCerts_InvalidCertPaths(t *testing.T) {
	_, err := NewWatcher("", "invalid_cert.pem", "invalid_key.pem")
	if err == nil {
		t.Fatalf("expected error but got none")
	}
}

func TestLoadingTLSCerts_SuccessfullyLoaded(t *testing.T) {
	certPem, keyPem, err := generateRSAKeyCertPair()
	if err != nil {
		t.Fatalf("failed to generate cert/key pair: %v", err)
	}
	certDir, err := os.MkdirTemp("", "certs")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	certFile, keyFile, err := generateAndSaveCertKey(certDir, certPem, keyPem)
	if err != nil {
		t.Fatalf("failed to generate and save cert/key: %v", err)
	}
	defer func() {
		os.Remove(certFile)
		os.Remove(keyFile)
	}()

	watcher, err := NewWatcher("", certFile, keyFile)
	if err != nil {
		t.Fatalf("failed to create TLS secret watcher: %v", err)
	}

	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		t.Fatalf("failed to create X509 key pair: %v", err)
	}
	if !certificatesEqual(cert, *watcher.ratifyServerTLSCert.Load()) {
		t.Fatalf("expected cert to be %v, got %v", cert, *watcher.ratifyServerTLSCert.Load())
	}
}

func TestLoadingTLSCerts_InvalidCACertProvided(t *testing.T) {
	if _, err := NewWatcher("invalidCACert.pem", "invalidCert.pem", "invalidKey.pem"); err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestLoadingTLSCerts_GKCACertProvided(t *testing.T) {
	certPem, keyPem, err := generateRSAKeyCertPair()
	if err != nil {
		t.Fatalf("failed to generate cert/key pair: %v", err)
	}
	caCertPem, caKeyPem, err := generateRSAKeyCertPair()
	if err != nil {
		t.Fatalf("failed to generate CA cert/key pair: %v", err)
	}

	certDir, err := os.MkdirTemp("", "certs")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	certFile, keyFile, err := generateAndSaveCertKey(certDir, certPem, keyPem)
	if err != nil {
		t.Fatalf("failed to generate and save cert/key: %v", err)
	}
	defer func() {
		os.Remove(certFile)
		os.Remove(keyFile)
	}()
	caCertFile, caKeyFile, err := generateAndSaveCertKey(certDir, caCertPem, caKeyPem)
	if err != nil {
		t.Fatalf("failed to generate and save CA cert/key: %v", err)
	}
	defer func() {
		os.Remove(caCertFile)
		os.Remove(caKeyFile)
	}()

	gatekeeperCACertFile := caCertFile
	watcher, err := NewWatcher(gatekeeperCACertFile, certFile, keyFile)
	if err != nil {
		t.Fatalf("failed to create TLS secret watcher: %v", err)
	}

	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		t.Fatalf("failed to create X509 key pair: %v", err)
	}
	if !certificatesEqual(cert, *watcher.ratifyServerTLSCert.Load()) {
		t.Fatalf("expected cert to be %v, got %v", cert, *watcher.ratifyServerTLSCert.Load())
	}

	clientCAs := x509.NewCertPool()
	clientCAs.AppendCertsFromPEM(caCertPem)
	if !clientCAs.Equal(watcher.clientCAs.Load()) {
		t.Fatalf("expected client CA certs to be %v, got %v", clientCAs, watcher.clientCAs.Load())
	}

	err = watcher.Start()
	if err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}
	defer watcher.Stop()
}

func TestLoadingTLSCerts_StartWatching(t *testing.T) {
	certPem, keyPem, err := generateRSAKeyCertPair()
	if err != nil {
		t.Fatalf("failed to generate cert/key pair: %v", err)
	}
	certDir, err := os.MkdirTemp("", "certs")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	certFile, keyFile, err := generateAndSaveCertKey(certDir, certPem, keyPem)
	if err != nil {
		t.Fatalf("failed to generate and save cert/key: %v", err)
	}
	defer func() {
		os.Remove(certFile)
		os.Remove(keyFile)
	}()

	watcher, err := NewWatcher("", certFile, keyFile)
	if err != nil {
		t.Fatalf("failed to create TLS secret watcher: %v", err)
	}

	err = watcher.Start()
	if err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}
	defer watcher.Stop()

	time.Sleep(1 * time.Second)

	newCertPem, newKeyPem, err := generateRSAKeyCertPair()
	if err != nil {
		t.Fatalf("failed to generate new cert/key pair: %v", err)
	}

	err = os.WriteFile(certFile, newCertPem, 0600)
	if err != nil {
		t.Fatalf("failed to write new cert file: %v", err)
	}
	err = os.WriteFile(keyFile, newKeyPem, 0600)
	if err != nil {
		t.Fatalf("failed to write new key file: %v", err)
	}

	time.Sleep(1 * time.Second)

	newCert, _ := tls.X509KeyPair(newCertPem, newKeyPem)

	if !certificatesEqual(newCert, *watcher.ratifyServerTLSCert.Load()) {
		t.Fatalf("expected cert to be %v, got %v", newCert, *watcher.ratifyServerTLSCert.Load())
	}
}

func TestGetConfigForClient(t *testing.T) {
	certPem, keyPem, err := generateRSAKeyCertPair()
	if err != nil {
		t.Fatalf("failed to generate cert/key pair: %v", err)
	}
	caCertPem, caKeyPem, err := generateRSAKeyCertPair()
	if err != nil {
		t.Fatalf("failed to generate CA cert/key pair: %v", err)
	}

	certDir, err := os.MkdirTemp("", "certs")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	certFile, keyFile, err := generateAndSaveCertKey(certDir, certPem, keyPem)
	if err != nil {
		t.Fatalf("failed to generate and save cert/key: %v", err)
	}
	defer func() {
		os.Remove(certFile)
		os.Remove(keyFile)
	}()
	caCertFile, caKeyFile, err := generateAndSaveCertKey(certDir, caCertPem, caKeyPem)
	if err != nil {
		t.Fatalf("failed to generate and save CA cert/key: %v", err)
	}
	defer func() {
		os.Remove(caCertFile)
		os.Remove(caKeyFile)
	}()

	watcher, err := NewWatcher(caCertFile, certFile, keyFile)
	if err != nil {
		t.Fatalf("failed to create TLS secret watcher: %v", err)
	}

	config, err := watcher.GetConfigForClient(nil)
	if err != nil {
		t.Fatalf("failed to get config for client: %v", err)
	}

	if config.MinVersion != tls.VersionTLS13 {
		t.Fatalf("expected min version to be %v, got %v", tls.VersionTLS13, config.MinVersion)
	}
	if config.GetConfigForClient == nil {
		t.Fatal("expected GetConfigForClient to be set")
	}
}

// generateAndSaveCertKey generates a new RSA key/cert pair and saves them to temp files.
// It returns the cert file path, key file path, and any error encountered.
func generateAndSaveCertKey(dir string, certPEM, keyPEM []byte) (certFile string, keyFile string, err error) {
	certTmp, err := os.CreateTemp(dir, "cert.pem")
	if err != nil {
		return "", "", err
	}
	defer certTmp.Close()

	keyTmp, err := os.CreateTemp(dir, "key.pem")
	if err != nil {
		os.Remove(certTmp.Name())
		return "", "", err
	}
	defer keyTmp.Close()

	if _, err := certTmp.Write(certPEM); err != nil {
		os.Remove(certTmp.Name())
		os.Remove(keyTmp.Name())
		return "", "", err
	}
	if _, err := keyTmp.Write(keyPEM); err != nil {
		os.Remove(certTmp.Name())
		os.Remove(keyTmp.Name())
		return "", "", err
	}

	return certTmp.Name(), keyTmp.Name(), nil
}

// generateRSAKeyCertPair generates a new RSA private key and a self-signed certificate.
func generateRSAKeyCertPair() ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return certPEM, keyPEM, nil
}

func certificatesEqual(a, b tls.Certificate) bool {
	if !certChainsEqual(a.Certificate, b.Certificate) {
		return false
	}
	if !privateKeysEqual(a.PrivateKey, b.PrivateKey) {
		return false
	}
	return true
}

func certChainsEqual(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bytes.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

func privateKeysEqual(a, b interface{}) bool {
	ak, aok := a.(*rsa.PrivateKey)
	bk, bok := b.(*rsa.PrivateKey)
	if aok && bok {
		return ak.PublicKey.Equal(&bk.PublicKey) && ak.D.Cmp(bk.D) == 0
	}
	// Add other key types if needed (e.g., ECDSA, Ed25519)
	return reflect.DeepEqual(a, b) // fallback, but unsafe for keys
}
