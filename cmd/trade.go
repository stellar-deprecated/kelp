package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/plugins"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/lightyeario/kelp/trader"
	"github.com/spf13/cobra"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/config"
)

const tradeExamples = `  kelp trade --botConf trader.cfg --strategy buysell --stratConf buysell.cfg
  kelp trade --botConf trader.cfg --strategy buysell --stratConf buysell.cfg --sim`

var tradeCmd = &cobra.Command{
	Use:     "trade",
	Short:   "Trades against the Stellar universal marketplace using the specified strategy",
	Example: tradeExamples,
}

func requiredFlag(flag string) {
	e := tradeCmd.MarkFlagRequired(flag)
	if e != nil {
		panic(e)
	}
}

func hiddenFlag(flag string) {
	e := tradeCmd.Flags().MarkHidden(flag)
	if e != nil {
		panic(e)
	}
}

func init() {
	// short flags
	botConfigPath := tradeCmd.Flags().StringP("botConf", "c", "", "(required) trading bot's basic config file path")
	strategy := tradeCmd.Flags().StringP("strategy", "s", "", "(required) type of strategy to run")
	stratConfigPath := tradeCmd.Flags().StringP("stratConf", "f", "", "strategy config file path")
	// long-only flags
	fractionalReserveMultiplier := tradeCmd.Flags().Int8("fractionalReserveMultiplier", 1, "fractional multiplier for XLM reserves")
	operationalBuffer := tradeCmd.Flags().Float64("operationalBuffer", 20, "buffer of native XLM to maintain beyond minimum account balance requirement")
	simMode := tradeCmd.Flags().Bool("sim", false, "simulate the bot's actions without placing any trades")

	requiredFlag("botConf")
	requiredFlag("strategy")
	hiddenFlag("fractionalReserveMultiplier")
	hiddenFlag("operationalBuffer")
	tradeCmd.Flags().SortFlags = false

	tradeCmd.Run = func(ccmd *cobra.Command, args []string) {
		startupMessage := "Starting Kelp Trader: v0.6"
		if *simMode {
			startupMessage += " (simulation mode)"
		}
		log.Println(startupMessage)

		var botConfig trader.BotConfig
		e := config.Read(*botConfigPath, &botConfig)
		utils.CheckConfigError(botConfig, e, *botConfigPath)
		e = botConfig.Init()
		if e != nil {
			log.Println()
			log.Fatal(e)
		}
		log.Printf("Trading %s:%s for %s:%s\n", botConfig.ASSET_CODE_A, botConfig.ISSUER_A, botConfig.ASSET_CODE_B, botConfig.ISSUER_B)

		client := &horizon.Client{
			URL:  botConfig.HORIZON_URL,
			HTTP: http.DefaultClient,
		}
		validateTrustlines(client, &botConfig)

		// --- start initialization of objects ----
		sdex := plugins.MakeSDEX(
			client,
			botConfig.SOURCE_SECRET_SEED,
			botConfig.TRADING_SECRET_SEED,
			botConfig.SourceAccount(),
			botConfig.TradingAccount(),
			utils.ParseNetwork(botConfig.HORIZON_URL),
			*fractionalReserveMultiplier,
			*operationalBuffer,
			*simMode,
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
			log.Println("Restarting the trader bot...")
		}
	}
}

func validateTrustlines(client *horizon.Client, botConfig *trader.BotConfig) {
	account, e := client.LoadAccount(botConfig.TradingAccount())
	if e != nil {
		log.Println()
		log.Fatal(e)
	}

	missingTrustlines := []string{}
	if botConfig.ISSUER_A != "" {
		balance := utils.GetCreditBalance(account, botConfig.ASSET_CODE_A, botConfig.ISSUER_A)
		if balance == nil {
			missingTrustlines = append(missingTrustlines, fmt.Sprintf("%s:%s", botConfig.ASSET_CODE_A, botConfig.ISSUER_A))
		}
	}

	if botConfig.ISSUER_B != "" {
		balance := utils.GetCreditBalance(account, botConfig.ASSET_CODE_B, botConfig.ISSUER_B)
		if balance == nil {
			missingTrustlines = append(missingTrustlines, fmt.Sprintf("%s:%s", botConfig.ASSET_CODE_B, botConfig.ISSUER_B))
		}
	}

	if len(missingTrustlines) > 0 {
		log.Println()
		log.Fatalf("error: your trading account does not have the required trustlines: %v\n", missingTrustlines)
	}
}
