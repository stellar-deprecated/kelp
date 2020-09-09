package cmd

import (
	"fmt"

	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/sdk"

	"github.com/spf13/cobra"
)

var exchangesCmd = &cobra.Command{
	Use:   "exchanges",
	Short: "Lists the available exchange integrations",
}

func init() {
	exchangesCmd.Run = func(ccmd *cobra.Command, args []string) {
		checkInitRootFlags()
		// call sdk.GetExchangeList() here so we pre-load exchanges before displaying the table
		sdk.GetExchangeList()
		fmt.Printf("  Exchange\t\t\tTested\t\tTrading\t\tAtomic Post-Only\tTrade Has OrderID\t\tDescription\n")
		fmt.Printf("  -----------------------------------------------------------------------------------------------------------------------------\n")
		exchanges := plugins.Exchanges()
		for _, name := range sortedExchangeKeys(exchanges) {
			fmt.Printf("  %-24s\t%v\t\t%v\t\t%v\t\t\t%v\t\t%s\n", name, exchanges[name].Tested, exchanges[name].TradeEnabled, exchanges[name].AtomicPostOnly, exchanges[name].TradeHasOrderId, exchanges[name].Description)
		}
	}
}

func sortedExchangeKeys(m map[string]plugins.ExchangeContainer) []string {
	keys := make([]string, len(m))
	for k, v := range m {
		if len(keys[v.SortOrder]) > 0 && keys[v.SortOrder] != k {
			panic(fmt.Errorf("invalid sort order specified for strategies, SortOrder that was repeated: %d", v.SortOrder))
		}
		keys[v.SortOrder] = k
	}
	return keys
}
