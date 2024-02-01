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
	"os"
	"strings"
	"testing"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/pkg/controllers"
	"github.com/deislabs/ratify/pkg/customresources/verifiers"
	test "github.com/deislabs/ratify/pkg/utils"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	notFoundErr   = "notFound"
	failure       = "failure"
	invalidConfig = "invalidConfig"
	success       = "success"
	verifierName  = "verifierName"
	sampleName    = "sampleName"
)

func TestMain(m *testing.M) {
	// make sure to reset verifierMap before each test run
	controllers.VerifierMap = verifiers.NewActiveVerifiers()
	code := m.Run()
	os.Exit(code)
}

func TestVerifierAdd_EmptyParameter(t *testing.T) {
	resetVerifierMap()
	dirPath, err := test.CreatePlugin(sampleName)
	if err != nil {
		t.Fatalf("createPlugin() expected no error, actual %v", err)
	}
	defer os.RemoveAll(dirPath)

	var testVerifierSpec = configv1beta1.ClusterVerifierSpec{
		Name:          sampleName,
		ArtifactTypes: "application/vnd.cncf.notary.signature",
		Address:       dirPath,
	}

	if err := verifierAddOrReplace(testVerifierSpec, sampleName); err != nil {
		t.Fatalf("verifierAddOrReplace() expected no error, actual %v", err)
	}
	if controllers.VerifierMap.GetVerifierCount() != 1 {
		t.Fatalf("Verifier map expected size 1, actual %v", controllers.VerifierMap.GetVerifierCount())
	}
}

func TestVerifierAddOrReplace_PluginNotFound(t *testing.T) {
	resetVerifierMap()
	var resource = "invalidplugin"
	expectedMsg := "plugin not found"
	var testVerifierSpec = getInvalidVerifierSpec()
	err := verifierAddOrReplace(testVerifierSpec, resource)

	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("TestVerifierAddOrReplace_PluginNotFound expected msg: '%v', actual %v", expectedMsg, err.Error())
	}
}

func TestVerifierAdd_InvalidConfig(t *testing.T) {
	resetVerifierMap()
	var testVerifierSpec = configv1beta1.ClusterVerifierSpec{
		Name:          "notation",
		ArtifactTypes: "application/vnd.cncf.notary.signature",
		Parameters: runtime.RawExtension{
			Raw: []byte("test\n"),
		},
	}
	var resource = "notation"

	if err := verifierAddOrReplace(testVerifierSpec, resource); err == nil {
		t.Fatalf("Expected an error but returned nil")
	}
}

func TestWriteVerifierStatus(t *testing.T) {
	logger := logrus.WithContext(context.Background())
	testCases := []struct {
		name       string
		isSuccess  bool
		verifier   *configv1beta1.ClusterVerifier
		errString  string
		reconciler client.StatusClient
	}{
		{
			name:       "success status",
			isSuccess:  true,
			errString:  "",
			verifier:   &configv1beta1.ClusterVerifier{},
			reconciler: &mockStatusClient{},
		},
		{
			name:       "error status",
			isSuccess:  false,
			verifier:   &configv1beta1.ClusterVerifier{},
			errString:  "a long error string that exceeds the max length of 30 characters",
			reconciler: &mockStatusClient{},
		},
		{
			name:      "status update failed",
			isSuccess: true,
			verifier:  &configv1beta1.ClusterVerifier{},
			reconciler: &mockStatusClient{
				updateFailed: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			writeVerifierStatus(context.Background(), tc.reconciler, tc.verifier, logger, tc.isSuccess, tc.errString)

			if tc.verifier.Status.IsSuccess != tc.isSuccess {
				t.Fatalf("Expected isSuccess to be %+v , actual %+v", tc.isSuccess, tc.verifier.Status.IsSuccess)
			}

			if tc.verifier.Status.Error != tc.errString {
				t.Fatalf("Expected Error to be %+v , actual %+v", tc.errString, tc.verifier.Status.Error)
			}
		})
	}
}

func TestVerifierReconcile(t *testing.T) {
	dirPath, err := test.CreatePlugin(sampleName)
	if err != nil {
		t.Fatalf("createPlugin() expected no error, actual %v", err)
	}
	defer os.RemoveAll(dirPath)

	tests := []struct {
		name                  string
		verifier              *configv1beta1.ClusterVerifier
		req                   *reconcile.Request
		expectedErr           bool
		expectedVerifierCount int
	}{
		{
			name: "nonexistent verifier",
			req: &reconcile.Request{
				NamespacedName: types.NamespacedName{Name: "nonexistent"},
			},
			verifier: &configv1beta1.ClusterVerifier{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "",
					Name:      verifierName,
				},
			},
			expectedErr:           false,
			expectedVerifierCount: 0,
		},
		{
			name: "invalid parameters",
			verifier: &configv1beta1.ClusterVerifier{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "",
					Name:      verifierName,
				},
				Spec: configv1beta1.ClusterVerifierSpec{
					Parameters: runtime.RawExtension{
						Raw: []byte("test"),
					},
				},
			},
			expectedErr:           true,
			expectedVerifierCount: 0,
		},
		{
			name: "valid spec",
			verifier: &configv1beta1.ClusterVerifier{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "",
					Name:      verifierName,
				},
				Spec: configv1beta1.ClusterVerifierSpec{
					Name:    sampleName,
					Address: dirPath,
				},
			},
			expectedErr:           false,
			expectedVerifierCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetVerifierMap()
			scheme, _ := test.CreateScheme()
			client := fake.NewClientBuilder().WithScheme(scheme)
			client.WithObjects(tt.verifier)
			r := &VerifierReconciler{
				Scheme: scheme,
				Client: client.Build(),
			}
			var req reconcile.Request
			if tt.req != nil {
				req = *tt.req
			} else {
				req = reconcile.Request{
					NamespacedName: test.KeyFor(tt.verifier),
				}
			}

			_, err := r.Reconcile(context.Background(), req)
			if tt.expectedErr != (err != nil) {
				t.Fatalf("Reconcile() expected error %v, actual %v", tt.expectedErr, err)
			}
			if controllers.VerifierMap.GetVerifierCount() != tt.expectedVerifierCount {
				t.Fatalf("Verifier map expected size %v, actual %v", tt.expectedVerifierCount, controllers.StoreMap.GetStoreCount())
			}
		})
	}
}

func resetVerifierMap() {
	controllers.VerifierMap = verifiers.NewActiveVerifiers()
}

func getInvalidVerifierSpec() configv1beta1.ClusterVerifierSpec {
	return configv1beta1.ClusterVerifierSpec{
		Name:          "pluginnotfound",
		ArtifactTypes: "application/vnd.ratify.spdx.v0",
		Address:       "test/path",
	}
}
