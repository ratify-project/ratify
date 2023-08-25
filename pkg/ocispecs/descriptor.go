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

const MediaTypeArtifactManifest = "application/vnd.oci.artifact.manifest.v1+json"

// ReferenceDescriptor represents a descriptor for an artifact manifest
type ReferenceDescriptor struct {
	oci.Descriptor

	ArtifactType string `json:"artifactType,omitempty"`
}

// ReferenceManifest describes an artifact manifest
type ReferenceManifest struct {
	MediaType    string            `json:"mediaType"`
	ArtifactType string            `json:"artifactType,omitempty"`
	Blobs        []oci.Descriptor  `json:"blobs"`
	Subject      *oci.Descriptor   `json:"subject,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

type SubjectDescriptor struct {
	oci.Descriptor
}
