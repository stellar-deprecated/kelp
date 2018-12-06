package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves and API for kelp",
	Run: func(ccmd *cobra.Command, args []string) {
		fmt.Printf("Starting server\n")
		fmt.Printf("  --------------------------------------------------------------------------------\n")
	},
}
