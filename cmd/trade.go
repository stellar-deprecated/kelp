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

	"github.com/nikhilsaraf/go-tools/multithreading"
	"github.com/spf13/cobra"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	horizonclient "github.com/stellar/go/exp/clients/horizon"
	"github.com/stellar/go/support/config"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/logger"
	"github.com/stellar/kelp/support/monitoring"
	"github.com/stellar/kelp/support/networking"
	"github.com/stellar/kelp/support/prefs"
	"github.com/stellar/kelp/support/sdk"
	"github.com/stellar/kelp/support/utils"
	"github.com/stellar/kelp/trader"
)

const tradeExamples = `  kelp trade --botConf ./path/trader.cfg --strategy buysell --stratConf ./path/buysell.cfg
  kelp trade --botConf ./path/trader.cfg --strategy buysell --stratConf ./path/buysell.cfg --sim`

const prefsFilename = "kelp.prefs"

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

type inputs struct {
	botConfigPath                 *string
	strategy                      *string
	stratConfigPath               *string
	operationalBuffer             *float64
	operationalBufferNonNativePct *float64
	simMode                       *bool
	logPrefix                     *string
	fixedIterations               *uint64
	noHeaders                     *bool
}

func validateCliParams(l logger.Logger, options inputs) {
	if *options.operationalBuffer < 0 {
		panic(fmt.Sprintf("invalid operationalBuffer argument, must be non-negative: %f", *options.operationalBuffer))
	}

	if *options.operationalBufferNonNativePct < 0 || *options.operationalBufferNonNativePct > 1 {
		panic(fmt.Sprintf("invalid operationalBufferNonNativePct argument, must be between 0 and 1 inclusive: %f", *options.operationalBufferNonNativePct))
	}

	if *options.fixedIterations == 0 {
		options.fixedIterations = nil
		l.Info("will run unbounded iterations")
	} else {
		l.Infof("will run only %d update iterations\n", *options.fixedIterations)
	}
}

func validateBotConfig(l logger.Logger, botConfig trader.BotConfig) {
	if botConfig.IsTradingSdex() && botConfig.Fee == nil {
		logger.Fatal(l, fmt.Errorf("The `FEE` object needs to exist in the trader config file when trading on SDEX"))
	}

	if !botConfig.IsTradingSdex() && botConfig.CentralizedMinBaseVolumeOverride != nil && *botConfig.CentralizedMinBaseVolumeOverride <= 0.0 {
		logger.Fatal(l, fmt.Errorf("need to specify positive CENTRALIZED_MIN_BASE_VOLUME_OVERRIDE config param in trader config file when not trading on SDEX"))
	}
	if !botConfig.IsTradingSdex() && botConfig.CentralizedMinQuoteVolumeOverride != nil && *botConfig.CentralizedMinQuoteVolumeOverride <= 0.0 {
		logger.Fatal(l, fmt.Errorf("need to specify positive CENTRALIZED_MIN_QUOTE_VOLUME_OVERRIDE config param in trader config file when not trading on SDEX"))
	}
	validatePrecisionConfig(l, botConfig.IsTradingSdex(), botConfig.CentralizedVolumePrecisionOverride, "CENTRALIZED_VOLUME_PRECISION_OVERRIDE")
	validatePrecisionConfig(l, botConfig.IsTradingSdex(), botConfig.CentralizedPricePrecisionOverride, "CENTRALIZED_PRICE_PRECISION_OVERRIDE")
}

func validatePrecisionConfig(l logger.Logger, isTradingSdex bool, precisionField *int8, name string) {
	if !isTradingSdex && precisionField != nil && *precisionField < 0 {
		logger.Fatal(l, fmt.Errorf("need to specify non-negative %s config param in trader config file when not trading on SDEX", name))
	}
}

func init() {
	options := inputs{}
	// short flags
	options.botConfigPath = tradeCmd.Flags().StringP("botConf", "c", "", "(required) trading bot's basic config file path")
	options.strategy = tradeCmd.Flags().StringP("strategy", "s", "", "(required) type of strategy to run")
	options.stratConfigPath = tradeCmd.Flags().StringP("stratConf", "f", "", "strategy config file path")
	// long-only flags
	options.operationalBuffer = tradeCmd.Flags().Float64("operationalBuffer", 20, "buffer of native XLM to maintain beyond minimum account balance requirement")
	options.operationalBufferNonNativePct = tradeCmd.Flags().Float64("operationalBufferNonNativePct", 0.001, "buffer of non-native assets to maintain as a percentage (0.001 = 0.1%)")
	options.simMode = tradeCmd.Flags().Bool("sim", false, "simulate the bot's actions without placing any trades")
	options.logPrefix = tradeCmd.Flags().StringP("log", "l", "", "log to a file (and stdout) with this prefix for the filename")
	options.fixedIterations = tradeCmd.Flags().Uint64("iter", 0, "only run the bot for the first N iterations (defaults value 0 runs unboundedly)")
	options.noHeaders = tradeCmd.Flags().Bool("no-headers", false, "do not set X-App-Name and X-App-Version headers on requests to horizon")

	requiredFlag("botConf")
	requiredFlag("strategy")
	hiddenFlag("operationalBuffer")
	hiddenFlag("operationalBufferNonNativePct")
	tradeCmd.Flags().SortFlags = false

	tradeCmd.Run = func(ccmd *cobra.Command, args []string) {
		runTradeCmd(options)
	}
}

func makeStartupMessage(options inputs) string {
	startupMessage := "Starting Kelp Trader: " + version + " [" + gitHash + "]"
	if *options.simMode {
		startupMessage += " (simulation mode)"
	}
	return startupMessage
}

func makeFeeFn(l logger.Logger, botConfig trader.BotConfig, newClient *horizonclient.Client) plugins.OpFeeStroops {
	if !botConfig.IsTradingSdex() {
		return plugins.SdexFixedFeeFn(0)
	}

	feeFn, e := plugins.SdexFeeFnFromStats(
		botConfig.Fee.CapacityTrigger,
		botConfig.Fee.Percentile,
		botConfig.Fee.MaxOpFeeStroops,
		newClient,
	)
	if e != nil {
		logger.Fatal(l, fmt.Errorf("could not set up feeFn correctly: %s", e))
	}
	return feeFn
}

func readBotConfig(l logger.Logger, options inputs) trader.BotConfig {
	var botConfig trader.BotConfig
	e := config.Read(*options.botConfigPath, &botConfig)
	utils.CheckConfigError(botConfig, e, *options.botConfigPath)
	e = botConfig.Init()
	if e != nil {
		logger.Fatal(l, e)
	}

	if *options.logPrefix != "" {
		setLogFile(l, options, botConfig)
	}

	l.Info(makeStartupMessage(options))
	// now that we've got the basic messages logged, validate the cli params
	validateCliParams(l, options)

	// only log botConfig file here so it can be included in the log file
	utils.LogConfig(botConfig)
	validateBotConfig(l, botConfig)

	return botConfig
}

func makeExchangeShimSdex(
	l logger.Logger,
	botConfig trader.BotConfig,
	options inputs,
	client *horizon.Client,
	newClient *horizonclient.Client,
	ieif *plugins.IEIF,
	network build.Network,
	threadTracker *multithreading.ThreadTracker,
	tradingPair *model.TradingPair,
) (api.ExchangeShim, *plugins.SDEX) {
	var e error
	var exchangeShim api.ExchangeShim
	if !botConfig.IsTradingSdex() {
		exchangeAPIKeys := []api.ExchangeAPIKey{}
		for _, apiKey := range botConfig.ExchangeAPIKeys {
			exchangeAPIKeys = append(exchangeAPIKeys, api.ExchangeAPIKey{
				Key:    apiKey.Key,
				Secret: apiKey.Secret,
			})
		}

		exchangeParams := []api.ExchangeParam{}
		for _, param := range botConfig.ExchangeParams {
			exchangeParams = append(exchangeParams, api.ExchangeParam{
				Param: param.Param,
				Value: param.Value,
			})
		}

		exchangeHeaders := []api.ExchangeHeader{}
		for _, header := range botConfig.ExchangeHeaders {
			exchangeHeaders = append(exchangeHeaders, api.ExchangeHeader{
				Header: header.Header,
				Value:  header.Value,
			})
		}

		var exchangeAPI api.Exchange
		exchangeAPI, e = plugins.MakeTradingExchange(botConfig.TradingExchange, exchangeAPIKeys, exchangeParams, exchangeHeaders, *options.simMode)
		if e != nil {
			logger.Fatal(l, fmt.Errorf("unable to make trading exchange: %s", e))
			return nil, nil
		}

		exchangeShim = plugins.MakeBatchedExchange(exchangeAPI, *options.simMode, botConfig.AssetBase(), botConfig.AssetQuote(), botConfig.TradingAccount())

		// update precision overrides
		exchangeShim.OverrideOrderConstraints(tradingPair, model.MakeOrderConstraintsOverride(
			botConfig.CentralizedPricePrecisionOverride,
			botConfig.CentralizedVolumePrecisionOverride,
			nil,
			nil,
		))
		if botConfig.CentralizedMinBaseVolumeOverride != nil {
			// use updated precision overrides to convert the minCentralizedBaseVolume to a model.Number
			exchangeShim.OverrideOrderConstraints(tradingPair, model.MakeOrderConstraintsOverride(
				nil,
				nil,
				model.NumberFromFloat(*botConfig.CentralizedMinBaseVolumeOverride, exchangeShim.GetOrderConstraints(tradingPair).VolumePrecision),
				nil,
			))
		}
		if botConfig.CentralizedMinQuoteVolumeOverride != nil {
			// use updated precision overrides to convert the minCentralizedQuoteVolume to a model.Number
			minQuoteVolume := model.NumberFromFloat(*botConfig.CentralizedMinQuoteVolumeOverride, exchangeShim.GetOrderConstraints(tradingPair).VolumePrecision)
			exchangeShim.OverrideOrderConstraints(tradingPair, model.MakeOrderConstraintsOverride(
				nil,
				nil,
				nil,
				&minQuoteVolume,
			))
		}
	}

	sdexAssetMap := map[model.Asset]horizon.Asset{
		tradingPair.Base:  botConfig.AssetBase(),
		tradingPair.Quote: botConfig.AssetQuote(),
	}
	feeFn := makeFeeFn(l, botConfig, newClient)
	sdex := plugins.MakeSDEX(
		client,
		ieif,
		exchangeShim,
		botConfig.SourceSecretSeed,
		botConfig.TradingSecretSeed,
		botConfig.SourceAccount(),
		botConfig.TradingAccount(),
		network,
		threadTracker,
		*options.operationalBuffer,
		*options.operationalBufferNonNativePct,
		*options.simMode,
		tradingPair,
		sdexAssetMap,
		feeFn,
	)

	if botConfig.IsTradingSdex() {
		exchangeShim = sdex
	}
	return exchangeShim, sdex
}

func makeStrategy(
	l logger.Logger,
	network build.Network,
	botConfig trader.BotConfig,
	client *horizon.Client,
	sdex *plugins.SDEX,
	exchangeShim api.ExchangeShim,
	assetBase horizon.Asset,
	assetQuote horizon.Asset,
	ieif *plugins.IEIF,
	tradingPair *model.TradingPair,
	options inputs,
	threadTracker *multithreading.ThreadTracker,
) api.Strategy {
	// setting the temp hack variables for the sdex price feeds
	e := plugins.SetPrivateSdexHack(client, plugins.MakeIEIF(true), network)
	if e != nil {
		l.Info("")
		l.Errorf("%s", e)
		// we want to delete all the offers and exit here since there is something wrong with our setup
		deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker)
	}

	strategy, e := plugins.MakeStrategy(sdex, ieif, tradingPair, &assetBase, &assetQuote, *options.strategy, *options.stratConfigPath, *options.simMode)
	if e != nil {
		l.Info("")
		l.Errorf("%s", e)
		// we want to delete all the offers and exit here since there is something wrong with our setup
		deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker)
	}
	return strategy
}

func makeBot(
	l logger.Logger,
	botConfig trader.BotConfig,
	client *horizon.Client,
	sdex *plugins.SDEX,
	exchangeShim api.ExchangeShim,
	ieif *plugins.IEIF,
	tradingPair *model.TradingPair,
	strategy api.Strategy,
	threadTracker *multithreading.ThreadTracker,
	options inputs,
) *trader.Trader {
	timeController := plugins.MakeIntervalTimeController(
		time.Duration(botConfig.TickIntervalSeconds)*time.Second,
		botConfig.MaxTickDelayMillis,
	)
	submitMode, e := api.ParseSubmitMode(botConfig.SubmitMode)
	if e != nil {
		log.Println()
		log.Println(e)
		// we want to delete all the offers and exit here since there is something wrong with our setup
		deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker)
	}
	dataKey := model.MakeSortedBotKey(botConfig.AssetBase(), botConfig.AssetQuote())
	alert, e := monitoring.MakeAlert(botConfig.AlertType, botConfig.AlertAPIKey)
	if e != nil {
		l.Infof("Unable to set up monitoring for alert type '%s' with the given API key\n", botConfig.AlertType)
	}
	bot := trader.MakeBot(
		client,
		ieif,
		botConfig.AssetBase(),
		botConfig.AssetQuote(),
		tradingPair,
		botConfig.TradingAccount(),
		sdex,
		exchangeShim,
		strategy,
		timeController,
		botConfig.DeleteCyclesThreshold,
		submitMode,
		threadTracker,
		options.fixedIterations,
		dataKey,
		alert,
	)
	return bot
}

func convertDeprecatedBotConfigValues(l logger.Logger, botConfig trader.BotConfig) trader.BotConfig {
	if botConfig.CentralizedMinBaseVolumeOverride != nil && botConfig.MinCentralizedBaseVolumeDeprecated != nil {
		l.Infof("deprecation warning: cannot set both '%s' (deprecated) and '%s' in the trader config, using value from '%s'\n", "MIN_CENTRALIZED_BASE_VOLUME", "CENTRALIZED_MIN_BASE_VOLUME_OVERRIDE", "CENTRALIZED_MIN_BASE_VOLUME_OVERRIDE")
	} else if botConfig.MinCentralizedBaseVolumeDeprecated != nil {
		l.Infof("deprecation warning: '%s' is deprecated, use the field '%s' in the trader config instead, see sample_trader.cfg as an example\n", "MIN_CENTRALIZED_BASE_VOLUME", "CENTRALIZED_MIN_BASE_VOLUME_OVERRIDE")
	}
	if botConfig.CentralizedMinBaseVolumeOverride == nil {
		botConfig.CentralizedMinBaseVolumeOverride = botConfig.MinCentralizedBaseVolumeDeprecated
	}
	return botConfig
}

func runTradeCmd(options inputs) {
	l := logger.MakeBasicLogger()
	botConfig := readBotConfig(l, options)
	botConfig = convertDeprecatedBotConfigValues(l, botConfig)
	l.Infof("Trading %s:%s for %s:%s\n", botConfig.AssetCodeA, botConfig.IssuerA, botConfig.AssetCodeB, botConfig.IssuerB)

	// --- start initialization of objects ----
	threadTracker := multithreading.MakeThreadTracker()
	assetBase := botConfig.AssetBase()
	assetQuote := botConfig.AssetQuote()
	tradingPair := &model.TradingPair{
		Base:  model.Asset(utils.Asset2CodeString(assetBase)),
		Quote: model.Asset(utils.Asset2CodeString(assetQuote)),
	}

	client := &horizon.Client{
		URL:  botConfig.HorizonURL,
		HTTP: http.DefaultClient,
	}
	newClient := &horizonclient.Client{
		// TODO horizonclient.Client has a bug in it where it does not use "/" to separate the horizonURL from the fee_stats endpoint
		HorizonURL: botConfig.HorizonURL + "/",
		HTTP:       http.DefaultClient,
	}
	if !*options.noHeaders {
		client.AppName = "kelp"
		client.AppVersion = version
		newClient.AppName = "kelp"
		newClient.AppVersion = version

		p := prefs.Make(prefsFilename)
		if p.FirstTime() {
			log.Printf("Kelp sets the `X-App-Name` and `X-App-Version` headers on requests made to Horizon. These headers help us track overall Kelp usage, so that we can learn about general usage patterns and adapt Kelp to be more useful in the future. These can be turned off using the `--no-headers` flag. See `kelp trade --help` for more information.\n")
			e := p.SetNotFirstTime()
			if e != nil {
				l.Info("")
				l.Errorf("unable to create preferences file: %s", e)
				// we can still proceed with this error
			}
		}
	}

	if botConfig.CcxtRestURL != nil {
		e := sdk.SetBaseURL(*botConfig.CcxtRestURL)
		if e != nil {
			logger.Fatal(l, fmt.Errorf("unable to set CCXT-rest URL to '%s': %s", *botConfig.CcxtRestURL, e))
		}
	}
	l.Infof("using CCXT-rest URL: %s\n", sdk.GetBaseURL())

	ieif := plugins.MakeIEIF(botConfig.IsTradingSdex())
	network := utils.ParseNetwork(botConfig.HorizonURL)
	exchangeShim, sdex := makeExchangeShimSdex(
		l,
		botConfig,
		options,
		client,
		newClient,
		ieif,
		network,
		threadTracker,
		tradingPair,
	)
	strategy := makeStrategy(
		l,
		network,
		botConfig,
		client,
		sdex,
		exchangeShim,
		assetBase,
		assetQuote,
		ieif,
		tradingPair,
		options,
		threadTracker,
	)
	bot := makeBot(
		l,
		botConfig,
		client,
		sdex,
		exchangeShim,
		ieif,
		tradingPair,
		strategy,
		threadTracker,
		options,
	)
	// --- end initialization of objects ---
	// --- start initialization of services ---
	validateTrustlines(l, client, &botConfig)
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
				deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker)
			}
		}()
	}
	startFillTracking(
		l,
		strategy,
		botConfig,
		client,
		sdex,
		exchangeShim,
		tradingPair,
		threadTracker,
	)
	// --- end initialization of services ---

	l.Info("Starting the trader bot...")
	bot.Start()
}

func startMonitoringServer(l logger.Logger, botConfig trader.BotConfig) error {
	healthMetrics, e := monitoring.MakeMetricsRecorder(map[string]interface{}{"success": true})
	if e != nil {
		return fmt.Errorf("unable to make metrics recorder for the /health endpoint: %s", e)
	}
	healthEndpoint, e := monitoring.MakeMetricsEndpoint("/health", healthMetrics, networking.NoAuth)
	if e != nil {
		return fmt.Errorf("unable to make /health endpoint: %s", e)
	}

	kelpMetrics, e := monitoring.MakeMetricsRecorder(nil)
	if e != nil {
		return fmt.Errorf("unable to make metrics recorder for the /metrics endpoint: %s", e)
	}
	metricsAuth := networking.NoAuth
	if botConfig.GoogleClientID != "" || botConfig.GoogleClientSecret != "" {
		metricsAuth = networking.GoogleAuth
	}
	metricsEndpoint, e := monitoring.MakeMetricsEndpoint("/metrics", kelpMetrics, metricsAuth)
	if e != nil {
		return fmt.Errorf("unable to make /metrics endpoint: %s", e)
	}

	serverConfig := &networking.Config{
		GoogleClientID:     botConfig.GoogleClientID,
		GoogleClientSecret: botConfig.GoogleClientSecret,
		PermittedEmails:    map[string]bool{},
	}
	for _, email := range strings.Split(botConfig.AcceptableEmails, ",") {
		serverConfig.PermittedEmails[email] = true
	}
	server, e := networking.MakeServer(serverConfig, []networking.Endpoint{healthEndpoint, metricsEndpoint})
	if e != nil {
		return fmt.Errorf("unable to initialize the metrics server: %s", e)
	}

	l.Infof("Starting monitoring server on port %d\n", botConfig.MonitoringPort)
	return server.StartServer(botConfig.MonitoringPort, botConfig.MonitoringTLSCert, botConfig.MonitoringTLSKey)
}

func startFillTracking(
	l logger.Logger,
	strategy api.Strategy,
	botConfig trader.BotConfig,
	client *horizon.Client,
	sdex *plugins.SDEX,
	exchangeShim api.ExchangeShim,
	tradingPair *model.TradingPair,
	threadTracker *multithreading.ThreadTracker,
) {
	strategyFillHandlers, e := strategy.GetFillHandlers()
	if e != nil {
		l.Info("")
		l.Info("problem encountered while instantiating the fill tracker:")
		l.Errorf("%s", e)
		deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker)
	}

	if botConfig.FillTrackerSleepMillis != 0 {
		fillTracker := plugins.MakeFillTracker(tradingPair, threadTracker, exchangeShim, botConfig.FillTrackerSleepMillis, botConfig.FillTrackerDeleteCyclesThreshold)
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
				l.Errorf("problem encountered while running the fill tracker: %s", e)
				// we want to delete all the offers and exit here because we don't want the bot to run if fill tracking isn't working
				deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker)
			}
		}()
	} else if strategyFillHandlers != nil && len(strategyFillHandlers) > 0 {
		l.Info("")
		l.Error("error: strategy has FillHandlers but fill tracking was disabled (set FILL_TRACKER_SLEEP_MILLIS to a non-zero value)")
		// we want to delete all the offers and exit here because we don't want the bot to run if fill tracking isn't working
		deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker)
	}
}

func validateTrustlines(l logger.Logger, client *horizon.Client, botConfig *trader.BotConfig) {
	if !botConfig.IsTradingSdex() {
		l.Info("no need to validate trustlines because we're not using SDEX as the trading exchange")
		return
	}

	log.Printf("validating trustlines...\n")
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
	l.Info("trustlines valid")
}

func deleteAllOffersAndExit(
	l logger.Logger,
	botConfig trader.BotConfig,
	client *horizon.Client,
	sdex *plugins.SDEX,
	exchangeShim api.ExchangeShim,
	threadTracker *multithreading.ThreadTracker,
) {
	l.Info("")
	l.Infof("waiting for all outstanding threads (%d) to finish before loading offers to be deleted...", threadTracker.NumActiveThreads())
	threadTracker.Stop(multithreading.StopModeError)
	threadTracker.Wait()
	l.Info("...all outstanding threads finished")

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
		e := exchangeShim.SubmitOpsSynch(dOps, func(hash string, e error) {
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
			l.Infof("sleeping for %d seconds until our deletion is confirmed and we exit...(should never reach this line since we submit delete ops synchronously)\n", sleepSeconds)
			time.Sleep(time.Duration(sleepSeconds) * time.Second)
		}
	} else {
		logger.Fatal(l, fmt.Errorf("...nothing to delete, exiting"))
	}
}

func setLogFile(l logger.Logger, options inputs, botConfig trader.BotConfig) {
	t := time.Now().Format("20060102T150405MST")
	fileName := fmt.Sprintf("%s_%s_%s_%s_%s_%s.log", *options.logPrefix, botConfig.AssetCodeA, botConfig.IssuerA, botConfig.AssetCodeB, botConfig.IssuerB, t)
	if !botConfig.IsTradingSdex() {
		fileName = fmt.Sprintf("%s_%s_%s_%s.log", *options.logPrefix, botConfig.AssetCodeA, botConfig.AssetCodeB, t)
	}

	f, e := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if e != nil {
		logger.Fatal(l, fmt.Errorf("failed to set log file: %s", e))
		return
	}
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	l.Infof("logging to file: %s\n", fileName)
	// we want to create a deferred recovery function here that will log panics to the log file and then exit
	defer logPanic(l)
}
