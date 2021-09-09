package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/deislabs/hora/pkg/referrerstore"
	"github.com/deislabs/hora/pkg/verifier"
	"github.com/deislabs/hora/pkg/verifier/plugin/skel"
)

type PluginConfig struct {
	Name string `json:"name"`
}

type PluginInputConfig struct {
	Config PluginConfig `json:"config"`
}

type SbomContents struct {
	Contents string `json:"contents"`
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

		var content SbomContents
		if err := json.Unmarshal(refBlob, &content); err != nil {
			return nil, fmt.Errorf("failed to parse sbom: %v", err)
		}

		if content.Contents == "good" {
			return &verifier.VerifierResult{
				Name:      input.Name,
				IsSuccess: true,
				Results:   []string{fmt.Sprintf("SBOM verification completed. contents %s", content.Contents)},
			}, nil
		}
	}

	return &verifier.VerifierResult{
		Name:      input.Name,
		IsSuccess: false,
		Results:   []string{fmt.Sprintf("SBOM verification completed. verification failed.")},
	}, nil
}
