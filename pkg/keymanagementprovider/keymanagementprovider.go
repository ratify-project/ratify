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
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"sync"

	"github.com/deislabs/ratify/errors"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
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
	// Returns an array of keys and the provider specific key attributes
	GetKeys(ctx context.Context) (map[KMPMapKey]crypto.PublicKey, KeyManagementProviderStatus, error)
}

// static concurrency-safe map to store certificates fetched from key management provider
// layout:
//
//	map["<namespace>/<name>"] = map[KMPMapKey][]*x509.Certificate
//	where KMPMapKey is dimensioned by the name and version of the certificate.
//	Array of x509 Certificates for certificate chain scenarios
var certificatesMap sync.Map

// static concurrency-safe map to store keys fetched from key management provider
// layout:
//
//	map["<namespace>/<name>"] = map[KMPMapKey]PublicKey
//	where KMPMapKey is dimensioned by the name and version of the public key.
var keyMap sync.Map

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

// DecodeKey takes in a PEM encoded byte array and returns a public key
// PEM encoded byte array is expected to be a single public key. If multiple
// are provided, the first one is returned
func DecodeKey(value []byte) (crypto.PublicKey, error) {
	pk, err := cryptoutils.UnmarshalPEMToPublicKey(value)
	if err != nil {
		return nil, errors.ErrorCodeKeyInvalid.WithComponentType(errors.KeyManagementProvider).WithDetail("error parsing public key").WithError(err)
	}
	return pk, nil
}

// SetCertificatesInMap sets the certificates in the map
// it is concurrency-safe
func SetCertificatesInMap(resource string, certs map[KMPMapKey][]*x509.Certificate) {
	certificatesMap.Store(resource, certs)
}

// GetCertificatesFromMap gets the certificates from the map and returns an empty map of certificate arrays if not found
func GetCertificatesFromMap(ctx context.Context, resource string) map[KMPMapKey][]*x509.Certificate {
	if !isCompatibleNamespace(ctx, resource) {
		return map[KMPMapKey][]*x509.Certificate{}
	}
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
	var items []*x509.Certificate
	for _, val := range certMap {
		items = append(items, val...)
	}
	return items
}

// FlattenKMPMapKeys flattens the map of keys fetched for a single key management provider resource and returns a single array
func FlattenKMPMapKeys(keyMap map[KMPMapKey]crypto.PublicKey) []crypto.PublicKey {
	items := []crypto.PublicKey{}
	for _, val := range keyMap {
		items = append(items, val)
	}
	return items
}

// SetKeysInMap sets the keys in the map
func SetKeysInMap(resource string, keys map[KMPMapKey]crypto.PublicKey) {
	keyMap.Store(resource, keys)
}

// GetKeysFromMap gets the keys from the map and returns an empty map of keys if not found
func GetKeysFromMap(ctx context.Context, resource string) map[KMPMapKey]crypto.PublicKey {
	if !isCompatibleNamespace(ctx, resource) {
		return map[KMPMapKey]crypto.PublicKey{}
	}
	keys, ok := keyMap.Load(resource)
	if !ok {
		return map[KMPMapKey]crypto.PublicKey{}
	}
	return keys.(map[KMPMapKey]crypto.PublicKey)
}

// DeleteKeysFromMap deletes the keys from the map
func DeleteKeysFromMap(resource string) {
	keyMap.Delete(resource)
}

// Namespaced verifiers could access both cluster-scoped and namespaced certStores.
// But cluster-wide verifiers could only access cluster-scoped certStores.
// TODO: current implementation always returns true. Check the namespace once we support multi-tenancy.
func isCompatibleNamespace(_ context.Context, _ string) bool {
	return true
}
