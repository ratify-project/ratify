package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/deislabs/ratify/plugins/verifier/licensechecker/utils"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	_ "github.com/deislabs/ratify/pkg/referrerstore/oras"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/deislabs/ratify/pkg/verifier/plugin/skel"
)

type PluginConfig struct {
	Name            string   `json:"name"`
	AllowedLicenses []string `json:"allowedLicenses"`
}

type PluginInputConfig struct {
	Config PluginConfig `json:"config"`
}

func main() {
	skel.PluginMain("licensechecker", "1.0.0", VerifyReference, []string{"1.0.0"})
}

func parseInput(stdin []byte) (*PluginConfig, error) {
	conf := PluginInputConfig{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse stdin for input: %v", err)
	}

	return &conf.Config, nil
}

func VerifyReference(args *skel.CmdArgs, subjectReference common.Reference, descriptor ocispecs.ReferenceDescriptor, store referrerstore.ReferrerStore) (*verifier.VerifierResult, error) {
	input, err := parseInput(args.StdinData)
	if err != nil {
		return nil, err
	}

	allowedLicenses := utils.LoadAllowedLicenses(input.AllowedLicenses)

	ctx := context.Background()
	referenceManifest, err := store.GetReferenceManifest(ctx, subjectReference, descriptor)
	if err != nil {
		return nil, err
	}

	for _, blobDesc := range referenceManifest.Blobs {
		refBlob, err := store.GetBlobContent(ctx, subjectReference, blobDesc.Digest)
		if err != nil {
			return nil, err
		}

		spdxDoc, err := utils.BlobToSPDX(refBlob)
		if err != nil {
			return nil, err
		}

		packageLicenses := utils.GetPackageLicenses(*spdxDoc)
		disallowedLicenses := utils.FilterPackageLicenses(packageLicenses, allowedLicenses)

		if len(disallowedLicenses) > 0 {
			resultStrings := []string{"License Check: FAILED"}
			for _, disallowedLicense := range disallowedLicenses {
				s := fmt.Sprintf("package '%s' has unpermitted license '%s'", disallowedLicense.PackageName, disallowedLicense.PackageLicense)
				resultStrings = append(resultStrings, s)
			}
			return &verifier.VerifierResult{
				Name:      input.Name,
				IsSuccess: false,
				Results:   resultStrings,
			}, nil
		}
	}

	return &verifier.VerifierResult{
		Name:      input.Name,
		IsSuccess: true,
		Results:   []string{"License Check: SUCCESS", "All packages have allowed licenses"},
	}, nil
}
