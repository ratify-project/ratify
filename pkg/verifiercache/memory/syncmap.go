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

package memory

import (
	"sync"
	"time"
)

const defaultEvictionPercentage int = 5 //The default eviction percentage used when map reaches its capacity at insertion

// itemWithAge is a cache item whose age is tracked.
type itemWithAge struct {
	item   interface{}
	expiry time.Time
}

// expired indicates if the cached item is expired and should be deleted.
func (et itemWithAge) expired() bool {
	return time.Now().After(et.expiry)
}

// SyncMapWithExpiration enhances SyncMap with expiry-tracked cached items.
type SyncMapWithExpiration struct {
	// The underlying SyncMap.
	s *SyncMap
}

// GetEntry gets an active entry.
func (sme *SyncMapWithExpiration) GetEntry(key string) (interface{}, bool) {
	if entry, ok := sme.s.GetEntry(key); ok {
		if trackedEntry, ok := entry.(itemWithAge); ok {
			if !trackedEntry.expired() {
				return trackedEntry.item, true
			}
			// Clean up expired cached item.
			sme.s.DeleteEntry(key)
		}
	}
	return nil, false
}

// SetEntry upserts a new cache entry, with the given ttl.
func (sme *SyncMapWithExpiration) SetEntry(key string, item interface{}, ttl time.Duration) {
	if ttl <= 0 {
		return
	}

	trackedEntry := itemWithAge{
		item:   item,
		expiry: time.Now().Add(ttl),
	}
	sme.s.SetEntry(key, trackedEntry)
}

// SyncMap is a map with synchronized access support
type SyncMap struct {
	mapObj             *map[string]interface{} //map containing data
	lock               *sync.RWMutex           //Simple mutex used to guard the read/write operation.
	capacity           int                     //Max capacity of the map
	evictionPercentage int                     //Eviction percentage used when the map is full during insertion
}

// GetEntry gets an entry
func (sm *SyncMap) GetEntry(key string) (entry interface{}, ok bool) {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	entry, ok = (*sm.mapObj)[key]
	return
}

// SetEntry adds a new entry or update existing entry
func (sm *SyncMap) SetEntry(key string, entry interface{}) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	if _, ok := (*sm.mapObj)[key]; !ok { //We will need to add an entry
		if numEntries := len(*sm.mapObj); numEntries >= sm.capacity { //exceeding capacity, remove evictionPercentage of the entries
			numToEvict := numEntries * sm.evictionPercentage / 100
			if numToEvict <= 1 { //We will evict one as the minimum
				numToEvict = 1
			}
			numEvicted := 0
			for k := range *sm.mapObj { // GO map iterator will randomize the order. We just delete the first in the iterator
				delete(*sm.mapObj, k)
				numEvicted++
				if numEvicted >= numToEvict {
					break
				}
			}
		}
	}

	(*sm.mapObj)[key] = entry
	return
}

// DeleteEntry delete an entry
// TODO evaluate need for delete
func (sm *SyncMap) DeleteEntry(key string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	delete(*sm.mapObj, key)
	return
}

// SetMapObj sets the whole mapping object directly
func (sm *SyncMap) SetMapObj(newMap *map[string]interface{}) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	sm.mapObj = newMap
	return
}

// GetLength get the number of items in the map
func (sm *SyncMap) GetLength() int {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	return len(*sm.mapObj)
}

// SetEvictionPercentage sets the new percentage of entries to be evicted when map reaches its capacity
// When the map is full during insertion, the minimum number of entries to be evicted will be 1
func (sm *SyncMap) SetEvictionPercentage(newPercentage int) {
	if newPercentage < 0 {
		newPercentage = 0
	} else if newPercentage > 100 {
		newPercentage = 100
	}

	sm.lock.Lock()
	defer sm.lock.Unlock()
	sm.evictionPercentage = newPercentage
	return
}

// MakeSyncMap create a new SyncMap
// if maxEntries is 0, the capacity if override as 1
func MakeSyncMap(maxEntries int) SyncMap {
	if maxEntries <= 0 {
		maxEntries = 1
	}
	return SyncMap{mapObj: &map[string]interface{}{},
		lock:               &sync.RWMutex{},
		capacity:           maxEntries,
		evictionPercentage: defaultEvictionPercentage}
}

// NewSyncMapWithExpiration creates a new SyncMapWithExpiration.
func NewSyncMapWithExpiration(maxEntries int) *SyncMapWithExpiration {
	s := MakeSyncMap(maxEntries)
	return &SyncMapWithExpiration{s: &s}
}
