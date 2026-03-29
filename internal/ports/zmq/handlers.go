package zmq

import (
	"github.com/amirhnajafiz/bedrock-api/pkg/models"

	"github.com/zeromq/goczmq"
	"go.uber.org/zap"
)

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

func (z ZMQServer) socketSender(router *goczmq.Sock, channel chan [][]byte) {
	for event := range channel {
		if err := router.SendMessage(event); err != nil {
			z.Logr.Warn("failed to send message", zap.Error(err))
			continue
		}
	}
}

func (z ZMQServer) socketHandler(in chan [][]byte, out chan [][]byte) {
	for event := range in {
		msg, err := models.PacketFromBytes(event[1])
		if err != nil {
			z.Logr.Warn("failed to parse event", zap.Error(err))
			continue
		}

		if len(msg.Headers) == 0 {
			out <- [][]byte{event[0], msg.ToBytes()}
			continue
		}

		if val, ok := msg.Headers["register_daemon"]; ok {
			z.Scheduler.Append(val)
			z.Logr.Info("new daemon registered", zap.String("name", val))
			out <- [][]byte{event[0], msg.ToBytes()}
			continue
		}
	}
}
