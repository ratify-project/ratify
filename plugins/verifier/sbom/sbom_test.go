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
package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deislabs/ratify/plugins/verifier/sbom/utils"
)

func TestProcessSPDXJsonMediaType(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "bom.json"))
	if err != nil {
		t.Fatalf("error reading %s", filepath.Join("testdata", "bom.json"))
	}
	vr, err := processSpdxJSONMediaType("test", b, nil, nil)
	if err != nil {
		t.Fatalf("expected to process spdx json file: %s", filepath.Join("testdata", "bom.json"))
	}
	if !vr.IsSuccess {
		t.Fatalf("expected to successfully verify schema")
	}
}

func TestProcessInvalidSPDXJsonMediaType(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "invalid-bom.json"))
	if err != nil {
		t.Fatalf("error reading %s", filepath.Join("testdata", "invalid-bom.json"))
	}
	_, err = processSpdxJSONMediaType("test", b, nil, nil)
	if err == nil {
		t.Fatalf("expected to have an error processing spdx json file: %s", filepath.Join("testdata", "bom.json"))
	}
}

func TestFormatPackageLicense(t *testing.T) {
	bash := utils.PackageLicense{
		PackageName:    "bash",
		PackageLicense: "License",
		PackageVersion: "4.4.18-2ubuntu1.2",
	}

	testdata := []utils.PackageLicense{bash}

	result :=
		formatPackageLicense(testdata)
	if result == "nil" {
		t.Fatalf("expected to have an error processing spdx json file: %s", filepath.Join("testdata", "bom.json"))
	}
}

func TestGetViolations(t *testing.T) {
	disallowedPackage := utils.PackageInfo{
		Name:    "libcrypto3",
		Version: "3.0.7-r2",
	}

	packageViolation := utils.PackageLicense{
		PackageName:    "libcrypto3",
		PackageVersion: "3.0.7-r2",
		PackageLicense: "Apache-2.0",
	}

	licenseViolation := utils.PackageLicense{
		PackageName:    "libc-utils",
		PackageVersion: "0.7.2-r3",
		PackageLicense: "BSD-2-Clause AND LicenseRef-AND AND BSD-3-Clause",
	}

	violation2 := utils.PackageLicense{
		PackageName:    "zlib",
		PackageLicense: "Zlib",
		PackageVersion: "1.2.13-r0",
	}

	disallowedPackage2 := utils.PackageInfo{
		Name:    "libcrypto3",
		Version: "3.0.7-r3",
	}

	b, err := os.ReadFile(filepath.Join("testdata", "syftbom.spdx.json"))
	if err != nil {
		t.Fatalf("error reading %s", filepath.Join("testdata", "syftbom.spdx.json"))
	}

	cases := []struct {
		description               string
		disallowedLicenses        []string
		disallowedPackages        []utils.PackageInfo
		expectedLicenseViolations []utils.PackageLicense
		expectedPackageViolations []utils.PackageLicense
		enabled                   bool
	}{
		{
			description:               "package and license violation found",
			disallowedLicenses:        []string{"BSD-3-Clause", "Zlib"},
			disallowedPackages:        []utils.PackageInfo{disallowedPackage},
			expectedLicenseViolations: []utils.PackageLicense{licenseViolation, violation2},
			expectedPackageViolations: []utils.PackageLicense{packageViolation},
		},
		{
			description:               "license violation found",
			disallowedLicenses:        []string{"Zlib"},
			disallowedPackages:        []utils.PackageInfo{},
			expectedLicenseViolations: []utils.PackageLicense{violation2},
			expectedPackageViolations: []utils.PackageLicense{},
		},
		{
			description:               "license violation case insensitive",
			disallowedLicenses:        []string{"zlib"},
			disallowedPackages:        []utils.PackageInfo{},
			expectedLicenseViolations: []utils.PackageLicense{violation2},
			expectedPackageViolations: []utils.PackageLicense{},
		},
		{
			description:               "package violation not found",
			disallowedLicenses:        []string{},
			disallowedPackages:        []utils.PackageInfo{disallowedPackage2},
			expectedLicenseViolations: []utils.PackageLicense{},
			expectedPackageViolations: []utils.PackageLicense{},
			enabled:                   true,
		},
		{
			description:               "license violation not found",
			disallowedLicenses:        []string{"GPL-3.0-only"},
			disallowedPackages:        []utils.PackageInfo{},
			expectedLicenseViolations: []utils.PackageLicense{},
			expectedPackageViolations: []utils.PackageLicense{},
		},
	}

	for _, tc := range cases {
		t.Run("test scenario", func(t *testing.T) {
			report, err := processSpdxJSONMediaType("test", b, tc.disallowedLicenses, tc.disallowedPackages)
			if err != nil {
				t.Fatalf("unexpected error processing spdx json file: %s", filepath.Join("testdata", "bom.json"))
			}

			if !report.IsSuccess {
				violations := report.Extensions.(map[string]interface{})
				packageViolation := violations[PackageViolation].([]utils.PackageLicense)
				licensesViolation := violations[LicenseViolation].([]utils.PackageLicense)

				AssertEquals(tc.expectedPackageViolations, packageViolation, tc.description, t)
				AssertEquals(tc.expectedLicenseViolations, licensesViolation, tc.description, t)
			}
		})
	}
}

func AssertEquals(expected []utils.PackageLicense, actual []utils.PackageLicense, description string, t *testing.T) {
	if len(expected) != len(actual) {
		t.Fatalf("Test %s failed. Expected len of expectedPackageViolations %v, got: %v", description, len(expected), len(actual))
	}

	for i, packageInfo := range expected {
		if packageInfo.PackageName != actual[i].PackageName {
			t.Fatalf("Test %s failed. Expected PackageName: %s, got: %s", description, packageInfo.PackageName, actual[i].PackageName)
		}
		if packageInfo.PackageVersion != actual[i].PackageVersion {
			t.Fatalf("Test %s Failed. expected PackageVersion: %s, got: %s", description, packageInfo.PackageVersion, actual[i].PackageVersion)
		}
		if packageInfo.PackageLicense != actual[i].PackageLicense {
			t.Fatalf("Test %s Failed. expected PackageLicense: %s, got: %s", description, packageInfo.PackageLicense, actual[i].PackageLicense)
		}
	}
}
