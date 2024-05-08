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

package context

import (
	"context"
	"testing"
)

const (
	testNamespace = "testNamespace"
	testKey       = "testKey"
)

func TestSetContext(t *testing.T) {
	ctx := context.Background()
	ctx = SetContextWithNamespace(ctx, testNamespace)
	namespace := ctx.Value(ContextKeyNamespace).(string)
	if namespace != testNamespace {
		t.Fatalf("expected namespace %s, got %s", testNamespace, namespace)
	}
}

func TestCreateCacheKey(t *testing.T) {
	testCases := []struct {
		name         string
		key          string
		namespaceSet bool
		namespace    string
		expectedKey  string
	}{
		{
			name:         "no namespace",
			key:          testKey,
			namespaceSet: false,
			expectedKey:  testKey,
		},
		{
			name:         "empty namespace",
			key:          testKey,
			namespaceSet: true,
			namespace:    "",
			expectedKey:  testKey,
		},
		{
			name:         "with namespace",
			key:          testKey,
			namespaceSet: true,
			namespace:    testNamespace,
			expectedKey:  "testNamespace:testKey",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.namespaceSet {
				ctx = SetContextWithNamespace(ctx, tc.namespace)
			}

			key := CreateCacheKey(ctx, tc.key)
			if key != tc.expectedKey {
				t.Fatalf("expected key %s, got %s", tc.expectedKey, key)
			}
		})
	}
}

func TestGetNamespace(t *testing.T) {
	testCases := []struct {
		name         string
		namespaceSet bool
		namespace    string
	}{
		{
			name:         "no namespace",
			namespaceSet: false,
		},
		{
			name:         "empty namespace",
			namespaceSet: true,
			namespace:    "",
		},
		{
			name:         "with namespace",
			namespaceSet: true,
			namespace:    testNamespace,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.namespaceSet {
				ctx = SetContextWithNamespace(ctx, tc.namespace)
			}

			namespace := GetNamespace(ctx)
			if namespace != tc.namespace {
				t.Fatalf("expected namespace %s, got %s", tc.namespace, namespace)
			}
		})
	}
}
