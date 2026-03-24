package zclient

import (
	"fmt"

	"github.com/zeromq/goczmq"
)

// SendEvent sends an event to the specified address and waits for a response with a timeout.
func SendEvent(address string, event []byte, timeout int) ([]byte, error) {
	// create a dealer
	dealer, err := goczmq.NewDealer(address)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZMQ dealer instance: %v", err)
	}
	defer dealer.Destroy()

	// set receive timeout
	dealer.SetConnectTimeout(timeout)
	dealer.SetRcvtimeo(timeout)

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
