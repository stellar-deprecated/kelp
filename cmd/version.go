package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version and build information",
	Run: func(ccmd *cobra.Command, args []string) {
		fmt.Printf("  cli version: %s\n", version)
		fmt.Printf("  gui version: %s\n", guiVersion)
		fmt.Printf("  git branch: %s\n", gitBranch)
		fmt.Printf("  git hash: %s\n", gitHash)
		fmt.Printf("  build date: %s\n", buildDate)
		fmt.Printf("  build type: %s\n", buildType)
		fmt.Printf("  env: %s\n", env)
		fmt.Printf("  GOOS: %s\n", runtime.GOOS)
		fmt.Printf("  GOARCH: %s\n", runtime.GOARCH)
	},
}
