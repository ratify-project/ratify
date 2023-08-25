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

package query

import (
	"context"
	"testing"
)

const (
	policy1 = `
  package ratify.policy
  
  default valid := false
  
  valid {
	  input.method == "GET"
  }
  `
	policy2 = "package"
)

func TestCreate(t *testing.T) {
	testcases := []struct {
		name        string
		policy      string
		expectErr   bool
		expectQuery bool
	}{
		{
			name:        "valid policy",
			policy:      policy1,
			expectErr:   false,
			expectQuery: true,
		},
		{
			name:        "invalid policy",
			policy:      policy2,
			expectErr:   true,
			expectQuery: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			factory := &RegoFactory{}
			query, err := factory.Create(tc.policy)
			if tc.expectErr != (err != nil) {
				t.Fatalf("error = %v, expectErr = %v", err, tc.expectErr)
			}
			if tc.expectQuery != (query != nil) {
				t.Fatalf("query = %v, expectQuery = %v", query, tc.expectQuery)
			}
		})
	}
}

func TestEvaluate(t *testing.T) {
	factory := &RegoFactory{}
	query, err := factory.Create(policy1)
	if err != nil {
		t.Fatalf("err = %v", err)
	}

	testcases := []struct {
		name         string
		input        map[string]interface{}
		expectResult bool
		expectErr    bool
	}{
		{
			name:         "empty input",
			input:        nil,
			expectResult: false,
			expectErr:    false,
		},
		{
			name: "input with false result",
			input: map[string]interface{}{
				"method": "POST",
			},
			expectResult: false,
			expectErr:    false,
		},
		{
			name: "valid input",
			input: map[string]interface{}{
				"method": "GET",
			},
			expectResult: true,
			expectErr:    false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := query.Evaluate(context.Background(), tc.input)
			if tc.expectErr != (err != nil) {
				t.Fatalf("error = %v, expectErr = %v", err, tc.expectErr)
			}
			if tc.expectResult != result {
				t.Fatalf("result = %v, expectResult = %v", result, tc.expectResult)
			}
		})
	}
}
