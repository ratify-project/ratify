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
	"github.com/dgraph-io/ristretto"
	"github.com/dgraph-io/ristretto/z"
	"github.com/sirupsen/logrus"
)

// operation defines the API operations of ReferrerStore.
type operation int

const (
	operationName operation = iota
	operationGetConfig
	operationListReferrers
	operationGetBlobContent
	operationGetReferenceManifest
	operationGetSubjectDescriptor
)

type orasStoreWithInMemoryCache struct {
	referrerstore.ReferrerStore
	cache     *ristretto.Cache
	cacheConf *CacheConf
}

type CacheConf struct {
	Enabled bool `json:"cacheEnabled"`
	Ttl     int  `json:"ttl"`
}

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

// createCachedStore creates a new oras store decorated with in-memory cache to cache
// results of ListReferrers API.
func createCachedStore(storeBase referrerstore.ReferrerStore, cacheConf *CacheConf) (referrerstore.ReferrerStore, error) {
	memoryCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e5,
		MaxCost:     1 << 30,
		BufferItems: 64,
		KeyToHash:   keyToHash,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create cache for referrers, err: %w", err)
	}
	return &orasStoreWithInMemoryCache{
		ReferrerStore: storeBase,
		cache:         memoryCache,
		cacheConf:     cacheConf,
	}, nil
}

func (store *orasStoreWithInMemoryCache) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string, subjectDesc *ocispecs.SubjectDescriptor) (referrerstore.ListReferrersResult, error) {
	val, found := store.cache.Get(getCacheKey(operationListReferrers, subjectReference))
	if found {
		if result, ok := val.(referrerstore.ListReferrersResult); ok {
			return result, nil
		}
	}

	result, err := store.ReferrerStore.ListReferrers(ctx, subjectReference, artifactTypes, nextToken, subjectDesc)
	
	if err == nil {
		if added := store.cache.SetWithTTL(getCacheKey(operationListReferrers, subjectReference), result, 1, time.Duration(store.cacheConf.Ttl)*time.Second); !added {
			logrus.WithContext(ctx).Warnf("failed to add cache with key: %+v, val: %+v", subjectReference, result)
		}
	}

	return result, err
}

func toCacheConfig(storePluginConfig map[string]interface{}) (*CacheConf, error) {
	bytes, err := json.Marshal(storePluginConfig)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling store plugin config: %+v to bytes, err: %w", storePluginConfig, err)
	}

	cacheConf := &CacheConf{}
	if err := json.Unmarshal(bytes, &cacheConf); err != nil {
		return nil, fmt.Errorf("failed unmarshalling to Oras cache config, err: %w", err)
	}

	return cacheConf, nil
}

func getCacheKey(op operation, ref common.Reference) string {
	return fmt.Sprintf("%d+%s@%s", op, ref.Path, ref.Digest)
}
