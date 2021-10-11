package oras

import (
	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/deislabs/hora/plugins/referrerstore/ociregistry/registry"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sigstore/cosign/pkg/cosign"
)

const CosignArtifactType = "org.sigstore.cosign.v1"

func getCosignReferences(client *registry.Client, subjectReference common.Reference) (*[]ocispecs.ReferenceDescriptor, error) {
	var references []ocispecs.ReferenceDescriptor
	ref, err := name.ParseReference(subjectReference.Original)
	if err != nil {
		return nil, err
	}
	hash := v1.Hash{
		Algorithm: subjectReference.Digest.Algorithm().String(),
		Hex:       subjectReference.Digest.Hex(),
	}
	signatureTag := cosign.AttachedImageTag(ref.Context(), hash, cosign.SignatureTagSuffix)
	tagRef := common.Reference{
		Path: subjectReference.Path,
		Tag:  signatureTag.TagStr(),
	}
	desc, err := client.GetManifestMetadata(tagRef)

	if err != nil && err != registry.ManifestNotFound {
		return nil, err
	}

	if err == nil {
		references = append(references, ocispecs.ReferenceDescriptor{
			ArtifactType: CosignArtifactType,
			Descriptor: oci.Descriptor{
				MediaType: desc.MediaType,
				Digest:    desc.Digest,
				Size:      desc.Size,
			},
		})
	}

	return &references, nil
}
