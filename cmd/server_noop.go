package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// load the basic serverCmd for all architectures. Default action is a noop action.
// This is overriden for platforms that support this command, example: server_amd64.go
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Serves the Kelp GUI",
	Run: func(ccmd *cobra.Command, args []string) {
		log.Printf("Kelp GUI Server unsupported in this version: %s [%s]\n", version, gitHash)
	},
}
