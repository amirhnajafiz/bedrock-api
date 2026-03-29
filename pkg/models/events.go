package models

type EventType string

const (
	EventCreate EventType = "create"
	EventPatch  EventType = "patch"
)

// Event represents a generic event.
type Event interface {
	Type() EventType
}

// EventCreatePayload represents the payload for creating a new session.
type EventCreatePayload struct {
	SessionId string `json:"sessionId"`
	Image     string `json:"image"`
	Command   string `json:"command"`
	TTL       int    `json:"ttl"`
}

func (p EventCreatePayload) Type() EventType {
	return EventCreate
}

// EventPatchPayload represents the payload for patching an existing session.
type EventPatchPayload struct {
	SessionId string `json:"sessionId"`
	Status    string `json:"status"`
	TTL       int    `json:"ttl"`
}

func (p EventPatchPayload) Type() EventType {
	return EventPatch
}
