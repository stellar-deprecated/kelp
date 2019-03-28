package trader

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/nikhilsaraf/go-tools/multithreading"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/utils"
)

const maxLumenTrust float64 = math.MaxFloat64

// Trader represents a market making bot, which is composed of various parts include the strategy and various APIs.
type Trader struct {
	api                   *horizon.Client
	ieif                  *plugins.IEIF
	assetBase             horizon.Asset
	assetQuote            horizon.Asset
	tradingAccount        string
	sdex                  *plugins.SDEX
	exchangeShim          api.ExchangeShim
	strategy              api.Strategy // the instance of this bot is bound to this strategy
	timeController        api.TimeController
	deleteCyclesThreshold int64
	submitFilters         []plugins.SubmitFilter
	threadTracker         *multithreading.ThreadTracker
	fixedIterations       *uint64
	dataKey               *model.BotKey
	alert                 api.Alert

	// initialized runtime vars
	deleteCycles int64

	// uninitialized runtime vars
	maxAssetA      float64
	maxAssetB      float64
	trustAssetA    float64
	trustAssetB    float64
	buyingAOffers  []horizon.Offer // quoted A/B
	sellingAOffers []horizon.Offer // quoted B/A
}

// MakeBot is the factory method for the Trader struct
func MakeBot(
	api *horizon.Client,
	ieif *plugins.IEIF,
	assetBase horizon.Asset,
	assetQuote horizon.Asset,
	tradingPair *model.TradingPair,
	minBaseVolume *float64,
	tradingAccount string,
	sdex *plugins.SDEX,
	exchangeShim api.ExchangeShim,
	strategy api.Strategy,
	timeController api.TimeController,
	deleteCyclesThreshold int64,
	submitMode api.SubmitMode,
	threadTracker *multithreading.ThreadTracker,
	fixedIterations *uint64,
	dataKey *model.BotKey,
	alert api.Alert,
) *Trader {
	submitFilters := []plugins.SubmitFilter{}

	oc := exchangeShim.GetOrderConstraints(tradingPair)
	if minBaseVolume != nil {
		oc.MinBaseVolume = *model.NumberFromFloat(*minBaseVolume, oc.VolumePrecision)
	}
	orderConstraintsFilter := plugins.MakeFilterOrderConstraints(oc, assetBase, assetQuote)
	submitFilters = append(submitFilters, orderConstraintsFilter)

	sdexSubmitFilter := plugins.MakeFilterMakerMode(submitMode, exchangeShim, sdex, tradingPair)
	if sdexSubmitFilter != nil {
		submitFilters = append(submitFilters, sdexSubmitFilter)
	}

	return &Trader{
		api:                   api,
		ieif:                  ieif,
		assetBase:             assetBase,
		assetQuote:            assetQuote,
		tradingAccount:        tradingAccount,
		sdex:                  sdex,
		exchangeShim:          exchangeShim,
		strategy:              strategy,
		timeController:        timeController,
		deleteCyclesThreshold: deleteCyclesThreshold,
		submitFilters:         submitFilters,
		threadTracker:         threadTracker,
		fixedIterations:       fixedIterations,
		dataKey:               dataKey,
		alert:                 alert,
		// initialized runtime vars
		deleteCycles: 0,
	}
}

// Start starts the bot with the injected strategy
func (t *Trader) Start() {
	log.Println("----------------------------------------------------------------------------------------------------")
	var lastUpdateTime time.Time

	for {
		currentUpdateTime := time.Now()
		if lastUpdateTime.IsZero() || t.timeController.ShouldUpdate(lastUpdateTime, currentUpdateTime) {
			t.update()
			if t.fixedIterations != nil {
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
	dOps := []build.TransactionMutator{}
	dOps = append(dOps, t.sdex.DeleteAllOffers(t.sellingAOffers)...)
	t.sellingAOffers = []horizon.Offer{}
	dOps = append(dOps, t.sdex.DeleteAllOffers(t.buyingAOffers)...)
	t.buyingAOffers = []horizon.Offer{}

	log.Printf("created %d operations to delete offers\n", len(dOps))
	if len(dOps) > 0 {
		e := t.exchangeShim.SubmitOps(dOps, nil)
		if e != nil {
			log.Println(e)
			return
		}
	}
}

// time to update the order book and possibly readjust the offers
func (t *Trader) update() {
	var e error
	t.load()
	t.loadExistingOffers()

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
		return
	}

	// strategy has a chance to set any state it needs
	e = t.strategy.PreUpdate(t.maxAssetA, t.maxAssetB, t.trustAssetA, t.trustAssetB)
	if e != nil {
		log.Println(e)
		t.deleteAllOffers()
		return
	}

	// delete excess offers
	var pruneOps []build.TransactionMutator
	pruneOps, t.buyingAOffers, t.sellingAOffers = t.strategy.PruneExistingOffers(t.buyingAOffers, t.sellingAOffers)
	log.Printf("created %d operations to prune excess offers\n", len(pruneOps))
	if len(pruneOps) > 0 {
		e = t.exchangeShim.SubmitOps(pruneOps, nil)
		if e != nil {
			log.Println(e)
			t.deleteAllOffers()
			return
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
		return
	}

	ops, e := t.strategy.UpdateWithOps(t.buyingAOffers, t.sellingAOffers)
	log.Printf("liabilities at the end of a call to UpdateWithOps\n")
	t.sdex.IEIF().LogAllLiabilities(t.assetBase, t.assetQuote)
	if e != nil {
		log.Println(e)
		log.Printf("liabilities (force recomputed) after encountering an error after a call to UpdateWithOps\n")
		t.sdex.IEIF().RecomputeAndLogCachedLiabilities(t.assetBase, t.assetQuote)
		t.deleteAllOffers()
		return
	}

	for i, filter := range t.submitFilters {
		ops, e = filter.Apply(ops, t.sellingAOffers, t.buyingAOffers)
		if e != nil {
			log.Printf("error in filter index %d: %s\n", i, e)
			t.deleteAllOffers()
			return
		}
	}

	log.Printf("created %d operations to update existing offers\n", len(ops))
	if len(ops) > 0 {
		e = t.exchangeShim.SubmitOps(ops, nil)
		if e != nil {
			log.Println(e)
			t.deleteAllOffers()
			return
		}
	}

	e = t.strategy.PostUpdate()
	if e != nil {
		log.Println(e)
		t.deleteAllOffers()
		return
	}

	// reset deleteCycles on every successful run
	t.deleteCycles = 0
}

func (t *Trader) load() {
	// load the maximum amounts we can offer for each asset
	baseBalance, e := t.exchangeShim.GetBalanceHack(t.assetBase)
	if e != nil {
		log.Println(e)
		return
	}
	quoteBalance, e := t.exchangeShim.GetBalanceHack(t.assetQuote)
	if e != nil {
		log.Println(e)
		return
	}

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
}

func (t *Trader) loadExistingOffers() {
	offers, e := t.exchangeShim.LoadOffersHack()
	if e != nil {
		log.Println(e)
		return
	}
	t.sellingAOffers, t.buyingAOffers = utils.FilterOffers(offers, t.assetBase, t.assetQuote)

	sort.Sort(utils.ByPrice(t.buyingAOffers))
	sort.Sort(utils.ByPrice(t.sellingAOffers)) // don't need to reverse since the prices are inverse
}
