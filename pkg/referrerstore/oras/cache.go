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
	"time"

	"github.com/deislabs/ratify/cache"
	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"

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

const (
	defaultTTL      = 10
	defaultCapacity = 100 * 1024 * 1024 // 100 Megabytes
)

type orasStoreWithInMemoryCache struct {
	referrerstore.ReferrerStore
	cacheConf *cacheConf
}

type cacheConf struct {
	Enabled bool `json:"cacheEnabled"`
	TTL     int  `json:"ttl"`
}

// createCachedStore creates a new oras store decorated with in-memory cache to cache
// results of ListReferrers API.
func createCachedStore(storeBase referrerstore.ReferrerStore, cacheConf *cacheConf) (referrerstore.ReferrerStore, error) {
	return &orasStoreWithInMemoryCache{
		ReferrerStore: storeBase,
		cacheConf:     cacheConf,
	}, nil
}

func (store *orasStoreWithInMemoryCache) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string, subjectDesc *ocispecs.SubjectDescriptor) (referrerstore.ListReferrersResult, error) {
	cacheProvider, err := cache.GetCacheProvider()
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}
	val, found := cacheProvider.Get(fmt.Sprintf(cache.CacheKeyListReferrers, subjectReference.Original))
	if found {
		if result, ok := val.(referrerstore.ListReferrersResult); ok {
			return result, nil
		}
	}

	result, err := store.ReferrerStore.ListReferrers(ctx, subjectReference, artifactTypes, nextToken, subjectDesc)
	if err == nil {
		cacheKey := fmt.Sprintf(cache.CacheKeyListReferrers, subjectReference.Original)
		if added := cacheProvider.SetWithTTL(cacheKey, result, time.Duration(store.cacheConf.TTL)*time.Second); !added { // TODO: convert ttl to duration in helm values
			logrus.WithContext(ctx).Warnf("failed to add cache with key: %+v, val: %+v", cacheKey, result)
		}
	}

	return result, err
}

func (store *orasStoreWithInMemoryCache) GetSubjectDescriptor(ctx context.Context, subjectReference common.Reference) (*ocispecs.SubjectDescriptor, error) {
	cacheProvider, err := cache.GetCacheProvider()
	if err != nil {
		return nil, err
	}
	val, found := cacheProvider.Get(fmt.Sprintf(cache.CacheKeySubjectDescriptor, subjectReference.Digest))
	if found {
		if result, ok := val.(ocispecs.SubjectDescriptor); ok {
			return &result, nil
		}
	}
	logrus.Debugf("no digest provided for reference %s. attempting to resolve...", subjectReference.Original)
	result, err := store.ReferrerStore.GetSubjectDescriptor(ctx, subjectReference)
	if err == nil {
		cacheKey := fmt.Sprintf(cache.CacheKeySubjectDescriptor, result.Digest)
		if added := cacheProvider.Set(cacheKey, *result); !added {
			logrus.WithContext(ctx).Warnf("failed to add cache with key: %+v, val: %+v", cacheKey, result)
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

	return cacheConf, nil
}
