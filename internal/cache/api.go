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

package cache

import (
	"context"
	"errors"
)

var (
	ErrNotFound   = errors.New("cache not found")
	ErrInvalidTTL = errors.New("invalid TTL provided")
	ErrAddFailed  = errors.New("failed to add key/value to cache")
)

// Cache is the main interface for a generic key-value cache.
type Cache interface {
	// Get returns the value associated with the key, or an error if not found.
	Get(ctx context.Context, key string) (any, error)

	// Set stores a value with the specified key.
	Set(ctx context.Context, key string, value any) error

	// Delete removes the specified key/value from the cache.
	Delete(ctx context.Context, key string) error
}
