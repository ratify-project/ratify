package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	// This import is required to utilize the oras built-in referrer store
	_ "github.com/deislabs/ratify/pkg/referrerstore/oras"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/deislabs/ratify/pkg/verifier/plugin/skel"
)

type PluginConfig struct {
	Name             string `json:"name"`
	AlpineMinVersion string `json:"alpineMinVersion"`
}

type PluginInputConfig struct {
	Config PluginConfig `json:"config"`
}

type PackageInfo struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"versionInfo,omitempty"`
}

type SbomContents struct {
	Contents string        `json:"contents"`
	Packages []PackageInfo `json:"packages,omitempty"`
}

func main() {
	skel.PluginMain("sbom", "1.0.0", VerifyReference, []string{"1.0.0"})
}

func parseInput(stdin []byte) (*PluginConfig, error) {
	conf := PluginInputConfig{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse stdin for the input: %v", err)
	}

	return &conf.Config, nil
}

func VerifyReference(args *skel.CmdArgs, subjectReference common.Reference, referenceDescriptor ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) (*verifier.VerifierResult, error) {
	input, err := parseInput(args.StdinData)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	referenceManifest, err := referrerStore.GetReferenceManifest(ctx, subjectReference, referenceDescriptor)

	if err != nil {
		return nil, err
	}

	for _, blobDesc := range referenceManifest.Blobs {
		refBlob, err := referrerStore.GetBlobContent(ctx, subjectReference, blobDesc.Digest)
		if err != nil {
			return nil, err
		}

		var sbomBlob SbomContents
		if err := json.Unmarshal(refBlob, &sbomBlob); err != nil {
			return nil, fmt.Errorf("failed to parse sbom: %v", err)
		}

		for _, p := range sbomBlob.Packages {
			if strings.HasPrefix(p.Name, "alpine-base") && p.Version < input.AlpineMinVersion {
				return &verifier.VerifierResult{
					Name:      input.Name,
					IsSuccess: false,
					Results:   []string{fmt.Sprintf("SBOM verification failed. The artifact has base image 'alpine' with version below '%s' and not compliant.", input.AlpineMinVersion)},
				}, nil
			}
		}
	}

	return &verifier.VerifierResult{
		Name:      input.Name,
		IsSuccess: true,
		Results:   []string{"SBOM verification success."},
	}, nil

}
