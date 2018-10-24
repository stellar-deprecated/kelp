package cmd

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/plugins"
	"github.com/lightyeario/kelp/support/monitoring"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/lightyeario/kelp/trader"
	"github.com/spf13/cobra"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/config"
)

const tradeExamples = `  kelp trade --botConf ./path/trader.cfg --strategy buysell --stratConf ./path/buysell.cfg
  kelp trade --botConf ./path/trader.cfg --strategy buysell --stratConf ./path/buysell.cfg --sim`

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
	operationalBuffer := tradeCmd.Flags().Float64("operationalBuffer", 20, "buffer of native XLM to maintain beyond minimum account balance requirement")
	simMode := tradeCmd.Flags().Bool("sim", false, "simulate the bot's actions without placing any trades")

	requiredFlag("botConf")
	requiredFlag("strategy")
	hiddenFlag("operationalBuffer")
	tradeCmd.Flags().SortFlags = false

	tradeCmd.Run = func(ccmd *cobra.Command, args []string) {
		startupMessage := "Starting Kelp Trader: " + version + " [" + gitHash + "]"
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

		alert, e := monitoring.MakeAlert(botConfig.ALERT_TYPE, botConfig.ALERT_API_KEY)
		if e != nil {
			log.Printf("Unable to set up monitoring for alert type '%s' with the given API key\n", botConfig.ALERT_TYPE)
		}
		// --- start initialization of objects ----
		sdex := plugins.MakeSDEX(
			client,
			botConfig.SOURCE_SECRET_SEED,
			botConfig.TRADING_SECRET_SEED,
			botConfig.SourceAccount(),
			botConfig.TradingAccount(),
			utils.ParseNetwork(botConfig.HORIZON_URL),
			*operationalBuffer,
			*simMode,
		)

		assetBase := botConfig.AssetBase()
		assetQuote := botConfig.AssetQuote()
		dataKey := model.MakeSortedBotKey(assetBase, assetQuote)
		strat, e := plugins.MakeStrategy(sdex, &assetBase, &assetQuote, *strategy, *stratConfigPath)
		if e != nil {
			log.Println()
			log.Println(e)
			// we want to delete all the offers and exit here since there is something wrong with our setup
			deleteAllOffersAndExit(botConfig, client, sdex)
		}
		bot := trader.MakeBot(
			client,
			botConfig.AssetBase(),
			botConfig.AssetQuote(),
			botConfig.TradingAccount(),
			sdex,
			strat,
			botConfig.TICK_INTERVAL_SECONDS,
			dataKey,
			alert,
		)
		// --- end initialization of objects ---

		log.Printf("validating trustlines...\n")
		validateTrustlines(client, &botConfig)
		log.Printf("trustlines valid\n")

		log.Println("Starting the trader bot...")
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

func deleteAllOffersAndExit(botConfig trader.BotConfig, client *horizon.Client, sdex *plugins.SDEX) {
	log.Println()
	log.Printf("deleting all offers and then exiting...\n")

	offers, e := utils.LoadAllOffers(botConfig.TradingAccount(), client)
	if e != nil {
		log.Println()
		log.Fatal(e)
		return
	}
	sellingAOffers, buyingAOffers := utils.FilterOffers(offers, botConfig.AssetBase(), botConfig.AssetQuote())
	allOffers := append(sellingAOffers, buyingAOffers...)

	dOps := sdex.DeleteAllOffers(allOffers)
	log.Printf("created %d operations to delete offers\n", len(dOps))

	if len(dOps) > 0 {
		e := sdex.SubmitOps(dOps, func(hash string, e error) {
			if e != nil {
				log.Println()
				log.Fatal(e)
				return
			}
			log.Fatal("...deleted all offers, exiting")
		})
		if e != nil {
			log.Println()
			log.Fatal(e)
			return
		}

		for {
			sleepSeconds := 10
			log.Printf("sleeping for %d seconds until our deletion is confirmed and we exit...\n", sleepSeconds)
			time.Sleep(time.Duration(sleepSeconds) * time.Second)
		}
	} else {
		log.Fatal("...nothing to delete, exiting")
	}
}
