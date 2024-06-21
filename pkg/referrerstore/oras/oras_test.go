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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/ratify-project/ratify/pkg/cache"
	"github.com/ratify-project/ratify/pkg/common"
	"github.com/ratify-project/ratify/pkg/ocispecs"
	"github.com/ratify-project/ratify/pkg/referrerstore/config"
	"github.com/ratify-project/ratify/pkg/referrerstore/oras/mocks"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote/errcode"
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
				ArtifactType: "application/vnd.cncf.notary.signature",
			},
		},
	}
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, error) {
		return testRepo, nil
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

	if referrers.Referrers[0].ArtifactType != "application/vnd.cncf.notary.signature" {
		t.Fatalf("expected artifact type %s, got %s", "application/vnd.cncf.notary.signature", referrers.Referrers[0].ArtifactType)
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
				ArtifactType: "application/vnd.cncf.notary.signature",
			},
		},
	}
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, error) {
		return testRepo, nil
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

	if referrers.Referrers[0].ArtifactType != "application/vnd.cncf.notary.signature" {
		t.Fatalf("expected artifact type %s, got %s", "application/vnd.cncf.notary.signature", referrers.Referrers[0].ArtifactType)
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
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, error) {
		return testRepo, nil
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
			MediaType: ocispecs.MediaTypeArtifactManifest,
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
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, error) {
		return testRepo, nil
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
			MediaType: ocispecs.MediaTypeArtifactManifest,
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
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, error) {
		return testRepo, nil
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
	store.createRepository = func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, error) {
		return testRepo, nil
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

func Test_EvictOnError(t *testing.T) {
	ctx := context.Background()
	var err error

	cacheProvider := cache.GetCacheProvider()
	if cacheProvider == nil {
		// if no cache provider has been initialized, initialize one
		cacheProvider, err = cache.NewCacheProvider(ctx, cache.DefaultCacheType, cache.DefaultCacheName, cache.DefaultCacheSize)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}

	testcases := []struct {
		Method           string
		StatusCode       int
		subjectReference string
	}{
		{
			Method:     "GET",
			StatusCode: 401,
		},
		{
			Method:     "GET",
			StatusCode: 403,
		},
	}

	for _, testcase := range testcases {
		subjectReference := "testSubjectRef"
		cacheKey := fmt.Sprintf(cache.CacheKeyOrasAuth, subjectReference)
		cacheProvider.Set(ctx, cacheKey, "hello")
		time.Sleep(1 * time.Second)
		// validate the entry exists
		_, ok := cacheProvider.Get(ctx, cacheKey)
		if !ok {
			t.Fatalf("failed to add entry to auth cache")
		}

		mockErrResponse := errcode.ErrorResponse{
			Method:     testcase.Method,
			StatusCode: testcase.StatusCode,
		}
		evictOnError(ctx, &mockErrResponse, subjectReference+"/test")
		time.Sleep(1 * time.Second)
		// validate the entry should no longer exist
		_, ok = cacheProvider.Get(ctx, cacheKey)
		if ok {
			t.Fatalf("Auth cache entry should have been evicted")
		}
	}
}

// Test_ORASRetryClient tests that the retry client retries on 429 for specified number of retries
func Test_ORASRetryClient(t *testing.T) {
	// Create a test server
	count := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/v2/test/manifests/latest":
			w.WriteHeader(http.StatusTooManyRequests)
			count++
			return
		default:
			w.WriteHeader(http.StatusForbidden)
		}
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	conf := config.StorePluginConfig{
		"name":    "oras",
		"useHttp": true,
	}
	store, err := createBaseStore("1.0.0", conf)
	if err != nil {
		t.Fatalf("failed to create oras store: %v", err)
	}
	uri, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("invalid test http server: %v", err)
	}
	_, err = store.GetSubjectDescriptor(context.Background(), common.Reference{
		Original: uri.Host + "/test:latest",
		Tag:      "latest",
		Path:     uri.Host + "/test",
	})
	if err != nil {
		var ec errcode.Error
		if errors.As(err, &ec) && (ec.Code != fmt.Sprint(http.StatusTooManyRequests)) {
			t.Fatalf("expected error code %d, got %s", http.StatusTooManyRequests, ec.Code)
		}
	}

	// Verify that the retry client was used by checking the number of retries
	// The retry client will retry 5 times, so we expect 6 total requests
	if count != 6 {
		t.Fatalf("expected 6 retries, got %d", count)
	}
}

func TestORASCreate_CreateBaseStore_Failure(t *testing.T) {
	conf := config.StorePluginConfig{
		"name": "oras",
		"authProvider": map[string]interface{}{
			"name": "mock",
		},
	}
	factory := orasStoreFactory{}
	_, err := factory.Create("1.0.0", conf)
	if err == nil {
		t.Fatalf("expected error creating oras store")
	}
}
func TestORASCreate_CacheProvider_Nil(t *testing.T) {
	conf := config.StorePluginConfig{
		"name": "oras",
	}
	factory := orasStoreFactory{}
	store, err := factory.Create("1.0.0", conf)
	if err != nil {
		t.Fatalf("failed to create oras store: %v", err)
	}
	if _, ok := store.(*orasStore); !ok {
		t.Fatalf("expected oras store")
	}
}
