package cmd

import (
	"fmt"
	"sort"

	"github.com/interstellar/kelp/plugins"

	"github.com/spf13/cobra"
)

var exchanagesCmd = &cobra.Command{
	Use:   "exchanges",
	Short: "Lists the available exchange integrations",
}

func init() {
	exchanagesCmd.Run = func(ccmd *cobra.Command, args []string) {
		fmt.Printf("  Exchange\tDescription\n")
		fmt.Printf("  --------------------------------------------------------------------------------\n")
		exchanges := plugins.Exchanges()
		for _, name := range sortedExchangeKeys(exchanges) {
			fmt.Printf("  %-14s%s\n", name, exchanges[name])
		}
	}
}

func sortedExchangeKeys(m map[string]string) []string {
	keys := []string{}
	for name := range m {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}
