package storage

// SessionStore provides domain-specific access to session data.
// Sessions are stored as opaque byte slices so the storage layer remains
// agnostic of the serialisation format chosen by higher-level code.
type SessionStore interface {
	// SaveSession persists raw session bytes under the given id.
	// Calling SaveSession with an id that already exists overwrites the entry.
	SaveSession(id string, data []byte) error
	// GetSession retrieves the raw bytes for id. Returns ErrNotFound when absent.
	GetSession(id string) ([]byte, error)
	// ListSessions returns every stored session keyed by its id.
	ListSessions() (map[string][]byte, error)
	// DeleteSession removes the session for id. It is a no-op when id is unknown.
	DeleteSession(id string) error
}
