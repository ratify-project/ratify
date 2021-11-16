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
	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/plugin/skel"
	"github.com/opencontainers/go-digest"
)

func main() {
	skel.PluginMain("sample", "1.0.0", ListReferrers, GetBlobContent, GetReferenceManifest, ResolveTag, []string{"1.0.0"})
}

func ListReferrers(args *skel.CmdArgs, subjectReference common.Reference, artifactTypes []string, nextToken string) (*referrerstore.ListReferrersResult, error) {
	artifactType := ""
	if len(artifactTypes) > 0 {
		artifactType = artifactTypes[0]
	}
	return &referrerstore.ListReferrersResult{
		Referrers: []ocispecs.ReferenceDescriptor{{
			ArtifactType: artifactType,
		}},
		NextToken: "",
	}, nil
}

func GetBlobContent(args *skel.CmdArgs, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	return []byte(digest.String()), nil
}

func ResolveTag(args *skel.CmdArgs, subjectReference common.Reference) (digest.Digest, error) {
	dig := subjectReference.Digest
	if dig == "" {
		dig = digest.FromString(subjectReference.Tag)
	}
	return dig, nil
}

func GetReferenceManifest(args *skel.CmdArgs, subjectReference common.Reference, digest digest.Digest) (ocispecs.ReferenceManifest, error) {
	return ocispecs.ReferenceManifest{MediaType: "testMediaType", ArtifactType: "testArtifactType"}, nil
}
