package trader

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/nikhilsaraf/go-tools/multithreading"

	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/utils"
)

const maxLumenTrust float64 = math.MaxFloat64

// Trader represents a market making bot, which is composed of various parts include the strategy and various APIs.
type Trader struct {
	api                            *horizonclient.Client
	ieif                           *plugins.IEIF
	assetBase                      hProtocol.Asset
	assetQuote                     hProtocol.Asset
	valueBaseFeed                  api.PriceFeed
	valueQuoteFeed                 api.PriceFeed
	tradingAccount                 string
	sdex                           *plugins.SDEX
	exchangeShim                   api.ExchangeShim
	strategy                       api.Strategy // the instance of this bot is bound to this strategy
	timeController                 api.TimeController
	sleepMode                      SleepMode
	synchronizeStateLoadEnable     bool
	synchronizeStateLoadMaxRetries int
	fillTracker                    api.FillTracker
	deleteCyclesThreshold          int64
	submitMode                     api.SubmitMode
	submitFilters                  []plugins.SubmitFilter
	threadTracker                  *multithreading.ThreadTracker
	fixedIterations                *uint64
	dataKey                        *model.BotKey
	alert                          api.Alert
	metricsTracker                 *plugins.MetricsTracker
	startTime                      time.Time

	// initialized runtime vars
	deleteCycles int64

	// uninitialized runtime vars
	maxAssetA      float64
	maxAssetB      float64
	trustAssetA    float64
	trustAssetB    float64
	buyingAOffers  []hProtocol.Offer // quoted A/B
	sellingAOffers []hProtocol.Offer // quoted B/A
}

// MakeTrader is the factory method for the Trader struct
func MakeTrader(
	api *horizonclient.Client,
	ieif *plugins.IEIF,
	assetBase hProtocol.Asset,
	assetQuote hProtocol.Asset,
	valueBaseFeed api.PriceFeed,
	valueQuoteFeed api.PriceFeed,
	tradingAccount string,
	sdex *plugins.SDEX,
	exchangeShim api.ExchangeShim,
	strategy api.Strategy,
	timeController api.TimeController,
	sleepMode SleepMode,
	synchronizeStateLoadEnable bool,
	synchronizeStateLoadMaxRetries int,
	fillTracker api.FillTracker,
	deleteCyclesThreshold int64,
	submitMode api.SubmitMode,
	submitFilters []plugins.SubmitFilter,
	threadTracker *multithreading.ThreadTracker,
	fixedIterations *uint64,
	dataKey *model.BotKey,
	alert api.Alert,
	metricsTracker *plugins.MetricsTracker,
	startTime time.Time,
) *Trader {
	return &Trader{
		api:                            api,
		ieif:                           ieif,
		assetBase:                      assetBase,
		assetQuote:                     assetQuote,
		valueBaseFeed:                  valueBaseFeed,
		valueQuoteFeed:                 valueQuoteFeed,
		tradingAccount:                 tradingAccount,
		sdex:                           sdex,
		exchangeShim:                   exchangeShim,
		strategy:                       strategy,
		timeController:                 timeController,
		sleepMode:                      sleepMode,
		synchronizeStateLoadEnable:     synchronizeStateLoadEnable,
		synchronizeStateLoadMaxRetries: synchronizeStateLoadMaxRetries,
		fillTracker:                    fillTracker,
		deleteCyclesThreshold:          deleteCyclesThreshold,
		submitMode:                     submitMode,
		submitFilters:                  submitFilters,
		threadTracker:                  threadTracker,
		fixedIterations:                fixedIterations,
		dataKey:                        dataKey,
		alert:                          alert,
		metricsTracker:                 metricsTracker,
		startTime:                      startTime,
		// initialized runtime vars
		deleteCycles: 0,
	}
}

// Start starts the bot with the injected strategy
func (t *Trader) Start() {
	log.Println("----------------------------------------------------------------------------------------------------")
	// lastUpdateStartTime is the start time of the last update
	var lastUpdateStartTime time.Time
	// lastUpdateEndTime is the end time of the last update
	var lastUpdateEndTime time.Time

	for {
		// ref time for shouldUpdate depends on the sleepMode
		updateRefTime := lastUpdateStartTime
		if t.sleepMode.shouldSleepAtBeginning() {
			// use lastUpdateEndTime here because we want to sleep starting for the time after the last cycle ended (i.e. we want to sleep in the beginning)
			updateRefTime = lastUpdateEndTime
		}

		// skip first sleep cycle if sleeping first so there is no delay when running the bot in the first iteration
		if t.sleepMode.shouldSleepAtBeginning() && !lastUpdateEndTime.IsZero() {
			t.doSleep(lastUpdateEndTime)
		}

		currentUpdateTime := time.Now()
		if updateRefTime.IsZero() || t.timeController.ShouldUpdate(updateRefTime, currentUpdateTime) {
			updateResult := t.update()
			millisForUpdate := time.Since(currentUpdateTime).Milliseconds()
			log.Printf("time taken for update loop: %d millis\n", millisForUpdate)
			if shouldSendUpdateMetric(t.startTime, currentUpdateTime, t.metricsTracker.GetUpdateEventSentTime()) {
				e := t.threadTracker.TriggerGoroutine(func(inputs []interface{}) {
					e := t.metricsTracker.SendUpdateEvent(currentUpdateTime, updateResult, millisForUpdate)
					if e != nil {
						log.Printf("failed to send update event metric: %s", e)
					}
				}, nil)
				if e != nil {
					log.Printf("failed to trigger goroutine for send update event: %s", e)
				}
			}

			if t.fixedIterations != nil && updateResult.Success {
				*t.fixedIterations = *t.fixedIterations - 1
				if *t.fixedIterations <= 0 {
					log.Printf("finished requested number of iterations, waiting for all threads to finish...\n")
					t.threadTracker.Wait()
					log.Printf("...all threads finished, stopping bot update loop\n")
					return
				}
			}

			// wait for any goroutines from the current update to finish so we don't have inconsistent state reads
			t.threadTracker.Wait()
			log.Println("----------------------------------------------------------------------------------------------------")
			lastUpdateStartTime = currentUpdateTime
			// lastUpdateEndTime uses the real time.Now() because we want to capture the actual end time
			lastUpdateEndTime = time.Now()
		}

		if !t.sleepMode.shouldSleepAtBeginning() {
			// this needs to synchronize with the time of the last run attempt
			t.doSleep(lastUpdateStartTime)
		}
	}
}

func (t *Trader) doSleep(lastUpdateTime time.Time) {
	sleepTime := t.timeController.SleepTime(lastUpdateTime)
	log.Printf("sleeping for %s...\n", sleepTime)
	time.Sleep(sleepTime)
}

func shouldSendUpdateMetric(start time.Time, currentUpdate time.Time, lastMetricUpdate *time.Time) bool {
	if lastMetricUpdate == nil {
		return true
	}

	timeFromStart := currentUpdate.Sub(start)
	var refreshMetricInterval time.Duration
	switch {
	case timeFromStart <= 5*time.Minute:
		refreshMetricInterval = 5 * time.Second
	case timeFromStart <= 1*time.Hour:
		refreshMetricInterval = 10 * time.Minute
	default:
		refreshMetricInterval = 1 * time.Hour
	}

	timeSinceLastUpdate := currentUpdate.Sub(*lastMetricUpdate)
	return timeSinceLastUpdate >= refreshMetricInterval
}

// deletes all offers for the bot (not all offers on the account)
func (t *Trader) deleteAllOffers(isAsync bool) {
	logPrefix := ""
	if isAsync {
		logPrefix = "(async) "
	}
	if t.deleteCyclesThreshold < 0 {
		log.Printf("%snot deleting any offers because deleteCyclesThreshold is negative\n", logPrefix)
		return
	}

	t.deleteCycles++
	if t.deleteCycles <= t.deleteCyclesThreshold {
		log.Printf("%snot deleting any offers, deleteCycles (=%d) needs to exceed deleteCyclesThreshold (=%d)\n", logPrefix, t.deleteCycles, t.deleteCyclesThreshold)
		return
	}

	log.Printf("%sdeleting all offers, num. continuous update cycles with errors (including this one): %d; (deleteCyclesThreshold to be exceeded=%d)\n", logPrefix, t.deleteCycles, t.deleteCyclesThreshold)
	dOps := []txnbuild.Operation{}
	dOps = append(dOps, t.sdex.DeleteAllOffers(t.sellingAOffers)...)
	t.sellingAOffers = []hProtocol.Offer{}
	dOps = append(dOps, t.sdex.DeleteAllOffers(t.buyingAOffers)...)
	t.buyingAOffers = []hProtocol.Offer{}

	// LOH-3 - we want to guarantee that the bot crashes if the errors exceed deleteCyclesThreshold, so we start a new thread with a sleep timer to crash the bot as a safety
	defer func() {
		log.Printf("%sstarted thread to crash bot in 1 minute as a fallback (to respect deleteCyclesThreshold)\n", logPrefix)
		time.Sleep(time.Minute)
		log.Fatalf("%sbot should have crashed by now (programmer error?), crashing\n", logPrefix)
	}()

	log.Printf("%screated %d operations to delete offers\n", logPrefix, len(dOps))
	if len(dOps) > 0 {
		e := t.threadTracker.TriggerGoroutine(func(inputs []interface{}) {
			e := t.metricsTracker.SendDeleteEvent(false)
			if e != nil {
				log.Printf("failed to send update event metric: %s", e)
			}
		}, nil)
		if e != nil {
			log.Printf("failed to trigger goroutine for send delete event: %s", e)
			return
		}

		// to delete offers the submitMode doesn't matter, so use api.SubmitModeBoth as the default
		e = t.exchangeShim.SubmitOps(api.ConvertOperation2TM(dOps), api.SubmitModeBoth, func(hash string, e error) {
			log.Fatalf("(async) ...deleted %d offers, exiting (asyncCallback: hash=%s, e=%v)", len(dOps), hash, e)
		})
		if e != nil {
			log.Fatalf("%scontinuing to exit after showing error during submission of delete offer ops: %s", logPrefix, e)
			return
		}
	} else {
		log.Fatalf("%s...nothing to delete, exiting", logPrefix)
	}
}

// synchronizeFetchBalancesOffersTrades pivots checking the balances and offers around trades, ensuring that:
// 1) we fetch and process the latest trades and
// 2) the balances and offers are consistent with the fetched trades
//
// Note1: we cannot pivot around balances and/or offers by checking if if there are 0 trades because it's possible that the
//        background thread has fetched the trades during this time. This is why we check if the balances/offers have changed.
// Note2: if the trade API is not working (like sometimes on Kraken) then this will fail once but will not crash the bot (we
//        want the bot to crash in this scenario). We will end up retring here and subsequent runs will likely succeed to because
//        the bot allows occassional failures. The likelihood that a trade happens exactly during our critical section many times,
//        which would cause multiple failures, is unlikely. Even if that happens, it does not necessarily indicate a failed API as
//        that could just be a coincidence, which is exactly what this synchronization function is preventing against.
func (t *Trader) synchronizeFetchBalancesOffersTrades() error {
	if t.synchronizeStateLoadEnable && !t.fillTracker.IsRunningInBackground() {
		// this is purely an optimization block.
		// run the trades query here so the synchronization logic is cheaper.
		// we will catch any trades that occur here with 1 network call instead of having to retry the synchronization block below.
		// Moreover, we will not consume 1 attempt whenever a legitimate trade does occur (which otherwise is handled by the background
		// fill tracker thread)
		_, e := t.fillTracker.FillTrackSingleIteration()
		if e != nil {
			return fmt.Errorf("unable to get trades: %s", e)
		}
	}

	// run initial query for balanecs and offers
	baseBalance1, quoteBalance1, e := t.getBalances()
	if e != nil {
		return fmt.Errorf("unable to get balances1: %s", e)
	}
	sellingAOffers1, buyingAOffers1, e := t.getExistingOffers()
	if e != nil {
		return fmt.Errorf("unable to get offers1: %s", e)
	}

	if !t.synchronizeStateLoadEnable {
		log.Printf("synchronized state loading is disabled\n")
		t.setBalances(baseBalance1, quoteBalance1)
		t.setExistingOffers(sellingAOffers1, buyingAOffers1)
		return nil
	}

	// on the first iteration, and every subsequent iteration, we want to fetch trades, balances, and offers.
	// this ensures that we reuse the last fetch of balances and offers when retrying.
	for i := 0; i < t.synchronizeStateLoadMaxRetries+1; i++ {
		trades, e := t.fillTracker.FillTrackSingleIteration()
		if e != nil {
			return fmt.Errorf("unable to get trades, iteration %d of %d attempts (1-indexed): %s", i+1, t.synchronizeStateLoadMaxRetries+1, e)
		}

		// reset cache of balances to get actual balances from network
		t.sdex.IEIF().ResetCachedBalances()

		// run it again once we have fetched trades so we can compare that nothing changed and the data is in sync
		baseBalance2, quoteBalance2, e := t.getBalances()
		if e != nil {
			return fmt.Errorf("unable to get balances2, iteration %d of %d attempts (1-indexed): %s", i+1, t.synchronizeStateLoadMaxRetries+1, e)
		}
		sellingAOffers2, buyingAOffers2, e := t.getExistingOffers()
		if e != nil {
			return fmt.Errorf("unable to get offers2, iteration %d of %d attempts (1-indexed): %s", i+1, t.synchronizeStateLoadMaxRetries+1, e)
		}

		if isStateSynchronized(
			trades,
			baseBalance1,
			quoteBalance1,
			sellingAOffers1,
			buyingAOffers1,
			baseBalance2,
			quoteBalance2,
			sellingAOffers2,
			buyingAOffers2,
		) {
			// this is the only success case
			t.setBalances(baseBalance1, quoteBalance1)
			t.setExistingOffers(sellingAOffers1, buyingAOffers1)
			return nil
		}
		log.Printf("could not synchronize data in attempt %d of %d (1-indexed), trying again...\n", i+1, t.synchronizeStateLoadMaxRetries+1)

		// set recently fetched values of balances and offers as our first set of values for the next run
		baseBalance1, quoteBalance1 = baseBalance2, quoteBalance2
		sellingAOffers1, buyingAOffers1 = sellingAOffers2, buyingAOffers2
	}
	return fmt.Errorf("exhausted all %d attempts at synchronizing data when fetching trades, balances, and offers but all attempts failed", t.synchronizeStateLoadMaxRetries+1)
}

func isStateSynchronized(
	trades []model.Trade,
	baseBalance1 *api.Balance,
	quoteBalance1 *api.Balance,
	sellingAOffers1 []hProtocol.Offer,
	buyingAOffers1 []hProtocol.Offer,
	baseBalance2 *api.Balance,
	quoteBalance2 *api.Balance,
	sellingAOffers2 []hProtocol.Offer,
	buyingAOffers2 []hProtocol.Offer,
) bool {
	hasNewTrades := trades != nil && len(trades) > 0
	baseBalanceSame := baseBalance1.Balance == baseBalance2.Balance
	quoteBalanceSame := quoteBalance1.Balance == quoteBalance2.Balance
	sellOffersSame := len(sellingAOffers1) == len(sellingAOffers2)
	buyOffersSame := len(buyingAOffers1) == len(buyingAOffers2)

	isStateSynchronized := !hasNewTrades && baseBalanceSame && quoteBalanceSame && sellOffersSame && buyOffersSame
	if isStateSynchronized {
		log.Printf("isStateSynchronized is %v\n", isStateSynchronized)
	} else {
		log.Printf("isStateSynchronized is %v, values (all should be true for success): !hasNewTrades=%v, baseBalanceSame=%v, quoteBalanceSame=%v, sellOffersSame=%v, buyOffersSame=%v\n",
			isStateSynchronized, !hasNewTrades, baseBalanceSame, quoteBalanceSame, sellOffersSame, buyOffersSame)
	}
	return isStateSynchronized
}

// time to update the order book and possibly readjust the offers
// returns true if the update was successful, otherwise false
func (t *Trader) update() plugins.UpdateLoopResult {
	// initialize counts of types of ops
	numPruneOps := 0
	numUpdateOpsDelete := 0
	numUpdateOpsUpdate := 0
	numUpdateOpsCreate := 0

	e := t.synchronizeFetchBalancesOffersTrades()
	if e != nil {
		log.Println(e)
		t.deleteAllOffers(false)
		return plugins.UpdateLoopResult{
			Success:            false,
			NumPruneOps:        numPruneOps,
			NumUpdateOpsDelete: numUpdateOpsDelete,
			NumUpdateOpsUpdate: numUpdateOpsUpdate,
			NumUpdateOpsCreate: numUpdateOpsCreate,
		}
	}

	pair := &model.TradingPair{
		Base:  model.FromHorizonAsset(t.assetBase),
		Quote: model.FromHorizonAsset(t.assetQuote),
	}
	log.Printf("orderConstraints for trading pair %s: %s", pair, t.exchangeShim.GetOrderConstraints(pair))

	// TODO 2 streamline the request data instead of caching
	// reset cache of balances for this update cycle to reduce redundant requests to calculate asset balances
	t.sdex.IEIF().ResetCachedBalances()
	// reset and recompute cached liabilities for this update cycle
	e = t.sdex.IEIF().ResetCachedLiabilities(t.assetBase, t.assetQuote)
	log.Printf("liabilities after resetting\n")
	t.sdex.IEIF().LogAllLiabilities(t.assetBase, t.assetQuote)
	if e != nil {
		log.Println(e)
		t.deleteAllOffers(false)
		return plugins.UpdateLoopResult{
			Success:            false,
			NumPruneOps:        numPruneOps,
			NumUpdateOpsDelete: numUpdateOpsDelete,
			NumUpdateOpsUpdate: numUpdateOpsUpdate,
			NumUpdateOpsCreate: numUpdateOpsCreate,
		}
	}

	// strategy has a chance to set any state it needs
	e = t.strategy.PreUpdate(t.maxAssetA, t.maxAssetB, t.trustAssetA, t.trustAssetB)
	if e != nil {
		log.Println(e)
		t.deleteAllOffers(false)
		return plugins.UpdateLoopResult{
			Success:            false,
			NumPruneOps:        numPruneOps,
			NumUpdateOpsDelete: numUpdateOpsDelete,
			NumUpdateOpsUpdate: numUpdateOpsUpdate,
			NumUpdateOpsCreate: numUpdateOpsCreate,
		}
	}

	// delete excess offers
	var pruneOps []build.TransactionMutator
	pruneOps, t.buyingAOffers, t.sellingAOffers = t.strategy.PruneExistingOffers(t.buyingAOffers, t.sellingAOffers)
	numPruneOps = len(pruneOps)
	log.Printf("created %d operations to prune excess offers\n", numPruneOps)
	if numPruneOps > 0 {
		// to prune/delete offers the submitMode doesn't matter, so use api.SubmitModeBoth as the default
		e = t.exchangeShim.SubmitOps(pruneOps, api.SubmitModeBoth, nil)
		if e != nil {
			log.Println(e)
			t.deleteAllOffers(false)
			return plugins.UpdateLoopResult{
				Success:            false,
				NumPruneOps:        numPruneOps,
				NumUpdateOpsDelete: numUpdateOpsDelete,
				NumUpdateOpsUpdate: numUpdateOpsUpdate,
				NumUpdateOpsCreate: numUpdateOpsCreate,
			}
		}

		// TODO 2 streamline the request data instead of caching - may not need this since result of PruneOps is async
		// reset cache of balances for this update cycle to reduce redundant requests to calculate asset balances
		t.sdex.IEIF().ResetCachedBalances()
		// reset and recompute cached liabilities for this update cycle
		e = t.sdex.IEIF().ResetCachedLiabilities(t.assetBase, t.assetQuote)
		log.Printf("liabilities after resetting\n")
		t.sdex.IEIF().LogAllLiabilities(t.assetBase, t.assetQuote)
		if e != nil {
			log.Println(e)
			t.deleteAllOffers(false)
			return plugins.UpdateLoopResult{
				Success:            false,
				NumPruneOps:        numPruneOps,
				NumUpdateOpsDelete: numUpdateOpsDelete,
				NumUpdateOpsUpdate: numUpdateOpsUpdate,
				NumUpdateOpsCreate: numUpdateOpsCreate,
			}
		}
	}

	opsOld, e := t.strategy.UpdateWithOps(t.buyingAOffers, t.sellingAOffers)
	log.Printf("liabilities at the end of a call to UpdateWithOps\n")
	t.sdex.IEIF().LogAllLiabilities(t.assetBase, t.assetQuote)
	if e != nil {
		log.Println(e)
		log.Printf("liabilities (force recomputed) after encountering an error after a call to UpdateWithOps\n")
		t.sdex.IEIF().RecomputeAndLogCachedLiabilities(t.assetBase, t.assetQuote)
		t.deleteAllOffers(false)
		return plugins.UpdateLoopResult{
			Success:            false,
			NumPruneOps:        numPruneOps,
			NumUpdateOpsDelete: numUpdateOpsDelete,
			NumUpdateOpsUpdate: numUpdateOpsUpdate,
			NumUpdateOpsCreate: numUpdateOpsCreate,
		}
	}

	msos := api.ConvertTM2MSO(opsOld)
	numUpdateOpsDelete, numUpdateOpsUpdate, numUpdateOpsCreate, e = countOfferChangeTypes(msos)
	if e != nil {
		log.Println(e)
		t.deleteAllOffers(false)
		return plugins.UpdateLoopResult{
			Success:            false,
			NumPruneOps:        numPruneOps,
			NumUpdateOpsDelete: numUpdateOpsDelete,
			NumUpdateOpsUpdate: numUpdateOpsUpdate,
			NumUpdateOpsCreate: numUpdateOpsCreate,
		}
	}

	ops := api.ConvertMSO2Ops(msos)
	for i, filter := range t.submitFilters {
		ops, e = filter.Apply(ops, t.sellingAOffers, t.buyingAOffers)
		if e != nil {
			log.Printf("error in filter index %d: %s\n", i, e)
			t.deleteAllOffers(false)
			return plugins.UpdateLoopResult{
				Success:            false,
				NumPruneOps:        numPruneOps,
				NumUpdateOpsDelete: numUpdateOpsDelete,
				NumUpdateOpsUpdate: numUpdateOpsUpdate,
				NumUpdateOpsCreate: numUpdateOpsCreate,
			}
		}
	}

	log.Printf("created %d operations to update existing offers\n", len(ops))
	if len(ops) > 0 {
		e = t.exchangeShim.SubmitOps(api.ConvertOperation2TM(ops), t.submitMode, func(hash string, e error) {
			// if there is an error we want it to count towards the delete cycles threshold, so run the check
			if e != nil {
				t.deleteAllOffers(true)
			}
		})
		if e != nil {
			log.Println(e)
			t.deleteAllOffers(false)
			return plugins.UpdateLoopResult{
				Success:            false,
				NumPruneOps:        numPruneOps,
				NumUpdateOpsDelete: numUpdateOpsDelete,
				NumUpdateOpsUpdate: numUpdateOpsUpdate,
				NumUpdateOpsCreate: numUpdateOpsCreate,
			}
		}
	}

	e = t.strategy.PostUpdate()
	if e != nil {
		log.Println(e)
		t.deleteAllOffers(false)
		return plugins.UpdateLoopResult{
			Success:            false,
			NumPruneOps:        numPruneOps,
			NumUpdateOpsDelete: numUpdateOpsDelete,
			NumUpdateOpsUpdate: numUpdateOpsUpdate,
			NumUpdateOpsCreate: numUpdateOpsCreate,
		}
	}

	// reset deleteCycles on every successful run
	t.deleteCycles = 0
	return plugins.UpdateLoopResult{
		Success:            true,
		NumPruneOps:        numPruneOps,
		NumUpdateOpsDelete: numUpdateOpsDelete,
		NumUpdateOpsUpdate: numUpdateOpsUpdate,
		NumUpdateOpsCreate: numUpdateOpsCreate,
	}
}

func (t *Trader) getBalances() (*api.Balance /*baseBalance*/, *api.Balance /*quoteBalance*/, error) {
	baseBalance, e := t.exchangeShim.GetBalanceHack(t.assetBase)
	if e != nil {
		return nil, nil, fmt.Errorf("error fetching base balance: %s", e)
	}

	quoteBalance, e := t.exchangeShim.GetBalanceHack(t.assetQuote)
	if e != nil {
		return nil, nil, fmt.Errorf("error fetching quote balance: %s", e)
	}

	return baseBalance, quoteBalance, nil
}

func (t *Trader) setBalances(baseBalance *api.Balance, quoteBalance *api.Balance) {
	t.maxAssetA = baseBalance.Balance
	t.maxAssetB = quoteBalance.Balance
	t.trustAssetA = baseBalance.Trust
	t.trustAssetB = quoteBalance.Trust

	trustAString := "math.MaxFloat64"
	if t.assetBase.Type != utils.Native {
		trustAString = fmt.Sprintf("%.8f", t.trustAssetA)
	}
	trustBString := "math.MaxFloat64"
	if t.assetQuote.Type != utils.Native {
		trustBString = fmt.Sprintf("%.8f", t.trustAssetB)
	}

	log.Printf(" (base) assetA=%s, maxA=%.8f, trustA=%s\n", utils.Asset2String(t.assetBase), t.maxAssetA, trustAString)
	log.Printf("(quote) assetB=%s, maxB=%.8f, trustB=%s\n", utils.Asset2String(t.assetQuote), t.maxAssetB, trustBString)

	if t.valueBaseFeed != nil && t.valueQuoteFeed != nil {
		baseUsdPrice, e := t.valueBaseFeed.GetPrice()
		if e != nil {
			log.Println(e)
			return
		}
		quoteUsdPrice, e := t.valueQuoteFeed.GetPrice()
		if e != nil {
			log.Println(e)
			return
		}

		totalUSDValue := (t.maxAssetA * baseUsdPrice) + (t.maxAssetB * quoteUsdPrice)
		log.Printf("value of total assets in terms of USD=%.12f, base=%.12f, quote=%.12f, baseUSDPrice=%.12f, quoteUSDPrice=%.12f, baseQuotePrice=%.12f\n",
			totalUSDValue,
			totalUSDValue/baseUsdPrice,
			totalUSDValue/quoteUsdPrice,
			baseUsdPrice,
			quoteUsdPrice,
			baseUsdPrice/quoteUsdPrice,
		)
	}
}

func (t *Trader) getExistingOffers() ([]hProtocol.Offer /*sellingAOffers*/, []hProtocol.Offer /*buyingAOffers*/, error) {
	offers, e := t.exchangeShim.LoadOffersHack()
	if e != nil {
		return nil, nil, fmt.Errorf("unable to load existing offers: %s", e)
	}
	sellingAOffers, buyingAOffers := utils.FilterOffers(offers, t.assetBase, t.assetQuote)

	sort.Sort(utils.ByPrice(buyingAOffers))
	sort.Sort(utils.ByPrice(sellingAOffers)) // don't reverse since prices are inverse
	return sellingAOffers, buyingAOffers, nil
}

func (t *Trader) setExistingOffers(sellingAOffers []hProtocol.Offer, buyingAOffers []hProtocol.Offer) {
	t.sellingAOffers, t.buyingAOffers = sellingAOffers, buyingAOffers
}

func countOfferChangeTypes(offers []*txnbuild.ManageSellOffer) (int /*numDelete*/, int /*numUpdate*/, int /*numCreate*/, error) {
	numDelete, numUpdate, numCreate := 0, 0, 0
	for i, o := range offers {
		if o == nil {
			return 0, 0, 0, fmt.Errorf("offer at index %d was not of expected type ManageSellOffer (actual type = %T): %+v", i, o, o)
		}

		opAmount, e := strconv.ParseFloat(o.Amount, 64)
		if e != nil {
			return 0, 0, 0, fmt.Errorf("invalid operation amount (%s) could not be parsed as float for operation at index %d: %v", o.Amount, i, o)
		}

		// 0 amount represents deletion
		// 0 offer id represents creating a new offer
		// anything else represents updating an extiing offer
		if opAmount == 0 {
			numDelete++
		} else if o.OfferID == 0 {
			numCreate++
		} else {
			numUpdate++
		}
	}

	return numDelete, numUpdate, numCreate, nil
}
