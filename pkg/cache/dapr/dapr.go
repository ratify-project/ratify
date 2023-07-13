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

package dapr

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/deislabs/ratify/pkg/cache"
	"github.com/deislabs/ratify/pkg/featureflag"
	"github.com/sirupsen/logrus"
)

const DaprCacheType = "dapr"

type factory struct{}

type daprCache struct {
	daprClient client.Client
	cacheName  string
}

func init() {
	cache.Register(DaprCacheType, &factory{})
}

func (factory *factory) Create(_ context.Context, cacheName string, _ int) (cache.CacheProvider, error) {
	if !featureflag.DaprCacheProvider.Enabled {
		return nil, fmt.Errorf("Dapr cache provider is not enabled. Please set the environment variable RATIFY_DAPR_CACHE_PROVIDER to enable it")
	}
	daprClient, err := client.NewClient()
	if err != nil {
		return nil, err
	}

	return &daprCache{
		daprClient: daprClient,
		cacheName:  cacheName,
	}, nil
}

func (d *daprCache) Get(ctx context.Context, key string) (string, bool) {
	item, err := d.daprClient.GetState(ctx, d.cacheName, key, nil)
	if err != nil {
		return "", false
	}
	return string(item.Value), true
}

func (d *daprCache) Set(ctx context.Context, key string, value interface{}) bool {
	bytes, err := json.Marshal(value)
	if err != nil {
		logrus.Error("Error marshalling value for redis: ", err)
		return false
	}
	if err := d.daprClient.SaveState(ctx, d.cacheName, key, bytes, nil); err != nil {
		logrus.Error("Error saving value to redis: ", err)
		return false
	}
	return true
}

func (d *daprCache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) bool {
	bytes, err := json.Marshal(value)
	if err != nil {
		logrus.Error("Error marshalling value for redis: ", err)
		return false
	}
	ttlString := strconv.Itoa(int(ttl.Seconds()))
	md := map[string]string{"ttlInSeconds": ttlString}
	if err := d.daprClient.SaveState(ctx, d.cacheName, key, bytes, md); err != nil {
		logrus.Error("Error saving value to redis: ", err)
		return false
	}
	return true
}

func (d *daprCache) Delete(ctx context.Context, key string) bool {
	if err := d.daprClient.DeleteState(ctx, d.cacheName, key, nil); err != nil {
		logrus.Error("Error deleting value from redis: ", err)
		return false
	}
	return true
}
