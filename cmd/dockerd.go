package cmd

import (
	"github.com/amirhnajafiz/bedrock-api/internal/configs"

	"github.com/spf13/cobra"
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
			d.main()
		},
	}
}

func (d Dockerd) main() {}
