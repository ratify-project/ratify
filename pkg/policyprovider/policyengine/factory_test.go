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

package policyengine

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type mockEngine struct{}

func (e *mockEngine) Evaluate(_ context.Context, _ map[string]interface{}) (bool, error) {
	return true, nil
}

type mockFactory struct {
	returnErr bool
}

func (f *mockFactory) Create(_ string, _ string) (PolicyEngine, error) {
	if f.returnErr {
		return nil, errors.New("error")
	}
	return &mockEngine{}, nil
}

func TestRegister(t *testing.T) {
	testcases := []struct {
		name        string
		engineName  string
		factory     EngineFactory
		factories   map[string]EngineFactory
		expectPanic bool
	}{
		{
			name:        "nil factory",
			engineName:  "test",
			factory:     nil,
			expectPanic: true,
		},
		{
			name:       "engine duplicated",
			engineName: "test",
			factory:    &mockFactory{},
			factories: map[string]EngineFactory{
				"test": &mockFactory{},
			},
			expectPanic: true,
		},
		{
			name:        "engine registered",
			engineName:  "test",
			factory:     &mockFactory{},
			factories:   make(map[string]EngineFactory),
			expectPanic: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tc.expectPanic {
						t.Errorf("expected no panic, got %v", r)
					}
				}
			}()
			engineFactories = tc.factories
			Register(tc.engineName, tc.factory)
		})
	}
}

func TestCreateEngineFromConf(t *testing.T) {
	testcases := []struct {
		name         string
		config       Config
		factories    map[string]EngineFactory
		expectErr    bool
		expectEngine PolicyEngine
	}{
		{
			name: "empty engine name",
			config: Config{
				Name: "",
			},
			expectErr:    true,
			expectEngine: nil,
		},
		{
			name: "engine not found",
			config: Config{
				Name: "test",
			},
			factories:    map[string]EngineFactory{},
			expectErr:    true,
			expectEngine: nil,
		},
		{
			name: "failed creating engine",
			config: Config{
				Name: "test",
			},
			factories: map[string]EngineFactory{
				"test": &mockFactory{
					returnErr: true,
				},
			},
			expectErr:    true,
			expectEngine: nil,
		},
		{
			name: "engine created",
			config: Config{
				Name: "test",
			},
			factories: map[string]EngineFactory{
				"test": &mockFactory{},
			},
			expectErr:    false,
			expectEngine: &mockEngine{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			engineFactories = tc.factories
			engine, err := CreateEngineFromConfig(tc.config)
			if tc.expectErr != (err != nil) {
				t.Errorf("expected error %v, got %v", tc.expectErr, err)
			}

			if !reflect.DeepEqual(tc.expectEngine, engine) {
				t.Errorf("expected engine %v, got %v", tc.expectEngine, engine)
			}
		})
	}
}
