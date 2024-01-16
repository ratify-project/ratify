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

package clustered

import (
	"context"
	"errors"
	"testing"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/pkg/controllers"
	"github.com/deislabs/ratify/pkg/controllers/test"
	"github.com/deislabs/ratify/pkg/customresources/policies"
	_ "github.com/deislabs/ratify/pkg/policyprovider/configpolicy"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const policyName1 = "policyName1"

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

func TestPolicyReconcile(t *testing.T) {
	tests := []struct {
		name                string
		policy              *configv1beta1.ClusterPolicy
		req                 *reconcile.Request
		expectedErr         bool
		expectedPolicyCount int
	}{
		{
			name: "nonexistent policy",
			req: &reconcile.Request{
				NamespacedName: types.NamespacedName{Name: "nonexistent"},
			},
			policy: &configv1beta1.ClusterPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "",
					Name:      policyName1,
				},
			},
			expectedErr:         false,
			expectedPolicyCount: 0,
		},
		{
			name: "no policy parameters provided",
			policy: &configv1beta1.ClusterPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "",
					Name:      "ratify-policy",
				},
				Spec: configv1beta1.ClusterPolicySpec{
					Type: "regopolicy",
				},
			},
			expectedErr:         true,
			expectedPolicyCount: 0,
		},
		{
			name: "wrong policy name",
			policy: &configv1beta1.ClusterPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "",
					Name:      "ratify-policy2",
				},
				Spec: configv1beta1.ClusterPolicySpec{
					Type: "regopolicy",
				},
			},
			expectedErr:         false,
			expectedPolicyCount: 0,
		},
		{
			name: "invalid params",
			policy: &configv1beta1.ClusterPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "",
					Name:      "ratify-policy",
				},
				Spec: configv1beta1.ClusterPolicySpec{
					Type: "regopolicy",
					Parameters: runtime.RawExtension{
						Raw: []byte("test"),
					},
				},
			},
			expectedErr:         true,
			expectedPolicyCount: 0,
		},
		{
			name: "valid params",
			policy: &configv1beta1.ClusterPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "",
					Name:      "ratify-policy",
				},
				Spec: configv1beta1.ClusterPolicySpec{
					Type: "configpolicy",
					Parameters: runtime.RawExtension{
						Raw: []byte("{\"passthroughEnabled:\": false}"),
					},
				},
			},
			expectedErr:         false,
			expectedPolicyCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetPolicyMap()
			scheme, err := test.CreateScheme()
			if err != nil {
				t.Fatalf("CreateScheme() expected no error, actual %v", err)
			}
			client := fake.NewClientBuilder().WithScheme(scheme)
			client.WithObjects(tt.policy)
			r := &PolicyReconciler{
				Scheme: scheme,
				Client: client.Build(),
			}
			var req reconcile.Request
			if tt.req != nil {
				req = *tt.req
			} else {
				req = reconcile.Request{
					NamespacedName: test.KeyFor(tt.policy),
				}
			}
			_, err = r.Reconcile(context.Background(), req)
			if tt.expectedErr != (err != nil) {
				t.Fatalf("Reconcile() expected error to be %t, actual %t", tt.expectedErr, err != nil)
			}
			if len(controllers.ActivePolicies.NamespacedPolicies) != tt.expectedPolicyCount {
				t.Fatalf("Expected policy count to be %d, actual %d", tt.expectedPolicyCount, len(controllers.ActivePolicies.NamespacedPolicies))
			}
		})
	}
}

func TestPolicySetupWithManager(t *testing.T) {
	scheme, err := test.CreateScheme()
	if err != nil {
		t.Fatalf("CreateScheme() expected no error, actual %v", err)
	}
	client := fake.NewClientBuilder().WithScheme(scheme)
	r := &PolicyReconciler{
		Scheme: scheme,
		Client: client.Build(),
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		t.Fatalf("NewManager() expected no error, actual %v", err)
	}

	if err := r.SetupWithManager(mgr); err != nil {
		t.Fatalf("SetupWithManager() expected no error, actual %v", err)
	}
}

func TestWritePolicyStatus(t *testing.T) {
	logger := logrus.WithContext(context.Background())
	testCases := []struct {
		name       string
		isSuccess  bool
		policy     *configv1beta1.ClusterPolicy
		errString  string
		reconciler client.StatusClient
	}{
		{
			name:       "success status",
			isSuccess:  true,
			policy:     &configv1beta1.ClusterPolicy{},
			reconciler: &mockStatusClient{},
		},
		{
			name:       "error status",
			isSuccess:  false,
			policy:     &configv1beta1.ClusterPolicy{},
			errString:  "a long error string that exceeds the max length of 30 characters",
			reconciler: &mockStatusClient{},
		},
		{
			name:      "status update failed",
			isSuccess: true,
			policy:    &configv1beta1.ClusterPolicy{},
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

func resetPolicyMap() {
	controllers.ActivePolicies = policies.NewActivePolicies()
}
