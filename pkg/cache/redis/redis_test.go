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

package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/deislabs/ratify/pkg/cache"
	"github.com/redis/go-redis/v9"
)

// TestSet_Expected tests the Set function
func TestSet_Expected(t *testing.T) {
	var err error
	ctx := context.Background()
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to create miniredis server")
	}
	//defer s.Close()
	time.Sleep(1 * time.Second)
	// first attempt to get the cache provider if it's already been initialized
	cacheProvider := cache.GetCacheProvider()
	if cacheProvider == nil {
		// if no cache provider has been initialized, initialize one
		cacheProvider, err = cache.NewCacheProvider(ctx, "redis", s.Addr(), cache.DefaultCacheSize)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}
	ok := cacheProvider.Set(ctx, "test", "test")
	if !ok {
		t.Errorf("Expected ok to be true")
	}

	redisCache := cacheProvider.(*redisCache)
	val, err := redisCache.redisClient.Get(ctx, "test").Result()
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if val != "\"test\"" {
		t.Errorf("Expected value to be test, but got %s", val)
	}
}

// TestSetWithTTL_Expected tests the SetWithTTL function
func TestSetWithTTL_Expected(t *testing.T) {
	var err error
	ctx := context.Background()
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to create miniredis server")
	}
	time.Sleep(1 * time.Second)
	// first attempt to get the cache provider if it's already been initialized
	cacheProvider := cache.GetCacheProvider()
	if cacheProvider == nil {
		// if no cache provider has been initialized, initialize one
		cacheProvider, err = cache.NewCacheProvider(ctx, "redis", s.Addr(), cache.DefaultCacheSize)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}
	ok := cacheProvider.SetWithTTL(ctx, "test_ttl", "test", 2*time.Second)
	if !ok {
		t.Errorf("Expected ok to be true")
	}

	redisCache := cacheProvider.(*redisCache)
	val, err := redisCache.redisClient.Get(ctx, "test_ttl").Result()
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if val != "\"test\"" {
		t.Errorf("Expected value to be test, but got %s", val)
	}
}

// TestGet_Expected tests the Get function
func TestGet_Expected(t *testing.T) {
	var err error
	ctx := context.Background()
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to create miniredis server")
	}
	time.Sleep(1 * time.Second)
	// first attempt to get the cache provider if it's already been initialized
	cacheProvider := cache.GetCacheProvider()
	if cacheProvider == nil {
		// if no cache provider has been initialized, initialize one
		cacheProvider, err = cache.NewCacheProvider(ctx, "redis", s.Addr(), cache.DefaultCacheSize)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}
	redisCache := cacheProvider.(*redisCache)
	if err = redisCache.redisClient.Set(ctx, "test-get", "test-get-val", 1).Err(); err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	val, ok := cacheProvider.Get(ctx, "test-get")
	if !ok {
		t.Errorf("Expected ok to be true")
	}
	if val != "test-get-val" {
		t.Errorf("Expected value to be test-get-val, but got %s", val)
	}
}

// TestDelete_Expected tests the Delete function
func TestDelete_Expected(t *testing.T) {
	var err error
	ctx := context.Background()
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to create miniredis server")
	}
	time.Sleep(1 * time.Second)
	// first attempt to get the cache provider if it's already been initialized
	cacheProvider := cache.GetCacheProvider()
	if cacheProvider == nil {
		// if no cache provider has been initialized, initialize one
		cacheProvider, err = cache.NewCacheProvider(ctx, "redis", s.Addr(), cache.DefaultCacheSize)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}
	redisCache := cacheProvider.(*redisCache)
	if err = redisCache.redisClient.Set(ctx, "test-delete", "test-delete-val", 1).Err(); err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	ok := cacheProvider.Delete(ctx, "test-delete")
	if !ok {
		t.Errorf("Expected delete operation to succeed")
	}

	if err = redisCache.redisClient.Get(ctx, "test-delete").Err(); !errors.Is(err, redis.Nil) {
		t.Errorf("Expected key to be deleted from redis")
	}
}

// TestClear_Expected tests the Clear function
func TestClear_Expected(t *testing.T) {
	var err error
	ctx := context.Background()
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to create miniredis server")
	}
	time.Sleep(1 * time.Second)
	// first attempt to get the cache provider if it's already been initialized
	cacheProvider := cache.GetCacheProvider()
	if cacheProvider == nil {
		// if no cache provider has been initialized, initialize one
		cacheProvider, err = cache.NewCacheProvider(ctx, "redis", s.Addr(), cache.DefaultCacheSize)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}
	redisCache := cacheProvider.(*redisCache)
	if err = redisCache.redisClient.Set(ctx, "test-clear", "test-clear-val", 1).Err(); err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	ok := cacheProvider.Clear(ctx)
	if !ok {
		t.Errorf("Expected clear operation to succeed")
	}

	if err = redisCache.redisClient.Get(ctx, "test-clear").Err(); !errors.Is(err, redis.Nil) {
		t.Errorf("Expected key to be deleted from redis")
	}
}
