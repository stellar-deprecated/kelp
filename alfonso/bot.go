package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/lightyeario/kelp/alfonso/strategy"
	kelp "github.com/lightyeario/kelp/support"
	"github.com/lightyeario/kelp/support/datamodel"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/log"
)

// Bot represents a market making bot, which contains it's strategy
// the Bot is meant to contain all the non-strategy specific logic
type Bot struct {
	// TODO make all pointers
	api                 horizon.Client
	assetA              horizon.Asset // TODO call this base and quote
	assetB              horizon.Asset
	tradingAccount      string
	txButler            *kelp.TxButler
	strat               strategy.Strategy // the instance of this bot is bound to this strategy
	tickIntervalSeconds int32
	dataKey             *datamodel.BotKey
	writeUniqueKey      bool

	// uninitialized
	maxAssetA      float64
	maxAssetB      float64
	buyingAOffers  []horizon.Offer // quoted A/B
	sellingAOffers []horizon.Offer // quoted B/A
}

// MakeBot is the factory method for the Bot struct
func MakeBot(
	api horizon.Client,
	assetA horizon.Asset,
	assetB horizon.Asset,
	tradingAccount string,
	txButler *kelp.TxButler,
	strat strategy.Strategy,
	tickIntervalSeconds int32,
	dataKey *datamodel.BotKey,
	writeUniqueKey bool,
) *Bot {
	return &Bot{
		api:                 api,
		assetA:              assetA,
		assetB:              assetB,
		tradingAccount:      tradingAccount,
		txButler:            txButler,
		strat:               strat,
		tickIntervalSeconds: tickIntervalSeconds,
		dataKey:             dataKey,
		writeUniqueKey:      writeUniqueKey,
	}
}

// Start starts the bot with the injected strategy
func (b *Bot) Start() {
	for {
		b.update()
		log.Info(fmt.Sprintf("sleeping for %d seconds...", b.tickIntervalSeconds))
		time.Sleep(time.Duration(b.tickIntervalSeconds) * time.Second)
	}
}

// deletes all offers for this bot (not all offers on the account)
func (b *Bot) deleteAllOffers() {
	dOps := []build.TransactionMutator{}

	dOps = append(dOps, b.txButler.DeleteAllOffers(b.sellingAOffers)...)
	b.sellingAOffers = []horizon.Offer{}
	dOps = append(dOps, b.txButler.DeleteAllOffers(b.buyingAOffers)...)
	b.buyingAOffers = []horizon.Offer{}

	log.Info(fmt.Sprintf("deleting %d offers", len(dOps)))
	if len(dOps) > 0 {
		e := b.txButler.SubmitOps(dOps)
		if e != nil {
			log.Warn(e)
			return
		}
	}
}

// time to update the order book and possibly readjust your offers
// TODO make sure we aren't crossing existing orders when we place these
func (b *Bot) update() {
	var e error
	b.load()
	b.loadExistingOffers()

	// strategy has a chance to set any state it needs
	e = b.strat.PreUpdate(b.maxAssetA, b.maxAssetB)
	if e != nil {
		log.Warn(e)
		b.deleteAllOffers()
		return
	}

	// delete excess offers
	var pruneOps []build.TransactionMutator
	pruneOps, b.buyingAOffers, b.sellingAOffers = b.strat.PruneExistingOffers(b.buyingAOffers, b.sellingAOffers)
	if len(pruneOps) > 0 {
		e = b.txButler.SubmitOps(pruneOps)
		if e != nil {
			log.Warn(e)
			b.deleteAllOffers()
			return
		}
	}

	// reset cached xlm exposure here so we only compute it once per update
	// TODO 2 - calculate this here and pass it in
	b.txButler.ResetCachedXlmExposure()
	ops, e := b.strat.UpdateWithOps(b.buyingAOffers, b.sellingAOffers)
	if e != nil {
		log.Warn(e)
		b.deleteAllOffers()
		return
	}

	// append manageDataOps to update timestamp along with the update ops
	mdOps, e := b.makeManageDataOps(time.Now())
	if e != nil {
		log.Warn(e)
		b.deleteAllOffers()
		return
	}
	ops = append(ops, mdOps...)
	if len(ops) > 0 {
		e = b.txButler.SubmitOps(ops)
		if e != nil {
			log.Warn(e)
			b.deleteAllOffers()
			return
		}
		// TODO 3 - should verify that the async submission actually succeeded before setting to false
		b.writeUniqueKey = false
	}

	e = b.strat.PostUpdate()
	if e != nil {
		log.Warn(e)
		b.deleteAllOffers()
		return
	}
}

// makeManageDataOps writes data to the account to track when this bot successfully updated its orderbook
// looks like this: hash=sortedAsset1/sortedAsset2/timeMillis
func (b *Bot) makeManageDataOps(t time.Time) ([]build.TransactionMutator, error) {
	ops := []build.TransactionMutator{}
	tradingSourceAccount := build.SourceAccount{AddressOrSeed: b.tradingAccount}

	// always write timestamp
	millis := t.UnixNano() / 1000000
	millisStr := fmt.Sprintf("%d", millis)
	millisData := []byte(millisStr)
	millisOp := build.SetData(b.dataKey.FullKey(0), millisData, tradingSourceAccount)
	ops = append(ops, &millisOp)

	if b.writeUniqueKey {
		ops = append(ops, build.SetData(b.dataKey.FullKey(1), []byte(b.dataKey.AssetBaseCode), tradingSourceAccount))
		ops = append(ops, build.SetData(b.dataKey.FullKey(2), []byte(b.dataKey.AssetBaseIssuer), tradingSourceAccount))
		ops = append(ops, build.SetData(b.dataKey.FullKey(3), []byte(b.dataKey.AssetQuoteCode), tradingSourceAccount))
		ops = append(ops, build.SetData(b.dataKey.FullKey(4), []byte(b.dataKey.AssetQuoteIssuer), tradingSourceAccount))
	}
	log.Info("setting data with key = " + b.dataKey.Key() + " | dataKeyPrefix = " + b.dataKey.HashWithPrefix() + " | millis = " + millisStr)

	return ops, nil
}

func (b *Bot) load() {
	// load the maximum amounts we can offer for each asset
	account, e := b.api.LoadAccount(b.tradingAccount)
	if e != nil {
		log.Info(e)
		return
	}

	var maxA float64
	var maxB float64
	for _, balance := range account.Balances {
		if balance.Asset == b.assetA {
			maxA = kelp.AmountStringAsFloat(balance.Balance)
			log.Info("maxA:", maxA)
		} else if balance.Asset == b.assetB {
			maxB = kelp.AmountStringAsFloat(balance.Balance)
			log.Info("maxB:", maxB)
		}
	}
	b.maxAssetA = maxA
	b.maxAssetB = maxB
}

// get complete list of offers
// look at the offers for this pair
// delete any extra ones
func (b *Bot) loadExistingOffers() {
	// TODO 2 pass in reference to horizon.Client
	offers, e := kelp.LoadAllOffers(b.tradingAccount, b.api)
	if e != nil {
		log.Warn(e)
		return
	}
	b.sellingAOffers, b.buyingAOffers = kelp.FilterOffers(offers, b.assetA, b.assetB)

	sort.Sort(kelp.ByPrice(b.buyingAOffers))
	sort.Sort(kelp.ByPrice(b.sellingAOffers)) // don't need to reverse since the prices are inverse
}
