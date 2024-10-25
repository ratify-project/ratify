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

package crlprovider

import (
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"fmt"

	corecrl "github.com/notaryproject/notation-core-go/revocation/crl"
)

// oidFreshestCRL is the object identifier for the distribution point
// for the delta CRL. (See RFC 5280, Section 5.2.6)
var oidFreshestCRL = asn1.ObjectIdentifier{2, 5, 29, 46}

// maxCRLSize is the maximum size of CRL in bytes
//
// The 32 MiB limit is based on investigation that even the largest CRLs
// are less than 16 MiB. The limit is set to 32 MiB to prevent
const maxCRLSize = 32 * 1024 * 1024 // 32 MiB

// BytesFetcher implements the Fetcher interface to fetch CRL from the given bytes
type BytesFetcher struct {
	// Cache stores fetched CRLs and reuses them until the CRL reaches the
	// NextUpdate time.
	// If Cache is nil, no cache is used.
	Cache corecrl.Cache

	// DiscardCacheError specifies whether to discard any error on cache.
	//
	// ErrCacheMiss is not considered as an failure and will not be returned as
	// an error if DiscardCacheError is false.
	DiscardCacheError bool
}

// NewBytesFetcher creates a new BytesFetcher with the given HTTP client
func NewBytesFetcher() (*BytesFetcher, error) {
	return &BytesFetcher{}, nil
}

// Fetch retrieves the CRL from the given bytes
//
// If cache is not nil, try to get the CRL from the cache first. On failure
// (e.g. cache miss), it will load the CRL from the bytes and store it to the
// cache.
func (f *BytesFetcher) Fetch(data []byte) (*corecrl.Bundle, error) {
	base, err := f.fetch(data)
	if err != nil {
		return nil, err
	}
	// check delta CRL
	// TODO: support delta CRL https://github.com/notaryproject/notation-core-go/issues/228
	for _, ext := range base.Extensions {
		if ext.Id.Equal(oidFreshestCRL) {
			return nil, errors.New("delta CRL is not supported")
		}
	}

	return &corecrl.Bundle{
		BaseCRL: base,
	}, nil
}

// fetch retrieves the CRL from the given bytes
func (f *BytesFetcher) fetch(data []byte) (*x509.RevocationList, error) {
	if len(data) >= maxCRLSize {
		return nil, fmt.Errorf("CRL size exceeds the limit: %d", maxCRLSize)
	}
	// parse CRL
	return x509.ParseRevocationList(data)
}
