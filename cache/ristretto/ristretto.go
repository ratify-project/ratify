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

package ristretto

import (
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/deislabs/ratify/cache"
	"github.com/dgraph-io/ristretto"
	"github.com/dgraph-io/ristretto/z"
	"github.com/sirupsen/logrus"
)

type ristrettoFactory struct {
	once sync.Once
}

type ristrettoCache struct {
	memoryCache *ristretto.Cache
}

func init() {
	cache.Register("ristretto", &ristrettoFactory{})
}

func (f *ristrettoFactory) Create(maxSize int, keyNumber int) (cache.CacheProvider, error) {
	var err error
	var memoryCache *ristretto.Cache
	f.once.Do(func() {
		memoryCache, err = ristretto.NewCache(&ristretto.Config{
			NumCounters: int64(keyNumber) * 10,        // number of keys to track frequency. Recommended to 10x the number of total item count in the cache.
			MaxCost:     int64(maxSize) * 1024 * 1024, // Max size in Megabytes.
			BufferItems: 64,                           // number of keys per Get buffer. 64 is recommended by the ristretto library.
			KeyToHash:   keyToHash,
		})
	})
	if err != nil {
		logrus.Errorf("could not create cache, err: %v", err)
		return &ristrettoCache{}, err
	}

	return &ristrettoCache{
		memoryCache: memoryCache,
	}, nil
}

func (r *ristrettoCache) Get(key interface{}) (interface{}, bool) {
	return r.memoryCache.Get(key)
}

func (r *ristrettoCache) Set(key interface{}, value interface{}) bool {
	return r.SetWithTTL(key, value, 0*time.Second)
}

func (r *ristrettoCache) SetWithTTL(key interface{}, value interface{}, ttl time.Duration) bool {
	return r.memoryCache.SetWithTTL(key, value, 1, ttl)
}

func (r *ristrettoCache) Delete(key interface{}) {
	r.memoryCache.Del(key)
}

func (r *ristrettoCache) Clear() {
	r.memoryCache.Clear()
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
