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

package ocispecs

import (
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

const MediaTypeArtifactManifest = "application/vnd.cncf.oras.artifact.manifest.v1+json"

type ReferenceDescriptor struct {
	oci.Descriptor

	ArtifactType string `json:"artifactType,omitempty"`
}

type ReferrersResponse struct {
	Digest    string     `json:"digest"`
	Referrers []Referrer `json:"references"`
}

type Referrer struct {
	Digest   string            `json:"digest"`
	Manifest ReferenceManifest `json:"manifest"`
}

type ReferenceManifest1 struct {
	oci.Manifest

	SubjectDecriptor oci.Descriptor `json:"subjectDesc"`
}

type ReferenceManifest struct {
	MediaType    string           `json:"mediaType"`
	ArtifactType string           `json:"artifactType"`
	Blobs        []oci.Descriptor `json:"blobs"`
	Subjects     []oci.Descriptor `json:"manifests"`
}
