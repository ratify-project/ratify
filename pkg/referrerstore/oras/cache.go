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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"

	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/ristretto"
	"github.com/dgraph-io/ristretto/z"
	"github.com/sirupsen/logrus"
)

var (
	memoryCache *ristretto.Cache
	once        sync.Once
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

const (
	defaultTTL       = 10
	defaultCapacity  = 100 * 1024 * 1024 // 100 Megabytes
	defaultKeyNumber = 10000
)

type orasStoreWithInMemoryCache struct {
	referrerstore.ReferrerStore
	cacheConf *cacheConf
}

type cacheConf struct {
	Enabled   bool `json:"cacheEnabled"`
	TTL       int  `json:"ttl"`
	Capacity  int  `json:"capacity"`
	KeyNumber int  `json:"keyNumber"`
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
func createCachedStore(storeBase referrerstore.ReferrerStore, cacheConf *cacheConf) (referrerstore.ReferrerStore, error) {
	var err error
	once.Do(func() {
		memoryCache, err = ristretto.NewCache(&ristretto.Config{
			NumCounters: int64(cacheConf.KeyNumber) * 10,         // number of keys to track frequency. Recommended to 10x the number of total item count in the cache.
			MaxCost:     int64(cacheConf.Capacity) * 1024 * 1024, // Max size in Megabytes.
			BufferItems: 64,                                      // number of keys per Get buffer. 64 is recommended by the ristretto library.
			KeyToHash:   keyToHash,
		})
	})
	if err != nil {
		logrus.Errorf("could not create cache for referrers, err: %v", err)
		return &orasStoreWithInMemoryCache{}, err
	}

	return &orasStoreWithInMemoryCache{
		ReferrerStore: storeBase,
		cacheConf:     cacheConf,
	}, nil
}

func (store *orasStoreWithInMemoryCache) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string, subjectDesc *ocispecs.SubjectDescriptor) (referrerstore.ListReferrersResult, error) {
	val, found := memoryCache.Get(getCacheKey(operationListReferrers, subjectReference))
	if found {
		if result, ok := val.(referrerstore.ListReferrersResult); ok {
			return result, nil
		}
	}

	result, err := store.ReferrerStore.ListReferrers(ctx, subjectReference, artifactTypes, nextToken, subjectDesc)
	if err == nil {
		if added := memoryCache.SetWithTTL(getCacheKey(operationListReferrers, subjectReference), result, 1, time.Duration(store.cacheConf.TTL)*time.Second); !added {
			logrus.WithContext(ctx).Warnf("failed to add cache with key: %+v, val: %+v", subjectReference, result)
		}
	}

	return result, err
}

func toCacheConfig(storePluginConfig map[string]interface{}) (*cacheConf, error) {
	bytes, err := json.Marshal(storePluginConfig)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling store plugin config: %+v to bytes, err: %w", storePluginConfig, err)
	}

	cacheConf := &cacheConf{}
	if err := json.Unmarshal(bytes, cacheConf); err != nil {
		return nil, fmt.Errorf("failed unmarshalling to Oras cache config, err: %w", err)
	}

	if cacheConf.TTL == 0 {
		cacheConf.TTL = defaultTTL
	}
	if cacheConf.Capacity == 0 {
		cacheConf.Capacity = defaultCapacity
	}
	if cacheConf.KeyNumber == 0 {
		cacheConf.KeyNumber = defaultKeyNumber
	}

	return cacheConf, nil
}

func getCacheKey(op operation, ref common.Reference) string {
	return fmt.Sprintf("%d+%s@%s", op, ref.Path, ref.Digest)
}
