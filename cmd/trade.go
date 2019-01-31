package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/interstellar/kelp/model"
	"github.com/interstellar/kelp/plugins"
	"github.com/interstellar/kelp/support/logger"
	"github.com/interstellar/kelp/support/monitoring"
	"github.com/interstellar/kelp/support/networking"
	"github.com/interstellar/kelp/support/utils"
	"github.com/interstellar/kelp/trader"
	"github.com/nikhilsaraf/go-tools/multithreading"
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

func logPanic(l logger.Logger) {
	if r := recover(); r != nil {
		st := debug.Stack()
		l.Errorf("PANIC!! recovered to log it in the file\npanic: %v\n\n%s\n", r, string(st))
	}
}

func init() {
	// short flags
	botConfigPath := tradeCmd.Flags().StringP("botConf", "c", "", "(required) trading bot's basic config file path")
	strategy := tradeCmd.Flags().StringP("strategy", "s", "", "(required) type of strategy to run")
	stratConfigPath := tradeCmd.Flags().StringP("stratConf", "f", "", "strategy config file path")
	// long-only flags
	operationalBuffer := tradeCmd.Flags().Float64("operationalBuffer", 20, "buffer of native XLM to maintain beyond minimum account balance requirement")
	operationalBufferNonNativePct := tradeCmd.Flags().Float64("operationalBufferNonNativePct", 0.001, "buffer of non-native assets to maintain as a percentage (0.001 = 0.1%)")
	simMode := tradeCmd.Flags().Bool("sim", false, "simulate the bot's actions without placing any trades")
	logPrefix := tradeCmd.Flags().StringP("log", "l", "", "log to a file (and stdout) with this prefix for the filename")
	fixedIterations := tradeCmd.Flags().Uint64("iter", 0, "only run the bot for the first N iterations (defaults value 0 runs unboundedly)")

	requiredFlag("botConf")
	requiredFlag("strategy")
	hiddenFlag("operationalBuffer")
	hiddenFlag("operationalBufferNonNativePct")
	tradeCmd.Flags().SortFlags = false

	validateCliParams := func(l logger.Logger) {
		if *operationalBuffer < 0 {
			panic(fmt.Sprintf("invalid operationalBuffer argument, must be non-negative: %f", *operationalBuffer))
		}

		if *operationalBufferNonNativePct < 0 || *operationalBufferNonNativePct > 1 {
			panic(fmt.Sprintf("invalid operationalBufferNonNativePct argument, must be between 0 and 1 inclusive: %f", *operationalBufferNonNativePct))
		}

		if *fixedIterations == 0 {
			fixedIterations = nil
			l.Info("will run unbounded iterations")
		} else {
			l.Infof("will run only %d update iterations\n", *fixedIterations)
		}
	}

	tradeCmd.Run = func(ccmd *cobra.Command, args []string) {
		l := logger.MakeBasicLogger()
		var botConfig trader.BotConfig
		e := config.Read(*botConfigPath, &botConfig)
		utils.CheckConfigError(botConfig, e, *botConfigPath)
		e = botConfig.Init()
		if e != nil {
			logger.Fatal(l, e)
		}

		if *logPrefix != "" {
			t := time.Now().Format("20060102T150405MST")
			fileName := fmt.Sprintf("%s_%s_%s_%s_%s_%s.log", *logPrefix, botConfig.AssetCodeA, botConfig.IssuerA, botConfig.AssetCodeB, botConfig.IssuerB, t)
			e = setLogFile(fileName)
			if e != nil {
				logger.Fatal(l, e)
				return
			}
			l.Infof("logging to file: %s\n", fileName)

			// we want to create a deferred recovery function here that will log panics to the log file and then exit
			defer logPanic(l)
		}

		startupMessage := "Starting Kelp Trader: " + version + " [" + gitHash + "]"
		if *simMode {
			startupMessage += " (simulation mode)"
		}
		l.Info(startupMessage)

		// now that we've got the basic messages logged, validate the cli params
		validateCliParams(l)

		// only log botConfig file here so it can be included in the log file
		utils.LogConfig(botConfig)
		l.Infof("Trading %s:%s for %s:%s\n", botConfig.AssetCodeA, botConfig.IssuerA, botConfig.AssetCodeB, botConfig.IssuerB)

		client := &horizon.Client{
			URL:  botConfig.HorizonURL,
			HTTP: http.DefaultClient,
		}

		alert, e := monitoring.MakeAlert(botConfig.AlertType, botConfig.AlertAPIKey)
		if e != nil {
			l.Infof("Unable to set up monitoring for alert type '%s' with the given API key\n", botConfig.AlertType)
		}
		// --- start initialization of objects ----
		threadTracker := multithreading.MakeThreadTracker()

		assetBase := botConfig.AssetBase()
		assetQuote := botConfig.AssetQuote()
		tradingPair := &model.TradingPair{
			Base:  model.Asset(utils.Asset2CodeString(assetBase)),
			Quote: model.Asset(utils.Asset2CodeString(assetQuote)),
		}

		sdexAssetMap := map[model.Asset]horizon.Asset{
			tradingPair.Base:  assetBase,
			tradingPair.Quote: assetQuote,
		}
		sdex := plugins.MakeSDEX(
			client,
			botConfig.SourceSecretSeed,
			botConfig.TradingSecretSeed,
			botConfig.SourceAccount(),
			botConfig.TradingAccount(),
			utils.ParseNetwork(botConfig.HorizonURL),
			threadTracker,
			*operationalBuffer,
			*operationalBufferNonNativePct,
			*simMode,
			tradingPair,
			sdexAssetMap,
		)

		// setting the temp hack variables for the sdex price feeds
		e = plugins.SetPrivateSdexHack(client, utils.ParseNetwork(botConfig.HorizonURL))
		if e != nil {
			l.Info("")
			l.Errorf("%s", e)
			// we want to delete all the offers and exit here since there is something wrong with our setup
			deleteAllOffersAndExit(l, botConfig, client, sdex)
		}

		dataKey := model.MakeSortedBotKey(assetBase, assetQuote)
		strat, e := plugins.MakeStrategy(sdex, tradingPair, &assetBase, &assetQuote, *strategy, *stratConfigPath, *simMode)
		if e != nil {
			l.Info("")
			l.Errorf("%s", e)
			// we want to delete all the offers and exit here since there is something wrong with our setup
			deleteAllOffersAndExit(l, botConfig, client, sdex)
		}

		timeController := plugins.MakeIntervalTimeController(
			time.Duration(botConfig.TickIntervalSeconds)*time.Second,
			botConfig.MaxTickDelayMillis,
		)
		bot := trader.MakeBot(
			client,
			botConfig.AssetBase(),
			botConfig.AssetQuote(),
			botConfig.TradingAccount(),
			sdex,
			strat,
			timeController,
			botConfig.DeleteCyclesThreshold,
			threadTracker,
			fixedIterations,
			dataKey,
			alert,
		)
		// --- end initialization of objects ---

		l.Info("validating trustlines...")
		validateTrustlines(l, client, &botConfig)
		l.Info("trustlines valid")

		// --- start initialization of services ---
		if botConfig.MonitoringPort != 0 {
			go func() {
				e := startMonitoringServer(l, botConfig)
				if e != nil {
					l.Info("")
					l.Info("unable to start the monitoring server or problem encountered while running server:")
					l.Errorf("%s", e)
					// we want to delete all the offers and exit here because we don't want the bot to run if monitoring isn't working
					// if monitoring is desired but not working properly, we want the bot to be shut down and guarantee that there
					// aren't outstanding offers.
					deleteAllOffersAndExit(l, botConfig, client, sdex)
				}
			}()
		}

		strategyFillHandlers, e := strat.GetFillHandlers()
		if e != nil {
			l.Info("")
			l.Info("problem encountered while instantiating the fill tracker:")
			l.Errorf("%s", e)
			deleteAllOffersAndExit(l, botConfig, client, sdex)
		}
		if botConfig.FillTrackerSleepMillis != 0 {
			fillTracker := plugins.MakeFillTracker(tradingPair, threadTracker, sdex, botConfig.FillTrackerSleepMillis)
			fillLogger := plugins.MakeFillLogger()
			fillTracker.RegisterHandler(fillLogger)
			if strategyFillHandlers != nil {
				for _, h := range strategyFillHandlers {
					fillTracker.RegisterHandler(h)
				}
			}

			l.Infof("Starting fill tracker with %d handlers\n", fillTracker.NumHandlers())
			go func() {
				e := fillTracker.TrackFills()
				if e != nil {
					l.Info("")
					l.Info("problem encountered while running the fill tracker:")
					l.Errorf("%s", e)
					// we want to delete all the offers and exit here because we don't want the bot to run if fill tracking isn't working
					deleteAllOffersAndExit(l, botConfig, client, sdex)
				}
			}()
		} else if strategyFillHandlers != nil && len(strategyFillHandlers) > 0 {
			l.Info("")
			l.Error("error: strategy has FillHandlers but fill tracking was disabled (set FILL_TRACKER_SLEEP_MILLIS to a non-zero value)")
			// we want to delete all the offers and exit here because we don't want the bot to run if fill tracking isn't working
			deleteAllOffersAndExit(l, botConfig, client, sdex)
		}
		// --- end initialization of services ---

		l.Info("Starting the trader bot...")
		bot.Start()
	}
}

func startMonitoringServer(l logger.Logger, botConfig trader.BotConfig) error {
	serverConfig := &networking.Config{
		GoogleClientID:     botConfig.GoogleClientID,
		GoogleClientSecret: botConfig.GoogleClientSecret,
		PermittedEmails:    map[string]bool{},
	}
	// Load acceptable Google emails into the map
	for _, email := range strings.Split(botConfig.AcceptableEmails, ",") {
		serverConfig.PermittedEmails[email] = true
	}

	healthMetrics, e := monitoring.MakeMetricsRecorder(map[string]interface{}{"success": true})
	if e != nil {
		return fmt.Errorf("unable to make metrics recorder for the health endpoint: %s", e)
	}

	healthEndpoint, e := monitoring.MakeMetricsEndpoint("/health", healthMetrics, networking.NoAuth)
	if e != nil {
		return fmt.Errorf("unable to make /health endpoint: %s", e)
	}
	kelpMetrics, e := monitoring.MakeMetricsRecorder(nil)
	if e != nil {
		return fmt.Errorf("unable to make metrics recorder for the /metrics endpoint: %s", e)
	}

	metricsEndpoint, e := monitoring.MakeMetricsEndpoint("/metrics", kelpMetrics, networking.GoogleAuth)
	if e != nil {
		return fmt.Errorf("unable to make /metrics endpoint: %s", e)
	}
	server, e := networking.MakeServer(serverConfig, []networking.Endpoint{healthEndpoint, metricsEndpoint})
	if e != nil {
		return fmt.Errorf("unable to initialize the metrics server: %s", e)
	}

	l.Infof("Starting monitoring server on port %d\n", botConfig.MonitoringPort)
	return server.StartServer(botConfig.MonitoringPort, botConfig.MonitoringTLSCert, botConfig.MonitoringTLSKey)
}

func validateTrustlines(l logger.Logger, client *horizon.Client, botConfig *trader.BotConfig) {
	account, e := client.LoadAccount(botConfig.TradingAccount())
	if e != nil {
		logger.Fatal(l, e)
	}

	missingTrustlines := []string{}
	if botConfig.IssuerA != "" {
		balance := utils.GetCreditBalance(account, botConfig.AssetCodeA, botConfig.IssuerA)
		if balance == nil {
			missingTrustlines = append(missingTrustlines, fmt.Sprintf("%s:%s", botConfig.AssetCodeA, botConfig.IssuerA))
		}
	}

	if botConfig.IssuerB != "" {
		balance := utils.GetCreditBalance(account, botConfig.AssetCodeB, botConfig.IssuerB)
		if balance == nil {
			missingTrustlines = append(missingTrustlines, fmt.Sprintf("%s:%s", botConfig.AssetCodeB, botConfig.IssuerB))
		}
	}

	if len(missingTrustlines) > 0 {
		logger.Fatal(l, fmt.Errorf("error: your trading account does not have the required trustlines: %v", missingTrustlines))
	}
}

func deleteAllOffersAndExit(l logger.Logger, botConfig trader.BotConfig, client *horizon.Client, sdex *plugins.SDEX) {
	l.Info("")
	l.Info("deleting all offers and then exiting...")

	offers, e := utils.LoadAllOffers(botConfig.TradingAccount(), client)
	if e != nil {
		logger.Fatal(l, e)
		return
	}
	sellingAOffers, buyingAOffers := utils.FilterOffers(offers, botConfig.AssetBase(), botConfig.AssetQuote())
	allOffers := append(sellingAOffers, buyingAOffers...)

	dOps := sdex.DeleteAllOffers(allOffers)
	l.Infof("created %d operations to delete offers\n", len(dOps))

	if len(dOps) > 0 {
		e := sdex.SubmitOps(dOps, func(hash string, e error) {
			if e != nil {
				logger.Fatal(l, e)
				return
			}
			logger.Fatal(l, fmt.Errorf("...deleted all offers, exiting"))
		})
		if e != nil {
			logger.Fatal(l, e)
			return
		}

		for {
			sleepSeconds := 10
			l.Infof("sleeping for %d seconds until our deletion is confirmed and we exit...\n", sleepSeconds)
			time.Sleep(time.Duration(sleepSeconds) * time.Second)
		}
	} else {
		logger.Fatal(l, fmt.Errorf("...nothing to delete, exiting"))
	}
}

func setLogFile(fileName string) error {
	f, e := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if e != nil {
		return fmt.Errorf("failed to set log file: %s", e)
	}
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
	return nil
}
