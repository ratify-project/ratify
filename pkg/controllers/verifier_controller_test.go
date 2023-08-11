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
	"os"
	"testing"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	vr "github.com/deislabs/ratify/pkg/verifier"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestMain(m *testing.M) {
	// make sure to reset verifierMap before each test run
	VerifierMap = map[string]vr.ReferenceVerifier{}
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

	if err := verifierAddOrReplace(testVerifierSpec, resource); err != nil {
		t.Fatalf("verifierAddOrReplace() expected no error, actual %v", err)
	}
	if len(VerifierMap) != 1 {
		t.Fatalf("Verifier map expected size 1, actual %v", len(VerifierMap))
	}
}

func TestVerifierAdd_WithParameters(t *testing.T) {
	resetVerifierMap()
	if len(VerifierMap) != 0 {
		t.Fatalf("Verifier map expected size 0, actual %v", len(VerifierMap))
	}

	var testVerifierSpec = getDefaultLicenseCheckerSpec()

	if err := verifierAddOrReplace(testVerifierSpec, "testObject"); err != nil {
		t.Fatalf("verifierAddOrReplace() expected no error, actual %v", err)
	}
	if len(VerifierMap) != 1 {
		t.Fatalf("Verifier map expected size 1, actual %v", len(VerifierMap))
	}
}

func TestVerifier_UpdateAndDelete(t *testing.T) {
	resetVerifierMap()

	var resource = "licensechecker"
	var testVerifierSpec = getDefaultLicenseCheckerSpec()

	// add a verifier
	if err := verifierAddOrReplace(testVerifierSpec, resource); err != nil {
		t.Fatalf("verifierAddOrReplace() expected no error, actual %v", err)
	}
	if len(VerifierMap) != 1 {
		t.Fatalf("Verifier map expected size 1, actual %v", len(VerifierMap))
	}

	// modify the verifier
	var parametersString = "{\"allowedLicenses\":[\"MIT\",\"GNU\"]}"
	testVerifierSpec = getLicenseCheckerFromParam(parametersString)
	if err := verifierAddOrReplace(testVerifierSpec, resource); err != nil {
		t.Fatalf("verifierAddOrReplace() expected no error, actual %v", err)
	}

	// validate no verifier has been added
	if len(VerifierMap) != 1 {
		t.Fatalf("Verifier map should be 1 after replacement, actual %v", len(VerifierMap))
	}

	verifierRemove(resource)

	if len(VerifierMap) != 0 {
		t.Fatalf("Verifier map should be 0 after deletion, actual %v", len(VerifierMap))
	}
}

func resetVerifierMap() {
	VerifierMap = map[string]vr.ReferenceVerifier{}
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
