package cmd

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/nikhilsaraf/go-tools/multithreading"
	"github.com/spf13/cobra"

	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/support/config"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/kelpdb"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/constants"
	"github.com/stellar/kelp/support/database"
	"github.com/stellar/kelp/support/logger"
	"github.com/stellar/kelp/support/monitoring"
	"github.com/stellar/kelp/support/networking"
	"github.com/stellar/kelp/support/prefs"
	"github.com/stellar/kelp/support/sdk"
	"github.com/stellar/kelp/support/utils"
	"github.com/stellar/kelp/trader"
)

var upgradeScripts = []*database.UpgradeScript{
	database.MakeUpgradeScript(1,
		database.SqlDbVersionTableCreate,
	),
	database.MakeUpgradeScript(2,
		kelpdb.SqlMarketsTableCreate,
		kelpdb.SqlTradesTableCreate,
		kelpdb.SqlTradesIndexCreate,
	),
	database.MakeUpgradeScript(3,
		kelpdb.SqlTradesIndexDrop,
		kelpdb.SqlTradesIndexCreate2,
	),
	database.MakeUpgradeScript(4,
		database.SqlDbVersionTableAlter1,
	),
	database.MakeUpgradeScript(5,
		kelpdb.SqlTradesTableAlter1,
		kelpdb.SqlTradesIndexCreate3,
	),
	database.MakeUpgradeScript(6,
		kelpdb.SqlStrategyMirrorTradeTriggersTableCreate,
		kelpdb.SqlTradesTableAlter2,
	),
}

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

func logPanic(l logger.Logger, fatalOnError bool) {
	if r := recover(); r != nil {
		st := debug.Stack()
		l.Errorf("PANIC!! recovered to log it in the file\npanic: %v\n\n%s\n", r, string(st))
		if fatalOnError {
			logger.Fatal(l, fmt.Errorf("PANIC!! recovered to log it in the file\npanic: %v\n\n%s\n", r, string(st)))
		}
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
	trigger                       *string
	guiUserID                     *string
	cpuProfile                    *string
	memProfile                    *string
}

func validateCliParams(l logger.Logger, options inputs) {
	checkInitRootFlags()

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

	if *options.trigger != constants.TriggerDefault && *options.trigger != constants.TriggerUI && *options.trigger != constants.TriggerKaas {
		panic(fmt.Sprintf("invalid trigger argument: '%s'", *options.trigger))
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

	if botConfig.SleepMode != "" && botConfig.SleepMode != trader.SleepModeBegin.String() && botConfig.SleepMode != trader.SleepModeEnd.String() {
		logger.Fatal(l, fmt.Errorf("SLEEP_MODE needs to be set to either '%s' or '%s'", trader.SleepModeBegin, trader.SleepModeEnd))
	}
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
	options.noHeaders = tradeCmd.Flags().Bool("no-headers", false, "do not use Amplitude or set X-App-Name and X-App-Version headers on requests to horizon")
	options.trigger = tradeCmd.Flags().String("trigger", constants.TriggerDefault, fmt.Sprintf("indicates a bot that is triggered from a parent process ('%s' or '%s')", constants.TriggerUI, constants.TriggerKaas))
	options.guiUserID = tradeCmd.Flags().String("gui-user-id", "", "specifies the guiUserID associated with this bot to use for metric tracking")
	options.cpuProfile = tradeCmd.Flags().String("cpuprofile", "", "write cpu profile to `file`")
	options.memProfile = tradeCmd.Flags().String("memprofile", "", "write memory profile to `file`")

	requiredFlag("botConf")
	requiredFlag("strategy")
	hiddenFlag("operationalBuffer")
	hiddenFlag("operationalBufferNonNativePct")
	hiddenFlag("trigger")
	hiddenFlag("gui-user-id")
	tradeCmd.Flags().SortFlags = false

	tradeCmd.Run = func(ccmd *cobra.Command, args []string) {
		// TODO NS - profiling fails if we call os.Exit
		if *options.cpuProfile != "" {
			f, e := os.Create(*options.cpuProfile)
			if e != nil {
				log.Fatal("could not create CPU profile: ", e)
			}
			defer func() {
				e := f.Close()
				if e != nil {
					log.Fatalf("could not close file: %s", e)
				}
			}()
			if e := pprof.StartCPUProfile(f); e != nil {
				log.Fatal("could not start CPU profile: ", e)
			}
			defer pprof.StopCPUProfile()
		}

		runTradeCmd(options)

		if *options.memProfile != "" {
			f, e := os.Create(*options.memProfile)
			if e != nil {
				log.Fatal("could not create memory profile: ", e)
			}
			defer func() {
				e := f.Close()
				if e != nil {
					log.Fatalf("could not close file: %s", e)
				}
			}()
			runtime.GC() // get up-to-date statistics
			if e := pprof.WriteHeapProfile(f); e != nil {
				log.Fatal("could not write memory profile: ", e)
			}
		}
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

func readBotConfig(l logger.Logger, options inputs, botStartTime time.Time) trader.BotConfig {
	var botConfig trader.BotConfig
	e := config.Read(*options.botConfigPath, &botConfig)
	utils.CheckConfigError(botConfig, e, *options.botConfigPath)
	e = botConfig.Init()
	if e != nil {
		logger.Fatal(l, e)
	}

	if *options.logPrefix != "" {
		logFilename := makeLogFilename(*options.logPrefix, botConfig, botStartTime)
		setLogFile(l, logFilename)
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
	client *horizonclient.Client,
	ieif *plugins.IEIF,
	network string,
	threadTracker *multithreading.ThreadTracker,
	tradingPair *model.TradingPair,
	sdexAssetMap map[model.Asset]hProtocol.Asset,
) (api.ExchangeShim, *plugins.SDEX) {
	var e error
	var exchangeShim api.ExchangeShim
	if !botConfig.IsTradingSdex() {
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

		exchangeAPIKeys := botConfig.ExchangeAPIKeys.ToExchangeAPIKeys()
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

	feeFn := makeFeeFn(l, botConfig, client)
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
	network string,
	botConfig trader.BotConfig,
	client *horizonclient.Client,
	sdex *plugins.SDEX,
	exchangeShim api.ExchangeShim,
	assetBase hProtocol.Asset,
	assetQuote hProtocol.Asset,
	marketID string,
	ieif *plugins.IEIF,
	tradingPair *model.TradingPair,
	filterFactory *plugins.FilterFactory,
	options inputs,
	threadTracker *multithreading.ThreadTracker,
	db *sql.DB,
	metricsTracker *plugins.MetricsTracker,
) api.Strategy {
	// setting the temp hack variables for the sdex price feeds
	e := plugins.SetPrivateSdexHack(client, plugins.MakeIEIF(true), network)
	if e != nil {
		l.Info("")
		l.Errorf("%s", e)
		// we want to delete all the offers and exit here since there is something wrong with our setup
		deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker, metricsTracker)
	}

	strategy, e := plugins.MakeStrategy(
		sdex,
		exchangeShim,
		exchangeShim,
		ieif,
		tradingPair,
		&assetBase,
		&assetQuote,
		marketID,
		*options.strategy,
		*options.stratConfigPath,
		*options.simMode,
		botConfig.IsTradingSdex(),
		filterFactory,
		db,
	)
	if e != nil {
		l.Info("")
		l.Errorf("%s", e)
		// we want to delete all the offers and exit here since there is something wrong with our setup
		deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker, metricsTracker)
	}
	return strategy
}

func makeBot(
	l logger.Logger,
	botConfig trader.BotConfig,
	client *horizonclient.Client,
	sdex *plugins.SDEX,
	exchangeShim api.ExchangeShim,
	ieif *plugins.IEIF,
	tradingPair *model.TradingPair,
	filterFactory *plugins.FilterFactory,
	strategy api.Strategy,
	fillTracker api.FillTracker,
	threadTracker *multithreading.ThreadTracker,
	options inputs,
	metricsTracker *plugins.MetricsTracker,
	botStartTime time.Time,
) *trader.Trader {
	timeController := plugins.MakeIntervalTimeController(
		time.Duration(botConfig.TickIntervalMillis)*time.Millisecond,
		botConfig.MaxTickDelayMillis,
	)
	submitMode, e := api.ParseSubmitMode(botConfig.SubmitMode)
	if e != nil {
		log.Println()
		log.Println(e)
		// we want to delete all the offers and exit here since there is something wrong with our setup
		deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker, metricsTracker)
	}

	if botConfig.SynchronizeStateLoadEnable && botConfig.SynchronizeStateLoadMaxRetries < 0 {
		log.Println()
		utils.PrintErrorHintf("SYNCHRONIZE_STATE_LOAD_MAX_RETRIES needs to be greater than or equal to 0 when SYNCHRONIZE_STATE_LOAD_ENABLE is set to true")
		// we want to delete all the offers and exit here since there is something wrong with our setup
		deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker, metricsTracker)
	}

	assetBase := botConfig.AssetBase()
	assetQuote := botConfig.AssetQuote()
	dataKey := model.MakeSortedBotKey(assetBase, assetQuote)
	alert, e := monitoring.MakeAlert(botConfig.AlertType, botConfig.AlertAPIKey)
	if e != nil {
		l.Infof("Unable to set up monitoring for alert type '%s' with the given API key\n", botConfig.AlertType)
	}

	var valueBaseFeed api.PriceFeed
	var valueQuoteFeed api.PriceFeed
	if botConfig.DollarValueFeedBaseAsset != "" && botConfig.DollarValueFeedQuoteAsset != "" {
		valueBaseFeed, e = parseValueFeed(botConfig.DollarValueFeedBaseAsset)
		if e != nil {
			log.Println()
			log.Println(e)
			// we want to delete all the offers and exit here since there is something wrong with our setup
			deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker, metricsTracker)
		}

		valueQuoteFeed, e = parseValueFeed(botConfig.DollarValueFeedQuoteAsset)
		if e != nil {
			log.Println()
			log.Println(e)
			// we want to delete all the offers and exit here since there is something wrong with our setup
			deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker, metricsTracker)
		}
	}

	// start make filters
	submitFilters := []plugins.SubmitFilter{}
	if submitMode == api.SubmitModeMakerOnly {
		submitFilters = append(submitFilters,
			plugins.MakeFilterMakerMode(exchangeShim, sdex, tradingPair),
		)
	}
	if len(botConfig.Filters) > 0 && *options.strategy != "sell" && *options.strategy != "sell_twap" && *options.strategy != "buy_twap" && *options.strategy != "delete" {
		log.Println()
		utils.PrintErrorHintf("FILTERS currently only supported on 'sell', 'sell_twap', 'buy_twap', 'delete' strategies, remove FILTERS from the trader config file")
		// we want to delete all the offers and exit here since there is something wrong with our setup
		deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker, metricsTracker)
	}
	for _, filterString := range botConfig.Filters {
		filter, e := filterFactory.MakeFilter(filterString)
		if e != nil {
			log.Println()
			log.Println(e)
			// we want to delete all the offers and exit here since there is something wrong with our setup
			deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker, metricsTracker)
		}
		submitFilters = append(submitFilters, filter)
	}
	// exchange constraints filter is last so we catch any modifications made by previous filters. this ensures that the exchange is
	// less likely to reject our updates
	submitFilters = append(submitFilters,
		plugins.MakeFilterOrderConstraints(exchangeShim.GetOrderConstraints(tradingPair), assetBase, assetQuote),
	)
	// end make filters

	return trader.MakeTrader(
		client,
		ieif,
		assetBase,
		assetQuote,
		valueBaseFeed,
		valueQuoteFeed,
		botConfig.TradingAccount(),
		sdex,
		exchangeShim,
		strategy,
		timeController,
		trader.ParseSleepMode(botConfig.SleepMode),
		botConfig.SynchronizeStateLoadEnable,
		botConfig.SynchronizeStateLoadMaxRetries,
		fillTracker,
		botConfig.DeleteCyclesThreshold,
		submitMode,
		submitFilters,
		threadTracker,
		options.fixedIterations,
		dataKey,
		alert,
		metricsTracker,
		botStartTime,
	)
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

	if botConfig.TickIntervalMillis != 0 && botConfig.TickIntervalSecondsDeprecated != 0 {
		l.Infof("deprecation warning: cannot set both '%s' (deprecated) and '%s' in the trader config, using value from '%s'\n", "TICK_INTERVAL_SECONDS", "TICK_INTERVAL_MILLIS", "TICK_INTERVAL_MILLIS")
	} else if botConfig.TickIntervalSecondsDeprecated != 0 {
		l.Infof("deprecation warning: '%s' is deprecated, use the field '%s' in the trader config instead, see sample_trader.cfg as an example\n", "TICK_INTERVAL_SECONDS", "TICK_INTERVAL_MILLIS")
	}
	if botConfig.TickIntervalMillis == 0 {
		botConfig.TickIntervalMillis = botConfig.TickIntervalSecondsDeprecated * 1000
	}

	return botConfig
}

func runTradeCmd(options inputs) {
	l := logger.MakeBasicLogger()
	botStartTime := time.Now()
	botConfig := readBotConfig(l, options, botStartTime)
	botConfig = convertDeprecatedBotConfigValues(l, botConfig)
	l.Infof("Trading %s:%s for %s:%s\n", botConfig.AssetCodeA, botConfig.IssuerA, botConfig.AssetCodeB, botConfig.IssuerB)

	var guiVersionFlag string
	if *options.trigger == constants.TriggerUI || *options.trigger == constants.TriggerKaas {
		guiVersionFlag = guiVersion
	}

	userID, e := getUserID(l, botConfig)
	if e != nil {
		logger.Fatal(l, fmt.Errorf("could not get user id: %s", e))
	}
	deviceID, e := machineid.ID()
	if e != nil {
		logger.Fatal(l, fmt.Errorf("could not generate machine id: %s", e))
	}
	isTestnet := strings.Contains(botConfig.HorizonURL, "test") && botConfig.IsTradingSdex()
	metricsTracker, e := plugins.MakeMetricsTracker(
		http.DefaultClient,
		amplitudeAPIKey,
		userID,
		*options.guiUserID,
		deviceID,
		botStartTime,
		*options.noHeaders, // disable metrics if the CLI specified no headers
		plugins.MakeCommonProps(
			version,
			gitHash,
			env,
			runtime.GOOS,
			runtime.GOARCH,
			goarm,
			runtime.Version(),
			0,
			isTestnet,
			guiVersionFlag,
		),
		plugins.MakeCliProps(
			*options.strategy,
			float64(botConfig.TickIntervalMillis)/1000,
			botConfig.TradingExchange,
			botConfig.TradingPair(),
			botConfig.MaxTickDelayMillis,
			botConfig.SubmitMode,
			botConfig.DeleteCyclesThreshold,
			botConfig.FillTrackerSleepMillis,
			botConfig.FillTrackerDeleteCyclesThreshold,
			botConfig.SynchronizeStateLoadEnable,
			botConfig.SynchronizeStateLoadMaxRetries,
			botConfig.DollarValueFeedBaseAsset != "" && botConfig.DollarValueFeedQuoteAsset != "",
			botConfig.AlertType,
			int(botConfig.MonitoringPort) != 0,
			len(botConfig.Filters) > 0,
			botConfig.PostgresDbConfig != nil,
			*options.logPrefix != "",
			*options.operationalBuffer,
			*options.operationalBufferNonNativePct,
			*options.simMode,
			*options.fixedIterations,
		),
	)
	if e != nil {
		logger.Fatal(l, fmt.Errorf("could not generate metrics tracker: %s", e))
	}

	e = metricsTracker.SendStartupEvent(time.Now())
	if e != nil {
		l.Infof("metric - could not send startup event metric: %s", e)
	}

	// --- start initialization of objects ----
	threadTracker := multithreading.MakeThreadTracker()
	assetBase := botConfig.AssetBase()
	assetQuote := botConfig.AssetQuote()
	tradingPair := &model.TradingPair{
		Base:  model.Asset(utils.Asset2CodeString(assetBase)),
		Quote: model.Asset(utils.Asset2CodeString(assetQuote)),
	}

	client := &horizonclient.Client{
		HorizonURL: botConfig.HorizonURL,
		HTTP:       http.DefaultClient,
	}
	if !*options.noHeaders {
		client.AppName = "kelp--cli--bot"
		if *options.trigger == constants.TriggerUI {
			client.AppName = "kelp--gui-desktop--bot"
		} else if *options.trigger == constants.TriggerKaas {
			client.AppName = "kelp--gui-kaas--bot"
		}
		client.AppVersion = version

		p := prefs.Make(prefsFilename)
		if p.FirstTime() {
			log.Printf("Kelp sets the `X-App-Name` and `X-App-Version` headers on requests made to Horizon. These headers help us track overall Kelp usage, so that we can learn about general usage patterns and adapt Kelp to be more useful in the future. Kelp also uses Amplitude for metric tracking. These can be turned off using the `--no-headers` flag. See `kelp trade --help` for more information.\n")
			e := p.SetNotFirstTime()
			if e != nil {
				l.Info("")
				l.Errorf("unable to create preferences file: %s", e)
				// we can still proceed with this error
			}
		}
	}
	log.Printf("using client.AppName = %s", client.AppName)

	if *rootCcxtRestURL == "" && botConfig.CcxtRestURL != nil {
		e := sdk.SetBaseURL(*botConfig.CcxtRestURL)
		if e != nil {
			logger.Fatal(l, fmt.Errorf("unable to set CCXT-rest URL to '%s': %s", *botConfig.CcxtRestURL, e))
		}
	}
	l.Infof("using CCXT-rest URL: %s\n", sdk.GetBaseURL())

	ieif := plugins.MakeIEIF(botConfig.IsTradingSdex())
	network := utils.ParseNetwork(botConfig.HorizonURL)
	sdexAssetMap := map[model.Asset]hProtocol.Asset{
		tradingPair.Base:  botConfig.AssetBase(),
		tradingPair.Quote: botConfig.AssetQuote(),
	}
	assetDisplayFn := model.MakePassthroughAssetDisplayFn()
	if botConfig.IsTradingSdex() {
		assetDisplayFn = model.MakeSdexMappedAssetDisplayFn(sdexAssetMap)
	}

	var db *sql.DB
	if botConfig.PostgresDbConfig != nil {
		if !botConfig.SynchronizeStateLoadEnable && botConfig.FillTrackerSleepMillis == 0 {
			log.Println()
			utils.PrintErrorHintf("SYNCHRONIZE_STATE_LOAD_ENABLE needs to be enabled and/or FILL_TRACKER_SLEEP_MILLIS needs to be set in the trader.cfg file when the POSTGRES_DB is enabled so we can fetch trades to be saved in the db")
			logger.Fatal(l, fmt.Errorf("invalid trader.cfg config, need to set SYNCHRONIZE_STATE_LOAD_ENABLE and/or FILL_TRACKER_SLEEP_MILLIS"))
		}

		if botConfig.DbOverrideAccountID == "" {
			log.Println()
			utils.PrintErrorHintf("DB_OVERRIDE__ACCOUNT_ID needs to be set in the trader.cfg file when the POSTGRES_DB is enabled so we can assign an account_id to trades that are fetched before writing them in the db")
			logger.Fatal(l, fmt.Errorf("invalid trader.cfg config, need to set DB_OVERRIDE__ACCOUNT_ID"))
		}

		var e error
		db, e = database.ConnectInitializedDatabase(botConfig.PostgresDbConfig, upgradeScripts, version)
		if e != nil {
			logger.Fatal(l, fmt.Errorf("problem encountered while initializing the db: %s", e))
		}
		log.Printf("made db instance with config: %s\n", botConfig.PostgresDbConfig.MakeConnectString())
	}
	exchangeShim, sdex := makeExchangeShimSdex(
		l,
		botConfig,
		options,
		client,
		ieif,
		network,
		threadTracker,
		tradingPair,
		sdexAssetMap,
	)
	filterFactory := &plugins.FilterFactory{
		ExchangeName:   botConfig.TradingExchangeName(),
		TradingPair:    tradingPair,
		AssetDisplayFn: assetDisplayFn,
		BaseAsset:      assetBase,
		QuoteAsset:     assetQuote,
		DB:             db,
	}
	baseString, e := assetDisplayFn(tradingPair.Base)
	if e != nil {
		logger.Fatal(l, fmt.Errorf("could not convert base trading pair to string: %s", e))
	}
	quoteString, e := assetDisplayFn(tradingPair.Quote)
	if e != nil {
		logger.Fatal(l, fmt.Errorf("could not convert quote trading pair to string: %s", e))
	}
	marketID := plugins.MakeMarketID(botConfig.TradingExchangeName(), baseString, quoteString)
	strategy := makeStrategy(
		l,
		network,
		botConfig,
		client,
		sdex,
		exchangeShim,
		assetBase,
		assetQuote,
		marketID,
		ieif,
		tradingPair,
		filterFactory,
		options,
		threadTracker,
		db,
		metricsTracker,
	)
	fillTracker := makeFillTracker(
		l,
		strategy,
		botConfig,
		client,
		sdex,
		exchangeShim,
		tradingPair,
		assetDisplayFn,
		db,
		threadTracker,
		botConfig.DbOverrideAccountID,
		metricsTracker,
	)
	bot := makeBot(
		l,
		botConfig,
		client,
		sdex,
		exchangeShim,
		ieif,
		tradingPair,
		filterFactory,
		strategy,
		fillTracker,
		threadTracker,
		options,
		metricsTracker,
		botStartTime,
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
				deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker, metricsTracker)
			}
		}()
	}
	if fillTracker != nil && botConfig.FillTrackerSleepMillis != 0 {
		l.Infof("Starting fill tracker with %d handlers\n", fillTracker.NumHandlers())
		go func() {
			e := fillTracker.TrackFills()
			if e != nil {
				l.Info("")
				l.Errorf("problem encountered while running the fill tracker: %s", e)
				// we want to delete all the offers and exit here because we don't want the bot to run if fill tracking isn't working
				deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker, metricsTracker)
			}
		}()
	}
	// --- end initialization of services ---

	l.Info("Starting the trader bot...")
	bot.Start()
}

func getUserID(l logger.Logger, botConfig trader.BotConfig) (string, error) {
	var userIDPrehash string
	if botConfig.IsTradingSdex() {
		userIDPrehash = botConfig.TradingAccount()
	} else {
		exchangeAPIKeys := botConfig.ExchangeAPIKeys.ToExchangeAPIKeys()
		if len(exchangeAPIKeys) == 0 {
			return "", fmt.Errorf("could not find exchange API key on bot config")
		}

		userIDPrehash = exchangeAPIKeys[0].Key
	}

	// hash avoids exposing the user account or api key
	userIDHashed, e := utils.HashString(userIDPrehash)
	if e != nil {
		return "", fmt.Errorf("could not create user id: %s", e)
	}

	return fmt.Sprint(userIDHashed), nil
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
	server, e := networking.MakeServerWithGoogleAuth(serverConfig, []networking.Endpoint{healthEndpoint, metricsEndpoint})
	if e != nil {
		return fmt.Errorf("unable to initialize the metrics server: %s", e)
	}

	l.Infof("Starting monitoring server on port %d\n", botConfig.MonitoringPort)
	return server.StartServer(botConfig.MonitoringPort, botConfig.MonitoringTLSCert, botConfig.MonitoringTLSKey)
}

func makeFillTracker(
	l logger.Logger,
	strategy api.Strategy,
	botConfig trader.BotConfig,
	client *horizonclient.Client,
	sdex *plugins.SDEX,
	exchangeShim api.ExchangeShim,
	tradingPair *model.TradingPair,
	assetDisplayFn model.AssetDisplayFn,
	db *sql.DB,
	threadTracker *multithreading.ThreadTracker,
	accountID string,
	metricsTracker *plugins.MetricsTracker,
) api.FillTracker {
	strategyFillHandlers, e := strategy.GetFillHandlers()
	if e != nil {
		l.Info("")
		l.Info("problem encountered while instantiating the fill tracker:")
		l.Errorf("%s", e)
		deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker, metricsTracker)
	}

	fillTrackerEnabled := botConfig.SynchronizeStateLoadEnable || botConfig.FillTrackerSleepMillis != 0
	if !fillTrackerEnabled && strategyFillHandlers != nil && len(strategyFillHandlers) > 0 {
		l.Info("")
		l.Error("error: strategy has FillHandlers but fill tracking was disabled (set FILL_TRACKER_SLEEP_MILLIS to a non-zero value)")
		// we want to delete all the offers and exit here because we don't want the bot to run if fill tracking isn't working
		deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker, metricsTracker)
	} else if !fillTrackerEnabled {
		return nil
	}

	// start initializing the fill tracker
	var lastCursor interface{}
	if botConfig.FillTrackerLastTradeCursorOverride == "" {
		// loads cursor by fetching from exchange
		lastCursor, e = exchangeShim.GetLatestTradeCursor()
		if e != nil {
			l.Info("")
			l.Error(fmt.Sprintf("could not get last trade cursor from exchangeShim: %s", e))
			// we want to delete all the offers and exit here because we don't want the bot to run if fill tracking isn't working correctly
			deleteAllOffersAndExit(l, botConfig, client, sdex, exchangeShim, threadTracker, metricsTracker)
		}
		log.Printf("set latest trade cursor from where to start tracking fills (no override specified): %v\n", lastCursor)
	} else {
		// loads cursor from config file
		lastCursor = botConfig.FillTrackerLastTradeCursorOverride
		log.Printf("set latest trade cursor from where to start tracking fills (used override value): %v\n", lastCursor)
	}

	fillTracker := plugins.MakeFillTracker(tradingPair, threadTracker, exchangeShim, botConfig.FillTrackerSleepMillis, botConfig.FillTrackerDeleteCyclesThreshold, lastCursor)
	fillLogger := plugins.MakeFillLogger()
	fillTracker.RegisterHandler(fillLogger)
	if db != nil {
		fillDBWriter := plugins.MakeFillDBWriter(db, assetDisplayFn, botConfig.TradingExchangeName(), accountID)
		fillTracker.RegisterHandler(fillDBWriter)
	}
	if strategyFillHandlers != nil {
		for _, h := range strategyFillHandlers {
			fillTracker.RegisterHandler(h)
		}
	}

	return fillTracker
}

func validateTrustlines(l logger.Logger, client *horizonclient.Client, botConfig *trader.BotConfig) {
	if !botConfig.IsTradingSdex() {
		l.Info("no need to validate trustlines because we're not using SDEX as the trading exchange")
		return
	}

	log.Printf("validating trustlines...\n")
	acctReq := horizonclient.AccountRequest{AccountID: botConfig.TradingAccount()}
	account, e := client.AccountDetail(acctReq)
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
	client *horizonclient.Client,
	sdex *plugins.SDEX,
	exchangeShim api.ExchangeShim,
	threadTracker *multithreading.ThreadTracker,
	metricsTracker *plugins.MetricsTracker,
) {
	// synchronous event to guarantee execution. we want to know whenever we enter the delete all offers logic. this function
	// waits for all threads to be synchronous, which is equivalent to sending synchronously. we use
	e := metricsTracker.SendDeleteEvent(true)
	if e != nil {
		// We don't want to crash upon failure, so offers will be deleted regardless of metric send.
		l.Infof("metric - could not send delete event metric: %s", e)
	}

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
		// to delete offers the submitMode doesn't matter, so use api.SubmitModeBoth as the default
		e := exchangeShim.SubmitOpsSynch(api.ConvertOperation2TM(dOps), api.SubmitModeBoth, func(hash string, e error) {
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

func setLogFile(l logger.Logger, filename string) {
	f, e := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if e != nil {
		logger.Fatal(l, fmt.Errorf("failed to set log file: %s", e))
		return
	}
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	l.Infof("logging to file: %s\n", filename)
	// we want to create a deferred recovery function here that will log panics to the log file and then exit
	defer logPanic(l, false)
}

func makeLogFilename(logPrefix string, botConfig trader.BotConfig, botStartTime time.Time) string {
	botStartStr := botStartTime.Format("20060102T150405MST")
	if botConfig.IsTradingSdex() {
		return fmt.Sprintf("%s_%s_%s_%s_%s_%s.log", logPrefix, botConfig.AssetCodeA, botConfig.IssuerA, botConfig.AssetCodeB, botConfig.IssuerB, botStartStr)
	}
	return fmt.Sprintf("%s_%s_%s_%s.log", logPrefix, botConfig.AssetCodeA, botConfig.AssetCodeB, botStartStr)
}

func parseValueFeed(valueFeed string) (api.PriceFeed, error) {
	parts := strings.Split(valueFeed, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("could not parse value feed '%s'", valueFeed)
	}

	pf, e := plugins.MakePriceFeed(parts[0], parts[1])
	if e != nil {
		return nil, fmt.Errorf("could not make value price feed '%s': %s", valueFeed, e)
	}

	return pf, nil
}
