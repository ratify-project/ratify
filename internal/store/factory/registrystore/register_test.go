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

package registrystore

import (
	"context"
	"testing"

	"github.com/notaryproject/ratify-go"
	"github.com/notaryproject/ratify/v2/internal/store/factory"
)

func TestNewStore(t *testing.T) {
	tests := []struct {
		name      string
		opts      factory.NewStoreOptions
		expectErr bool
	}{
		{
			name: "Unsupported params",
			opts: factory.NewStoreOptions{
				Type:       registryStoreType,
				Parameters: make(chan int),
			},
			expectErr: true,
		},
		{
			name: "Malformed params",
			opts: factory.NewStoreOptions{
				Type:       registryStoreType,
				Parameters: "{",
			},
			expectErr: true,
		},
		{
			name: "Valid registry params",
			opts: factory.NewStoreOptions{
				Type:       registryStoreType,
				Parameters: map[string]interface{}{},
			},
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := factory.NewStore(test.opts)
			if (err != nil) != test.expectErr {
				t.Errorf("expected error: %v, got: %v", test.expectErr, err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name         string
		username     string
		password     string
		expectErr    bool
		expectedCred ratify.RegistryCredential
	}{
		{
			name:      "username/password provided",
			username:  "testuser",
			password:  "testpassword",
			expectErr: false,
			expectedCred: ratify.RegistryCredential{
				Username: "testuser",
				Password: "testpassword",
			},
		},
		{
			name:      "only password provided",
			username:  "",
			password:  "token",
			expectErr: false,
			expectedCred: ratify.RegistryCredential{
				RefreshToken: "token",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			getter := &defaultCredGetter{
				username: test.username,
				password: test.password,
			}
			got, err := getter.Get(context.Background(), "")
			if (err != nil) != test.expectErr {
				t.Errorf("expected error: %v, got: %v", test.expectErr, err)
			}
			if got != test.expectedCred {
				t.Errorf("expected credential: %v, got: %v", test.expectedCred, got)
			}
		})
	}
}
