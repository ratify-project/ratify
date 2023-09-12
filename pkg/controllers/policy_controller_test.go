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

package controllers

import (
	"reflect"
	"testing"

	configv1alpha1 "github.com/deislabs/ratify/api/v1alpha1"
	"github.com/deislabs/ratify/pkg/policyprovider/config"
	_ "github.com/deislabs/ratify/pkg/policyprovider/configpolicy"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	policyName1 = "policy1"
	policyName2 = "policy2"
)

func TestDeletePolicy(t *testing.T) {
	testCases := []struct {
		name             string
		policyName       string
		expectPolicyName string
	}{
		{
			name:             "Delete same name",
			policyName:       policyName1,
			expectPolicyName: "",
		},
		{
			name:             "Delete different name",
			policyName:       policyName2,
			expectPolicyName: policyName1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policy := &policy{
				Name: policyName1,
			}
			policy.deletePolicy(tc.policyName)
			if policy.Name != tc.expectPolicyName {
				t.Fatalf("Expected policy name to be %s, got %s", tc.expectPolicyName, policy.Name)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	testCases := []struct {
		name   string
		policy *policy
		expect bool
	}{
		{
			name:   "Empty policy",
			policy: &policy{},
			expect: true,
		},
		{
			name: "Non-empty policy",
			policy: &policy{
				Name: policyName1,
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isEmpty := tc.policy.IsEmpty()
			if isEmpty != tc.expect {
				t.Fatalf("Expected to be %t, got %t", tc.expect, isEmpty)
			}
		})
	}
}

func TestRawToPolicyConfig(t *testing.T) {
	testCases := []struct {
		name         string
		raw          []byte
		expectErr    bool
		expectConfig config.PoliciesConfig
	}{
		{
			name:         "empty Raw",
			raw:          []byte{},
			expectErr:    true,
			expectConfig: config.PoliciesConfig{},
		},
		{
			name:         "unmarshal failure",
			raw:          []byte("invalid"),
			expectErr:    true,
			expectConfig: config.PoliciesConfig{},
		},
		{
			name:      "valid Raw",
			raw:       []byte("{\"name\": \"policy1\"}"),
			expectErr: false,
			expectConfig: config.PoliciesConfig{
				PolicyPlugin: config.PolicyPluginConfig{
					"name": policyName1,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := rawToPolicyConfig(tc.raw, policyName1)

			if tc.expectErr != (err != nil) {
				t.Fatalf("Expected error to be %t, got %t", tc.expectErr, err != nil)
			}
			if !reflect.DeepEqual(config, tc.expectConfig) {
				t.Fatalf("Expected config to be %v, got %v", tc.expectConfig, config)
			}
		})
	}
}

func TestSpecToPolicyEnforcer(t *testing.T) {
	testCases := []struct {
		name           string
		policyName     string
		spec           configv1alpha1.PolicySpec
		expectErr      bool
		expectProvider bool
	}{
		{
			name:       "invalid spec",
			policyName: policyName1,
			spec: configv1alpha1.PolicySpec{
				Type: policyName1,
			},
			expectErr:      true,
			expectProvider: false,
		},
		{
			name: "non-supported policy",
			spec: configv1alpha1.PolicySpec{
				Parameters: runtime.RawExtension{
					Raw: []byte("{\"name\": \"policy1\"}"),
				},
				Type: policyName1,
			},
			expectErr:      true,
			expectProvider: false,
		},
		{
			name: "valid spec",
			spec: configv1alpha1.PolicySpec{
				Parameters: runtime.RawExtension{
					Raw: []byte("{\"name\": \"configpolicy\"}"),
				},
				Type: "configpolicy",
			},
			expectErr:      false,
			expectProvider: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider, err := specToPolicyEnforcer(tc.spec)

			if tc.expectErr != (err != nil) {
				t.Fatalf("Expected error to be %t, got %t", tc.expectErr, err != nil)
			}
			if tc.expectProvider != (provider != nil) {
				t.Fatalf("expected provider to be %t, got %t", tc.expectProvider, provider != nil)
			}
		})
	}
}

func TestPolicyAddOrReplace(t *testing.T) {
	testCases := []struct {
		name       string
		spec       configv1alpha1.PolicySpec
		policyName string
		expectErr  bool
	}{
		{
			name: "invalid spec",
			spec: configv1alpha1.PolicySpec{
				Type: policyName1,
			},
			expectErr: true,
		},
		{
			name: "valid spec",
			spec: configv1alpha1.PolicySpec{
				Parameters: runtime.RawExtension{
					Raw: []byte("{\"name\": \"configpolicy\"}"),
				},
				Type: "configpolicy",
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := policyAddOrReplace(tc.spec)

			if tc.expectErr != (err != nil) {
				t.Fatalf("Expected error to be %t, got %t", tc.expectErr, err != nil)
			}
		})
	}
}
