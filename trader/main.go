package main

import (
	"net/http"
	"os"

	"github.com/lightyeario/kelp/support/datamodel"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/lightyeario/kelp/trader/strategy"
	"github.com/spf13/cobra"
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
*/
var rootCmd = &cobra.Command{
	Use:   "trader",
	Short: "Market Making bot for Stellar",
}
var botConfigPath = rootCmd.PersistentFlags().String("botConf", "./trader.cfg", "trading bot's basic config file path")
var botConfig BotConfig
var stratType = rootCmd.PersistentFlags().String("stratType", "buysell", "type of strategy to run")
var stratConfigPath = rootCmd.PersistentFlags().String("stratConf", "./trader.cfg", "strategy config file path")
var fractionalReserveMagnifier = rootCmd.PersistentFlags().Int8("fractionalReserveMultiplier", 1, "(optional) fractional multiplier for XLM reserves")
var operationalBuffer = rootCmd.PersistentFlags().Float64("operationalBuffer", 2000, "(optional) operational buffer for min number of lumens needed in XLM reserves")

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
	log.Info("Starting trader: v0.5")
	err := config.Read(*botConfigPath, &botConfig)
	utils.CheckConfigError(botConfig, err)
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
	txB := utils.MakeTxButler(
		client,
		botConfig.SOURCE_SECRET_SEED,
		botConfig.TRADING_SECRET_SEED,
		botConfig.SourceAccount(),
		botConfig.TradingAccount(),
		utils.ParseNetwork(botConfig.HORIZON_URL),
		*fractionalReserveMagnifier,
		*operationalBuffer,
	)

	assetBase := botConfig.AssetBase()
	assetQuote := botConfig.AssetQuote()
	dataKey := datamodel.MakeSortedBotKey(assetBase, assetQuote)
	strat := strategy.StratFactory(txB, &assetBase, &assetQuote, *stratType, *stratConfigPath)
	bot := MakeBot(
		client,
		botConfig.AssetBase(),
		botConfig.AssetQuote(),
		botConfig.TradingAccount(),
		txB,
		strat,
		botConfig.TICK_INTERVAL_SECONDS,
		dataKey,
	)
	// --- end initialization of objects ----

	for {
		bot.Start()
		log.Info("Restarting strat")
	}
}
