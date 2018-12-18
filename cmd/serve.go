package cmd

import (
	"fmt"
	"github.com/interstellar/kelp/server"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Provides an API for managing kelp processes",
}

func init() {
	serveCmd.Run = func(ccmd *cobra.Command, args []string) {
		fmt.Printf("Starting server\n")
		fmt.Printf("  --------------------------------------------------------------------------------\n")

		server.Start()
	}
}
