package cmd

import (
	"fmt"

	"github.com/lightyeario/kelp/plugins"
	"github.com/lightyeario/kelp/support/utils"

	"github.com/spf13/cobra"
)

var exchanagesCmd = &cobra.Command{
	Use:   "exchanges",
	Short: "Lists the available exchange integrations",
}

func init() {
	exchanagesCmd.Run = func(ccmd *cobra.Command, args []string) {
		exchanges := plugins.Exchanges()
		for _, name := range utils.GetSortedKeys(exchanges) {
			fmt.Printf("  %-15s %s\n", name, exchanges[name])
		}
	}
}
