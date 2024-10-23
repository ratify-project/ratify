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
	"context"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	crl "github.com/notaryproject/notation-core-go/revocation/crl"
)

// oidFreshestCRL is the object identifier for the distribution point
// for the delta CRL. (See RFC 5280, Section 5.2.6)
var oidFreshestCRL = asn1.ObjectIdentifier{2, 5, 29, 46}

// maxCRLSize is the maximum size of CRL in bytes
//
// The 32 MiB limit is based on investigation that even the largest CRLs
// are less than 16 MiB. The limit is set to 32 MiB to prevent
const maxCRLSize = 32 * 1024 * 1024 // 32 MiB

// Options specifies values that are needed to check revocation
type Options struct {
	RevocationList []byte
}

// Fetcher is an interface that specifies methods used for fetching CRL
// from the given URL or raw date.
type Fetcher interface {
	// Fetch retrieves the CRL from the given URL.
	Fetch(ctx context.Context, url string) (*crl.Bundle, error)

	// FetchWithOptions retrieves the CRL from the given raw data.
	FetchWithOptions(ctx context.Context, url string, opts Options) (*crl.Bundle, error)
}

// CRLFetcher is a Fetcher implementation that fetches CRL from the given URL
type CRLFetcher struct {
	// Cache stores fetched CRLs and reuses them until the CRL reaches the
	// NextUpdate time.
	// If Cache is nil, no cache is used.
	Cache crl.Cache

	// DiscardCacheError specifies whether to discard any error on cache.
	//
	// ErrCacheMiss is not considered as an failure and will not be returned as
	// an error if DiscardCacheError is false.
	DiscardCacheError bool

	httpClient *http.Client
}

// NewCRLFetcher creates a new HTTPFetcher with the given HTTP client
func NewCRLFetcher(httpClient *http.Client) (*CRLFetcher, error) {
	if httpClient == nil {
		return nil, errors.New("httpClient cannot be nil")
	}

	return &CRLFetcher{
		httpClient: httpClient,
	}, nil
}

// Fetch retrieves the CRL from the given URL
//
// If cache is not nil, try to get the CRL from the cache first. On failure
// (e.g. cache miss), it will download the CRL from the URL and store it to the
// cache.
func (f *CRLFetcher) Fetch(ctx context.Context, url string) (*crl.Bundle, error) {
	if url == "" {
		return nil, errors.New("CRL URL cannot be empty")
	}

	if f.Cache != nil {
		bundle, err := f.Cache.Get(ctx, url)
		if err == nil {
			// check expiry
			nextUpdate := bundle.BaseCRL.NextUpdate
			if !nextUpdate.IsZero() && !time.Now().After(nextUpdate) {
				return bundle, nil
			}
		} else if !errors.Is(err, crl.ErrCacheMiss) && !f.DiscardCacheError {
			return nil, fmt.Errorf("failed to retrieve CRL from cache: %w", err)
		}
	}

	bundle, err := f.fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve CRL: %w", err)
	}

	if f.Cache != nil {
		err = f.Cache.Set(ctx, url, bundle)
		if err != nil && !f.DiscardCacheError {
			return nil, fmt.Errorf("failed to store CRL to cache: %w", err)
		}
	}

	return bundle, nil
}

// fetch downloads the CRL from the given URL.
func (f *CRLFetcher) fetch(ctx context.Context, url string) (*crl.Bundle, error) {
	// fetch base CRL
	base, err := fetchCRL(ctx, url, f.httpClient)
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

	return &crl.Bundle{
		BaseCRL: base,
	}, nil
}

func fetchCRL(ctx context.Context, crlURL string, client *http.Client) (*x509.RevocationList, error) {
	// validate URL
	parsedURL, err := url.Parse(crlURL)
	if err != nil {
		return nil, fmt.Errorf("invalid CRL URL: %w", err)
	}
	if parsedURL.Scheme != "http" {
		return nil, fmt.Errorf("unsupported scheme: %s. Only supports CRL URL in HTTP protocol", parsedURL.Scheme)
	}

	// download CRL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, crlURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRL request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to download with status code: %d", resp.StatusCode)
	}
	// read with size limit
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxCRLSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read CRL response: %w", err)
	}
	if len(data) == maxCRLSize {
		return nil, fmt.Errorf("CRL size exceeds the limit: %d", maxCRLSize)
	}

	// parse CRL
	return x509.ParseRevocationList(data)
}

// FetchWithOptions retrieves the CRL from the given raw data
func (f *CRLFetcher) FetchWithOptions(ctx context.Context, url string, opts Options) (*crl.Bundle, error) {
	if url == "" {
		return nil, errors.New("CRL URL cannot be empty")
	}

	// parse CRL
	base, err := x509.ParseRevocationList(opts.RevocationList)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CRL: %w", err)
	}

	return &crl.Bundle{
		BaseCRL: base,
	}, nil
}
