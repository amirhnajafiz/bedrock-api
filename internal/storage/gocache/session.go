package gocache

import (
	"strings"

	"github.com/amirhnajafiz/bedrock-api/internal/storage"
)

// sessionPrefix namespaces all session keys inside the shared KVStorage backend.
const sessionPrefix = "sessions/"

// SessionStore wraps any storage.KVStorage backend and exposes session-specific
// operations.  Using an interface rather than a concrete type means that any
// future backend (Redis, BadgerDB, …) can be swapped in without touching this file.
type SessionStore struct {
	backend storage.KVStorage
}

// NewSessionStore returns a SessionStore backed by the provided KVStorage.
func NewSessionStore(backend storage.KVStorage) *SessionStore {
	return &SessionStore{backend: backend}
}

// SaveSession persists raw session bytes under id.
func (s *SessionStore) SaveSession(id string, data []byte) error {
	return s.backend.Set(sessionPrefix+id, data)
}

// GetSession retrieves the raw bytes for id.
// Returns storage.ErrNotFound when the session does not exist.
func (s *SessionStore) GetSession(id string) ([]byte, error) {
	return s.backend.Get(sessionPrefix + id)
}

// ListSessions returns every stored session keyed by its id (prefix stripped).
func (s *SessionStore) ListSessions() (map[string][]byte, error) {
	raw, err := s.backend.List(sessionPrefix)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte, len(raw))
	for k, v := range raw {
		result[strings.TrimPrefix(k, sessionPrefix)] = v
	}

	return result, nil
}

// DeleteSession removes the session for id.  It is a no-op when id is unknown.
func (s *SessionStore) DeleteSession(id string) error {
	return s.backend.Delete(sessionPrefix + id)
}
