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

package azure

import (
	"testing"
)

func TestValidateEndpoints(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    string
		expectedErr bool
	}{
		{
			name:        "global wildcard",
			endpoint:    "*",
			expectedErr: true,
		},
		{
			name:        "multiple wildcard",
			endpoint:    "*.example.*",
			expectedErr: true,
		},
		{
			name:        "no subdomain",
			endpoint:    "*.",
			expectedErr: true,
		},
		{
			name:        "full qualified domain",
			endpoint:    "example.com",
			expectedErr: false,
		},
		{
			name:        "valid wildcard domain",
			endpoint:    "*.example.com",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseEndpoints([]string{tt.endpoint})
			if tt.expectedErr != (err != nil) {
				t.Fatalf("expected error: %v, got error: %v", tt.expectedErr, err)
			}
		})
	}
}

func TestValidateHost(t *testing.T) {
	endpoints := []string{
		"*.azurecr.io",
		"example.azurecr.io",
	}
	tests := []struct {
		name        string
		host        string
		expectedErr bool
	}{
		{
			name:        "empty host",
			host:        "",
			expectedErr: true,
		},
		{
			name:        "valid host",
			host:        "example.azurecr.io",
			expectedErr: false,
		},
		{
			name:        "no subdomain",
			host:        "azurecr.io",
			expectedErr: true,
		},
		{
			name:        "multiple subdomains",
			host:        "example.test.azurecr.io",
			expectedErr: true,
		},
		{
			name:        "matched host",
			host:        "test.azurecr.io",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHost(tt.host, endpoints)
			if tt.expectedErr != (err != nil) {
				t.Fatalf("expected error: %v, got error: %v", tt.expectedErr, err)
			}
		})
	}
}
