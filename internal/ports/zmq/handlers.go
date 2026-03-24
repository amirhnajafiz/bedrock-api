package zmq

import "github.com/zeromq/goczmq"

func (z ZMQServer) socketReceiver(router *goczmq.Sock, channel chan [][]byte) {
	for {
		request, err := router.RecvMessage()
		if err != nil {
			continue
		}

		channel <- request
	}
}

func (z ZMQServer) socketSender(router *goczmq.Sock, channel chan [][]byte) {
	for event := range channel {
		if err := router.SendMessage(event); err != nil {
			continue
		}
	}
}

func (z ZMQServer) socketHandler(in chan [][]byte, out chan [][]byte) {
	for event := range in {
		out <- event
	}
}
