package ocispecs

import (
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

const MediaTypeArtifactManifest = "application/vnd.oci.artifact.manifest.v1+json"

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
