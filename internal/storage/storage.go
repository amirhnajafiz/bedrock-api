package storage

import (
	"time"

	"github.com/amirhnajafiz/bedrock-api/internal/storage/gocache"
)

var (
	// goCacheBackendInstance is a singleton instance of the go-cache backend for KVStorage.
	goCacheBackendInstance KVStorage
)

// KVStorage represents a key-value storage backend.
// All implementations must be safe for concurrent use.
type KVStorage interface {
	// Set stores value under key, overwriting any existing entry.
	Set(key string, value []byte) error
	// Get retrieves the value for key. Returns ErrNotFound when the key is absent.
	Get(key string) ([]byte, error)
	// Delete removes the entry for key. It is a no-op when the key does not exist.
	Delete(key string) error
	// List retrieves the values for a matching wildcard with the key.
	List(wildcard string) ([][]byte, error)
}

// NewGoCache returns a singleton instance of a KVStorage implementation using go-cache as the backend.
func NewGoCache() KVStorage {
	if goCacheBackendInstance == nil {
		goCacheBackendInstance = gocache.NewBackend(1 * time.Minute)
	}

	return goCacheBackendInstance
}
