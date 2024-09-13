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
package refresh

import (
	"context"
	"testing"
)

type MockRefresher struct{}

func (f *MockRefresher) Create(_ map[string]interface{}) (Refresher, error) {
	return &MockRefresher{}, nil
}

func (f *MockRefresher) Refresh(_ context.Context) error {
	return nil
}

func (f *MockRefresher) GetResult() interface{} {
	return nil
}

func TestRefreshFactory_Create(t *testing.T) {
	Register("mockRefresher", &MockRefresher{})
	refresherConfig := map[string]interface{}{
		"type": "mockRefresher",
	}
	factory := refresherFactories["mockRefresher"]
	refresher, err := factory.Create(refresherConfig)
	// refresher, err := CreateRefresherFromConfig(refresherConfig)
	if _, ok := refresher.(*MockRefresher); !ok {
		t.Errorf("Expected refresher to be of type MockRefresher, got %v", refresher)
	}
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestRegister_InvalidFactory(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, got nil")
		}
	}()

	Register("invalidRefresher", nil)
}

func TestRegister_DuplicateFactory(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, got nil")
		}
	}()

	Register("duplicateRefresher", &MockRefresher{})
	Register("duplicateRefresher", &MockRefresher{})
}

func TestRegister_ValidFactory(t *testing.T) {
	refresherFactories = make(map[string]RefresherFactory)
	Register("validRefresher", &MockRefresher{})
	if len(refresherFactories) != 1 {
		t.Errorf("Expected 1 factory to be registered, got %d", len(refresherFactories))
	}
}

func TestCreateRefresherFromConfig(t *testing.T) {
	Register("mockRefresher", &MockRefresher{})
	tests := []struct {
		name          string
		refresherType string
		expectedError bool
	}{
		{
			name:          "Valid Refresher Type",
			refresherType: "mockRefresher",
			expectedError: false,
		},
		{
			name:          "Invalid Refresher Type",
			refresherType: "invalidRefresher",
			expectedError: true,
		},
		{
			name:          "Empty Refresher Type",
			refresherType: "",
			expectedError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refresherConfig := map[string]interface{}{
				"type": tt.refresherType,
			}
			_, err := CreateRefresherFromConfig(refresherConfig)
			if tt.expectedError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}
