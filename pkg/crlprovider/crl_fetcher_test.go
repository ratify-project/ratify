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
	"bytes"
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	crl "github.com/notaryproject/notation-core-go/revocation/crl"
	"github.com/notaryproject/notation-core-go/testhelper"
)

func TestNewCRLFetcher(t *testing.T) {
	t.Run("httpClient is nil", func(t *testing.T) {
		_, err := NewCRLFetcher(nil)
		if err.Error() != "httpClient cannot be nil" {
			t.Errorf("NewCRLFetcher() error = %v, want %v", err, "httpClient cannot be nil")
		}
	})
}

func TestFetch(t *testing.T) {
	// prepare crl
	certChain := testhelper.GetRevokableRSAChainWithRevocations(2, false, true)
	crlBytes, err := x509.CreateRevocationList(rand.Reader, &x509.RevocationList{
		Number:     big.NewInt(1),
		NextUpdate: time.Now().Add(1 * time.Hour),
	}, certChain[1].Cert, certChain[1].PrivateKey)
	if err != nil {
		t.Fatalf("failed to create base CRL: %v", err)
	}
	baseCRL, err := x509.ParseRevocationList(crlBytes)
	if err != nil {
		t.Fatalf("failed to parse base CRL: %v", err)
	}
	const exampleURL = "http://example.com"
	const uncachedURL = "http://uncached.com"

	bundle := &crl.Bundle{
		BaseCRL: baseCRL,
	}

	t.Run("url is empty", func(t *testing.T) {
		c := &memoryCache{}
		httpClient := &http.Client{}
		f, err := NewCRLFetcher(httpClient)
		if err != nil {
			t.Errorf("NewCRLFetcher() error = %v, want nil", err)
		}
		f.Cache = c
		_, err = f.Fetch(context.Background(), "")
		if err.Error() != "CRL URL cannot be empty" {
			t.Fatalf("Fetcher.Fetch() error = %v, want CRL URL cannot be empty", err)
		}
	})

	t.Run("fetch without cache", func(t *testing.T) {
		httpClient := &http.Client{
			Transport: expectedRoundTripperMock{Body: baseCRL.Raw},
		}
		f, err := NewCRLFetcher(httpClient)
		if err != nil {
			t.Errorf("NewCRLFetcher() error = %v, want nil", err)
		}
		bundle, err := f.Fetch(context.Background(), exampleURL)
		if err != nil {
			t.Errorf("Fetcher.Fetch() error = %v, want nil", err)
		}
		if !bytes.Equal(bundle.BaseCRL.Raw, baseCRL.Raw) {
			t.Errorf("Fetcher.Fetch() base.Raw = %v, want %v", bundle.BaseCRL.Raw, baseCRL.Raw)
		}
	})

	t.Run("cache hit", func(t *testing.T) {
		// set the cache
		c := &memoryCache{}
		if err := c.Set(context.Background(), exampleURL, bundle); err != nil {
			t.Errorf("Cache.Set() error = %v, want nil", err)
		}

		httpClient := &http.Client{}
		f, err := NewCRLFetcher(httpClient)
		if err != nil {
			t.Errorf("NewCRLFetcher() error = %v, want nil", err)
		}
		f.Cache = c
		bundle, err := f.Fetch(context.Background(), exampleURL)
		if err != nil {
			t.Errorf("Fetcher.Fetch() error = %v, want nil", err)
		}
		if !bytes.Equal(bundle.BaseCRL.Raw, baseCRL.Raw) {
			t.Errorf("Fetcher.Fetch() base.Raw = %v, want %v", bundle.BaseCRL.Raw, baseCRL.Raw)
		}
	})

	t.Run("cache miss and download failed error", func(t *testing.T) {
		c := &memoryCache{}
		httpClient := &http.Client{
			Transport: errorRoundTripperMock{},
		}
		f, err := NewCRLFetcher(httpClient)
		f.Cache = c
		if err != nil {
			t.Errorf("NewCRLFetcher() error = %v, want nil", err)
		}
		_, err = f.Fetch(context.Background(), uncachedURL)
		if err == nil {
			t.Errorf("Fetcher.Fetch() error = nil, want not nil")
		}
	})

	t.Run("cache miss", func(t *testing.T) {
		c := &memoryCache{}
		httpClient := &http.Client{
			Transport: expectedRoundTripperMock{Body: baseCRL.Raw},
		}
		f, err := NewCRLFetcher(httpClient)
		if err != nil {
			t.Errorf("NewCRLFetcher() error = %v, want nil", err)
		}
		f.Cache = c
		f.DiscardCacheError = false
		bundle, err := f.Fetch(context.Background(), uncachedURL)
		if err != nil {
			t.Errorf("Fetcher.Fetch() error = %v, want nil", err)
		}
		if !bytes.Equal(bundle.BaseCRL.Raw, baseCRL.Raw) {
			t.Errorf("Fetcher.Fetch() base.Raw = %v, want %v", bundle.BaseCRL.Raw, baseCRL.Raw)
		}
	})

	t.Run("cache expired", func(t *testing.T) {
		c := &memoryCache{}
		// prepare an expired CRL
		certChain := testhelper.GetRevokableRSAChainWithRevocations(2, false, true)
		expiredCRLBytes, err := x509.CreateRevocationList(rand.Reader, &x509.RevocationList{
			Number:     big.NewInt(1),
			NextUpdate: time.Now().Add(-1 * time.Hour),
		}, certChain[1].Cert, certChain[1].PrivateKey)
		if err != nil {
			t.Fatalf("failed to create base CRL: %v", err)
		}
		expiredCRL, err := x509.ParseRevocationList(expiredCRLBytes)
		if err != nil {
			t.Fatalf("failed to parse base CRL: %v", err)
		}
		// store the expired CRL
		const expiredCRLURL = "http://example.com/expired"
		bundle := &crl.Bundle{
			BaseCRL: expiredCRL,
		}
		if err := c.Set(context.Background(), expiredCRLURL, bundle); err != nil {
			t.Errorf("Cache.Set() error = %v, want nil", err)
		}

		// fetch the expired CRL
		httpClient := &http.Client{
			Transport: expectedRoundTripperMock{Body: baseCRL.Raw},
		}
		f, err := NewCRLFetcher(httpClient)
		if err != nil {
			t.Errorf("NewCRLFetcher() error = %v, want nil", err)
		}
		f.Cache = c
		f.DiscardCacheError = true
		bundle, err = f.Fetch(context.Background(), expiredCRLURL)
		if err != nil {
			t.Errorf("Fetcher.Fetch() error = %v, want nil", err)
		}
		// should re-download the CRL
		if !bytes.Equal(bundle.BaseCRL.Raw, baseCRL.Raw) {
			t.Errorf("Fetcher.Fetch() base.Raw = %v, want %v", bundle.BaseCRL.Raw, baseCRL.Raw)
		}
	})

	t.Run("delta CRL is not supported", func(t *testing.T) {
		c := &memoryCache{}
		// prepare a CRL with refresh CRL extension
		certChain := testhelper.GetRevokableRSAChainWithRevocations(2, false, true)
		expiredCRLBytes, err := x509.CreateRevocationList(rand.Reader, &x509.RevocationList{
			Number:     big.NewInt(1),
			NextUpdate: time.Now().Add(-1 * time.Hour),
			ExtraExtensions: []pkix.Extension{
				{
					Id:    oidFreshestCRL,
					Value: []byte{0x01, 0x02, 0x03},
				},
			},
		}, certChain[1].Cert, certChain[1].PrivateKey)
		if err != nil {
			t.Fatalf("failed to create base CRL: %v", err)
		}

		httpClient := &http.Client{
			Transport: expectedRoundTripperMock{Body: expiredCRLBytes},
		}
		f, err := NewCRLFetcher(httpClient)
		if err != nil {
			t.Errorf("NewCRLFetcher() error = %v, want nil", err)
		}
		f.Cache = c
		f.DiscardCacheError = true
		_, err = f.Fetch(context.Background(), uncachedURL)
		if !strings.Contains(err.Error(), "delta CRL is not supported") {
			t.Errorf("Fetcher.Fetch() error = %v, want delta CRL is not supported", err)
		}
	})

	t.Run("Set cache error", func(t *testing.T) {
		c := &errorCache{
			GetError: crl.ErrCacheMiss,
			SetError: errors.New("cache error"),
		}
		httpClient := &http.Client{
			Transport: expectedRoundTripperMock{Body: baseCRL.Raw},
		}
		f, err := NewCRLFetcher(httpClient)
		if err != nil {
			t.Errorf("NewCRLFetcher() error = %v, want nil", err)
		}
		f.Cache = c
		f.DiscardCacheError = true
		bundle, err = f.Fetch(context.Background(), exampleURL)
		if err != nil {
			t.Errorf("Fetcher.Fetch() error = %v, want nil", err)
		}
		if !bytes.Equal(bundle.BaseCRL.Raw, baseCRL.Raw) {
			t.Errorf("Fetcher.Fetch() base.Raw = %v, want %v", bundle.BaseCRL.Raw, baseCRL.Raw)
		}
	})

	t.Run("Get error without discard", func(t *testing.T) {
		c := &errorCache{
			GetError: errors.New("cache error"),
		}
		httpClient := &http.Client{
			Transport: expectedRoundTripperMock{Body: baseCRL.Raw},
		}
		f, err := NewCRLFetcher(httpClient)
		if err != nil {
			t.Errorf("NewCRLFetcher() error = %v, want nil", err)
		}
		f.Cache = c
		f.DiscardCacheError = false
		_, err = f.Fetch(context.Background(), exampleURL)
		if !strings.HasPrefix(err.Error(), "failed to retrieve CRL from cache:") {
			t.Errorf("Fetcher.Fetch() error = %v, want failed to retrieve CRL from cache:", err)
		}
	})

	t.Run("Set error without discard", func(t *testing.T) {
		c := &errorCache{
			GetError: crl.ErrCacheMiss,
			SetError: errors.New("cache error"),
		}
		httpClient := &http.Client{
			Transport: expectedRoundTripperMock{Body: baseCRL.Raw},
		}
		f, err := NewCRLFetcher(httpClient)
		if err != nil {
			t.Errorf("NewCRLFetcher() error = %v, want nil", err)
		}
		f.Cache = c
		f.DiscardCacheError = false
		_, err = f.Fetch(context.Background(), exampleURL)
		if !strings.HasPrefix(err.Error(), "failed to store CRL to cache:") {
			t.Errorf("Fetcher.Fetch() error = %v, want failed to store CRL to cache:", err)
		}
	})
}

func TestFetchWithOptions(t *testing.T) {
	// prepare crl
	certChain := testhelper.GetRevokableRSAChainWithRevocations(2, false, true)
	crlBytes, err := x509.CreateRevocationList(rand.Reader, &x509.RevocationList{
		Number:     big.NewInt(1),
		NextUpdate: time.Now().Add(1 * time.Hour),
	}, certChain[1].Cert, certChain[1].PrivateKey)
	if err != nil {
		t.Fatalf("failed to create base CRL: %v", err)
	}
	// baseCRL, err := x509.ParseRevocationList(crlBytes)
	if err != nil {
		t.Fatalf("failed to parse base CRL: %v", err)
	}
	const exampleURL = "http://example.com"

	t.Run("default success path", func(t *testing.T) {
		c := &memoryCache{}
		httpClient := &http.Client{}
		f, err := NewCRLFetcher(httpClient)
		if err != nil {
			t.Errorf("NewCRLFetcher() error = %v, want nil", err)
		}
		f.Cache = c
		_, err = f.FetchWithOptions(context.Background(), exampleURL, Options{RevocationList: crlBytes})
		if err != nil {
			t.Errorf("FetchWithOptions() error = %v, want nil", err)
		}
	})

	t.Run("url is empty", func(t *testing.T) {
		c := &memoryCache{}
		httpClient := &http.Client{}
		f, err := NewCRLFetcher(httpClient)
		if err != nil {
			t.Errorf("NewCRLFetcher() error = %v, want nil", err)
		}
		f.Cache = c
		_, err = f.FetchWithOptions(context.Background(), "", Options{RevocationList: crlBytes})
		if err.Error() != "CRL URL cannot be empty" {
			t.Fatalf("Fetcher.Fetch() error = %v, want CRL URL cannot be empty", err)
		}
	})
}

func TestDownload(t *testing.T) {
	t.Run("parse url error", func(t *testing.T) {
		_, err := fetchCRL(context.Background(), ":", http.DefaultClient)
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("https download", func(t *testing.T) {
		_, err := fetchCRL(context.Background(), "https://example.com", http.DefaultClient)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("http.NewRequestWithContext error", func(t *testing.T) {
		var ctx context.Context = nil
		_, err := fetchCRL(ctx, "http://example.com", &http.Client{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("client.Do error", func(t *testing.T) {
		_, err := fetchCRL(context.Background(), "http://example.com", &http.Client{
			Transport: errorRoundTripperMock{},
		})

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("status code is not 2xx", func(t *testing.T) {
		_, err := fetchCRL(context.Background(), "http://example.com", &http.Client{
			Transport: serverErrorRoundTripperMock{},
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("readAll error", func(t *testing.T) {
		_, err := fetchCRL(context.Background(), "http://example.com", &http.Client{
			Transport: readFailedRoundTripperMock{},
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("exceed the size limit", func(t *testing.T) {
		_, err := fetchCRL(context.Background(), "http://example.com", &http.Client{
			Transport: expectedRoundTripperMock{Body: make([]byte, maxCRLSize+1)},
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid crl", func(t *testing.T) {
		_, err := fetchCRL(context.Background(), "http://example.com", &http.Client{
			Transport: expectedRoundTripperMock{Body: []byte("invalid crl")},
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

type errorRoundTripperMock struct{}

func (rt errorRoundTripperMock) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("error")
}

type serverErrorRoundTripperMock struct{}

func (rt serverErrorRoundTripperMock) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Request:    req,
		StatusCode: http.StatusInternalServerError,
	}, nil
}

type readFailedRoundTripperMock struct{}

func (rt readFailedRoundTripperMock) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       errorReaderMock{},
	}, nil
}

type errorReaderMock struct{}

func (r errorReaderMock) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("error")
}

func (r errorReaderMock) Close() error {
	return nil
}

type expectedRoundTripperMock struct {
	Body []byte
}

func (rt expectedRoundTripperMock) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Request:    req,
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(rt.Body)),
	}, nil
}

// memoryCache is an in-memory cache that stores CRL bundles for testing.
type memoryCache struct {
	store sync.Map
}

// Get retrieves the CRL from the memory store.
//
// - if the key does not exist, return ErrNotFound
// - if the CRL is expired, return ErrCacheMiss
func (c *memoryCache) Get(ctx context.Context, url string) (*crl.Bundle, error) {
	value, ok := c.store.Load(url)
	if !ok {
		return nil, crl.ErrCacheMiss
	}
	bundle, ok := value.(*crl.Bundle)
	if !ok {
		return nil, fmt.Errorf("invalid type: %T", value)
	}

	return bundle, nil
}

// Set stores the CRL in the memory store.
func (c *memoryCache) Set(ctx context.Context, url string, bundle *crl.Bundle) error {
	c.store.Store(url, bundle)
	return nil
}

type errorCache struct {
	GetError error
	SetError error
}

func (c *errorCache) Get(ctx context.Context, url string) (*crl.Bundle, error) {
	return nil, c.GetError
}

func (c *errorCache) Set(ctx context.Context, url string, bundle *crl.Bundle) error {
	return c.SetError
}
