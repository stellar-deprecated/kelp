package trader

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/plugins"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/nikhilsaraf/go-tools/multithreading"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

const maxLumenTrust float64 = math.MaxFloat64

// Trader represents a market making bot, which is composed of various parts include the strategy and various APIs.
type Trader struct {
	api             *horizon.Client
	assetBase       horizon.Asset
	assetQuote      horizon.Asset
	tradingAccount  string
	sdex            *plugins.SDEX
	strat           api.Strategy // the instance of this bot is bound to this strategy
	timeController  api.TimeController
	threadTracker   *multithreading.ThreadTracker
	fixedIterations *uint64
	dataKey         *model.BotKey
	alert           api.Alert

	// uninitialized
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
	assetBase horizon.Asset,
	assetQuote horizon.Asset,
	tradingAccount string,
	sdex *plugins.SDEX,
	strat api.Strategy,
	timeController api.TimeController,
	threadTracker *multithreading.ThreadTracker,
	fixedIterations *uint64,
	dataKey *model.BotKey,
	alert api.Alert,
) *Trader {
	return &Trader{
		api:             api,
		assetBase:       assetBase,
		assetQuote:      assetQuote,
		tradingAccount:  tradingAccount,
		sdex:            sdex,
		strat:           strat,
		timeController:  timeController,
		threadTracker:   threadTracker,
		fixedIterations: fixedIterations,
		dataKey:         dataKey,
		alert:           alert,
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
	log.Printf("deleting all offers\n")
	dOps := []build.TransactionMutator{}

	dOps = append(dOps, t.sdex.DeleteAllOffers(t.sellingAOffers)...)
	t.sellingAOffers = []horizon.Offer{}
	dOps = append(dOps, t.sdex.DeleteAllOffers(t.buyingAOffers)...)
	t.buyingAOffers = []horizon.Offer{}

	log.Printf("created %d operations to delete offers\n", len(dOps))
	if len(dOps) > 0 {
		e := t.sdex.SubmitOps(dOps, nil)
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
	t.sdex.ResetCachedBalances()
	// reset and recompute cached liabilities for this update cycle
	e = t.sdex.ResetCachedLiabilities(t.assetBase, t.assetQuote)
	log.Printf("liabilities after resetting\n")
	t.sdex.LogAllLiabilities(t.assetBase, t.assetQuote)
	if e != nil {
		log.Println(e)
		t.deleteAllOffers()
		return
	}

	// strategy has a chance to set any state it needs
	e = t.strat.PreUpdate(t.maxAssetA, t.maxAssetB, t.trustAssetA, t.trustAssetB)
	if e != nil {
		log.Println(e)
		t.deleteAllOffers()
		return
	}

	// delete excess offers
	var pruneOps []build.TransactionMutator
	pruneOps, t.buyingAOffers, t.sellingAOffers = t.strat.PruneExistingOffers(t.buyingAOffers, t.sellingAOffers)
	log.Printf("created %d operations to prune excess offers\n", len(pruneOps))
	if len(pruneOps) > 0 {
		e = t.sdex.SubmitOps(pruneOps, nil)
		if e != nil {
			log.Println(e)
			t.deleteAllOffers()
			return
		}
	}

	// TODO 2 streamline the request data instead of caching
	// reset cache of balances for this update cycle to reduce redundant requests to calculate asset balances
	t.sdex.ResetCachedBalances()
	// reset and recompute cached liabilities for this update cycle
	e = t.sdex.ResetCachedLiabilities(t.assetBase, t.assetQuote)
	log.Printf("liabilities after resetting\n")
	t.sdex.LogAllLiabilities(t.assetBase, t.assetQuote)
	if e != nil {
		log.Println(e)
		t.deleteAllOffers()
		return
	}

	ops, e := t.strat.UpdateWithOps(t.buyingAOffers, t.sellingAOffers)
	log.Printf("liabilities at the end of a call to UpdateWithOps\n")
	t.sdex.LogAllLiabilities(t.assetBase, t.assetQuote)
	if e != nil {
		log.Println(e)
		log.Printf("liabilities (force recomputed) after encountering an error after a call to UpdateWithOps\n")
		t.sdex.RecomputeAndLogCachedLiabilities(t.assetBase, t.assetQuote)
		t.deleteAllOffers()
		return
	}

	log.Printf("created %d operations to update existing offers\n", len(ops))
	if len(ops) > 0 {
		e = t.sdex.SubmitOps(ops, nil)
		if e != nil {
			log.Println(e)
			t.deleteAllOffers()
			return
		}
	}

	e = t.strat.PostUpdate()
	if e != nil {
		log.Println(e)
		t.deleteAllOffers()
		return
	}
}

func (t *Trader) load() {
	// load the maximum amounts we can offer for each asset
	account, e := t.api.LoadAccount(t.tradingAccount)
	if e != nil {
		log.Println(e)
		return
	}

	var maxA float64
	var maxB float64
	var trustA float64
	var trustB float64
	var trustAString string
	var trustBString string
	for _, balance := range account.Balances {
		trust := maxLumenTrust
		trustString := "math.MaxFloat64"
		if balance.Asset.Type != utils.Native {
			trust = utils.AmountStringAsFloat(balance.Limit)
			trustString = fmt.Sprintf("%.7f", trust)
		}

		if utils.AssetsEqual(balance.Asset, t.assetBase) {
			maxA = utils.AmountStringAsFloat(balance.Balance)
			trustA = trust
			trustAString = trustString
		} else if utils.AssetsEqual(balance.Asset, t.assetQuote) {
			maxB = utils.AmountStringAsFloat(balance.Balance)
			trustB = trust
			trustBString = trustString
		}
	}
	t.maxAssetA = maxA
	t.maxAssetB = maxB
	t.trustAssetA = trustA
	t.trustAssetB = trustB

	log.Printf(" (base) assetA=%s, maxA=%.7f, trustA=%s\n", utils.Asset2String(t.assetBase), maxA, trustAString)
	log.Printf("(quote) assetB=%s, maxB=%.7f, trustB=%s\n", utils.Asset2String(t.assetQuote), maxB, trustBString)
}

func (t *Trader) loadExistingOffers() {
	offers, e := utils.LoadAllOffers(t.tradingAccount, t.api)
	if e != nil {
		log.Println(e)
		return
	}
	t.sellingAOffers, t.buyingAOffers = utils.FilterOffers(offers, t.assetBase, t.assetQuote)

	sort.Sort(utils.ByPrice(t.buyingAOffers))
	sort.Sort(utils.ByPrice(t.sellingAOffers)) // don't need to reverse since the prices are inverse
}
