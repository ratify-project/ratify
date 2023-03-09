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
	"reflect"
	"testing"

	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestIsInsecureRegistry(t *testing.T) {
	testCases := []struct {
		desc     string
		registry string
		config   OrasStoreConf
		expected bool
	}{
		{
			desc:     "secure registry with http disabled",
			registry: "ghcr.io/test/registry:v0",
			config: OrasStoreConf{
				Name: "oras",
			},
			expected: false,
		},
		{
			desc:     "insecure registry with http enabled",
			registry: "registry:5000/test/registry:v0",
			config: OrasStoreConf{
				Name:    "oras",
				UseHttp: true,
			},
			expected: true,
		},
		{
			desc:     "localhost insecure registry with http not specified",
			registry: "localhost:5000/test/registry:v0",
			config: OrasStoreConf{
				Name: "oras",
			},
			expected: true,
		},
		{
			desc:     "loopback ipv4 insecure registry with http not specified",
			registry: "127.0.0.1:5000/test/registry:v0",
			config: OrasStoreConf{
				Name: "oras",
			},
			expected: true,
		},
		{
			desc:     "loopback ipv6 insecure registry with http not specified",
			registry: "::1:5000/test/registry:v0",
			config: OrasStoreConf{
				Name: "oras",
			},
			expected: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			output := isInsecureRegistry(testCase.registry, &testCase.config)
			if output != testCase.expected {
				t.Fatalf("mismatch of insecure registry type: expected %v, actual %v", testCase.expected, output)
			}
		})
	}
}

func TestOciDescriptorToReferenceDescriptor(t *testing.T) {
	input := oci.Descriptor{
		Digest:       digest.FromString("test"),
		Size:         5,
		ArtifactType: "test_type",
	}
	expected := ocispecs.ReferenceDescriptor{
		Descriptor:   input,
		ArtifactType: "test_type",
	}
	output := OciDescriptorToReferenceDescriptor(input)
	if output.ArtifactType != expected.ArtifactType || output.Descriptor.Digest.String() != expected.Descriptor.Digest.String() {
		t.Fatalf("mismatch of reference descriptor: expected %v, actual %v", expected, output)
	}
}

func TestOciManifestToReferenceManifest(t *testing.T) {
	type args struct {
		ociManifest oci.Manifest
	}
	tests := []struct {
		name string
		args args
		want ocispecs.ReferenceManifest
	}{
		{
			name: "empty",
			args: args{
				ociManifest: oci.Manifest{},
			},
			want: ocispecs.ReferenceManifest{
				MediaType:    "",
				ArtifactType: "",
			},
		},
		{
			name: "simple",
			args: args{
				ociManifest: oci.Manifest{
					MediaType: "application/vnd.oci.image.manifest.v1+json",
					Config: oci.Descriptor{
						MediaType: "application/vnd.oci.image.config.v1+json",
					},
				},
			},
			want: ocispecs.ReferenceManifest{
				MediaType:    "application/vnd.oci.image.manifest.v1+json",
				ArtifactType: "application/vnd.oci.image.config.v1+json",
			},
		},
		{
			name: "layers",
			args: args{
				ociManifest: oci.Manifest{
					MediaType: "application/vnd.oci.image.manifest.v1+json",
					Config: oci.Descriptor{
						MediaType: "application/vnd.oci.image.config.v1+json",
					},
					Layers: []oci.Descriptor{
						{
							MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
						},
					},
				},
			},
			want: ocispecs.ReferenceManifest{
				MediaType:    "application/vnd.oci.image.manifest.v1+json",
				ArtifactType: "application/vnd.oci.image.config.v1+json",
				Blobs: []oci.Descriptor{
					{
						MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OciManifestToReferenceManifest(tt.args.ociManifest); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OciManifestToReferenceManifest() = %v, want %v", got, tt.want)
			}
		})
	}
}
