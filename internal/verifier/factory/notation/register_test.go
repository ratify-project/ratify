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
	"testing"

	"github.com/ratify-project/ratify/v2/internal/verifier/factory"
)

const testName = "notation-test"

func TestNewVerifier(t *testing.T) {
	tests := []struct {
		name      string
		opts      factory.NewVerifierOptions
		expectErr bool
	}{
		{
			name: "Unsupported params",
			opts: factory.NewVerifierOptions{
				Type:       notationType,
				Name:       testName,
				Parameters: make(chan int),
			},
			expectErr: true,
		},
		{
			name: "Malformed params",
			opts: factory.NewVerifierOptions{
				Type:       notationType,
				Name:       testName,
				Parameters: "{",
			},
			expectErr: true,
		},
		{
			name: "Missing notation params",
			opts: factory.NewVerifierOptions{
				Type:       notationType,
				Name:       testName,
				Parameters: map[string]interface{}{},
			},
			expectErr: true,
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
