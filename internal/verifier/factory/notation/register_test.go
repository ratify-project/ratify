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

package notation

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"testing"

	"github.com/notaryproject/ratify/v2/internal/verifier/factory"
	"github.com/notaryproject/ratify/v2/internal/verifier/keyprovider"
)

const testName = "notation-test"
const mockKeyProviderName = "mock-key-provider"

type mockKeyProvider struct {
	returnErr bool
}

func (m *mockKeyProvider) GetCertificates(_ context.Context) ([]*x509.Certificate, error) {
	if m.returnErr {
		return nil, fmt.Errorf("mock error")
	}
	return []*x509.Certificate{
		{
			Subject: pkix.Name{
				CommonName: "test-cert",
			},
		}}, nil
}

func createMockKeyProvider(options any) (keyprovider.KeyProvider, error) {
	if options == nil {
		return &mockKeyProvider{}, nil
	}
	val, ok := options.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid options type")
	}
	_, ok = val["returnErr"]
	return &mockKeyProvider{
		returnErr: ok,
	}, nil
}

func TestNewVerifier(t *testing.T) {
	// Register the mock key provider
	keyprovider.RegisterKeyProvider(mockKeyProviderName, createMockKeyProvider)

	tests := []struct {
		name      string
		opts      *factory.NewVerifierOptions
		expectErr bool
	}{
		{
			name: "Unsupported params",
			opts: &factory.NewVerifierOptions{
				Type:       notationType,
				Name:       testName,
				Parameters: make(chan int),
			},
			expectErr: true,
		},
		{
			name: "Malformed params",
			opts: &factory.NewVerifierOptions{
				Type:       notationType,
				Name:       testName,
				Parameters: "{",
			},
			expectErr: true,
		},
		{
			name: "Missing trust store options",
			opts: &factory.NewVerifierOptions{
				Type:       notationType,
				Name:       testName,
				Parameters: options{},
			},
			expectErr: true,
		},
		{
			name: "Invalid trust store type",
			opts: &factory.NewVerifierOptions{
				Type: notationType,
				Name: testName,
				Parameters: options{
					Certificates: []trustStoreOptions{
						{
							"type": "invalid",
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Duplicate trust store type",
			opts: &factory.NewVerifierOptions{
				Type: notationType,
				Name: testName,
				Parameters: options{
					Certificates: []trustStoreOptions{
						{
							"type":              "ca",
							mockKeyProviderName: nil,
						},
						{
							"type":              "ca",
							mockKeyProviderName: nil,
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Non-registered key provider",
			opts: &factory.NewVerifierOptions{
				Type: notationType,
				Name: testName,
				Parameters: options{
					Certificates: []trustStoreOptions{
						{
							"type":           "ca",
							"non-registered": nil,
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Failed to get certificates from key provider",
			opts: &factory.NewVerifierOptions{
				Type: notationType,
				Name: testName,
				Parameters: options{
					Certificates: []trustStoreOptions{
						{
							"type": "ca",
							mockKeyProviderName: map[string]any{
								"returnErr": true,
							},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Valid notation options",
			opts: &factory.NewVerifierOptions{
				Type: notationType,
				Name: testName,
				Parameters: options{
					Certificates: []trustStoreOptions{
						{
							"type":              "ca",
							mockKeyProviderName: nil,
						},
					},
				},
			},
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := factory.NewVerifier(test.opts)
			if test.expectErr != (err != nil) {
				t.Fatalf("Expected error: %v, got: %v", test.expectErr, err)
			}
		})
	}
}
