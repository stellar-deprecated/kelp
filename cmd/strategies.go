package cmd

import (
	"fmt"
	"sort"

	"github.com/lightyeario/kelp/plugins"

	"github.com/spf13/cobra"
)

var strategiesCmd = &cobra.Command{
	Use:   "strategies",
	Short: "Lists the available strategies",
}

func init() {
	strategiesCmd.Run = func(ccmd *cobra.Command, args []string) {
		strategies := plugins.Strategies()
		for _, name := range getSortedKeys(strategies) {
			fmt.Printf("  %-15s %s\n", name, strategies[name])
		}
	}
}

func getSortedKeys(m map[string]string) []string {
	keys := []string{}
	for name := range m {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}
