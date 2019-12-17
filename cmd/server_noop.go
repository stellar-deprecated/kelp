package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var hasUICapability = false

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Serves the Kelp GUI",
	Run: func(ccmd *cobra.Command, args []string) {
		log.Printf("Kelp GUI Server unsupported in this version: %s [%s]\n", version, gitHash)
	},
}
