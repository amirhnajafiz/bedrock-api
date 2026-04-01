package zmq

import (
	"github.com/amirhnajafiz/bedrock-api/pkg/models"

	"github.com/zeromq/goczmq"
	"go.uber.org/zap"
)

// socket receiver reads input messages from router and sends them over handler channel.
func (z ZMQServer) socketReceiver(router *goczmq.Sock, channel chan [][]byte) {
	for {
		request, err := router.RecvMessage()
		if err != nil {
			z.Logr.Warn("failed to received message", zap.Error(err))
			continue
		}

		channel <- request
	}
}

// socket sender reads input from handler channel and sends them to router.
func (z ZMQServer) socketSender(router *goczmq.Sock, channel chan [][]byte) {
	for event := range channel {
		if err := router.SendMessage(event); err != nil {
			z.Logr.Warn("failed to send message", zap.Error(err))
			continue
		}
	}
}

// socket handler is the main loop of ZMQ server.
func (z ZMQServer) socketHandler(in chan [][]byte, out chan [][]byte) {
	for event := range in {
		// parse events into packets
		pkt, err := models.PacketFromBytes(event[1])
		if err != nil {
			z.Logr.Warn("failed to parse event", zap.Error(err))
			continue
		}

		// reply empty packets
		if pkt.IsEmpty() {
			out <- [][]byte{event[0], pkt.ToBytes()}
			continue
		}

		// check daemon registration
		if val, ok := pkt.Headers["register_daemon"]; ok {
			z.Scheduler.Append(val)
			z.Logr.Info("new daemon registered", zap.String("name", val))

			out <- [][]byte{event[0], pkt.ToBytes()}
			continue
		}

		// TODO: read sessions, update KV storage, respond with sender sessions
		dockerd := ""
		if val, ok := pkt.Headers["sender"]; !ok {
			z.Logr.Warn("sender header is missing")

			out <- [][]byte{event[0], pkt.ToBytes()}
			continue
		} else {
			dockerd = val
		}

		// read sessions from packet and update KV storage
		for _, session := range pkt.Sessions {
			tmp, err := z.SessionStore.GetSession(session.Id, dockerd)
			if err != nil {
				z.Logr.Warn("failed to get session", zap.Error(err))
				continue
			}

			// update the session in KV storage
			if err := z.SessionStore.SaveSession(tmp.Id, dockerd, tmp); err != nil {
				z.Logr.Warn("failed to update session", zap.Error(err))
				continue
			}
		}
	}
}
