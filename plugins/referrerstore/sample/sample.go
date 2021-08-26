package main

import (
	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/deislabs/hora/pkg/referrerstore"
	"github.com/deislabs/hora/pkg/referrerstore/plugin/skel"
	"github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

func main() {
	skel.PluginMain("sample", "1.0.0", ListReferrers, GetBlobContent, GetReferenceManifest, []string{"1.0.0"})
}

func ListReferrers(args *skel.CmdArgs, subjectReference common.Reference, artifactTypes []string, nextToken string) (referrerstore.ListReferrersResult, error) {
	artifactType := ""
	if len(artifactTypes) > 0 {
		artifactType = artifactTypes[0]
	}
	return referrerstore.ListReferrersResult{
		Referrers: []ocispecs.ReferenceDescriptor{{
			ArtifactType: artifactType,
		}},
		NextToken: "",
	}, nil
}

func GetBlobContent(args *skel.CmdArgs, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	return []byte(digest.String()), nil
}

func GetReferenceManifest(args *skel.CmdArgs, subjectReference common.Reference, digest digest.Digest) (ocispecs.ReferenceManifest, error) {
	return ocispecs.ReferenceManifest{SubjectDecriptor: oci.Descriptor{MediaType: "testMediaType"}}, nil
}
