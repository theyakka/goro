// cache.go
// Goro
//
// Created by Posse in NYC
// http://goposse.com
//
// Copyright (c) 2016 Posse Productions LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

import (
	"hash/fnv"
	"sync"
)

// RouteCache - temporary storage for routes
type RouteCache struct {
	// mutex - locking
	mutex sync.RWMutex

	// containsMap - indexed positions (starting at 1) for each of the hashed path values
	containsMap map[uint32]bool

	// pathHashes - ordered list of pathHashes
	pathHashes []uint32

	// Entries - ordered list of cache entries
	Entries []CacheEntry

	// MaxEntries - maximum number of items permitted in the cache
	MaxEntries int

	// ReorderOnAccess - move the last accessed item to the top
	ReorderOnAccess bool
}

// CacheEntry - an entry in the route cache
type CacheEntry struct {
	hasValue bool
	Params   map[string]interface{}
	Route    *Route
}

// NewRouteCache - creates a new default RouteCache
func NewRouteCache() *RouteCache {
	return &RouteCache{
		pathHashes:      []uint32{},
		Entries:         []CacheEntry{},
		containsMap:     map[uint32]bool{},
		MaxEntries:      100,
		ReorderOnAccess: true,
	}
}

// Get - fetch a cache entry (if exists)
func (rc *RouteCache) Get(path string) CacheEntry {
	cacheEntry := CacheEntry{}
	hash := fnv.New32a()
	if rc.ReorderOnAccess {
		rc.mutex.Lock()
		defer rc.mutex.Unlock()
	} else {
		rc.mutex.RLock()
		defer rc.mutex.RUnlock()
	}
	hash.Write([]byte(path))
	pathHash := hash.Sum32()
	var foundIdx int
	for idx, hashKey := range rc.pathHashes {
		if hashKey == pathHash {
			cacheEntry = rc.Entries[idx]
			foundIdx = idx
			break
		}
	}
	if rc.ReorderOnAccess {
		if len(rc.pathHashes) > 1 {
			rc.moveEntryToTop(pathHash, foundIdx)
		}
	}
	return cacheEntry
}

func (rc *RouteCache) moveEntryToTop(pathHash uint32, moveIndex int) {
	allHashes := rc.pathHashes
	allEntries := rc.Entries
	entry := allEntries[moveIndex]
	// remove entry
	lastIndex := moveIndex + 1
	allHashes = append(allHashes[:moveIndex], allHashes[lastIndex:]...)
	allEntries = append(allEntries[:moveIndex], allEntries[lastIndex:]...)
	// re-add entry
	allHashes = append([]uint32{pathHash}, allHashes...)
	allEntries = append([]CacheEntry{entry}, allEntries...)
	rc.pathHashes = allHashes
	rc.Entries = allEntries
}

// PutRoute - add a route into the route cache
func (rc *RouteCache) PutRoute(path string, route *Route) {
	entry := CacheEntry{
		Route: route,
	}
	rc.Put(path, entry)
}

// Put - add an item to the route cache
func (rc *RouteCache) Put(path string, entry CacheEntry) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	hash := fnv.New32a()
	hash.Write([]byte(path))
	allHashes := rc.pathHashes
	allEntries := rc.Entries
	pathHash := hash.Sum32()
	if rc.containsMap[pathHash] {
		return // don't add it again
	}
	if len(allHashes) == rc.MaxEntries {
		// remove the last element from the slices
		removeHash := allHashes[len(allHashes)-1]
		allHashes = allHashes[:len(allHashes)-1]
		allEntries = allEntries[:len(allEntries)-1]
		delete(rc.containsMap, removeHash)
	}
	allHashes = append([]uint32{pathHash}, allHashes...)
	allEntries = append([]CacheEntry{entry}, allEntries...)
	rc.containsMap[pathHash] = true
	rc.pathHashes = allHashes
	rc.Entries = allEntries
}

// Clear - reset the cache
func (rc *RouteCache) Clear() {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	rc.pathHashes = []uint32{}
	rc.Entries = []CacheEntry{}
	rc.containsMap = map[uint32]bool{}
}

// NotFoundCacheEntry - represents the inability to find an entry in the cache
func NotFoundCacheEntry() CacheEntry {
	return CacheEntry{
		hasValue: false,
	}
}
