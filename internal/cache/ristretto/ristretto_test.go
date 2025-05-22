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
	"context"
	"testing"
	"time"
)

const (
	testKey   = "testKey"
	testValue = "testValue"
)

func TestCache(t *testing.T) {
	// Create a cache with negative TTL.
	_, err := NewRistrettoCache(-1)
	if err == nil {
		t.Errorf("expected error for negative TTL, got nil")
	}

	// Create a cache with valid TTL.
	cache, err := NewRistrettoCache(1 * time.Second)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err = cache.Set(context.Background(), testKey, testValue); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	val, err := cache.Get(context.Background(), testKey)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if val != testValue {
		t.Errorf("expected %v, got %v", testValue, val)
	}
	if err = cache.Delete(context.Background(), testKey); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	_, err = cache.Get(context.Background(), testKey)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// Test cache expiration.
	if err = cache.Set(context.Background(), testKey, testValue); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	time.Sleep(3 * time.Second)
	val, err = cache.Get(context.Background(), testKey)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if val != nil {
		t.Errorf("expected nil, got %v", val)
	}
}
