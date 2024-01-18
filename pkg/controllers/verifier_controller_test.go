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
	"os"
	"path/filepath"
	"testing"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/internal/constants"
	"github.com/deislabs/ratify/pkg/utils"
	vr "github.com/deislabs/ratify/pkg/verifier"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		Name:          "sample",
		ArtifactTypes: "application/vnd.cncf.notary.signature",
		Address:       getVerifierPluginsDir(),
	}
	var resource = "sample"

	if err := verifierAddOrReplace(testVerifierSpec, resource, constants.EmptyNamespace); err != nil {
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

	if err := verifierAddOrReplace(testVerifierSpec, "testObject", constants.EmptyNamespace); err != nil {
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
	if err := verifierAddOrReplace(testVerifierSpec, resource, constants.EmptyNamespace); err != nil {
		t.Fatalf("verifierAddOrReplace() expected no error, actual %v", err)
	}
	if len(VerifierMap) != 1 {
		t.Fatalf("Verifier map expected size 1, actual %v", len(VerifierMap))
	}

	// modify the verifier
	var parametersString = "{\"allowedLicenses\":[\"MIT\",\"GNU\"]}"
	testVerifierSpec = getLicenseCheckerFromParam(parametersString)
	if err := verifierAddOrReplace(testVerifierSpec, resource, constants.EmptyNamespace); err != nil {
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
		})
	}
}

func TestGetCertStoreNamespace(t *testing.T) {
	// error scenario, everything is empty, expect error
	_, err := getCertStoreNamespace("")
	if err.Error() == "environment variable" {
		t.Fatalf("env not set should trigger an error")
	}

	ratifyDeployedNamespace := "sample"
	os.Setenv(utils.RatifyNamespaceEnvVar, ratifyDeployedNamespace)
	defer os.Unsetenv(utils.RatifyNamespaceEnvVar)

	// scenario1, when default namespace is provided, then we should expect default
	verifierNamespace := "verifierNamespace"
	ns, _ := getCertStoreNamespace(verifierNamespace)
	if ns != verifierNamespace {
		t.Fatalf("default namespace expected")
	}

	// scenario2, default is empty, should return ratify installed namespace
	ns, _ = getCertStoreNamespace("")
	if ns != ratifyDeployedNamespace {
		t.Fatalf("default namespace expected")
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
		Address:       getVerifierPluginsDir(),
		Parameters: runtime.RawExtension{
			Raw: allowedLicenses,
		},
	}
}

func getDefaultLicenseCheckerSpec() configv1beta1.VerifierSpec {
	var parametersString = "{\"allowedLicenses\":[\"MIT\",\"Apache\"]}"
	return getLicenseCheckerFromParam(parametersString)
}

func getVerifierPluginsDir() string {
	workingDir, _ := os.Getwd()
	pluginDir := filepath.Clean(filepath.Join(workingDir, "../..", "./bin/plugins"))
	return pluginDir
}
