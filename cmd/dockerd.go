package cmd

import (
	"fmt"

	"github.com/amirhnajafiz/bedrock-api/internal/configs"
	"github.com/amirhnajafiz/bedrock-api/internal/logger"
	"github.com/amirhnajafiz/bedrock-api/pkg/models"
	"github.com/amirhnajafiz/bedrock-api/pkg/zclient"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// Dockerd represents the Docker Daemon command.
type Dockerd struct {
	Cfg *configs.DockerdConfig
}

// Command returns the cobra command for Dockerd.
func (d Dockerd) Command() *cobra.Command {
	return &cobra.Command{
		Use:   "dockerd",
		Short: "Docker Daemon",
		Long:  "Docker Daemon is a containerization platform that allows you to build, ship, and run containers.",
		Run: func(cmd *cobra.Command, args []string) {
			StartDockerd(d.Cfg)
		},
	}
}

func StartDockerd(cfg *configs.DockerdConfig) {
	// create a new logger instance
	logr := logger.New(cfg.LogLevel)

	// TODO: generate a unique name
	name := "dd_instance"
	address := fmt.Sprintf("tcp://%s:%d", cfg.APISocketHost, cfg.APISocketPort)

	// register this docker daemon with API
	_, err := zclient.SendEvent(address, models.NewPacket().WithRegisterDaemon(name).ToBytes(), 20)
	if err != nil {
		logr.Warn("register daemon failed", zap.Error(err))
	}
}
