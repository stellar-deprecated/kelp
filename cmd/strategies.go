package cmd

import (
	"fmt"

	"github.com/stellar/kelp/plugins"

	"github.com/spf13/cobra"
)

var strategiesCmd = &cobra.Command{
	Use:   "strategies",
	Short: "Lists the available strategies",
}

func init() {
	strategiesCmd.Run = func(ccmd *cobra.Command, args []string) {
		checkInitRootFlags()
		fmt.Printf("  Strategy\tComplexity\tNeeds Config\tDescription\n")
		fmt.Printf("  --------------------------------------------------------------------------------\n")
		strategies := plugins.Strategies()
		for _, name := range sortedStrategyKeys(strategies) {
			fmt.Printf("  %-14s%s\t%v\t\t%s\n", name, strategies[name].Complexity, strategies[name].NeedsConfig, strategies[name].Description)
		}
	}
}

func sortedStrategyKeys(m map[string]plugins.StrategyContainer) []string {
	keys := make([]string, len(m))
	for k, v := range m {
		if len(keys[v.SortOrder]) > 0 && keys[v.SortOrder] != k {
			panic(fmt.Errorf("invalid sort order specified for strategies, SortOrder that was repeated: %d", v.SortOrder))
		}
		keys[v.SortOrder] = k
	}
	return keys
}
