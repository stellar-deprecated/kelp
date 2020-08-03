package trader

import (
	"fmt"
	"log"
	"math"
	"sort"
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
	api                   *horizonclient.Client
	ieif                  *plugins.IEIF
	assetBase             hProtocol.Asset
	assetQuote            hProtocol.Asset
	valueBaseFeed         api.PriceFeed
	valueQuoteFeed        api.PriceFeed
	tradingAccount        string
	sdex                  *plugins.SDEX
	exchangeShim          api.ExchangeShim
	strategy              api.Strategy // the instance of this bot is bound to this strategy
	timeController        api.TimeController
	deleteCyclesThreshold int64
	submitMode            api.SubmitMode
	submitFilters         []plugins.SubmitFilter
	threadTracker         *multithreading.ThreadTracker
	fixedIterations       *uint64
	dataKey               *model.BotKey
	alert                 api.Alert

	// initialized runtime vars
	deleteCycles        int64
	stateSyncMaxRetries int

	// uninitialized runtime vars
	maxAssetA          float64
	maxAssetB          float64
	trustAssetA        float64
	trustAssetB        float64
	buyingAOffers      []hProtocol.Offer // quoted A/B
	sellingAOffers     []hProtocol.Offer // quoted B/A
	triggerFillTracker func() ([]model.Trade, error)
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
	deleteCyclesThreshold int64,
	submitMode api.SubmitMode,
	submitFilters []plugins.SubmitFilter,
	threadTracker *multithreading.ThreadTracker,
	fixedIterations *uint64,
	dataKey *model.BotKey,
	alert api.Alert,
) *Trader {
	return &Trader{
		api:                   api,
		ieif:                  ieif,
		assetBase:             assetBase,
		assetQuote:            assetQuote,
		valueBaseFeed:         valueBaseFeed,
		valueQuoteFeed:        valueQuoteFeed,
		tradingAccount:        tradingAccount,
		sdex:                  sdex,
		exchangeShim:          exchangeShim,
		strategy:              strategy,
		timeController:        timeController,
		deleteCyclesThreshold: deleteCyclesThreshold,
		submitMode:            submitMode,
		submitFilters:         submitFilters,
		threadTracker:         threadTracker,
		fixedIterations:       fixedIterations,
		dataKey:               dataKey,
		alert:                 alert,
		// initialized runtime vars
		deleteCycles:        0,
		stateSyncMaxRetries: 3,
	}
}

// SetTriggerFillTracker sets the fill tracker on this bot
// LOH-3 - triggerFillTracker should be injected into the bot instead of being set after creation
func (t *Trader) SetTriggerFillTracker(triggerFillTracker func() ([]model.Trade, error)) error {
	if t.triggerFillTracker != nil {
		return fmt.Errorf("(programmer error?) triggerFillTracker is already set, cannot reset")
	}

	t.triggerFillTracker = triggerFillTracker
	return nil
}

// Start starts the bot with the injected strategy
func (t *Trader) Start() {
	log.Println("----------------------------------------------------------------------------------------------------")
	var lastUpdateTime time.Time

	for {
		currentUpdateTime := time.Now()
		if lastUpdateTime.IsZero() || t.timeController.ShouldUpdate(lastUpdateTime, currentUpdateTime) {
			success := t.update()
			if t.fixedIterations != nil && success {
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
			lastUpdateTime = currentUpdateTime
		}

		sleepTime := t.timeController.SleepTime(lastUpdateTime, currentUpdateTime)
		log.Printf("sleeping for %s...\n", sleepTime)
		time.Sleep(sleepTime)
	}
}

// deletes all offers for the bot (not all offers on the account)
func (t *Trader) deleteAllOffers() {
	if t.deleteCyclesThreshold < 0 {
		log.Printf("not deleting any offers because deleteCyclesThreshold is negative\n")
		return
	}

	t.deleteCycles++
	if t.deleteCycles <= t.deleteCyclesThreshold {
		log.Printf("not deleting any offers, deleteCycles (=%d) needs to exceed deleteCyclesThreshold (=%d)\n", t.deleteCycles, t.deleteCyclesThreshold)
		return
	}

	log.Printf("deleting all offers, num. continuous update cycles with errors (including this one): %d; (deleteCyclesThreshold to be exceeded=%d)\n", t.deleteCycles, t.deleteCyclesThreshold)
	dOps := []txnbuild.Operation{}
	dOps = append(dOps, t.sdex.DeleteAllOffers(t.sellingAOffers)...)
	t.sellingAOffers = []hProtocol.Offer{}
	dOps = append(dOps, t.sdex.DeleteAllOffers(t.buyingAOffers)...)
	t.buyingAOffers = []hProtocol.Offer{}

	log.Printf("created %d operations to delete offers\n", len(dOps))
	if len(dOps) > 0 {
		// to delete offers the submitMode doesn't matter, so use api.SubmitModeBoth as the default
		e := t.exchangeShim.SubmitOps(api.ConvertOperation2TM(dOps), api.SubmitModeBoth, nil)
		if e != nil {
			log.Println(e)
			return
		}
	}
}

func (t *Trader) synchronizeFetchBalancesOffersTrades() error {
	for i := 0; i < t.stateSyncMaxRetries+1; i++ {
		baseBalance1, quoteBalance1, e := t.getBalances()
		if e != nil {
			return fmt.Errorf("unable to get balances1, iteration %d of %d attempts (1-indexed): %s", i+1, t.stateSyncMaxRetries+1, e)
		}
		sellingAOffers1, buyingAOffers1, e := t.getExistingOffers()
		if e != nil {
			return fmt.Errorf("unable to get offers1, iteration %d of %d attempts (1-indexed): %s", i+1, t.stateSyncMaxRetries+1, e)
		}
		// f.triggerFillTracker should never be nil
		// we pivot balances and offers around trades to ensure nothing changed w.r.t. trades
		trades, e := t.triggerFillTracker()
		if e != nil {
			return fmt.Errorf("unable to get trades, iteration %d of %d attempts (1-indexed): %s", i+1, t.stateSyncMaxRetries+1, e)
		}

		// run it again once we have fetched trades so we can compare that nothing changed and the data is in sync
		baseBalance2, quoteBalance2, e := t.getBalances()
		if e != nil {
			return fmt.Errorf("unable to get balances2, iteration %d of %d attempts (1-indexed): %s", i+1, t.stateSyncMaxRetries+1, e)
		}
		sellingAOffers2, buyingAOffers2, e := t.getExistingOffers()
		if e != nil {
			return fmt.Errorf("unable to get offers2, iteration %d of %d attempts (1-indexed): %s", i+1, t.stateSyncMaxRetries+1, e)
		}

		hasNewTrades := len(trades) > 0
		baseBalanceSame := baseBalance1.Balance == baseBalance2.Balance
		quoteBalanceSame := quoteBalance1.Balance == quoteBalance2.Balance
		sellOffersSame := len(sellingAOffers1) == len(sellingAOffers2)
		buyOffersSame := len(buyingAOffers1) == len(buyingAOffers2)
		if !hasNewTrades && baseBalanceSame && quoteBalanceSame && sellOffersSame && buyOffersSame {
			t.setBalances(baseBalance1, quoteBalance1)
			t.setExistingOffers(sellingAOffers1, buyingAOffers1)
			// this is the only success case
			return nil
		}

		log.Printf("something changed when fetching trades (!hasNewTrades=%v), balances (baseBalanceSame=%v, quoteBalanceSame=%v), and offers (sellOffersSame=%v, buyOffersSame=%v) [all should be true for success] so could not synchronize data in attempt %d of %d (1-indexed), trying again...",
			!hasNewTrades, baseBalanceSame, quoteBalanceSame, sellOffersSame, buyOffersSame, i+1, t.stateSyncMaxRetries+1)
	}
	return fmt.Errorf("exhausted all %d attempts at synchronizing data when fetching trades, balances, and offers but all attempts failed", t.stateSyncMaxRetries+1)
}

// time to update the order book and possibly readjust the offers
// returns true if the update was successful, otherwise false
func (t *Trader) update() bool {
	e := t.synchronizeFetchBalancesOffersTrades()
	if e != nil {
		log.Println(e)
		t.deleteAllOffers()
		return false
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
		t.deleteAllOffers()
		return false
	}

	// strategy has a chance to set any state it needs
	e = t.strategy.PreUpdate(t.maxAssetA, t.maxAssetB, t.trustAssetA, t.trustAssetB)
	if e != nil {
		log.Println(e)
		t.deleteAllOffers()
		return false
	}

	// delete excess offers
	var pruneOps []build.TransactionMutator
	pruneOps, t.buyingAOffers, t.sellingAOffers = t.strategy.PruneExistingOffers(t.buyingAOffers, t.sellingAOffers)
	log.Printf("created %d operations to prune excess offers\n", len(pruneOps))
	if len(pruneOps) > 0 {
		// to prune/delete offers the submitMode doesn't matter, so use api.SubmitModeBoth as the default
		e = t.exchangeShim.SubmitOps(pruneOps, api.SubmitModeBoth, nil)
		if e != nil {
			log.Println(e)
			t.deleteAllOffers()
			return false
		}
	}

	// TODO 2 streamline the request data instead of caching
	// reset cache of balances for this update cycle to reduce redundant requests to calculate asset balances
	t.sdex.IEIF().ResetCachedBalances()
	// reset and recompute cached liabilities for this update cycle
	e = t.sdex.IEIF().ResetCachedLiabilities(t.assetBase, t.assetQuote)
	log.Printf("liabilities after resetting\n")
	t.sdex.IEIF().LogAllLiabilities(t.assetBase, t.assetQuote)
	if e != nil {
		log.Println(e)
		t.deleteAllOffers()
		return false
	}

	opsOld, e := t.strategy.UpdateWithOps(t.buyingAOffers, t.sellingAOffers)
	log.Printf("liabilities at the end of a call to UpdateWithOps\n")
	t.sdex.IEIF().LogAllLiabilities(t.assetBase, t.assetQuote)
	if e != nil {
		log.Println(e)
		log.Printf("liabilities (force recomputed) after encountering an error after a call to UpdateWithOps\n")
		t.sdex.IEIF().RecomputeAndLogCachedLiabilities(t.assetBase, t.assetQuote)
		t.deleteAllOffers()
		return false
	}

	ops := api.ConvertTM2Operation(opsOld)
	for i, filter := range t.submitFilters {
		ops, e = filter.Apply(ops, t.sellingAOffers, t.buyingAOffers)
		if e != nil {
			log.Printf("error in filter index %d: %s\n", i, e)
			t.deleteAllOffers()
			return false
		}
	}

	log.Printf("created %d operations to update existing offers\n", len(ops))
	if len(ops) > 0 {
		e = t.exchangeShim.SubmitOps(api.ConvertOperation2TM(ops), t.submitMode, nil)
		if e != nil {
			log.Println(e)
			t.deleteAllOffers()
			return false
		}
	}

	e = t.strategy.PostUpdate()
	if e != nil {
		log.Println(e)
		t.deleteAllOffers()
		return false
	}

	// reset deleteCycles on every successful run
	t.deleteCycles = 0
	return true
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
