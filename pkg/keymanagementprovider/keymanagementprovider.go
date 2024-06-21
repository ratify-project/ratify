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
	"fmt"
	"strings"
	"sync"

	"github.com/ratify-project/ratify/errors"
	"github.com/ratify-project/ratify/internal/constants"
	ctxUtils "github.com/ratify-project/ratify/internal/context"
	vu "github.com/ratify-project/ratify/pkg/verifier/utils"
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

type PublicKey struct {
	Key          crypto.PublicKey
	ProviderType string
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
//		map["<namespace>/<name>"] = map[KMPMapKey]PublicKey
//		where KMPMapKey is dimensioned by the name and version of the public key
//	 where PublicKey is a struct containing the public key and the provider type
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

	if len(certs) == 0 {
		return nil, errors.ErrorCodeCertInvalid.WithComponentType(errors.CertProvider).WithDetail("no certificates found in the pem block")
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

// GetCertificatesFromMap gets the certificates from the map and returns an empty map of certificate arrays if not found or an error happened.
func GetCertificatesFromMap(ctx context.Context, resource string) (map[KMPMapKey][]*x509.Certificate, error) {
	if !hasAccessToProvider(ctx, resource) {
		return map[KMPMapKey][]*x509.Certificate{}, fmt.Errorf("namespace: [%s] does not have access to key management provider: %s", ctxUtils.GetNamespace(ctx), resource)
	}
	if certs, ok := certificatesMap.Load(resource); ok {
		return certs.(map[KMPMapKey][]*x509.Certificate), nil
	}
	return map[KMPMapKey][]*x509.Certificate{}, fmt.Errorf("failed to access non-existent key management provider: %s", resource)
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

// SetKeysInMap sets the keys in the map
func SetKeysInMap(resource string, providerType string, keys map[KMPMapKey]crypto.PublicKey) {
	typedMap := make(map[KMPMapKey]PublicKey)
	for key, value := range keys {
		typedMap[key] = PublicKey{Key: value, ProviderType: providerType}
	}
	keyMap.Store(resource, typedMap)
}

// GetKeysFromMap gets the keys from the map and returns an empty map if not found or an error happened.
func GetKeysFromMap(ctx context.Context, resource string) (map[KMPMapKey]PublicKey, error) {
	// A cluster-wide operation can cluster-wide provider
	// A namespaced operation can only fetch the provider in the same namespace or cluster-wide provider.
	if !hasAccessToProvider(ctx, resource) {
		return map[KMPMapKey]PublicKey{}, fmt.Errorf("namespace: [%s] does not have access to key management provider: %s", ctxUtils.GetNamespace(ctx), resource)
	}
	if keys, ok := keyMap.Load(resource); ok {
		return keys.(map[KMPMapKey]PublicKey), nil
	}
	return map[KMPMapKey]PublicKey{}, fmt.Errorf("failed to access non-existent key management provider: %s", resource)
}

// DeleteKeysFromMap deletes the keys from the map
func DeleteKeysFromMap(resource string) {
	keyMap.Delete(resource)
}

// A namespaced verification request could access KMP in the same namespace or cluster-wide KMP.
// A cluster-wide (context namespace is "") verification request could only access cluster-wide KMP.
func hasAccessToProvider(ctx context.Context, provider string) bool {
	namespace := ctxUtils.GetNamespace(ctx)
	if namespace == constants.EmptyNamespace {
		return !vu.IsNamespacedNamed(provider)
	}
	return strings.HasPrefix(provider, namespace+constants.NamespaceSeperator) || !vu.IsNamespacedNamed(provider)
}
