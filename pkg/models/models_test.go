package models_test

import (
	"testing"

	"github.com/amirhnajafiz/bedrock-api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPacket tests the Packet struct and its methods.
func TestPacket(t *testing.T) {
	t.Run("Packet can be created and converted to bytes and back", func(t *testing.T) {
		original := models.NewPacket().
			WithSender("test-sender").
			WithSessions(
				models.Session{Id: "session1"},
				models.Session{Id: "session2"},
			)

		bytes := original.ToBytes()

		converted, err := models.PacketFromBytes(bytes)
		require.NoError(t, err)

		assert.Equal(t, original.Headers, converted.Headers)
		assert.Equal(t, original.Sessions, converted.Sessions)
	})

	t.Run("IsEmpty returns true for a packet with no headers", func(t *testing.T) {
		packet := models.NewPacket()
		assert.True(t, packet.IsEmpty())
	})

	t.Run("IsEmpty returns false for a packet with headers", func(t *testing.T) {
		packet := models.NewPacket().WithSender("test-sender")
		assert.False(t, packet.IsEmpty())
	})
}
