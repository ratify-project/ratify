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
	"errors"
	"reflect"
	"testing"

	"github.com/ratify-project/ratify-go"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	subject1         = "subject1"
	subject2         = "subject2"
	mockVerifierName = "mock-verifier-name"
	mockVerifierType = "mock-verifier-type"
)

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

func TestConvertResult(t *testing.T) {
	tests := []struct {
		name     string
		src      *ratify.ValidationResult
		expected *result
	}{
		{
			name:     "nil source",
			src:      nil,
			expected: nil,
		},
		{
			name: "nil ArtifactReports",
			src: &ratify.ValidationResult{
				ArtifactReports: nil,
			},
			expected: &result{
				ArtifactReports: nil,
			},
		},
		{
			name: "nonempty ArtifactReports",
			src: &ratify.ValidationResult{
				ArtifactReports: []*ratify.ValidationReport{
					nil,
					{
						Subject: subject1,
						Results: []*ratify.VerificationResult{
							nil,
							{
								Verifier: &mockVerifier{},
								Err:      errors.New("error"),
								Detail:   map[string]any{},
							},
						},
					},
				},
			},
			expected: &result{
				ArtifactReports: []*validationReport{
					nil,
					{
						Subject: subject1,
						Results: []*verificationResult{
							nil,
							{
								VerifierName: mockVerifierName,
								ErrorReason:  "error",
								Detail:       "{}",
							},
						},
					},
				},
			},
		},
		{
			name: "nonempty ArtifactReports with invalid detail",
			src: &ratify.ValidationResult{
				ArtifactReports: []*ratify.ValidationReport{
					nil,
					{
						Subject: subject1,
						Results: []*ratify.VerificationResult{
							nil,
							{
								Verifier: &mockVerifier{},
								Err:      errors.New("error"),
								Detail:   make(chan int),
							},
						},
					},
				},
			},
			expected: &result{
				ArtifactReports: []*validationReport{
					nil,
					{
						Subject: subject1,
						Results: []*verificationResult{
							nil,
							nil,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convertResult(test.src)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Expected result: %v, got: %v", test.expected, result)
			}
		})
	}
}
