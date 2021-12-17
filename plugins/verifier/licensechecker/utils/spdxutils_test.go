package utils

import "testing"

var spdxTestBytes = []byte("" +
	"SPDXVersion: SPDX-2.2\n" +
	"DataLicense: CC0-1.0\n" +
	"SPDXID: SPDXRef-DOCUMENT\n" +
	"DocumentName: localhost-5000/test-v1\n" +
	"DocumentNamespace: localhost-5000/test-v1-c0e2605b-0d32-45e2-9ed3-530611f8798e" +
	"\n" +
	"LicenseListVersion: 3.15\n" +
	"Creator: Organization: Test\n" +
	"Creator: Tool: test-0.0.0\n" +
	"Created: 2021-12-17T20:24:36Z\n" +
	"\n" +
	"##### Package: test-baselayout\n" +
	"\n" +
	"PackageName: test-baselayout\n" +
	"SPDXID: SPDXRef-Package-apk-test-baselayout\n" +
	"PackageVersion: 1.1.1-r1\n" +
	"PackageDownloadLocation: NOASSERTION\n" +
	"FilesAnalyzed: false\n" +
	"PackageLicenseConcluded: GPL-2.0-only\n" +
	"PackageLicenseDeclared: GPL-2.0-only\n" +
	"PackageCopyrightText: NOASSERTION\n")

func TestBlobToSPDX(t *testing.T) {
	spdxDoc, err := BlobToSPDX(spdxTestBytes)
	if err != nil {
		t.Fatalf("could not parse SPDX doc from bytes")
	}
	expected := "localhost-5000/test-v1"
	result := spdxDoc.CreationInfo.DocumentName
	if expected != result {
		t.Fatalf("expected: %s, got: %s", expected, result)
	}
	expectedLen := 1
	resultLen := len(spdxDoc.Packages)
	if expectedLen != resultLen {
		t.Fatalf("expected: %d, got: %d", expectedLen, resultLen)
	}
}

func TestGetPackageLicenses(t *testing.T) {
	spdxDoc, err := BlobToSPDX(spdxTestBytes)
	if err != nil {
		t.Fatalf("could not parse SPDX doc from bytes")
	}
	expected := "GPL-2.0-only"
	result := GetPackageLicenses(*spdxDoc)
	if len(result) != 1 {
		t.Fatalf("no packages parsed, expected 1")
	}
	if result[0].PackageLicense != expected {
		t.Fatalf("expected: %s, got: %s", expected, result[0].PackageLicense)
	}
}

func TestLoadAllowedLicenses(t *testing.T) {
	license := "GPL-2.0-only"
	licenses := LoadAllowedLicenses([]string{license})
	_, ok := licenses[license]
	if !ok {
		t.Fatalf("expected license but not present")
	}
}

func TestFilterPackageLicenses(t *testing.T) {
	spdxDoc, err := BlobToSPDX(spdxTestBytes)
	if err != nil {
		t.Fatalf("could not parse SPDX doc from bytes")
	}
	PackageLicenses := GetPackageLicenses(*spdxDoc)
	filterLicenses := LoadAllowedLicenses([]string{"GPL-2.0-only"})
	result := FilterPackageLicenses(PackageLicenses, filterLicenses)
	if len(result) != 0 {
		t.Fatalf("License not filtered correctly")
	}
}
