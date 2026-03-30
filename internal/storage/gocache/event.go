package gocache

import (
	"strings"

	"github.com/amirhnajafiz/bedrock-api/internal/storage"
)

const (
	// eventPrefix namespaces all event payload keys.
	eventPrefix = "events/"
	// statusSuffix is appended to an event key to store its processing status.
	statusSuffix = "/status"
)

// EventStore wraps any storage.KVStorage backend and exposes event-specific
// operations, including tracking a per-event processing status.
type EventStore struct {
	backend storage.KVStorage
}

// NewEventStore returns an EventStore backed by the provided KVStorage.
func NewEventStore(backend storage.KVStorage) *EventStore {
	return &EventStore{backend: backend}
}

// SaveEvent persists raw event bytes under id and initialises its status to
// EventStatusPending.  Calling SaveEvent on an existing id overwrites both the
// payload and resets the status back to pending.
func (e *EventStore) SaveEvent(id string, data []byte) error {
	if err := e.backend.Set(eventPrefix+id, data); err != nil {
		return err
	}

	return e.backend.Set(eventPrefix+id+statusSuffix, []byte(storage.EventStatusPending))
}

// GetEvent retrieves the raw payload bytes for id.
// Returns storage.ErrNotFound when the event does not exist.
func (e *EventStore) GetEvent(id string) ([]byte, error) {
	return e.backend.Get(eventPrefix + id)
}

// ListEvents returns every stored event keyed by its id (prefix stripped).
// Status entries are excluded from the result.
func (e *EventStore) ListEvents() (map[string][]byte, error) {
	raw, err := e.backend.List(eventPrefix)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte)
	for k, v := range raw {
		// Skip status keys – they share the same prefix but end with statusSuffix.
		if strings.HasSuffix(k, statusSuffix) {
			continue
		}

		result[strings.TrimPrefix(k, eventPrefix)] = v
	}

	return result, nil
}

// DeleteEvent removes the event payload and its associated status for id.
// It is a no-op when id is unknown.
func (e *EventStore) DeleteEvent(id string) error {
	if err := e.backend.Delete(eventPrefix + id); err != nil {
		return err
	}

	return e.backend.Delete(eventPrefix + id + statusSuffix)
}

// MarkProcessed transitions the event identified by id to EventStatusProcessed.
// Returns storage.ErrNotFound when the event does not exist.
func (e *EventStore) MarkProcessed(id string) error {
	return e.setStatus(id, storage.EventStatusProcessed)
}

// MarkPending transitions the event identified by id back to EventStatusPending.
// Returns storage.ErrNotFound when the event does not exist.
func (e *EventStore) MarkPending(id string) error {
	return e.setStatus(id, storage.EventStatusPending)
}

// GetStatus returns the current EventStatus for id.
// Returns storage.ErrNotFound when the event does not exist.
func (e *EventStore) GetStatus(id string) (storage.EventStatus, error) {
	raw, err := e.backend.Get(eventPrefix + id + statusSuffix)
	if err != nil {
		return "", err
	}

	return storage.EventStatus(raw), nil
}

// setStatus is the shared helper for MarkProcessed / MarkPending.
// It verifies the event payload exists before touching the status key so that
// callers cannot create orphaned status entries.
func (e *EventStore) setStatus(id string, status storage.EventStatus) error {
	// Confirm the event payload exists before updating its status.
	if _, err := e.backend.Get(eventPrefix + id); err != nil {
		return err
	}

	return e.backend.Set(eventPrefix+id+statusSuffix, []byte(status))
}
