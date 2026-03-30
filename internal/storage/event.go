package storage

// EventStatus represents the processing state of a stored event.
type EventStatus string

const (
	// EventStatusPending marks an event that has not yet been processed.
	EventStatusPending EventStatus = "pending"
	// EventStatusProcessed marks an event that has been successfully processed.
	EventStatusProcessed EventStatus = "processed"
)

// EventStore provides domain-specific access to internal API events.
// Events are stored as opaque byte slices, with their processing state
// tracked separately so the payload bytes remain untouched.
type EventStore interface {
	// SaveEvent persists raw event bytes under the given id.
	// Newly saved events start in the EventStatusPending state.
	// Calling SaveEvent with an existing id overwrites the payload but
	// resets the status back to pending.
	SaveEvent(id string, data []byte) error
	// GetEvent retrieves the raw bytes for id. Returns ErrNotFound when absent.
	GetEvent(id string) ([]byte, error)
	// ListEvents returns every stored event keyed by its id.
	ListEvents() (map[string][]byte, error)
	// DeleteEvent removes the event and its associated status for id.
	// It is a no-op when id is unknown.
	DeleteEvent(id string) error
	// MarkProcessed transitions the event identified by id to EventStatusProcessed.
	// Returns ErrNotFound when the event does not exist.
	MarkProcessed(id string) error
	// MarkPending transitions the event identified by id back to EventStatusPending.
	// Returns ErrNotFound when the event does not exist.
	MarkPending(id string) error
	// GetStatus returns the current EventStatus for id.
	// Returns ErrNotFound when the event does not exist.
	GetStatus(id string) (EventStatus, error)
}
