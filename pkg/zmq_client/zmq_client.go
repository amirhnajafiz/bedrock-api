package zclient

import (
	"fmt"

	"github.com/zeromq/goczmq"
)

// ZMQClient represents a ZeroMQ client that can send events to a specified address with a time-to-live
// for connections.
type ZMQClient struct {
	address string
}

// NewZMQClient creates a new ZMQClient with the given address and time-to-live (TTL) for connections.
func NewZMQClient(address string) *ZMQClient {
	return &ZMQClient{
		address: address,
	}
}

// Send an event to the server and waits for a response without a timeout.
func (z *ZMQClient) Send(event []byte) ([]byte, error) {
	return z.sendEvent(event, 0)
}

// SendWithTimeout sends an event to the server and waits for a response with a specified timeout.
func (z *ZMQClient) SendWithTimeout(event []byte, timeout int) ([]byte, error) {
	return z.sendEvent(event, timeout)
}

// send event to the server and waits for a response with a specified timeout.
func (z *ZMQClient) sendEvent(event []byte, timeout int) ([]byte, error) {
	// create a dealer
	dealer, err := goczmq.NewDealer(z.address)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZMQ dealer instance: %v", err)
	}
	defer dealer.Destroy()

	// set receive timeout
	if timeout > 0 {
		dealer.SetConnectTimeout(timeout)
		dealer.SetRcvtimeo(timeout)
	}

	// send the event
	if err := dealer.SendFrame(event, goczmq.FlagNone); err != nil {
		return nil, fmt.Errorf("failed to send event: %v", err)
	}

	// receive the response with a timeout
	response, err := dealer.RecvMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to receive response: %v", err)
	}

	return response[0], nil
}
