package storage

// KVStorage represents a key-value storage system.
type KVStorage interface {
	// Set sets the value for a given key.
	Set(key string, value []byte) error
	// Get retrieves the value for a given key.
	Get(key string) ([]byte, error)
	// Delete removes the key-value pair for a given key.
	Delete(key string) error
	// List retrieves the values for a matching wildcard with the key.
	List(wildcard string) ([][]byte, error)
}
