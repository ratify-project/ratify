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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/deislabs/ratify/pkg/cache"
	redislib "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type redisFactory struct{}

type redisCache struct {
	redisClient *redislib.Client
}

func init() {
	cache.Register("redis", &redisFactory{})
}

func (f *redisFactory) Create(ctx context.Context, cacheEndpoint string, cacheSize int) (cache.CacheProvider, error) {
	client := redislib.NewClient(&redislib.Options{
		Addr:     cacheEndpoint,
		Password: "",
		DB:       0,
	})
	pong, err := client.Ping(ctx).Result()
	logrus.Debug("Redis Ping: ", pong)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize redis cache client: %w", err)
	}
	return &redisCache{redisClient: client}, nil
}

func (r *redisCache) Get(ctx context.Context, key string) (string, bool) {
	val, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if !errors.Is(err, redislib.Nil) {
			logrus.Error("Error getting key from redis: ", err)
		}
		return "", false
	}
	return val, true
}

func (r *redisCache) Set(ctx context.Context, key string, value interface{}) bool {
	bytes, err := json.Marshal(value)
	if err != nil {
		logrus.Error("Error marshalling value for redis: ", err)
		return false
	}
	if err := r.redisClient.Set(ctx, key, bytes, 0).Err(); err != nil {
		logrus.Error("Error setting key in redis: ", err)
		return false
	}
	return true
}

func (r *redisCache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) bool {
	bytes, err := json.Marshal(value)
	if err != nil {
		logrus.Error("Error marshalling value for redis: ", err)
		return false
	}
	if err := r.redisClient.Set(ctx, key, bytes, ttl).Err(); err != nil {
		logrus.Error("Error setting key in redis with ttl: ", err)
		return false
	}
	return true
}

func (r *redisCache) Delete(ctx context.Context, key string) bool {
	if err := r.redisClient.Del(ctx, key).Err(); err != nil {
		logrus.Error("Error deleting key in redis: ", err)
		return false
	}
	return true
}

func (r *redisCache) Clear(ctx context.Context) bool {
	if err := r.redisClient.FlushDB(ctx).Err(); err != nil {
		logrus.Error("Error clearing redis: ", err)
		return false
	}
	return true
}
