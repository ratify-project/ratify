package oras

import (
	"github.com/deislabs/ratify/pkg/ocispecs"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	artifactspec "github.com/oras-project/artifacts-spec/specs-go/v1"
	"regexp"
	"strings"
)

// Detect the loopback IP (127.0.0.1)
var reLoopback = regexp.MustCompile(regexp.QuoteMeta("127.0.0.1"))

// Detect the loopback IPV6 (::1)
var reipv6Loopback = regexp.MustCompile(regexp.QuoteMeta("::1"))

func isInsecureRegistry(registry string, config *OrasStoreConf) bool {
	if config.UseHttp {
		return true
	}
	if strings.HasPrefix(registry, "localhost:") {
		return true
	}

	if reLoopback.MatchString(registry) {
		return true
	}
	if reipv6Loopback.MatchString(registry) {
		return true
	}

	return false
}

func ArtifactDescriptorToReferenceDescriptor(artifactDescriptor artifactspec.Descriptor) ocispecs.ReferenceDescriptor {
	return ocispecs.ReferenceDescriptor{
		Descriptor: oci.Descriptor{
			MediaType:   artifactDescriptor.MediaType,
			Digest:      artifactDescriptor.Digest,
			Size:        artifactDescriptor.Size,
			URLs:        artifactDescriptor.URLs,
			Annotations: artifactDescriptor.Annotations,
		},
		ArtifactType: artifactDescriptor.ArtifactType,
	}
}

func ArtifactManifestToReferenceManifest(artifactManifest artifactspec.Manifest) ocispecs.ReferenceManifest {
	var blobs []oci.Descriptor
	for _, blob := range artifactManifest.Blobs {
		ociBlob := oci.Descriptor{
			MediaType:   blob.MediaType,
			Digest:      blob.Digest,
			Size:        blob.Size,
			URLs:        blob.URLs,
			Annotations: blob.Annotations,
		}
		blobs = append(blobs, ociBlob)
	}

	subjects := []oci.Descriptor{{
		MediaType:   artifactManifest.Subject.MediaType,
		Digest:      artifactManifest.Subject.Digest,
		Size:        artifactManifest.Subject.Size,
		URLs:        artifactManifest.Subject.URLs,
		Annotations: artifactManifest.Subject.Annotations,
	}}

	return ocispecs.ReferenceManifest{
		MediaType:    ocispecs.MediaTypeArtifactManifest,
		ArtifactType: artifactManifest.ArtifactType,
		Blobs:        blobs,
		Subjects:     subjects,
	}
}
