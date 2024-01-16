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
	"testing"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/internal/constants"
	"github.com/deislabs/ratify/pkg/controllers"
	"github.com/deislabs/ratify/pkg/controllers/test"
	"github.com/deislabs/ratify/pkg/customresources/verifiers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	verifierName = "verifierName"
)

func TestMain(m *testing.M) {
	// make sure to reset verifierMap before each test run
	controllers.VerifierMap = verifiers.NewActiveVerifiers()
	code := m.Run()
	os.Exit(code)
}

func TestVerifierAdd_EmptyParameter(t *testing.T) {
	resetVerifierMap()
	var testVerifierSpec = configv1beta1.VerifierSpec{
		Name:          "notation",
		ArtifactTypes: "application/vnd.cncf.notary.signature",
	}
	var resource = "notation"

	if err := verifierAddOrReplace(testVerifierSpec, resource, constants.EmptyNamespace); err != nil {
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

	var testVerifierSpec = getDefaultLicenseCheckerSpec()

	if err := verifierAddOrReplace(testVerifierSpec, "testObject", constants.EmptyNamespace); err != nil {
		t.Fatalf("verifierAddOrReplace() expected no error, actual %v", err)
	}
	if controllers.VerifierMap.GetVerifierCount() != 1 {
		t.Fatalf("Verifier map expected size 1, actual %v", controllers.VerifierMap.GetVerifierCount())
	}
}

func TestVerifier_UpdateAndDelete(t *testing.T) {
	resetVerifierMap()

	var resource = "licensechecker"
	var testVerifierSpec = getDefaultLicenseCheckerSpec()

	// add a verifier
	if err := verifierAddOrReplace(testVerifierSpec, resource, constants.EmptyNamespace); err != nil {
		t.Fatalf("verifierAddOrReplace() expected no error, actual %v", err)
	}
	if controllers.VerifierMap.GetVerifierCount() != 1 {
		t.Fatalf("Verifier map expected size 1, actual %v", controllers.VerifierMap.GetVerifierCount())
	}

	// modify the verifier
	var parametersString = "{\"allowedLicenses\":[\"MIT\",\"GNU\"]}"
	testVerifierSpec = getLicenseCheckerFromParam(parametersString)
	if err := verifierAddOrReplace(testVerifierSpec, resource, constants.EmptyNamespace); err != nil {
		t.Fatalf("verifierAddOrReplace() expected no error, actual %v", err)
	}

	// validate no verifier has been added
	if controllers.VerifierMap.GetVerifierCount() != 1 {
		t.Fatalf("Verifier map should be 1 after replacement, actual %v", controllers.VerifierMap.GetVerifierCount())
	}

	controllers.VerifierMap.DeleteVerifier(constants.EmptyNamespace, resource)

	if controllers.VerifierMap.GetVerifierCount() != 0 {
		t.Fatalf("Verifier map should be 0 after deletion, actual %v", controllers.VerifierMap.GetVerifierCount())
	}
}
func TestVerifierReconcile(t *testing.T) {
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
					Name: testNamespace,
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

func getLicenseCheckerFromParam(parametersString string) configv1beta1.VerifierSpec {
	var allowedLicenses = []byte(parametersString)

	return configv1beta1.VerifierSpec{
		Name:          "licensechecker",
		ArtifactTypes: "application/vnd.ratify.spdx.v0",
		Parameters: runtime.RawExtension{
			Raw: allowedLicenses,
		},
	}
}

func getDefaultLicenseCheckerSpec() configv1beta1.VerifierSpec {
	var parametersString = "{\"allowedLicenses\":[\"MIT\",\"Apache\"]}"
	return getLicenseCheckerFromParam(parametersString)
}
