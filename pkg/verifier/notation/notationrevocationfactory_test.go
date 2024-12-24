// Copyright The Ratify Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package notation

import (
	"context"
	"net/http"
	"runtime"
	"testing"

	"github.com/notaryproject/notation-core-go/revocation"
	corecrl "github.com/notaryproject/notation-core-go/revocation/crl"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/crl"
	"github.com/stretchr/testify/assert"
)

func TestNewRevocationFactoryImpl(t *testing.T) {
	factory := NewCRLHandler()
	assert.NotNil(t, factory)
}

func TestNewFetcher(t *testing.T) {
	tests := []struct {
		name       string
		cacheRoot  string
		httpClient *http.Client
		wantErr    bool
	}{
		{
			name:       "valid fetcher",
			cacheRoot:  "/valid/path",
			httpClient: &http.Client{},
			wantErr:    false,
		},
		{
			name:       "invalid fetcher",
			cacheRoot:  "",
			httpClient: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &CRLHandler{httpClient: tt.httpClient}
			fetcher, err := factory.NewFetcher()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, fetcher)
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	factory := NewCRLHandler()
	opts := revocation.Options{}

	validator, err := factory.NewValidator(opts)
	assert.NoError(t, err)
	assert.NotNil(t, validator)
}
func TestNewFileCache(t *testing.T) {
	tests := []struct {
		name      string
		cacheRoot string
		wantErr   bool
	}{
		{
			name:      "valid cache root",
			cacheRoot: "/valid/path",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if runtime.GOOS == "windows" {
				t.Skip("skipping test on Windows")
			}
			cache, err := newFileCache(tt.cacheRoot)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cache)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cache)
			}
		})
	}
}
func TestConfigureCache(t *testing.T) {
	testCache, _ := crl.NewFileCache(dir.PathCRLCache)
	tests := []struct {
		name          string
		cacheDisabled bool
		fetcher       corecrl.Fetcher
		expectCache   bool
	}{
		{
			name:          "cache enabled",
			cacheDisabled: false,
			fetcher:       &corecrl.HTTPFetcher{Cache: testCache},
			expectCache:   true,
		},
		{
			name:          "cache disabled",
			cacheDisabled: true,
			fetcher:       &corecrl.HTTPFetcher{Cache: testCache},
			expectCache:   false,
		},
		{
			name:          "non-HTTP fetcher",
			cacheDisabled: true,
			fetcher:       &mockFetcher{},
			expectCache:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &CRLHandler{
				CacheDisabled: tt.cacheDisabled,
				Fetcher:       tt.fetcher,
			}
			handler.configureCache()

			if httpFetcher, ok := handler.Fetcher.(*corecrl.HTTPFetcher); ok {
				if tt.expectCache {
					assert.NotNil(t, httpFetcher.Cache)
				} else {
					assert.Nil(t, httpFetcher.Cache)
				}
			}
		})
	}
}

type mockFetcher struct{}

func (m *mockFetcher) Fetch(_ context.Context, _ string) (*corecrl.Bundle, error) {
	return nil, nil
}
