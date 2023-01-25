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
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/opencontainers/go-digest"

	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	testName = "testName"
)

var (
	ttl          = 10
	base         = &mockBase{}
	pluginConfig = map[string]interface{}{
		"cacheEnabled": true,
		"ttl":          ttl,
	}
	conf = &cacheConf{
		Enabled:   true,
		TTL:       ttl,
		Capacity:  100 * 1024 * 1024,
		KeyNumber: 10000,
	}
	testStoreConfig = &config.StoreConfig{}
	testBlob        = make([]byte, 0)
	testDigest      = digest.Digest("sha256:123456")
	testReference   = common.Reference{
		Path:   "testRegistry/testRepo",
		Digest: testDigest,
	}
	testDesc    = &ocispecs.SubjectDescriptor{}
	testResult1 = referrerstore.ListReferrersResult{
		Referrers: []ocispecs.ReferenceDescriptor{
			{
				Descriptor: oci.Descriptor{
					Digest: testDigest,
				},
			},
		},
	}
	testResult2 = referrerstore.ListReferrersResult{
		NextToken: testNextToken2,
	}
	testNextToken1 = "1"
	testNextToken2 = "2"
)

type mockBase struct{}

func (m *mockBase) Name() string {
	return testName
}

func (m *mockBase) GetConfig() *config.StoreConfig {
	return testStoreConfig
}

func (m *mockBase) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string, subjectDesc *ocispecs.SubjectDescriptor) (referrerstore.ListReferrersResult, error) {
	if nextToken == testNextToken1 {
		return testResult1, nil
	} else if nextToken == testNextToken2 {
		return testResult2, nil
	}
	return referrerstore.ListReferrersResult{}, nil
}

func (m *mockBase) GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	return testBlob, nil
}

func (m *mockBase) GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error) {
	return ocispecs.ReferenceManifest{}, nil
}

func (m *mockBase) GetSubjectDescriptor(ctx context.Context, subjectReference common.Reference) (*ocispecs.SubjectDescriptor, error) {
	return testDesc, nil
}

func TestCreateCachedStore(t *testing.T) {
	if _, err := createCachedStore(base, conf); err != nil {
		t.Fatalf("expect no error, got %v", err)
	}
}

func TestName(t *testing.T) {
	store, _ := createCachedStore(base, conf)

	name := store.Name()
	if name != testName {
		t.Fatalf("expect name: %s, got %s", testName, name)
	}
}

func TestGetConfig(t *testing.T) {
	store, _ := createCachedStore(base, conf)

	conf := store.GetConfig()
	if !reflect.DeepEqual(conf, testStoreConfig) {
		t.Fatalf("expect config: %+v, got %+v", testStoreConfig, conf)
	}
}

func TestGetBlobContent(t *testing.T) {
	store, _ := createCachedStore(base, conf)

	blob, err := store.GetBlobContent(context.Background(), testReference, testDigest)
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}
	if !reflect.DeepEqual(blob, testBlob) {
		t.Fatalf("expect blob: %v, got %v", testBlob, blob)
	}
}

func TestGetSubjectDescriptor(t *testing.T) {
	store, _ := createCachedStore(base, conf)

	desc, err := store.GetSubjectDescriptor(context.Background(), testReference)
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}
	if !reflect.DeepEqual(testDesc, desc) {
		t.Fatalf("expect desc: %v, got %v", desc, testDesc)
	}
}

func TestListReferrers_CacheHit(t *testing.T) {
	store, _ := createCachedStore(base, conf)

	result, _ := store.ListReferrers(context.Background(), testReference, []string{}, testNextToken1, nil)

	time.Sleep(time.Duration(ttl-5) * time.Second)

	cachedResult, err := store.ListReferrers(context.Background(), testReference, []string{}, testNextToken2, nil)
	if err != nil {
		t.Fatalf("err should be nil, but got %v", err)
	}
	if !reflect.DeepEqual(result, cachedResult) {
		t.Fatalf("cached result: %+v is different from result: %+v", cachedResult, result)
	}
}

func TestListReferrers_CacheMiss(t *testing.T) {
	store, _ := createCachedStore(base, conf)

	result, _ := store.ListReferrers(context.Background(), testReference, []string{}, testNextToken1, nil)

	time.Sleep(time.Duration(ttl+5) * time.Second)

	cachedResult, err := store.ListReferrers(context.Background(), testReference, []string{}, testNextToken2, nil)
	if err != nil {
		t.Fatalf("err should be nil, but got %v", err)
	}
	if reflect.DeepEqual(result, cachedResult) {
		t.Fatalf("cached result: %+v should be different from result: %+v", cachedResult, result)
	}
}

func TestToCacheConfig(t *testing.T) {
	resultCache, err := toCacheConfig(pluginConfig)
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}
	if !reflect.DeepEqual(conf, resultCache) {
		t.Fatalf("expect %v, got %v", conf, resultCache)
	}
}
