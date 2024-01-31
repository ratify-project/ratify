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

package namespaced

import (
	"context"
	"os"
	"strings"
	"testing"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/internal/constants"
	"github.com/deislabs/ratify/pkg/controllers"
	"github.com/deislabs/ratify/pkg/controllers/test"
	"github.com/deislabs/ratify/pkg/customresources/verifiers"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	verifierName   = "verifierName"
	licenseChecker = "licensechecker"
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

	var testVerifierSpec = configv1beta1.VerifierSpec{
		Name:          sampleName,
		ArtifactTypes: "application/vnd.cncf.notary.signature",
		Address:       dirPath,
	}

	if err := verifierAddOrReplace(testVerifierSpec, sampleName, constants.EmptyNamespace); err != nil {
		t.Fatalf("verifierAddOrReplace() expected no error, actual %v", err)
	}
	if controllers.VerifierMap.GetVerifierCount() != 1 {
		t.Fatalf("Verifier map expected size 1, actual %v", controllers.VerifierMap.GetVerifierCount())
	}
}

func TestVerifierAdd_InvalidParameter(t *testing.T) {
	resetVerifierMap()
	var testVerifierSpec = configv1beta1.VerifierSpec{
		Name:          "notation",
		ArtifactTypes: "application/vnd.cncf.notary.signature",
		Parameters: runtime.RawExtension{
			Raw: []byte("test"),
		},
	}
	var resource = "notation"

	if err := verifierAddOrReplace(testVerifierSpec, resource, constants.EmptyNamespace); err == nil {
		t.Fatalf("verifierAddOrReplace() expected error, actual nil")
	}
	if controllers.VerifierMap.GetVerifierCount() != 0 {
		t.Fatalf("Verifier map expected size 0, actual %v", controllers.VerifierMap.GetVerifierCount())
	}
}

func TestVerifierAdd_WithParameters(t *testing.T) {
	resetVerifierMap()
	if controllers.VerifierMap.GetVerifierCount() != 0 {
		t.Fatalf("Verifier map expected size 0, actual %v", controllers.VerifierMap.GetVerifierCount())
	}
	dirPath, err := test.CreatePlugin(licenseChecker)
	if err != nil {
		t.Fatalf("createPlugin() expected no error, actual %v", err)
	}
	defer os.RemoveAll(dirPath)

	var testVerifierSpec = getDefaultLicenseCheckerSpec(dirPath)

	if err := verifierAddOrReplace(testVerifierSpec, "testObject", constants.EmptyNamespace); err != nil {
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
	err := verifierAddOrReplace(testVerifierSpec, resource, constants.EmptyNamespace)

	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("TestVerifierAddOrReplace_PluginNotFound expected msg: '%v', actual %v", expectedMsg, err.Error())
	}
}

func TestVerifier_UpdateAndDelete(t *testing.T) {
	resetVerifierMap()
	dirPath, err := test.CreatePlugin(licenseChecker)
	if err != nil {
		t.Fatalf("createPlugin() expected no error, actual %v", err)
	}
	defer os.RemoveAll(dirPath)

	testVerifierSpec := getDefaultLicenseCheckerSpec(dirPath)

	// add a verifier
	if err := verifierAddOrReplace(testVerifierSpec, licenseChecker, constants.EmptyNamespace); err != nil {
		t.Fatalf("verifierAddOrReplace() expected no error, actual %v", err)
	}
	if controllers.VerifierMap.GetVerifierCount() != 1 {
		t.Fatalf("Verifier map expected size 1, actual %v", controllers.VerifierMap.GetVerifierCount())
	}

	// modify the verifier
	var parametersString = "{\"allowedLicenses\":[\"MIT\",\"GNU\"]}"
	testVerifierSpec = getLicenseCheckerFromParam(parametersString, dirPath)
	if err := verifierAddOrReplace(testVerifierSpec, licenseChecker, constants.EmptyNamespace); err != nil {
		t.Fatalf("verifierAddOrReplace() expected no error, actual %v", err)
	}

	// validate no verifier has been added
	if controllers.VerifierMap.GetVerifierCount() != 1 {
		t.Fatalf("Verifier map should be 1 after replacement, actual %v", controllers.VerifierMap.GetVerifierCount())
	}

	controllers.VerifierMap.DeleteVerifier(constants.EmptyNamespace, licenseChecker)

	if controllers.VerifierMap.GetVerifierCount() != 0 {
		t.Fatalf("Verifier map should be 0 after deletion, actual %v", controllers.VerifierMap.GetVerifierCount())
	}
}

func TestWriteVerifierStatus(t *testing.T) {
	logger := logrus.WithContext(context.Background())
	testCases := []struct {
		name       string
		isSuccess  bool
		verifier   *configv1beta1.Verifier
		errString  string
		reconciler client.StatusClient
	}{
		{
			name:       "success status",
			isSuccess:  true,
			errString:  "",
			verifier:   &configv1beta1.Verifier{},
			reconciler: &mockStatusClient{},
		},
		{
			name:       "error status",
			isSuccess:  false,
			verifier:   &configv1beta1.Verifier{},
			errString:  "a long error string that exceeds the max length of 30 characters",
			reconciler: &mockStatusClient{},
		},
		{
			name:      "status update failed",
			isSuccess: true,
			verifier:  &configv1beta1.Verifier{},
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
		verifier              *configv1beta1.Verifier
		req                   *reconcile.Request
		expectedErr           bool
		expectedVerifierCount int
	}{
		{
			name: "nonexistent verifier",
			req: &reconcile.Request{
				NamespacedName: types.NamespacedName{Name: "nonexistent"},
			},
			verifier: &configv1beta1.Verifier{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testNamespace,
					Name:      verifierName,
				},
			},
			expectedErr:           false,
			expectedVerifierCount: 0,
		},
		{
			name: "invalid parameters",
			verifier: &configv1beta1.Verifier{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testNamespace,
					Name:      verifierName,
				},
				Spec: configv1beta1.VerifierSpec{
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
			verifier: &configv1beta1.Verifier{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testNamespace,
					Name:      verifierName,
				},
				Spec: configv1beta1.VerifierSpec{
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

func TestVerifierSetupWithManager(t *testing.T) {
	scheme, err := test.CreateScheme()
	if err != nil {
		t.Fatalf("CreateScheme() expected no error, actual %v", err)
	}
	client := fake.NewClientBuilder().WithScheme(scheme)
	r := &VerifierReconciler{
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

func resetVerifierMap() {
	controllers.VerifierMap = verifiers.NewActiveVerifiers()
}

func getLicenseCheckerFromParam(parametersString, pluginPath string) configv1beta1.VerifierSpec {
	var allowedLicenses = []byte(parametersString)

	return configv1beta1.VerifierSpec{
		Name:          licenseChecker,
		ArtifactTypes: "application/vnd.ratify.spdx.v0",
		Address:       pluginPath,
		Parameters: runtime.RawExtension{
			Raw: allowedLicenses,
		},
	}
}

func getDefaultLicenseCheckerSpec(pluginPath string) configv1beta1.VerifierSpec {
	var parametersString = "{\"allowedLicenses\":[\"MIT\",\"Apache\"]}"
	return getLicenseCheckerFromParam(parametersString, pluginPath)
}

func getInvalidVerifierSpec() configv1beta1.VerifierSpec {
	return configv1beta1.VerifierSpec{
		Name:          "pluginnotfound",
		ArtifactTypes: "application/vnd.ratify.spdx.v0",
		Address:       "test/path",
	}
}
