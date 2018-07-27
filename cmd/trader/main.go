package main

import (
	"net/http"
	"os"

	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/plugins"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/lightyeario/kelp/trader"
	"github.com/spf13/cobra"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/config"
	"github.com/stellar/go/support/log"
)

var rootCmd = &cobra.Command{
	Use:   "trader",
	Short: "Kelp is a free and open-source trading bot for the Stellar universal marketplace.",
	Long: `Kelp is a free and open-source trading bot for the Stellar universal marketplace.

Kelp includes several configurable trading strategies and its modular design allows you to customize
your algorithms, exchange integrations, and assets. You can define your own parameters or create your
own plugins to quickly implement a trading bot.`,
}
var botConfigPath = rootCmd.PersistentFlags().String("botConf", "./trader.cfg", "trading bot's basic config file path")
var stratType = rootCmd.PersistentFlags().String("stratType", "buysell", "type of strategy to run")
var stratConfigPath = rootCmd.PersistentFlags().String("stratConf", "./buysell.cfg", "strategy config file path")
var fractionalReserveMagnifier = rootCmd.PersistentFlags().Int8("fractionalReserveMultiplier", 1, "fractional multiplier for XLM reserves")
var operationalBuffer = rootCmd.PersistentFlags().Float64("operationalBuffer", 2000, "operational buffer for min number of lumens needed in XLM reserves")

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

	var botConfig trader.BotConfig
	err := config.Read(*botConfigPath, &botConfig)
	utils.CheckConfigError(botConfig, err)
	err = botConfig.Init()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	log.Info("Trading ", botConfig.ASSET_CODE_A, " for ", botConfig.ASSET_CODE_B)

	// --- start initialization of objects ----
	client := &horizon.Client{
		URL:  botConfig.HORIZON_URL,
		HTTP: http.DefaultClient,
	}
	sdex := plugins.MakeSDEX(
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
	dataKey := model.MakeSortedBotKey(assetBase, assetQuote)
	strategy := plugins.MakeStrategy(sdex, &assetBase, &assetQuote, *stratType, *stratConfigPath)
	bot := trader.MakeBot(
		client,
		botConfig.AssetBase(),
		botConfig.AssetQuote(),
		botConfig.TradingAccount(),
		sdex,
		strategy,
		botConfig.TICK_INTERVAL_SECONDS,
		dataKey,
	)
	// --- end initialization of objects ---

	for {
		bot.Start()
		log.Info("Restarting the trader bot...")
	}
}
