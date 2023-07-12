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

package policyquery

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type mockQuery struct{}

func (q *mockQuery) Evaluate(_ context.Context, _ map[string]interface{}) (bool, error) {
	return true, nil
}

type mockFactory struct {
	returnErr bool
}

func (f *mockFactory) Create(_ string) (PolicyQuery, error) {
	if f.returnErr {
		return nil, errors.New("error")
	}
	return &mockQuery{}, nil
}

func TestRegister(t *testing.T) {
	testcases := []struct {
		name        string
		queryName   string
		factory     Factory
		factories   map[string]Factory
		expectPanic bool
	}{
		{
			name:        "nil fatory",
			queryName:   "test",
			factory:     nil,
			expectPanic: true,
		},
		{
			name:      "factory duplicated",
			queryName: "test",
			factory:   &mockFactory{},
			factories: map[string]Factory{
				"test": &mockFactory{},
			},
			expectPanic: true,
		},
		{
			name:        "factory registered",
			queryName:   "test",
			factory:     &mockFactory{},
			factories:   map[string]Factory{},
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
			policyQueryFactories = tc.factories
			Register(tc.queryName, tc.factory)
		})
	}
}

func TestCreateQueryFromConfig(t *testing.T) {
	testcases := []struct {
		name        string
		config      Config
		factories   map[string]Factory
		expectErr   bool
		expectQuery PolicyQuery
	}{
		{
			name: "empty query name",
			config: Config{
				Name: "",
			},
			expectErr:   true,
			expectQuery: nil,
		},
		{
			name: "query not found",
			config: Config{
				Name: "test",
			},
			factories:   map[string]Factory{},
			expectErr:   true,
			expectQuery: nil,
		},
		{
			name: "failed creating query",
			config: Config{
				Name: "test",
			},
			factories: map[string]Factory{
				"test": &mockFactory{
					returnErr: true,
				},
			},
			expectErr:   true,
			expectQuery: nil,
		},
		{
			name: "query created",
			config: Config{
				Name: "test",
			},
			factories: map[string]Factory{
				"test": &mockFactory{},
			},
			expectErr:   false,
			expectQuery: &mockQuery{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			policyQueryFactories = tc.factories
			query, err := CreateQueryFromConfig(tc.config)
			if tc.expectErr != (err != nil) {
				t.Errorf("expected error %v, got %v", tc.expectErr, err)
			}

			if !reflect.DeepEqual(tc.expectQuery, query) {
				t.Errorf("expected query %v, got %v", tc.expectQuery, query)
			}
		})
	}
}
