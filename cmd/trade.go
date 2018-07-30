package cmd

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

var tradeCmd = &cobra.Command{
	Use:   "trade",
	Short: "Trades with a specific strategy against the Stellar universal marketplace.",
}

func requiredFlag(flag string) {
	e := tradeCmd.MarkFlagRequired(flag)
	if e != nil {
		panic(e)
	}
}

func init() {
	var botConfigPath = tradeCmd.Flags().String("botConf", "./trader.cfg", "(required) trading bot's basic config file path")
	var strategy = tradeCmd.Flags().String("strategy", "buysell", "(required) type of strategy to run")
	var stratConfigPath = tradeCmd.Flags().String("stratConf", "./buysell.cfg", "strategy config file path")
	var fractionalReserveMagnifier = tradeCmd.Flags().Int8("fractionalReserveMultiplier", 1, "fractional multiplier for XLM reserves")
	var operationalBuffer = tradeCmd.Flags().Float64("operationalBuffer", 2000, "operational buffer for min number of lumens needed in XLM reserves")

	requiredFlag("botConf")
	requiredFlag("strategy")

	tradeCmd.Run = func(ccmd *cobra.Command, args []string) {
		log.SetLevel(log.DebugLevel)
		log.Info("Starting Kelp Trader: v0.6")

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
		strat := plugins.MakeStrategy(sdex, &assetBase, &assetQuote, *strategy, *stratConfigPath)
		bot := trader.MakeBot(
			client,
			botConfig.AssetBase(),
			botConfig.AssetQuote(),
			botConfig.TradingAccount(),
			sdex,
			strat,
			botConfig.TICK_INTERVAL_SECONDS,
			dataKey,
		)
		// --- end initialization of objects ---

		for {
			bot.Start()
			log.Info("Restarting the trader bot...")
		}
	}
}
