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

package store

import (
	"context"
	"testing"

	"github.com/notaryproject/ratify-go"
	"github.com/notaryproject/ratify/v2/internal/store/factory"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type mockStore struct{}

func (m *mockStore) Resolve(_ context.Context, _ string) (ocispec.Descriptor, error) {
	return ocispec.Descriptor{}, nil
}

func (m *mockStore) ListReferrers(_ context.Context, _ string, _ []string, _ func(referrers []ocispec.Descriptor) error) error {
	return nil
}

func (m *mockStore) FetchBlob(_ context.Context, _ string, _ ocispec.Descriptor) ([]byte, error) {
	return nil, nil
}

func (m *mockStore) FetchManifest(_ context.Context, _ string, _ ocispec.Descriptor) ([]byte, error) {
	return nil, nil
}

func newMockStore(_ factory.NewStoreOptions) (ratify.Store, error) {
	return &mockStore{}, nil
}

func TestNewStore(t *testing.T) {
	factory.RegisterStoreFactory("mock-store", newMockStore)
	tests := []struct {
		name          string
		opts          PatternOptions
		expectedError bool
	}{
		{
			name:          "empty store options",
			opts:          PatternOptions{},
			expectedError: true,
		},
		{
			name: "unregistered store options",
			opts: PatternOptions{
				"test": factory.NewStoreOptions{
					Type:       "mock",
					Parameters: map[string]any{},
				},
			},
			expectedError: true,
		},
		{
			name: "no pattern provided",
			opts: PatternOptions{
				"": factory.NewStoreOptions{
					Type:       "mock-store",
					Parameters: map[string]any{},
				},
			},
			expectedError: true,
		},
		{
			name: "valid store options",
			opts: PatternOptions{
				"test": factory.NewStoreOptions{
					Type:       "mock-store",
					Parameters: map[string]any{},
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStore(tt.opts)
			if (err != nil) != tt.expectedError {
				t.Errorf("NewStore() error = %v, expectedError %v", err, tt.expectedError)
			}
		})
	}
}
