package cmd

import (
	"log"
	"net/http"

	"github.com/lightyeario/kelp/plugins"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/lightyeario/kelp/terminator"
	"github.com/spf13/cobra"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/config"
)

var terminateCmd = &cobra.Command{
	Hidden: true,
	Use:    "terminator",
	Short:  "Monitors a Stellar Account and terminates offers across all inactive bots",
}

func init() {
	var configPath = terminateCmd.Flags().String("conf", "./terminator.cfg", "service's basic config file path")

	terminateCmd.Run = func(ccmd *cobra.Command, args []string) {
		log.Println("Starting Terminator: v1.0")

		var configFile terminator.Config
		err := config.Read(*configPath, &configFile)
		utils.CheckConfigError(configFile, err)
		err = configFile.Init()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Started Terminator for account: ", *configFile.TradingAccount)

		// --- start initialization of objects ----
		client := &horizon.Client{
			URL:  configFile.HORIZON_URL,
			HTTP: http.DefaultClient,
		}
		sdex := plugins.MakeSDEX(
			client,
			configFile.SOURCE_SECRET_SEED,
			configFile.TRADING_SECRET_SEED,
			*configFile.SourceAccount,
			*configFile.TradingAccount,
			utils.ParseNetwork(configFile.HORIZON_URL),
			-1, // not needed here
			-1, // not needed here
		)
		terminator := terminator.MakeTerminator(client, sdex, *configFile.TradingAccount, configFile.TICK_INTERVAL_SECONDS, configFile.ALLOW_INACTIVE_MINUTES)
		// --- end initialization of objects ----

		for {
			terminator.StartService()
			log.Println("Restarting terminator service")
		}
	}
}
