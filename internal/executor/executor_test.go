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

package executor

import (
	"context"
	"testing"

	"github.com/ratify-project/ratify-go"
	"github.com/ratify-project/ratify/v2/internal/store/factory"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	ef "github.com/ratify-project/ratify/v2/internal/policyenforcer/factory"
	vf "github.com/ratify-project/ratify/v2/internal/verifier/factory"
)

const (
	mockVerifierName       = "mock-verifier-name"
	mockVerifierType       = "mock-verifier-type"
	mockStoreType          = "mock-store"
	mockPolicyEnforcerType = "mock-policy-enforcer"
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

type mockPolicyEnforcer struct{}

func (m *mockPolicyEnforcer) Evaluator(_ context.Context, _ string) (ratify.Evaluator, error) {
	return nil, nil
}

func createPolicyEnforcer(_ *ef.NewPolicyEnforcerOptions) (ratify.PolicyEnforcer, error) {
	return &mockPolicyEnforcer{}, nil
}

type mockVerifier struct{}

func (m *mockVerifier) Name() string {
	return mockVerifierName
}
func (m *mockVerifier) Type() string {
	return mockVerifierType
}
func (m *mockVerifier) Verifiable(_ ocispec.Descriptor) bool {
	return true
}

func (m *mockVerifier) Verify(_ context.Context, _ *ratify.VerifyOptions) (*ratify.VerificationResult, error) {
	return &ratify.VerificationResult{}, nil
}

func createMockVerifier(_ vf.NewVerifierOptions) (ratify.Verifier, error) {
	return &mockVerifier{}, nil
}

func TestNewExecutor(t *testing.T) {
	factory.RegisterStoreFactory(mockStoreType, newMockStore)
	vf.RegisterVerifierFactory(mockVerifierType, createMockVerifier)
	ef.RegisterPolicyEnforcerFactory(mockPolicyEnforcerType, createPolicyEnforcer)

	tests := []struct {
		name           string
		opts           *Options
		expectErr      bool
		expectExecutor bool
	}{
		{
			name:           "nil options",
			opts:           nil,
			expectErr:      true,
			expectExecutor: false,
		},
		{
			name:           "failed to create verifiers",
			opts:           &Options{},
			expectErr:      true,
			expectExecutor: false,
		},
		{
			name: "failed to create store",
			opts: &Options{
				Verifiers: []vf.NewVerifierOptions{
					{
						Name: mockVerifierName,
						Type: mockVerifierType,
					},
				},
			},
			expectErr:      true,
			expectExecutor: false,
		},
		{
			name: "failed to create policy enforcer",
			opts: &Options{
				Verifiers: []vf.NewVerifierOptions{
					{
						Name: mockVerifierName,
						Type: mockVerifierType,
					},
				},
				Stores: map[string]factory.NewStoreOptions{
					"test": {
						Type: mockStoreType,
					},
				},
				Policy: &ef.NewPolicyEnforcerOptions{},
			},
			expectErr:      true,
			expectExecutor: false,
		},
		{
			name: "valid options",
			opts: &Options{
				Verifiers: []vf.NewVerifierOptions{
					{
						Name: mockVerifierName,
						Type: mockVerifierType,
					},
				},
				Stores: map[string]factory.NewStoreOptions{
					"test": {
						Type: mockStoreType,
					},
				},
				Policy: &ef.NewPolicyEnforcerOptions{
					Type: mockPolicyEnforcerType,
				},
			},
			expectErr:      false,
			expectExecutor: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor, err := NewExecutor(test.opts)
			if (err != nil) != test.expectErr {
				t.Errorf("expected error: %v, got: %v", test.expectErr, err)
			}
			if (executor != nil) != test.expectExecutor {
				t.Errorf("expected executor: %v, got: %v", test.expectExecutor, executor != nil)
			}
		})
	}
}
