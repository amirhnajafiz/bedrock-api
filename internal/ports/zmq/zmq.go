package zmq

import (
	"context"
	"fmt"

	"github.com/amirhnajafiz/bedrock-api/internal/components/sessions"
	statemachine "github.com/amirhnajafiz/bedrock-api/internal/state_machine"
	"github.com/amirhnajafiz/bedrock-api/internal/storage"

	"github.com/zeromq/goczmq"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// ZMQServer represents the ZeroMQ server that handles incoming messages from clients, interacts with the session store and scheduler,
// and sends responses back to clients.
type ZMQServer struct {
	// public shared modules
	DockerDHealthChannel chan string
	Logr                 *zap.Logger

	// private modules
	address       string
	eventHandlers int
	ctx           context.Context
	sessionStore  sessions.SessionStore
	stateMachine  *statemachine.StateMachine
}

// Build initializes the ZMQServer with the specified address and returns the server instance.
func (z ZMQServer) Build(address string, eventHandlers int, ctx context.Context) *ZMQServer {
	z.address = address
	z.eventHandlers = eventHandlers
	z.ctx = ctx

	z.sessionStore = sessions.NewSessionStore(storage.NewGoCache())
	z.stateMachine = statemachine.NewStateMachine()

	return &z
}

func (z ZMQServer) Serve() error {
	// create a router socket and bind it to the specified host and port
	router, err := goczmq.NewRouter(z.address)
	if err != nil {
		return fmt.Errorf("failed to start zmq server: %v", err)
	}
	defer router.Destroy()

	z.Logr.Info("server started", zap.String("address", z.address))

	// create an errgroup with the provided context
	erg, ctx := errgroup.WithContext(z.ctx)

	// start the socket receiver, handler, and sender goroutines
	in := make(chan [][]byte)
	out := make(chan [][]byte)

	erg.Go(func() error { return z.socketReceiver(ctx, router, in) })
	erg.Go(func() error { return z.socketSender(ctx, router, out) })

	// main loop to handle incoming messages and send responses
	for i := 0; i < z.eventHandlers; i++ {
		erg.Go(func() error { return z.socketHandler(ctx, in, out) })
	}

	return erg.Wait()
}
