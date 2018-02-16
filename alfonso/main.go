package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/lightyeario/kelp/alfonso/strategy"
	"github.com/lightyeario/kelp/support"
	"github.com/spf13/cobra"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/config"
	"github.com/stellar/go/support/log"
)

/*
Trades one pair.
Has one data feed
Has one account it is trading on behalf of
Has a depth curve it maintains around the price
treasury management


Types of data feed:
- fixed rate
- coinmarketcap
- fiat


TODO:
- should size the orders in proportion to the amount of imbalance on either side
- When it can't access feed to must cancle orders and wait




*/

var rootCmd = &cobra.Command{
	Use:   "alfonso",
	Short: "Simple Market Making bot for Stellar",
}
var botConfigPath = rootCmd.PersistentFlags().String("botConf", "./alfonso.cfg", "bot's basic config file path")
var botConfig strategy.BotConfig
var stratType = rootCmd.PersistentFlags().String("stratType", "simple", "type of strategy to run")
var stratConfigPath = rootCmd.PersistentFlags().String("stratConf", "./alfonso.cfg", "strategy config file path")

func main() {
	log.SetLevel(log.DebugLevel)
	rootCmd.Run = run
	rootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) {
	log.Info("Starting alfonso: v0.2")
	err := config.Read(*botConfigPath, &botConfig)
	strategy.CheckConfigError(botConfig, err)
	err = botConfig.Init()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	log.Info("Trading ", botConfig.ASSET_CODE_A, " for ", botConfig.ASSET_CODE_B)

	// start the initialization of objects
	client := &horizon.Client{
		URL:  botConfig.HORIZON_URL,
		HTTP: http.DefaultClient,
	}
	txB := makeTxButler(
		client,
		botConfig.SOURCE_SECRET_SEED,
		botConfig.TRADING_SECRET_SEED,
		botConfig.SourceAccount(),
		botConfig.TradingAccount(),
		parseNetwork(botConfig.HORIZON_URL),
	)
	// TODO move txbutler factory to file and move the init code into that function, potentially factory takes config directly
	txB.Init()

	assetA := botConfig.AssetA()
	assetB := botConfig.AssetB()
	strat := strategy.StratFactory(txB, &assetA, &assetB, *stratType, *stratConfigPath)
	bot := MakeBot(
		*client,
		botConfig.AssetA(),
		botConfig.AssetB(),
		botConfig.TradingAccount(),
		txB,
		strat,
		botConfig.TICK_INTERVAL_SECONDS,
	)
	// --- end initialization of objects ----

	for true {
		bot.Start()
		log.Info("Restarting strat")
	}
}

func parseNetwork(horizonURL string) build.Network {
	if strings.Contains(horizonURL, "test") {
		return build.TestNetwork
	}
	return build.PublicNetwork
}

func makeTxButler(
	client *horizon.Client,
	sourceSeed string,
	tradingSeed string,
	sourceAccount string,
	tradingAccount string,
	network build.Network,
) *kelp.TxButler {
	return &kelp.TxButler{
		// TODO TxButler needs to take in the reference
		API:            *client,
		SourceSeed:     sourceSeed,
		TradingSeed:    tradingSeed,
		SourceAccount:  sourceAccount,
		TradingAccount: tradingAccount,
		Network:        network,
	}
}
