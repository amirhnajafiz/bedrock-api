package main

import (
	"github.com/amirhnajafiz/bedrock-api/cmd"
	"github.com/amirhnajafiz/bedrock-api/internal/configs"

	"github.com/spf13/cobra"
)

func main() {
	// create root cmd
	root := &cobra.Command{}

	// load configuration values
	cfg, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}

	// add subcommands
	root.AddCommand(
		cmd.API{Cfg: cfg.API}.Command(),
		cmd.Dockerd{Cfg: cfg.Dockerd}.Command(),
		cmd.FileMD{Cfg: cfg.FileMD}.Command(),
	)

	// execute root cmd
	if err := root.Execute(); err != nil {
		panic(err)
	}
}
