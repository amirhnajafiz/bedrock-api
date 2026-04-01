package cmd

import (
	"context"
	"fmt"

	"github.com/amirhnajafiz/bedrock-api/internal/components/sessions"
	"github.com/amirhnajafiz/bedrock-api/internal/configs"
	"github.com/amirhnajafiz/bedrock-api/internal/logger"
	"github.com/amirhnajafiz/bedrock-api/internal/ports/http"
	"github.com/amirhnajafiz/bedrock-api/internal/ports/zmq"
	"github.com/amirhnajafiz/bedrock-api/internal/scheduler"
	"github.com/amirhnajafiz/bedrock-api/internal/storage"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// API represents the API command.
type API struct {
	Ctx context.Context
	Cfg *configs.APIConfig
}

// Command returns the cobra command for API.
func (a API) Command() *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "API Server",
		Long:  "API Server is a RESTful API server that provides endpoints for managing and interacting with the system.",
		Run: func(cmd *cobra.Command, args []string) {
			StartAPI(a.Ctx, a.Cfg)
		},
	}
}

func StartAPI(ctx context.Context, cfg *configs.APIConfig) {
	// create public shared modules
	logr := logger.New(cfg.LogLevel)
	rr := scheduler.NewRoundRobin()
	ss := sessions.NewSessionStore(storage.NewGoCache())

	// create an errgroup with the provided context
	erg, ctx := errgroup.WithContext(ctx)

	// build and start the ZMQ server
	zmqAddress := fmt.Sprintf("tcp://%s:%d", cfg.SocketHost, cfg.SocketPort)
	zmqServer := zmq.ZMQServer{
		Logr:         logr.Named("zmq"),
		Scheduler:    rr,
		SessionStore: ss,
	}.Build(
		zmqAddress,
		cfg.SocketHandlers,
		ctx,
	)
	erg.Go(func() error {
		return zmqServer.Serve()
	})

	// build and start the HTTP server
	httpServer := http.HTTPServer{
		Logr:         logr.Named("http"),
		Scheduler:    rr,
		SessionStore: ss,
	}.Build(
		fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort),
		zmqAddress,
	)
	erg.Go(func() error {
		return httpServer.Serve()
	})

	// wait for all servers to finish
	if err := erg.Wait(); err != nil {
		logr.Error("api failed", zap.Error(err))
	}
}
