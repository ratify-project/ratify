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
	"context"
	"encoding/json"
	"fmt"

	"github.com/ratify-project/ratify/pkg/common"
	"github.com/ratify-project/ratify/pkg/ocispecs"
	"github.com/ratify-project/ratify/pkg/referrerstore"
	_ "github.com/ratify-project/ratify/pkg/referrerstore/oras"
	"github.com/ratify-project/ratify/pkg/verifier"
	"github.com/ratify-project/ratify/pkg/verifier/plugin/skel"
	"github.com/ratify-project/ratify/plugins/verifier/schemavalidator/schemavalidation"
)

type PluginConfig struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Schemas map[string]string `json:"schemas"`
}

type PluginInputConfig struct {
	Config PluginConfig `json:"config"`
}

func main() {
	skel.PluginMain("schemavalidator", "1.0.0", VerifyReference, []string{"1.0.0"})
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
	verifierType := input.Name
	if input.Type != "" {
		verifierType = input.Type
	}
	schemaMap := input.Schemas
	ctx := context.Background()

	referenceManifest, err := referrerStore.GetReferenceManifest(ctx, subjectReference, referenceDescriptor)
	if err != nil {
		return nil, fmt.Errorf("error fetching reference manifest for subject: %s reference descriptor: %v", subjectReference, referenceDescriptor.Descriptor)
	}

	if len(referenceManifest.Blobs) == 0 {
		return &verifier.VerifierResult{
			Name:      input.Name,
			Type:      verifierType,
			IsSuccess: false,
			Message:   fmt.Sprintf("schema validation failed: no blobs found for referrer %s@%s", subjectReference.Path, referenceDescriptor.Digest.String()),
		}, nil
	}

	for _, blobDesc := range referenceManifest.Blobs {
		refBlob, err := referrerStore.GetBlobContent(ctx, subjectReference, blobDesc.Digest)
		if err != nil {
			return nil, fmt.Errorf("error fetching blob for subject:[%s] digest:[%s]", subjectReference, blobDesc.Digest)
		}

		err = processMediaType(schemaMap, blobDesc.MediaType, refBlob)
		if err != nil {
			return &verifier.VerifierResult{
				Name:      input.Name,
				Type:      verifierType,
				IsSuccess: false,
				Message:   fmt.Sprintf("schema validation failed for digest:[%s],media type:[%s],parse errors:[%v]", blobDesc.Digest, blobDesc.MediaType, err.Error()),
			}, nil
		}
	}

	return &verifier.VerifierResult{
		Name:      input.Name,
		Type:      verifierType,
		IsSuccess: true,
		Message:   "schema validation passed for configured media types",
	}, nil
}

func processMediaType(schemaMap map[string]string, mediaType string, refBlob []byte) error {
	if ok := len(schemaMap[mediaType]) > 0; ok {
		return schemavalidation.Validate(schemaMap[mediaType], refBlob)
	}
	return fmt.Errorf("media type not configured for plugin:[%s]", mediaType)
}
