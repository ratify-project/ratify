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
	"strings"
	"testing"

	"github.com/ratify-project/ratify/plugins/verifier/sbom/utils"
)

func TestProcessSPDXJsonMediaType(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "bom.json"))
	if err != nil {
		t.Fatalf("error reading %s", filepath.Join("testdata", "bom.json"))
	}
	vr := processSpdxJSONMediaType("test", "", b, nil, nil)
	if !vr.IsSuccess {
		t.Fatalf("expected to successfully verify schema")
	}
}

func TestProcessInvalidSPDXJsonMediaType(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "invalid-bom.json"))
	if err != nil {
		t.Fatalf("error reading %s", filepath.Join("testdata", "invalid-bom.json"))
	}
	report := processSpdxJSONMediaType("test", "", b, nil, nil)

	if !strings.Contains(report.Message, "SBOM failed to parse") {
		t.Fatalf("expected to have an error processing spdx json file: %s", filepath.Join("testdata", "bom.json"))
	}
}

func TestGetViolations(t *testing.T) {
	disallowedPackage := utils.PackageInfo{
		Name:    "libcrypto3",
		Version: "3.0.7-r2",
	}

	disallowedPackageNoName := utils.PackageInfo{
		Name:    "",
		Version: "3.0.7-r2",
	}

	disallowedPackageNoVersion := utils.PackageInfo{
		Name: "libcrypto3",
	}

	packageViolation := utils.PackageLicense{
		Name:    "libcrypto3",
		Version: "3.0.7-r2",
		License: "Apache-2.0",
	}

	licenseViolation := utils.PackageLicense{
		Name:    "libc-utils",
		Version: "0.7.2-r3",
		License: "BSD-2-Clause AND LicenseRef-AND AND BSD-3-Clause",
	}

	violation2 := utils.PackageLicense{
		Name:    "zlib",
		License: "Zlib",
		Version: "1.2.13-r0",
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
			description:               "disallowed packages with no version",
			disallowedLicenses:        []string{"MPL"},
			expectedLicenseViolations: nil,
			expectedPackageViolations: nil,
		},
		{
			description:               "disallowed packages with no version",
			disallowedPackages:        []utils.PackageInfo{disallowedPackageNoVersion},
			expectedLicenseViolations: []utils.PackageLicense{},
			expectedPackageViolations: []utils.PackageLicense{packageViolation},
		},
		{
			description:               "package and license violation found",
			disallowedLicenses:        []string{"BSD-3-Clause", "Zlib"},
			disallowedPackages:        []utils.PackageInfo{disallowedPackage},
			expectedLicenseViolations: []utils.PackageLicense{licenseViolation, violation2},
			expectedPackageViolations: []utils.PackageLicense{packageViolation},
		},
		{
			description:               "invalid disallow package",
			disallowedPackages:        []utils.PackageInfo{disallowedPackageNoName},
			expectedLicenseViolations: []utils.PackageLicense{},
			expectedPackageViolations: []utils.PackageLicense{},
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
			report := processSpdxJSONMediaType("test", "", b, tc.disallowedLicenses, tc.disallowedPackages)

			if len(tc.expectedPackageViolations) != 0 || len(tc.expectedLicenseViolations) != 0 {
				if report.IsSuccess {
					t.Fatalf("Test %s failed. Expected IsSuccess: true, got: false", tc.description)
				}
			}

			if len(tc.expectedPackageViolations) == 0 && len(tc.expectedLicenseViolations) == 0 {
				if !report.IsSuccess {
					t.Fatalf("Test %s failed. Expected IsSuccess: false, got: true", tc.description)
				}
			}

			if len(tc.expectedPackageViolations) != 0 {
				extensionData := report.Extensions.(map[string]interface{})
				packageViolation := extensionData[PackageViolation].([]utils.PackageLicense)
				AssertEquals(tc.expectedPackageViolations, packageViolation, tc.description, t)
			}

			if len(tc.expectedLicenseViolations) != 0 {
				extensionData := report.Extensions.(map[string]interface{})
				licensesViolation := extensionData[LicenseViolation].([]utils.PackageLicense)
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
		if packageInfo.Name != actual[i].Name {
			t.Fatalf("Test %s failed. Expected PackageName: %s, got: %s", description, packageInfo.Name, actual[i].Name)
		}
		if packageInfo.Version != actual[i].Version {
			t.Fatalf("Test %s Failed. expected PackageVersion: %s, got: %s", description, packageInfo.Version, actual[i].Version)
		}
		if packageInfo.License != actual[i].License {
			t.Fatalf("Test %s Failed. expected PackageLicense: %s, got: %s", description, packageInfo.License, actual[i].License)
		}
	}
}
