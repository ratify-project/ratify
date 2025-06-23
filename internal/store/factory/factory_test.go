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

package factory

import (
	"testing"

	"github.com/notaryproject/ratify-go"
)

const (
	testType = "test-type"
	testName = "test-name"
)

func createStore(_ *NewStoreOptions) (ratify.Store, error) {
	return nil, nil
}

func TestRegisterStoreFactory(t *testing.T) {
	t.Run("Registering an empty type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic when registering an empty type, but did not panic")
			}
		}()
		RegisterStoreFactory("", createStore)
	})

	t.Run("Registering a nil factory function", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic when registering a nil factory function, but did not panic")
			}
		}()
		RegisterStoreFactory(testType, nil)
	})

	t.Run("Registering a valid factory function", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Did not expect panic when registering a valid factory function, but got: %v", r)
			}
			delete(registeredStores, testType)
		}()
		RegisterStoreFactory(testType, createStore)
	})

	t.Run("Registering a duplicate factory function", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic when registering a duplicate factory function, but did not panic")
			}
			delete(registeredStores, testType)
		}()
		RegisterStoreFactory(testType, createStore)
		RegisterStoreFactory(testType, createStore)
	})
}

func TestNewStore(t *testing.T) {
	t.Run("Empty store options", func(t *testing.T) {
		_, err := NewStore(&NewStoreOptions{})
		if err == nil {
			t.Errorf("Expected error when creating a store with empty options, but got nil")
		}
	})

	t.Run("Unregistered store type", func(t *testing.T) {
		_, err := NewStore(&NewStoreOptions{Type: "unregistered"})
		if err == nil {
			t.Errorf("Expected error when creating a store with unregistered type, but got nil")
		}
	})

	t.Run("Valid store options", func(t *testing.T) {
		RegisterStoreFactory(testType, createStore)
		defer delete(registeredStores, testType)

		_, err := NewStore(&NewStoreOptions{Type: testType})
		if err != nil {
			t.Errorf("Did not expect error when creating a store with valid options, but got: %v", err)
		}
	})
}
