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
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"

	// This import is required to utilize the oras built-in referrer store
	_ "github.com/deislabs/ratify/pkg/referrerstore/oras"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/deislabs/ratify/pkg/verifier/plugin/skel"

	jsonLoader "github.com/spdx/tools-golang/json"
)

// PluginConfig describes the configuration of the sbom verifier
type PluginConfig struct {
	Name string `json:"name"`
}

type PluginInputConfig struct {
	Config PluginConfig `json:"config"`
}

type PackageInfo struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"versionInfo,omitempty"`
}

const (
	SpdxJsonMediaType string = "application/spdx+json"
)

func main() {
	skel.PluginMain("sbom", "1.0.0", VerifyReference, []string{"1.0.0"})
}

func parseInput(stdin []byte) (*PluginConfig, error) {
	conf := PluginInputConfig{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse stdin for the input: %w", err)
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
		return &verifier.VerifierResult{
			Name:      input.Name,
			IsSuccess: false,
			Message:   fmt.Sprintf("Error fetching reference manifest for subject: %s reference descriptor: %v", subjectReference, referenceDescriptor.Descriptor),
		}, err
	}

	var mediaType string
	for _, blobDesc := range referenceManifest.Blobs {
		mediaType = blobDesc.MediaType
		refBlob, err := referrerStore.GetBlobContent(ctx, subjectReference, blobDesc.Digest)

		if err != nil {
			return &verifier.VerifierResult{
				Name:      input.Name,
				IsSuccess: false,
				Message:   fmt.Sprintf("Error fetching blob for subject: %s digest: %s", subjectReference, blobDesc.Digest),
			}, err
		}

		switch mediaType {
		case SpdxJsonMediaType:
			return processSpdxJsonMediaType(input.Name, refBlob)
		default:
		}
	}

	return &verifier.VerifierResult{
		Name:      input.Name,
		IsSuccess: false,
		Message:   fmt.Sprintf("Unsupported mediaType: %s", mediaType),
	}, nil
}

func processSpdxJsonMediaType(name string, refBlob []byte) (*verifier.VerifierResult, error) {
	if doc, err := jsonLoader.Read(bytes.NewReader(refBlob)); doc != nil {
		return &verifier.VerifierResult{
			Name:       name,
			IsSuccess:  true,
			Extensions: doc.CreationInfo,
			Message:    "SBOM verification success. The schema is good.",
		}, err
	} else {
		return &verifier.VerifierResult{
			Name:      name,
			IsSuccess: false,
			Message:   fmt.Sprintf("SBOM failed to parse: %v", err),
		}, err
	}
}
