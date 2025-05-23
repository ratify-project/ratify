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

const testType = "test-type"

func createPolicyEnforcer(_ *NewPolicyEnforcerOptions) (ratify.PolicyEnforcer, error) {
	return nil, nil
}

func TestRegisterPolicyEnforcerFactory(t *testing.T) {
	t.Run("Registering an empty type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic when registering an empty type, but did not panic")
			}
		}()
		RegisterPolicyEnforcerFactory("", createPolicyEnforcer)
	})

	t.Run("Registering a nil factory function", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic when registering a nil factory function, but did not panic")
			}
		}()
		RegisterPolicyEnforcerFactory(testType, nil)
	})

	t.Run("Registering a valid factory function", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Did not expect panic when registering a valid factory function, but got: %v", r)
			}
			delete(registeredPolicyEnforcers, testType)
		}()
		RegisterPolicyEnforcerFactory(testType, createPolicyEnforcer)
	})

	t.Run("Registering the same type again", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic when registering the same type again, but did not panic")
			}
			delete(registeredPolicyEnforcers, testType)
		}()
		RegisterPolicyEnforcerFactory(testType, createPolicyEnforcer)
		RegisterPolicyEnforcerFactory(testType, createPolicyEnforcer)
	})
}

func TestNewPolicyEnforcer(t *testing.T) {
	t.Run("nil options", func(t *testing.T) {
		if _, err := NewPolicyEnforcer(nil); err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
	})

	t.Run("empty type", func(t *testing.T) {
		opts := &NewPolicyEnforcerOptions{
			Type: "",
		}
		if _, err := NewPolicyEnforcer(opts); err == nil {
			t.Errorf("Expected error, but got none")
		}
	})

	t.Run("unregistered type", func(t *testing.T) {
		opts := &NewPolicyEnforcerOptions{
			Type: "unregistered",
		}
		if _, err := NewPolicyEnforcer(opts); err == nil {
			t.Errorf("Expected error, but got none")
		}
	})

	t.Run("registered type", func(t *testing.T) {
		RegisterPolicyEnforcerFactory(testType, createPolicyEnforcer)
		defer delete(registeredPolicyEnforcers, testType)

		opts := &NewPolicyEnforcerOptions{
			Type: testType,
		}
		if _, err := NewPolicyEnforcer(opts); err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
	})
}
