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

package filesystemocistore

import (
	"testing"

	"github.com/notaryproject/ratify/v2/internal/store/factory"
)

func TestNewStore(t *testing.T) {
	tests := []struct {
		name      string
		opts      *factory.NewStoreOptions
		expectErr bool
	}{
		{
			name: "Nil params",
			opts: &factory.NewStoreOptions{
				Type:       filesystemOCIStoreType,
				Parameters: nil,
			},
			expectErr: true,
		},
		{
			name: "Unsupported params",
			opts: &factory.NewStoreOptions{
				Type:       filesystemOCIStoreType,
				Parameters: make(chan int),
			},
			expectErr: true,
		},
		{
			name: "Malformed params",
			opts: &factory.NewStoreOptions{
				Type:       filesystemOCIStoreType,
				Parameters: "{",
			},
			expectErr: true,
		},
		{
			name: "Missing path params",
			opts: &factory.NewStoreOptions{
				Type:       filesystemOCIStoreType,
				Parameters: map[string]interface{}{},
			},
			expectErr: true,
		},
		{
			name: "Empty Path value",
			opts: &factory.NewStoreOptions{
				Type: filesystemOCIStoreType,
				Parameters: map[string]interface{}{
					"path": "",
				},
			},
			expectErr: true,
		},
		{
			name: "Nonexistent path",
			opts: &factory.NewStoreOptions{
				Type: filesystemOCIStoreType,
				Parameters: map[string]interface{}{
					"path": "/nonexistent/path",
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := factory.NewStore(tt.opts)
			if (err != nil) != tt.expectErr {
				t.Errorf("NewStore() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
		})
	}
}
