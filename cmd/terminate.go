package cmd

import (
	"log"
	"net/http"

	"github.com/nikhilsaraf/go-tools/multithreading"
	"github.com/spf13/cobra"
	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/support/config"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/utils"
	"github.com/stellar/kelp/terminator"
)

var terminateCmd = &cobra.Command{
	Hidden: true,
	Use:    "terminate",
	Short:  "Monitors a Stellar Account and terminates offers across all inactive bots",
}

func init() {
	configPath := terminateCmd.Flags().StringP("conf", "c", "./terminator.cfg", "service's basic config file path")

	terminateCmd.Run = func(ccmd *cobra.Command, args []string) {
		log.Println("Starting Terminator: " + version + " [" + gitHash + "]")

		var configFile terminator.Config
		err := config.Read(*configPath, &configFile)
		utils.CheckConfigError(configFile, err, *configPath)
		err = configFile.Init()
		if err != nil {
			log.Fatal(err)
		}
		utils.LogConfig(configFile)
		log.Println("Started Terminator for account: ", *configFile.TradingAccount)

		// --- start initialization of objects ----
		client := &horizonclient.Client{
			HorizonURL: configFile.HorizonURL,
			HTTP:       http.DefaultClient,
			AppName:    "kelp",
			AppVersion: version,
		}
		sdex := plugins.MakeSDEX(
			client,
			plugins.MakeIEIF(true), // used true for now since it's only ever been tested on SDEX and uses SDEX's data for now
			nil,
			configFile.SourceSecretSeed,
			configFile.TradingSecretSeed,
			*configFile.SourceAccount,
			*configFile.TradingAccount,
			utils.ParseNetwork(configFile.HorizonURL),
			multithreading.MakeThreadTracker(),
			-1, // not needed here
			-1, // not needed here
			false,
			nil, // not needed here
			map[model.Asset]hProtocol.Asset{},
			plugins.SdexFixedFeeFn(0),
		)
		terminator := terminator.MakeTerminator(client, sdex, *configFile.TradingAccount, configFile.TickIntervalSeconds, configFile.AllowInactiveMinutes)
		// --- end initialization of objects ----

		for {
			terminator.StartService()
			log.Println("Restarting terminator service")
		}
	}
}
