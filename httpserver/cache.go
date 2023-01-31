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

package httpserver

import (
	"time"

	"github.com/deislabs/ratify/pkg/executor/types"
	utilcache "k8s.io/apimachinery/pkg/util/cache"
)

type simpleCache struct {
	cache   *utilcache.LRUExpireCache
	ttl     time.Duration
	maxSize int
}

func newSimpleCache(ttl time.Duration, maxSize int) *simpleCache {
	return &simpleCache{
		cache:   utilcache.NewLRUExpireCache(maxSize),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// given a key, return the item from the cache if it exists and has not expired
func (c *simpleCache) get(key string) *types.VerifyResult {
	if item, ok := c.cache.Get(key); ok {
		return item.(*types.VerifyResult)
	}
	return nil
}

// given a key, set the item in the cache
func (c *simpleCache) set(key string, item *types.VerifyResult) {
	c.cache.Add(key, item, c.ttl)
}
