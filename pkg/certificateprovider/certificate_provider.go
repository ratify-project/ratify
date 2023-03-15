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

package certificateprovider

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// This is a map containing Cert store configuration including name, tenantID, and cert object information
type CertStoreConfig map[string]string

// CertificateProvider is an interface that defines methods to be implemented by a each certificate provider
type CertificateProvider interface {
	// Returns an array of certificates based on certificate properties defined in attrib map
	GetCertificates(ctx context.Context, attrib map[string]string) ([]*x509.Certificate, error)
}

var certificateProviders = make(map[string]CertificateProvider)

// returns the internal certificate provider map
func GetCertificateProviders() map[string]CertificateProvider {
	return certificateProviders
}

// Register adds the factory to the built in providers map
func Register(name string, provider CertificateProvider) {
	if _, registered := certificateProviders[name]; registered {
		panic(fmt.Sprintf("cert provider named %s already registered", name))
	}

	certificateProviders[name] = provider
}

// Decode PEM-encoded bytes into an x509.Certificate chain.
func DecodeCertificates(value []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	block, rest := pem.Decode(value)
	if block == nil && len(rest) > 0 {
		return nil, fmt.Errorf("failed to decode pem block")
	}

	for block != nil {
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("error parsing x509 certificate: %w", err)
			}
			certs = append(certs, cert)
		}
		block, rest = pem.Decode(rest)
		if block == nil && len(rest) > 0 {
			return nil, fmt.Errorf("failed to decode pem block")
		}
	}

	return certs, nil
}
