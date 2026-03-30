// Package gocache provides an in-memory storage backend built on top of
// github.com/eko/gocache (backed by github.com/patrickmn/go-cache).
// It implements the storage.KVStorage interface and can therefore be used
// to back any higher-level store (SessionStore, EventStore, …) without those
// stores knowing about the underlying cache library.
package gocache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	lib "github.com/eko/gocache/lib/v4/cache"
	lib_store "github.com/eko/gocache/lib/v4/store"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	goc "github.com/patrickmn/go-cache"

	"github.com/amirhnajafiz/bedrock-api/internal/storage"
)

// Backend is a thread-safe, in-memory key-value store backed by eko/gocache
// with a go-cache store adapter.
// The zero value is not usable; create instances with NewBackend.
type Backend struct {
	cache  *lib.Cache[[]byte]
	client *goc.Cache // retained for List prefix scans via Items()
}

// NewBackend returns a Backend with no entry expiration.
// The underlying go-cache janitor runs every cleanupInterval to evict
// entries that were stored with an explicit TTL (none are in this project,
// but the option is there for future use).
func NewBackend(cleanupInterval time.Duration) *Backend {
	client := goc.New(goc.NoExpiration, cleanupInterval)
	store := gocache_store.NewGoCache(client)
	return &Backend{
		cache:  lib.New[[]byte](store),
		client: client,
	}
}

// Set stores value under key, overwriting any existing entry.
func (b *Backend) Set(key string, value []byte) error {
	return b.cache.Set(context.Background(), key, value)
}

// Get retrieves the raw bytes stored under key.
// Returns storage.ErrNotFound when the key is absent.
func (b *Backend) Get(key string) ([]byte, error) {
	val, err := b.cache.Get(context.Background(), key)
	if err != nil {
		if errors.Is(err, lib_store.NotFound{}) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return val, nil
}

// Delete removes the entry for key. It is a no-op when key is absent.
func (b *Backend) Delete(key string) error {
	return b.cache.Delete(context.Background(), key)
}

// List returns the values of all entries whose keys start with prefix.
// An empty prefix returns every value currently in the store.
// The returned slice is a snapshot; mutations to it do not affect the cache.
// eko/gocache does not expose a scan API, so this method accesses the
// underlying go-cache client directly via Items().
func (b *Backend) List(prefix string) ([][]byte, error) {
	items := b.client.Items()
	var result [][]byte

	for k, item := range items {
		if !strings.HasPrefix(k, prefix) {
			continue
		}

		data, ok := item.Object.([]byte)
		if !ok {
			return nil, fmt.Errorf("gocache: unexpected value type for key %q", k)
		}

		result = append(result, data)
	}

	return result, nil
}
