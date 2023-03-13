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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/deislabs/ratify/pkg/referrerstore/oras/mocks"
	"github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

const inputOriginalPath = "localhost:5000/net-monitor:v0"

// TestORASName tests the Name method of the oras store.
func TestORASName(t *testing.T) {
	conf := config.StorePluginConfig{
		"name": "oras",
	}
	store, err := createBaseStore("1.0.0", conf)
	if err != nil {
		t.Fatalf("failed to create oras store: %v", err)
	}
	output := store.Name()
	if output != "oras" {
		t.Fatalf("expected name to be 'oras', got '%s'", output)
	}
}

// TestORASGetConfig tests the GetConfig method of the oras store.
func TestORASGetConfig(t *testing.T) {
	conf := config.StorePluginConfig{
		"name": "oras",
	}
	expectedRawConfig := config.StoreConfig{Version: "1.0.0", Store: conf}
	store, err := createBaseStore("1.0.0", conf)
	if err != nil {
		t.Fatalf("failed to create oras store: %v", err)
	}
	output := store.GetConfig()
	if !reflect.DeepEqual(*output, expectedRawConfig) {
		t.Fatalf("expected raw config to be %v, got %v", expectedRawConfig, output)
	}
}

// TestORASGetSubjectDescriptor_CacheMiss tests the case where the subject descriptor cache is empty
func TestORASGetSubjectDescriptor_CacheMiss(t *testing.T) {
	conf := config.StorePluginConfig{
		"name": "oras",
	}
	ctx := context.Background()
	store, err := createBaseStore("1.0.0", conf)
	if err != nil {
		t.Fatalf("failed to create oras store: %v", err)
	}
	testRepo := mocks.TestRepository{
		ResolveMap: map[string]oci.Descriptor{
			inputOriginalPath: {
				Digest: digest.FromString("testDigest"),
			},
		},
	}
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, time.Time, error) {
		return testRepo, time.Now().Add(time.Minute), nil
	}
	inputRef := common.Reference{
		Original: inputOriginalPath,
		Digest:   digest.FromString("testDigest"),
	}
	subjDesc, err := store.GetSubjectDescriptor(ctx, inputRef)
	if err != nil {
		t.Fatalf("failed to get subject descriptor: %v", err)
	}
	if subjDesc.Digest != digest.FromString("testDigest") {
		t.Fatalf("expected digest %s, got %s", digest.FromString("testDigest"), subjDesc.Digest)
	}
}

// TestORASGetSubjectDescriptor_CacheHit tests that the subject descriptor cache is used
func TestORASGetSubjectDescriptor_CacheHit(t *testing.T) {
	conf := config.StorePluginConfig{
		"name": "oras",
	}
	ctx := context.Background()
	store, err := createBaseStore("1.0.0", conf)
	if err != nil {
		t.Fatalf("failed to create oras store: %v", err)
	}
	firstDigest := digest.FromString("firstDigest")
	secondDigest := digest.FromString("secondDigest")
	store.subjectDescriptorCache.Store(firstDigest, oci.Descriptor{Digest: secondDigest})
	testRepo := mocks.TestRepository{
		ResolveMap: map[string]oci.Descriptor{
			inputOriginalPath: {
				Digest: firstDigest,
			},
		},
	}
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, time.Time, error) {
		return testRepo, time.Now().Add(time.Minute), nil
	}
	inputRef := common.Reference{
		Original: inputOriginalPath,
		Digest:   firstDigest,
	}
	subjDesc, err := store.GetSubjectDescriptor(ctx, inputRef)
	if err != nil {
		t.Fatalf("failed to get subject descriptor: %v", err)
	}
	if subjDesc.Digest != secondDigest {
		t.Fatalf("expected digest %s, got %s", secondDigest, subjDesc.Digest)
	}
}

func TestORASListReferrers_SubjectDesc(t *testing.T) {
	conf := config.StorePluginConfig{
		"name":          "oras",
		"cosignEnabled": false,
	}
	ctx := context.Background()
	subjectDigest := digest.FromString("testDigest")
	referrerDigest := digest.FromString("testArtifactDigest")
	store, err := createBaseStore("1.0.0", conf)
	if err != nil {
		t.Fatalf("failed to create oras store: %v", err)
	}
	subjectDesc := ocispecs.SubjectDescriptor{
		Descriptor: oci.Descriptor{
			Digest: subjectDigest,
		},
	}
	testRepo := mocks.TestRepository{
		ResolveMap: map[string]oci.Descriptor{
			inputOriginalPath: subjectDesc.Descriptor,
		},
		ReferrersList: []oci.Descriptor{
			{
				Digest:       referrerDigest,
				ArtifactType: "application/vnd.cncf.notary.v2",
			},
		},
	}
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, time.Time, error) {
		return testRepo, time.Now().Add(time.Minute), nil
	}
	inputRef := common.Reference{
		Original: inputOriginalPath,
		Digest:   subjectDigest,
	}
	referrers, err := store.ListReferrers(ctx, inputRef, []string{}, "", &subjectDesc)
	if err != nil {
		t.Fatalf("failed to list referrers: %v", err)
	}

	if referrers.NextToken != "" {
		t.Fatalf("expected next token to be empty, got %s", referrers.NextToken)
	}

	if len(referrers.Referrers) != 1 {
		t.Fatalf("expected 1 referrer, got %d", len(referrers.Referrers))
	}

	if referrers.Referrers[0].Digest != referrerDigest {
		t.Fatalf("expected digest %s, got %s", referrerDigest, referrers.Referrers[0].Digest)
	}

	if referrers.Referrers[0].ArtifactType != "application/vnd.cncf.notary.v2" {
		t.Fatalf("expected artifact type %s, got %s", "application/vnd.cncf.notary.v2", referrers.Referrers[0].ArtifactType)
	}
}

func TestORASListReferrers_NoSubjectDesc(t *testing.T) {
	conf := config.StorePluginConfig{
		"name":          "oras",
		"cosignEnabled": false,
	}
	ctx := context.Background()
	subjectDigest := digest.FromString("testDigest")
	referrerDigest := digest.FromString("testArtifactDigest")
	store, err := createBaseStore("1.0.0", conf)
	if err != nil {
		t.Fatalf("failed to create oras store: %v", err)
	}
	subjectDesc := ocispecs.SubjectDescriptor{
		Descriptor: oci.Descriptor{
			Digest: subjectDigest,
		},
	}
	testRepo := mocks.TestRepository{
		ResolveMap: map[string]oci.Descriptor{
			inputOriginalPath: subjectDesc.Descriptor,
		},
		ReferrersList: []oci.Descriptor{
			{
				Digest:       referrerDigest,
				ArtifactType: "application/vnd.cncf.notary.v2",
			},
		},
	}
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, time.Time, error) {
		return testRepo, time.Now().Add(time.Minute), nil
	}
	inputRef := common.Reference{
		Original: inputOriginalPath,
		Digest:   subjectDigest,
	}
	referrers, err := store.ListReferrers(ctx, inputRef, []string{}, "", nil)
	if err != nil {
		t.Fatalf("failed to list referrers: %v", err)
	}

	if referrers.NextToken != "" {
		t.Fatalf("expected next token to be empty, got %s", referrers.NextToken)
	}

	if len(referrers.Referrers) != 1 {
		t.Fatalf("expected 1 referrer, got %d", len(referrers.Referrers))
	}

	if referrers.Referrers[0].Digest != referrerDigest {
		t.Fatalf("expected digest %s, got %s", referrerDigest, referrers.Referrers[0].Digest)
	}

	if referrers.Referrers[0].ArtifactType != "application/vnd.cncf.notary.v2" {
		t.Fatalf("expected artifact type %s, got %s", "application/vnd.cncf.notary.v2", referrers.Referrers[0].ArtifactType)
	}
}

// TODO: add cosign test for List Referrers

// TestORASGetReferenceManifest_CachedDesc tests that the reference manifest is returned from the cache if it exists
func TestORASGetReferenceManifest_CachedDesc(t *testing.T) {
	conf := config.StorePluginConfig{
		"name": "oras",
	}
	ctx := context.Background()
	firstDigest := digest.FromString("testDigest")
	artifactDigest := digest.FromString("testArtifactDigest")
	expectedReferenceMediatype := "application/vnd.oci.image.manifest.right.v1+json"
	wrongReferenceMediatype := "application/vnd.oci.image.manifest.wrong.v1+json"
	store, err := createBaseStore("1.0.0", conf)
	if err != nil {
		t.Fatalf("failed to create oras store: %v", err)
	}
	manifestCached := oci.Manifest{
		MediaType: expectedReferenceMediatype,
		Config:    oci.Descriptor{},
		Layers:    []oci.Descriptor{},
	}
	manifestCachedBytes, err := json.Marshal(manifestCached)
	if err != nil {
		t.Fatalf("failed to marshal cached manifest: %v", err)
	}
	manifestNotCached := oci.Manifest{
		MediaType: wrongReferenceMediatype,
		Config:    oci.Descriptor{},
		Layers:    []oci.Descriptor{},
	}
	manifestNotCachedBytes, err := json.Marshal(manifestNotCached)
	if err != nil {
		t.Fatalf("failed to marshal not cached manifest: %v", err)
	}
	testRepo := mocks.TestRepository{
		FetchMap: map[digest.Digest]io.ReadCloser{
			artifactDigest: io.NopCloser(bytes.NewReader(manifestNotCachedBytes)),
		},
	}
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, time.Time, error) {
		return testRepo, time.Now().Add(time.Minute), nil
	}
	store.localCache = mocks.TestStorage{
		ExistsMap: map[digest.Digest]io.Reader{
			artifactDigest: bytes.NewReader(manifestCachedBytes),
		},
	}
	inputRef := common.Reference{
		Original: inputOriginalPath,
		Digest:   firstDigest,
	}
	manifest, err := store.GetReferenceManifest(ctx, inputRef, ocispecs.ReferenceDescriptor{
		Descriptor: oci.Descriptor{
			MediaType: oci.MediaTypeArtifactManifest,
			Digest:    artifactDigest,
		},
	})
	if err != nil {
		t.Fatalf("failed to get reference manifest: %v", err)
	}
	if manifest.MediaType != expectedReferenceMediatype {
		t.Fatalf("expected media type %s, got %s", expectedReferenceMediatype, manifest.MediaType)
	}
}

// TestORASGetReferenceManifest_NotCachedDesc tests that the reference manifest is fetched from the registry if it is not cached
func TestORASGetReferenceManifest_NotCachedDesc(t *testing.T) {
	conf := config.StorePluginConfig{
		"name": "oras",
	}
	ctx := context.Background()
	firstDigest := digest.FromString("testDigest")
	artifactDigest := digest.FromString("testArtifactDigest")
	artifactDigestNotCached := digest.FromString("testArtifactDigestNotCached")
	expectedReferenceMediatype := "application/vnd.oci.image.manifest.right.v1+json"
	wrongReferenceMediatype := "application/vnd.oci.image.manifest.wrong.v1+json"
	store, err := createBaseStore("1.0.0", conf)
	if err != nil {
		t.Fatalf("failed to create oras store: %v", err)
	}
	manifestCached := oci.Manifest{
		MediaType: wrongReferenceMediatype,
		Config:    oci.Descriptor{},
		Layers:    []oci.Descriptor{},
	}
	manifestCachedBytes, err := json.Marshal(manifestCached)
	if err != nil {
		t.Fatalf("failed to marshal cached manifest: %v", err)
	}
	manifestNotCached := oci.Manifest{
		MediaType: expectedReferenceMediatype,
		Config:    oci.Descriptor{},
		Layers:    []oci.Descriptor{},
	}
	manifestNotCachedBytes, err := json.Marshal(manifestNotCached)
	if err != nil {
		t.Fatalf("failed to marshal not cached manifest: %v", err)
	}
	testRepo := mocks.TestRepository{
		FetchMap: map[digest.Digest]io.ReadCloser{
			artifactDigest: io.NopCloser(bytes.NewReader(manifestNotCachedBytes)),
		},
	}
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, time.Time, error) {
		return testRepo, time.Now().Add(time.Minute), nil
	}
	store.localCache = mocks.TestStorage{
		ExistsMap: map[digest.Digest]io.Reader{
			artifactDigestNotCached: bytes.NewReader(manifestCachedBytes),
		},
	}
	inputRef := common.Reference{
		Original: inputOriginalPath,
		Digest:   firstDigest,
	}
	manifest, err := store.GetReferenceManifest(ctx, inputRef, ocispecs.ReferenceDescriptor{
		Descriptor: oci.Descriptor{
			MediaType: oci.MediaTypeArtifactManifest,
			Digest:    artifactDigest,
		},
	})
	if err != nil {
		t.Fatalf("failed to get reference manifest: %v", err)
	}
	if manifest.MediaType != expectedReferenceMediatype {
		t.Fatalf("expected media type %s, got %s", expectedReferenceMediatype, manifest.MediaType)
	}
}

// TestORASGetBlobContent_CachedDesc tests that the blob content is fetched from the cache if it is cached
func TestORASGetBlobContent_CachedDesc(t *testing.T) {
	conf := config.StorePluginConfig{
		"name": "oras",
	}
	ctx := context.Background()
	firstDigest := digest.FromString("testDigest")
	blobDigest := digest.FromString("testBlobDigest")
	expectedContent := []byte("test content")
	inputRef := common.Reference{
		Original: inputOriginalPath,
		Path:     inputOriginalPath,
		Digest:   firstDigest,
	}
	store, err := createBaseStore("1.0.0", conf)
	if err != nil {
		t.Fatalf("failed to create oras store: %v", err)
	}
	testRepo := mocks.TestRepository{
		BlobStoreTest: mocks.TestBlobStore{
			BlobMap: map[string]mocks.BlobPair{
				fmt.Sprintf("%s@%s", inputRef.Path, blobDigest.String()): {
					Descriptor: oci.Descriptor{
						Digest: blobDigest,
					},
					Reader: io.NopCloser(bytes.NewReader(expectedContent)),
				},
			},
		},
	}
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, time.Time, error) {
		return testRepo, time.Now().Add(time.Minute), nil
	}
	store.localCache = mocks.TestStorage{
		ExistsMap: map[digest.Digest]io.Reader{
			blobDigest: bytes.NewReader(expectedContent),
		},
	}
	content, err := store.GetBlobContent(ctx, inputRef, blobDigest)
	if err != nil {
		t.Fatalf("failed to get blob content: %v", err)
	}
	if !bytes.Equal(content, expectedContent) {
		t.Fatalf("expected content %s, got %s", expectedContent, content)
	}
}

// TestORASGetBlobContent_NotCachedDesc tests that the blob content is fetched from the registry if it is not cached
func TestORASGetBlobContent_NotCachedDesc(t *testing.T) {
	conf := config.StorePluginConfig{
		"name": "oras",
	}
	ctx := context.Background()
	firstDigest := digest.FromString("testDigest")
	blobDigest := digest.FromString("testBlobDigest")
	expectedContent := []byte("test content")
	inputRef := common.Reference{
		Original: inputOriginalPath,
		Path:     inputOriginalPath,
		Digest:   firstDigest,
	}
	store, err := createBaseStore("1.0.0", conf)
	if err != nil {
		t.Fatalf("failed to create oras store: %v", err)
	}
	testRepo := mocks.TestRepository{
		BlobStoreTest: mocks.TestBlobStore{
			BlobMap: map[string]mocks.BlobPair{
				fmt.Sprintf("%s@%s", inputRef.Path, blobDigest.String()): {
					Descriptor: oci.Descriptor{
						Digest: blobDigest,
					},
					Reader: io.NopCloser(bytes.NewReader(expectedContent)),
				},
			},
		},
	}
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, time.Time, error) {
		return testRepo, time.Now().Add(time.Minute), nil
	}
	store.localCache = mocks.TestStorage{
		ExistsMap: map[digest.Digest]io.Reader{},
	}
	content, err := store.GetBlobContent(ctx, inputRef, blobDigest)
	if err != nil {
		t.Fatalf("failed to get blob content: %v", err)
	}
	if !bytes.Equal(content, expectedContent) {
		t.Fatalf("expected content %s, got %s", expectedContent, content)
	}
}
