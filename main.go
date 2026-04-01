package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amirhnajafiz/bedrock-api/cmd"
	"github.com/amirhnajafiz/bedrock-api/internal/configs"

	"github.com/spf13/cobra"
)

// initVars initializes the environment variables for the application.
func initVars() map[string]string {
	vars := make(map[string]string)

	vars["CONFIG_PATH"] = os.Getenv("CONFIG_PATH")
	if vars["CONFIG_PATH"] == "" {
		vars["CONFIG_PATH"] = "config.yaml"
	}

	return vars
}

func main() {
	// catch os signals for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	defer stop()

	// create root cmd
	root := &cobra.Command{}

	// initialize environment variables
	envVars := initVars()

	// load configuration values
	cfg, err := configs.LoadConfig(envVars["CONFIG_PATH"])
	if err != nil {
		panic(err)
	}
	fmt.Printf("configs:\n%s\n\n", cfg.String())

	// add subcommands
	root.AddCommand(
		cmd.API{Ctx: ctx, Cfg: cfg.API}.Command(),
		cmd.Dockerd{Ctx: ctx, Cfg: cfg.Dockerd}.Command(),
		cmd.FileMD{Ctx: ctx, Cfg: cfg.FileMD}.Command(),
	)

	// execute root cmd
	if err := root.Execute(); err != nil {
		panic(err)
	}
}
