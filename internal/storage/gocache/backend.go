package gocache

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/amirhnajafiz/bedrock-api/pkg/xerrors"

	lib "github.com/eko/gocache/lib/v4/cache"
	lib_store "github.com/eko/gocache/lib/v4/store"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	goc "github.com/patrickmn/go-cache"
)

// Backend is a thread-safe, in-memory key-value store backed by eko/gocache
// with a go-cache store adapter.
type Backend struct {
	cache  *lib.Cache[[]byte]
	client *goc.Cache
}

// NewBackend returns a new Backend instance with the specified cleanup interval for expired entries.
func NewBackend(cleanupInterval time.Duration) *Backend {
	client := goc.New(goc.NoExpiration, cleanupInterval)

	return &Backend{
		cache:  lib.New[[]byte](gocache_store.NewGoCache(client)),
		client: client,
	}
}

// Set stores value under key, overwriting any existing entry.
func (b *Backend) Set(key string, value []byte) error {
	return b.cache.Set(context.Background(), key, value)
}

// Get retrieves the raw bytes stored under key.
func (b *Backend) Get(key string) ([]byte, error) {
	val, err := b.cache.Get(context.Background(), key)
	if err != nil {
		if errors.Is(err, lib_store.NotFound{}) {
			return nil, xerrors.StorageErrNotFound
		}

		return nil, err
	}

	return val, nil
}

// Delete removes the entry for key. It is a no-op when key is absent.
func (b *Backend) Delete(key string) error {
	return b.cache.Delete(context.Background(), key)
}

// List returns the values of all entries whose keys match wildcard.
// If wildcard contains no glob tokens, it is treated as a key prefix.
// If wildcard contains glob tokens ('*', '?', '['), path.Match is used.
func (b *Backend) List(wildcard string) ([][]byte, error) {
	items := b.client.Items()

	var result [][]byte
	hasGlob := strings.ContainsAny(wildcard, "*?[")
	if hasGlob {
		if _, err := path.Match(wildcard, ""); err != nil {
			return nil, xerrors.StorageErrInvalidWildcard
		}
	}

	for k, item := range items {
		if wildcard != "" {
			if hasGlob {
				matched, _ := path.Match(wildcard, k)
				if !matched {
					continue
				}
			} else if !strings.HasPrefix(k, wildcard) {
				continue
			}
		}

		data, ok := item.Object.([]byte)
		if !ok {
			return nil, fmt.Errorf("gocache: unexpected value type for key %q", k)
		}

		result = append(result, data)
	}

	return result, nil
}
