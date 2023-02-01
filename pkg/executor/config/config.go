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

package config

// CacheConfig represents the configuration of the executor cache.
type CacheConfig struct {
	// TTL is default to 10 seconds.
	TTL int `json:"ttl"`
	// MaxSize is the maximum entries allowed in the cache. The default maxSize is 100.
	MaxSize int `json:"maxSize"`
}

// ExecutorConfig represents the configuration for the executor
type ExecutorConfig struct {
	// Gatekeeper default verification webhook timeout is 3 seconds. 100ms network buffer added
	VerificationRequestTimeout *int `json:"verificationRequestTimeout"`
	// Gatekeeper default mutation webhook timeout is 1 seconds. 50ms network buffer added
	MutationRequestTimeout *int `json:"mutationRequestTimeout"`
	// CacheConfig is the configuration of the cache.
	CacheConfig CacheConfig `json:"cacheConfig"`
}
