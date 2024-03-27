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

package keymanagementprovider

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"sync"

	"github.com/deislabs/ratify/errors"
)

// This is a map of properties for fetched certificates/keys
// The key and values are specific to each provider
//
//nolint:revive
type KeyManagementProviderStatus map[string]interface{}

// KMPMapKey is a key for the map of certificates fetched for a single key management provider resource
type KMPMapKey struct {
	Name    string
	Version string
}

// KeyManagementProvider is an interface that defines methods to be implemented by a each key management provider provider
type KeyManagementProvider interface {
	// Returns an array of certificates and the provider specific cert attributes
	GetCertificates(ctx context.Context) (map[KMPMapKey][]*x509.Certificate, KeyManagementProviderStatus, error)
}

// static concurreny-safe map to store certificates fetched from key management provider
var certificatesMap sync.Map

// DecodeCertificates decodes PEM-encoded bytes into an x509.Certificate chain.
func DecodeCertificates(value []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	block, rest := pem.Decode(value)
	if block == nil && len(rest) > 0 {
		return nil, errors.ErrorCodeCertInvalid.WithComponentType(errors.CertProvider).WithDetail("failed to decode pem block")
	}

	for block != nil {
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, errors.ErrorCodeCertInvalid.WithComponentType(errors.CertProvider).WithDetail("error parsing x509 certificate")
			}
			certs = append(certs, cert)
		}
		block, rest = pem.Decode(rest)
		if block == nil && len(rest) > 0 {
			return nil, errors.ErrorCodeCertInvalid.WithComponentType(errors.CertProvider).WithDetail("failed to decode pem block")
		}
	}

	return certs, nil
}

// SetCertificatesInMap sets the certificates in the map
// it is concurrency-safe
func SetCertificatesInMap(resource string, certs map[KMPMapKey][]*x509.Certificate) {
	certificatesMap.Store(resource, certs)
}

// GetCertificatesFromMap gets the certificates from the map and returns an empty map of certificate arrays if not found
func GetCertificatesFromMap(resource string) map[KMPMapKey][]*x509.Certificate {
	certs, ok := certificatesMap.Load(resource)
	if !ok {
		return map[KMPMapKey][]*x509.Certificate{}
	}
	return certs.(map[KMPMapKey][]*x509.Certificate)
}

// DeleteCertificatesFromMap deletes the certificates from the map
// it is concurrency-safe
func DeleteCertificatesFromMap(resource string) {
	certificatesMap.Delete(resource)
}

// FlattenKMPMap flattens the map of certificates fetched for a single key management provider resource and returns a single array
func FlattenKMPMap(certMap map[KMPMapKey][]*x509.Certificate) []*x509.Certificate {
	var certs []*x509.Certificate
	for _, v := range certMap {
		certs = append(certs, v...)
	}
	return certs
}
