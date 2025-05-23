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

package policyenforcer

import (
	"context"
	"testing"

	"github.com/notaryproject/ratify-go"
	"github.com/notaryproject/ratify/v2/internal/policyenforcer/factory"
)

const mockType = "mock-type"

type mockPolicyEnforcer struct{}

func (m *mockPolicyEnforcer) Evaluator(_ context.Context, _ string) (ratify.Evaluator, error) {
	return nil, nil
}

func createPolicyEnforcer(_ *factory.NewPolicyEnforcerOptions) (ratify.PolicyEnforcer, error) {
	return &mockPolicyEnforcer{}, nil
}

func TestNewPolicyEnforcer(t *testing.T) {
	t.Run("Registering a factory", func(t *testing.T) {
		factory.RegisterPolicyEnforcerFactory(mockType, createPolicyEnforcer)
		_, err := NewPolicyEnforcer(&factory.NewPolicyEnforcerOptions{
			Type: mockType,
		})
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
	})
}
