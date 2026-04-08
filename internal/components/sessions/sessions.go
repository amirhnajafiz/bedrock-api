package sessions

import (
	"github.com/amirhnajafiz/bedrock-api/internal/storage"
	"github.com/amirhnajafiz/bedrock-api/pkg/models"
)

// SessionStore provides domain-specific access to sessions data.
type SessionStore interface {
	// SaveSession persists raw session bytes under the given id, namespaced by
	// the owning Docker daemon's id. Calling SaveSession with the same id and
	// dockerdId overwrites the entry.
	SaveSession(id, dockerdId string, data *models.Session) error

	// GetSession retrieves the raw bytes for id within the given dockerdId namespace.
	// Returns ErrNotFound when absent.
	GetSession(id, dockerdId string) (*models.Session, error)

	// GetSessionById retrieves a session by id across all Docker daemon namespaces.
	// This is useful when the caller does not know which dockerd instance owns
	// the session and wants the store to resolve it. Returns ErrNotFound when
	// absent.
	GetSessionById(id string) (*models.Session, error)

	// ListSessions returns the raw bytes of every stored session across all daemons.
	ListSessions() ([]*models.Session, error)

	// ListSessionsByDockerDId returns the raw bytes of every session belonging to
	// the given Docker daemon instance. Returns an empty slice when none exist.
	ListSessionsByDockerDId(dockerdId string) ([]*models.Session, error)

	// DeleteSession removes the session for id within the given dockerdId namespace.
	// It is a no-op when the entry is unknown.
	DeleteSession(id, dockerdId string) error
}

// NewSessionStore returns a SessionStore backed by the provided KVStorage.
func NewSessionStore(backend storage.KVStorage) SessionStore {
	return &sessionStore{backend: backend}
}
