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
	"context"
	"errors"
	"reflect"
	"testing"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/pkg/policyprovider/config"
	_ "github.com/deislabs/ratify/pkg/policyprovider/configpolicy"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	policyName1 = "policy1"
	policyName2 = "policy2"
)

type mockResourceWriter struct {
	updateFailed bool
}

func (w mockResourceWriter) Create(_ context.Context, _ client.Object, _ client.Object, _ ...client.SubResourceCreateOption) error {
	return nil
}

func (w mockResourceWriter) Update(_ context.Context, _ client.Object, _ ...client.SubResourceUpdateOption) error {
	if w.updateFailed {
		return errors.New("update failed")
	}
	return nil
}

func (w mockResourceWriter) Patch(_ context.Context, _ client.Object, _ client.Patch, _ ...client.SubResourcePatchOption) error {
	return nil
}

type mockStatusClient struct {
	updateFailed bool
}

func (c mockStatusClient) Status() client.SubResourceWriter {
	writer := mockResourceWriter{}
	writer.updateFailed = c.updateFailed
	return writer
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
		spec           configv1beta1.PolicySpec
		expectErr      bool
		expectProvider bool
	}{
		{
			name:       "invalid spec",
			policyName: policyName1,
			spec: configv1beta1.PolicySpec{
				Type: policyName1,
			},
			expectErr:      true,
			expectProvider: false,
		},
		{
			name: "non-supported policy",
			spec: configv1beta1.PolicySpec{
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
			spec: configv1beta1.PolicySpec{
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
		spec       configv1beta1.PolicySpec
		policyName string
		expectErr  bool
	}{
		{
			name: "invalid spec",
			spec: configv1beta1.PolicySpec{
				Type: policyName1,
			},
			expectErr: true,
		},
		{
			name: "valid spec",
			spec: configv1beta1.PolicySpec{
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

func TestWritePolicyStatus(t *testing.T) {
	logger := logrus.WithContext(context.Background())
	testCases := []struct {
		name       string
		isSuccess  bool
		policy     *configv1beta1.Policy
		errString  string
		reconciler client.StatusClient
	}{
		{
			name:       "success status",
			isSuccess:  true,
			policy:     &configv1beta1.Policy{},
			reconciler: &mockStatusClient{},
		},
		{
			name:       "error status",
			isSuccess:  false,
			policy:     &configv1beta1.Policy{},
			errString:  "a long error string that exceeds the max length of 30 characters",
			reconciler: &mockStatusClient{},
		},
		{
			name:      "status update failed",
			isSuccess: true,
			policy:    &configv1beta1.Policy{},
			reconciler: &mockStatusClient{
				updateFailed: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			writePolicyStatus(context.Background(), tc.reconciler, tc.policy, logger, tc.isSuccess, tc.errString)
		})
	}
}
