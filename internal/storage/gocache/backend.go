// Package gocache provides an in-memory storage backend built on top of
// github.com/patrickmn/go-cache.  It implements the storage.KVStorage
// interface and can therefore be used to back any higher-level store
// (SessionStore, EventStore, …) without those stores knowing about the
// underlying cache library.
package gocache

import (
	"fmt"
	"strings"
	"time"

	goc "github.com/patrickmn/go-cache"

	"github.com/amirhnajafiz/bedrock-api/internal/storage"
)

// Backend is a thread-safe, in-memory key-value store backed by go-cache.
// The zero value is not usable; create instances with NewBackend.
type Backend struct {
	cache *goc.Cache
}

// NewBackend returns a Backend with no entry expiration.
// The underlying go-cache janitor runs every cleanupInterval to evict
// entries that were stored with an explicit TTL (none are in this project,
// but the option is there for future use).
func NewBackend(cleanupInterval time.Duration) *Backend {
	return &Backend{
		cache: goc.New(goc.NoExpiration, cleanupInterval),
	}
}

// Set stores value under key, overwriting any existing entry.
// go-cache is safe for concurrent use, so no additional locking is needed.
func (b *Backend) Set(key string, value []byte) error {
	b.cache.Set(key, value, goc.NoExpiration)
	return nil
}

// Get retrieves the raw bytes stored under key.
// Returns storage.ErrNotFound when the key is absent.
func (b *Backend) Get(key string) ([]byte, error) {
	raw, ok := b.cache.Get(key)
	if !ok {
		return nil, storage.ErrNotFound
	}

	data, ok := raw.([]byte)
	if !ok {
		// This should never happen because Set only stores []byte values,
		// but guard against external misuse of the underlying cache.
		return nil, fmt.Errorf("gocache: unexpected value type for key %q", key)
	}

	return data, nil
}

// Delete removes the entry for key.  It is a no-op when key is absent.
func (b *Backend) Delete(key string) error {
	b.cache.Delete(key)
	return nil
}

// List returns all key-value pairs whose keys start with prefix.
// An empty prefix returns every entry currently in the store.
// The returned map is a snapshot; mutations to it do not affect the cache.
func (b *Backend) List(prefix string) (map[string][]byte, error) {
	items := b.cache.Items()
	result := make(map[string][]byte)

	for k, item := range items {
		if !strings.HasPrefix(k, prefix) {
			continue
		}

		data, ok := item.Object.([]byte)
		if !ok {
			return nil, fmt.Errorf("gocache: unexpected value type for key %q", k)
		}

		result[k] = data
	}

	return result, nil
}
