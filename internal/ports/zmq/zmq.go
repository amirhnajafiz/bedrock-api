package zmq

import (
	"fmt"

	"github.com/zeromq/goczmq"
)

type ZMQServer struct {
}

func (z ZMQServer) Serve(host string, port int) error {
	// create a router socket and bind it to the specified host and port
	router, err := goczmq.NewRouter(fmt.Sprintf("tcp://%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("failed to start zemq server: %v", err)
	}
	defer router.Destroy()

	// start the socket receiver, handler, and sender goroutines
	in := make(chan [][]byte)
	out := make(chan [][]byte)

	go z.socketReceiver(router, in)
	go z.socketSender(router, out)

	// main loop to handle incoming messages and send responses
	z.socketHandler(in, out)

	return nil
}
