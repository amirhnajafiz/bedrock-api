package models

import "encoding/json"

// Packet represents a collection of sessions to be sent together over ZMQ.
type Packet struct {
	Headers  map[string]string `json:"headers"`
	Sessions []Session         `json:"sessions"`
}

// NewPacket creates and returns a new Packet instance.
func NewPacket() Packet {
	return Packet{
		Headers:  make(map[string]string),
		Sessions: nil,
	}
}

// WithSender adds the sender to the packet.
func (p Packet) WithSender(sender string) Packet {
	p.Headers["sender"] = sender
	return p
}

// WithRegisterDaemon adds the register daemon header.
func (p Packet) WithRegisterDaemon(name string) Packet {
	p.Headers["register_daemon"] = name
	return p
}

// WithSessions adds sessions to the packet.
func (p Packet) WithSessions(sessions ...Session) Packet {
	if p.Sessions == nil {
		p.Sessions = make([]Session, 0)
	}

	for _, session := range sessions {
		p.Sessions = append(p.Sessions, session)
	}

	return p
}

// ToBytes converts the Packet struct to a byte slice.
func (p Packet) ToBytes() []byte {
	b, _ := json.Marshal(p)
	return b
}

// PacketFromBytes converts a byte slice to a Packet struct.
func PacketFromBytes(bytes []byte) (*Packet, error) {
	var packet Packet

	err := json.Unmarshal(bytes, &packet)
	if err != nil {
		return nil, err
	}

	return &packet, nil
}
