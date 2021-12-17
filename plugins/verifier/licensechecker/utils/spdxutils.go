package utils

import (
	"strings"

	"github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/tvloader"
)

func BlobToSPDX(bytes []byte) (*spdx.Document2_2, error) {
	raw := string(bytes)
	reader := strings.NewReader(raw)
	return tvloader.Load2_2(reader)
}

func GetPackageLicenses(doc spdx.Document2_2) []PackageLicense {
	var output []PackageLicense
	for _, p := range doc.Packages {
		output = append(output, PackageLicense{
			PackageName:    p.PackageName,
			PackageLicense: p.PackageLicenseConcluded,
		})
	}
	return output
}

func LoadAllowedLicenses(licenses []string) map[string]struct{} {
	output := map[string]struct{}{}
	for _, license := range licenses {
		output[license] = struct{}{}
	}
	return output
}

func FilterPackageLicenses(packageLicenses []PackageLicense, allowedLicenses map[string]struct{}) []PackageLicense {
	var output []PackageLicense
	for _, packageLicense := range packageLicenses {
		_, ok := allowedLicenses[packageLicense.PackageLicense]
		if !ok {
			output = append(output, packageLicense)
		}
	}
	return output
}
