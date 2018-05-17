package main

import (
	"net/http"
	"os"

	"github.com/lightyeario/kelp/support"
	"github.com/spf13/cobra"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/config"
	"github.com/stellar/go/support/log"
)

var rootCmd = &cobra.Command{
	Use:   "terminator",
	Short: "Monitors a Stellar Account and terminates offers across all inactive bots",
}
var configPath = rootCmd.PersistentFlags().String("conf", "./terminator.cfg", "service's basic config file path")
var configFile Config

func main() {
	log.SetLevel(log.DebugLevel)
	rootCmd.Run = run
	e := rootCmd.Execute()
	if e != nil {
		log.Error(e)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	log.Info("Starting Terminator: v1.0")
	err := config.Read(*configPath, &configFile)
	kelp.CheckConfigError(configFile, err)
	err = configFile.Init()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	log.Info("Started Terminator for account: ", *configFile.tradingAccount)

	// start the initialization of objects
	client := &horizon.Client{
		URL:  configFile.HORIZON_URL,
		HTTP: http.DefaultClient,
	}
	txB := kelp.MakeTxButler(
		client,
		configFile.SOURCE_SECRET_SEED,
		configFile.TRADING_SECRET_SEED,
		*configFile.sourceAccount,
		*configFile.tradingAccount,
		kelp.ParseNetwork(configFile.HORIZON_URL),
		-1, // not needed here
		-1, // not needed here
	)
	terminator := MakeTerminator(client, txB, *configFile.tradingAccount, configFile.TICK_INTERVAL_SECONDS, configFile.ALLOW_INACTIVE_MINUTES)
	// --- end initialization of objects ----

	for {
		terminator.StartService()
		log.Info("Restarting terminator service")
	}
}
