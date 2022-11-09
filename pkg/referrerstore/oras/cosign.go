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

package oras

import (
	"net/http"
	"strings"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

const CosignArtifactType = "org.sigstore.cosign.v1"
const CosignSignatureTagSuffix = ".sig"

func getCosignReferences(subjectReference common.Reference) (*[]ocispecs.ReferenceDescriptor, error) {
	var references []ocispecs.ReferenceDescriptor
	ref, err := name.ParseReference(subjectReference.Original)
	if err != nil {
		return &references, err
	}
	hash := v1.Hash{
		Algorithm: subjectReference.Digest.Algorithm().String(),
		Hex:       subjectReference.Digest.Hex(),
	}

	signatureTag := attachedImageTag(ref.Context(), hash, CosignSignatureTagSuffix)

	desc, err := remote.Get(signatureTag)
	if terr, ok := err.(*transport.Error); ok && terr.StatusCode == http.StatusNotFound {
		return &references, nil
	}
	if err != nil {
		return &references, err
	}
	descDig, err := digest.Parse(desc.Digest.String())
	if err != nil {
		return &references, err
	}

	references = append(references, ocispecs.ReferenceDescriptor{
		ArtifactType: CosignArtifactType,
		Descriptor: oci.Descriptor{
			MediaType: string(desc.MediaType),
			Digest:    descDig,
			Size:      desc.Size,
		},
	})

	return &references, nil
}

func attachedImageTag(repo name.Repository, digest v1.Hash, tagSuffix string) name.Tag {
	// sha256:d34db33f -> sha256-d34db33f.suffix
	tagStr := strings.ReplaceAll(digest.String(), ":", "-") + tagSuffix
	return repo.Tag(tagStr)
}
