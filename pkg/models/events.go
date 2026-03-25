package models

import "encoding/json"

// Event represents a generic event.
type Event struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload,omitempty"`
}

// ToBytes converts the Event struct to a byte slice.
func (e Event) ToBytes() []byte {
	bytes, _ := json.Marshal(e)
	return bytes
}

// EventFromBytes converts a byte slice to an Event struct.
func EventFromBytes(bytes []byte) (*Event, error) {
	var event Event
	err := json.Unmarshal(bytes, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}
