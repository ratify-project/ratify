package oras

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/dgraph-io/ristretto"
	"github.com/dgraph-io/ristretto/z"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
)

type Operation int

const (
	Name Operation = iota
	GetConfig
	ListReferrers
	GetBlobContent
	GetReferenceManifest
	GetSubjectDescriptor
)

type orasStoreWithInMemoryCache struct {
	base      referrerstore.ReferrerStore
	cache     *ristretto.Cache
	cacheConf *OrasCacheConf
}

type OrasCacheConf struct {
	Enabled bool `json:"cacheEnabled"`
	Ttl     int  `json:"ttl"`
}

type orasStoreFactoryWithCache struct{}

func keyToHash(key interface{}) (uint64, uint64) {
	if key == nil {
		return 0, 0
	}
	switch k := key.(type) {
	case string:
		return z.MemHashString(k), xxhash.Sum64String(k)
	default:
		return 0, 0
	}
}

// Create creates a new oras store decorated with in-memory cache to cache
// results of ListReferrers API.
func (s *orasStoreFactoryWithCache) Create(storeBase referrerstore.ReferrerStore, cacheConf *OrasCacheConf) (referrerstore.ReferrerStore, error) {
	memoryCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e5,
		MaxCost:     1 << 30,
		BufferItems: 64,
		KeyToHash:   keyToHash,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create cache for referrers, err: %v", err)
	}
	return &orasStoreWithInMemoryCache{
		base:      storeBase,
		cache:     memoryCache,
		cacheConf: cacheConf,
	}, nil
}

func (store *orasStoreWithInMemoryCache) Name() string {
	return store.base.Name()
}

func (store *orasStoreWithInMemoryCache) GetConfig() *config.StoreConfig {
	return store.base.GetConfig()
}

func (store *orasStoreWithInMemoryCache) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string, subjectDesc *ocispecs.SubjectDescriptor) (referrerstore.ListReferrersResult, error) {
	val, found := store.cache.Get(getCacheKey(ListReferrers, subjectReference))
	if found {
		result, ok := val.(referrerstore.ListReferrersResult)
		if !ok {
			return referrerstore.ListReferrersResult{}, fmt.Errorf("failed to type assert result: %+v", val)
		}
		return result, nil
	}

	result, err := store.base.ListReferrers(ctx, subjectReference, artifactTypes, nextToken, subjectDesc)
	if err != nil {
		return result, err
	}

	if added := store.cache.SetWithTTL(getCacheKey(ListReferrers, subjectReference), result, 1, time.Duration(store.cacheConf.Ttl)*time.Second); !added {
		logrus.Warnf("failed to add cache with key: %+v, val: %+v", subjectReference, result)
	}

	return result, err
}

func (store *orasStoreWithInMemoryCache) GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	return store.base.GetBlobContent(ctx, subjectReference, digest)
}

func (store *orasStoreWithInMemoryCache) GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error) {
	return store.base.GetReferenceManifest(ctx, subjectReference, referenceDesc)
}

func (store *orasStoreWithInMemoryCache) GetSubjectDescriptor(ctx context.Context, subjectReference common.Reference) (*ocispecs.SubjectDescriptor, error) {
	return store.base.GetSubjectDescriptor(ctx, subjectReference)
}

func toCacheConfig(storePluginConfig map[string]interface{}) (*OrasCacheConf, error) {
	bytes, err := json.Marshal(storePluginConfig)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling store plugin config: %+v to bytes, err: %v", storePluginConfig, err)
	}

	cacheConf := &OrasCacheConf{}
	if err := json.Unmarshal(bytes, &cacheConf); err != nil {
		return nil, fmt.Errorf("failed unmarshalling to Oras cache config, err: %v", err)
	}

	return cacheConf, nil
}

func getCacheKey(op Operation, ref common.Reference) string {
	return fmt.Sprintf("%d+%s@%s", op, ref.Path, ref.Digest)
}
