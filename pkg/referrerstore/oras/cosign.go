package oras

import (
	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sigstore/cosign/pkg/cosign"
	"net/http"
	"strings"
)

const CosignArtifactType = "org.sigstore.cosign.v1"

func getCosignReferences(subjectReference common.Reference) (*[]ocispecs.ReferenceDescriptor, error) {
	var references []ocispecs.ReferenceDescriptor
	if strings.Split(subjectReference.Original, "@")[1] == "" {
		return &references, nil
	}
	ref, err := name.ParseReference(subjectReference.Original)
	if err != nil {
		return &references, err
	}
	hash := v1.Hash{
		Algorithm: subjectReference.Digest.Algorithm().String(),
		Hex:       subjectReference.Digest.Hex(),
	}
	signatureTag := cosign.AttachedImageTag(ref.Context(), hash, cosign.SignatureTagSuffix)

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
