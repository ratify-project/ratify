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
	"testing"

	"github.com/ratify-project/ratify/pkg/executor/types"
	pt "github.com/ratify-project/ratify/pkg/policyprovider/types"
)

func TestFromVerifyResult(t *testing.T) {
	result := types.VerifyResult{}
	testCases := []struct {
		name            string
		policyType      string
		expectedVersion string
	}{
		{
			name:            "Rego policy",
			policyType:      pt.RegoPolicy,
			expectedVersion: "1.1.0",
		},
		{
			name:            "Config policy",
			policyType:      pt.ConfigPolicy,
			expectedVersion: "0.2.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if res := fromVerifyResult(context.Background(), result, tc.policyType); res.Version != tc.expectedVersion {
				t.Fatalf("Expected version to be %s, got %s", tc.expectedVersion, res.Version)
			}
		})
	}
}
