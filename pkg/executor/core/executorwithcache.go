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

package core

import (
	"context"
	"time"

	"github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/executor/config"
	"github.com/deislabs/ratify/pkg/executor/types"
	"github.com/deislabs/ratify/pkg/verifiercache"
	"github.com/deislabs/ratify/pkg/verifiercache/memory"
)

const (
	defaultCacheMaxSize = 100
	defaultCacheTTL     = 10 * time.Second
)

// ExecutorWithCache wraps the executor with a verifier cache
type ExecutorWithCache struct {
	executor.Executor
	verifierCache          verifiercache.VerifierCache
	verfierCacheItemExpiry time.Duration
}

func newExecutorWithCache(executor executor.Executor, config config.CacheConfig) *ExecutorWithCache {
	maxSize := defaultCacheMaxSize
	if config.MaxSize != 0 {
		maxSize = config.MaxSize
	}
	ttl := defaultCacheTTL
	if config.TTL != 0 {
		ttl = time.Duration(config.TTL) * time.Second
	}

	cache := memory.NewMemoryCache(maxSize)
	return &ExecutorWithCache{
		Executor:               executor,
		verifierCache:          cache,
		verfierCacheItemExpiry: ttl,
	}
}

func (executor ExecutorWithCache) VerifySubject(ctx context.Context, verifyParameters executor.VerifyParameters) (types.VerifyResult, error) {
	// check the cache for the existence of item
	cachedResult, ok := executor.verifierCache.GetVerifyResult(ctx, verifyParameters.Subject)

	if ok {
		return cachedResult, nil
	}

	result, err := executor.Executor.VerifySubject(ctx, verifyParameters)

	if err == nil {
		executor.verifierCache.SetVerifyResult(ctx, verifyParameters.Subject, result, executor.verfierCacheItemExpiry)
	}

	return result, err
}
