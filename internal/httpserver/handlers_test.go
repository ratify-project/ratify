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

package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/notaryproject/ratify-go"
	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/singleflight"
)

type mockCache struct {
	entries map[string]string
}

func (c *mockCache) Get(_ context.Context, key string) (any, error) {
	if val, ok := c.entries[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("key not found")
}

func (c *mockCache) Set(_ context.Context, key string, value any) error {
	if strVal, ok := value.(string); ok {
		c.entries[key] = strVal
		return fmt.Errorf("duplicate key")
	}
	return nil
}

func (c *mockCache) Delete(_ context.Context, key string) error {
	delete(c.entries, key)
	return nil
}

func TestVerify(t *testing.T) {
	// cache, err := ristretto.NewRistrettoCache(defaultCacheTTL)
	// if err != nil {
	// 	t.Fatalf("failed to create cache: %v", err)
	// }
	server := &server{
		executor: &ratify.Executor{},
		cache:    &mockCache{entries: make(map[string]string)},
		sfGroup:  new(singleflight.Group),
	}

	tests := []struct {
		name          string
		requestBody   string
		expectedError bool
		cacheEntries  map[string]string
		expectedItems []externaldata.Item
	}{
		{
			name: "Valid request",
			requestBody: `{
				"request": {
					"keys": ["artifact1"]
				}
			}`,
			expectedError: false,
			expectedItems: []externaldata.Item{
				{
					Key:   "artifact1",
					Value: nil,
					Error: "store must be configured",
				},
			},
		},
		{
			name: "Valid request with cache hit",
			requestBody: `{
				"request": {
					"keys": ["artifact1"]
				}
			}`,
			cacheEntries: map[string]string{
				"verify_artifact1": "cachedValue",
			},
			expectedError: false,
			expectedItems: []externaldata.Item{
				{
					Key:   "artifact1",
					Value: "cachedValue",
				},
			},
		},
		{
			name:          "Invalid JSON",
			requestBody:   `{invalid-json}`,
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(test.requestBody))
			w := httptest.NewRecorder()

			if test.cacheEntries != nil {
				server.cache = &mockCache{entries: test.cacheEntries}
			}
			err := server.verify(context.Background(), w, req)
			if (err != nil) != test.expectedError {
				t.Errorf("expected error: %v, got: %v", test.expectedError, err)
			}

			if !test.expectedError {
				var response externaldata.ProviderResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if !reflect.DeepEqual(response.Response.Items, test.expectedItems) {
					t.Errorf("expected items: %v, got: %v", test.expectedItems, response.Response.Items)
				}
			}
		})
	}
}

func TestMutate(t *testing.T) {
	tests := []struct {
		name          string
		requestBody   string
		cacheEntries  map[string]string
		expectedError bool
		store         ratify.Store
		expectedItems []externaldata.Item
	}{
		{
			name: "Valid mutate request",
			requestBody: `{
				"request": {
					"keys": ["testrepo/testimage@sha256:498138d40d54f0fc20cd271e215366d3d8803f814b8f565b47c101480bbaaa88"]
				}
			}`,
			expectedError: false,
			expectedItems: []externaldata.Item{
				{
					Key:   "testrepo/testimage@sha256:498138d40d54f0fc20cd271e215366d3d8803f814b8f565b47c101480bbaaa88",
					Value: "testrepo/testimage@sha256:498138d40d54f0fc20cd271e215366d3d8803f814b8f565b47c101480bbaaa88",
				},
			},
		},
		{
			name:          "Invalid JSON mutate",
			requestBody:   `{invalid-json}`,
			expectedError: true,
		},
		{
			name: "Invalid reference",
			requestBody: `{
				"request": {
					"keys": ["testrepo"]
				}
			}`,
			expectedError: false,
			expectedItems: []externaldata.Item{
				{
					Key:   "testrepo",
					Value: "testrepo",
					Error: "failed to parse reference: invalid reference: missing registry or repository",
				},
			},
		},
		{
			name: "Cache hit",
			requestBody: `{
				"request": {
					"keys": ["testrepo/testimage:v1"]
				}
			}`,
			store: &mockStore{
				returnResolveErr: true,
			},
			cacheEntries: map[string]string{
				"mutate_testrepo/testimage:v1": "testrepo/testimage@sha256:498138d40d54f0fc20cd271e215366d3d8803f814b8f565b47c101480bbaaa88",
			},
			expectedError: false,
			expectedItems: []externaldata.Item{
				{
					Key:   "testrepo/testimage:v1",
					Value: "testrepo/testimage@sha256:498138d40d54f0fc20cd271e215366d3d8803f814b8f565b47c101480bbaaa88",
				},
			},
		},
		{
			name: "Store fails to resolve reference",
			requestBody: `{
				"request": {
					"keys": ["testrepo/testimage:v1"]
				}
			}`,
			store: &mockStore{
				returnResolveErr: true,
			},
			expectedError: false,
			expectedItems: []externaldata.Item{
				{
					Key:   "testrepo/testimage:v1",
					Value: "testrepo/testimage:v1",
					Error: "mock error",
				},
			},
		},
		{
			name: "Store resolves reference successfully",
			requestBody: `{
				"request": {
					"keys": ["testrepo/testimage:v1"]
				}
			}`,
			store: &mockStore{
				resolveMap: map[string]ocispec.Descriptor{
					"testrepo/testimage:v1": {
						Digest: "sha256:498138d40d54f0fc20cd271e215366d3d8803f814b8f565b47c101480bbaaa88",
					},
				},
			},
			expectedError: false,
			expectedItems: []externaldata.Item{
				{
					Key:   "testrepo/testimage:v1",
					Value: "testrepo/testimage@sha256:498138d40d54f0fc20cd271e215366d3d8803f814b8f565b47c101480bbaaa88",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/mutate", strings.NewReader(test.requestBody))
			w := httptest.NewRecorder()

			server := &server{
				executor: &ratify.Executor{
					Store: test.store,
				},
				cache:   &mockCache{entries: make(map[string]string)},
				sfGroup: new(singleflight.Group),
			}
			if test.cacheEntries != nil {
				server.cache = &mockCache{entries: test.cacheEntries}
			}
			if err := server.mutate(context.Background(), w, req); (err != nil) != test.expectedError {
				t.Errorf("expected error: %v, got: %v", test.expectedError, err)
			}

			if !test.expectedError {
				var response externaldata.ProviderResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if !reflect.DeepEqual(response.Response.Items, test.expectedItems) {
					t.Errorf("expected items: %v, got: %v", test.expectedItems, response.Response.Items)
				}
				if !response.Response.Idempotent {
					t.Errorf("expected Idempotent to be true for mutate, got false")
				}
			}
		})
	}
}
