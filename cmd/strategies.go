package cmd

import (
	"fmt"

	"github.com/lightyeario/kelp/plugins"
	"github.com/lightyeario/kelp/support/utils"

	"github.com/spf13/cobra"
)

var strategiesCmd = &cobra.Command{
	Use:   "strategies",
	Short: "Lists the available strategies",
}

func init() {
	strategiesCmd.Run = func(ccmd *cobra.Command, args []string) {
		strategies := plugins.Strategies()
		for _, name := range utils.GetSortedKeys(strategies) {
			fmt.Printf("  %-15s %s\n", name, strategies[name])
		}
	}
}
