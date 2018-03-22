package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/lightyeario/kelp/alfonso/strategy"
	kelp "github.com/lightyeario/kelp/support"
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
) *Bot {
	return &Bot{
		api:                 api,
		assetA:              assetA,
		assetB:              assetB,
		tradingAccount:      tradingAccount,
		txButler:            txButler,
		strat:               strat,
		tickIntervalSeconds: tickIntervalSeconds,
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
// TODO do in one transaction
func (b *Bot) deleteAllOffers() {
	for _, offer := range b.sellingAOffers {
		b.txButler.DeleteOffer(offer)
	}
	b.sellingAOffers = []horizon.Offer{}

	for _, offer := range b.buyingAOffers {
		b.txButler.DeleteOffer(offer)
	}
	b.buyingAOffers = []horizon.Offer{}
}

// time to update the order book and possibly readjust your offers
// TODO make sure we aren't crossing existing orders when we place these
func (b *Bot) update() {
	b.load()
	b.loadExistingOffers()
	// must delete excess offers
	b.buyingAOffers, b.sellingAOffers = b.strat.PruneExistingOffers(b.buyingAOffers, b.sellingAOffers)

	e := b.strat.PreUpdate(b.maxAssetA, b.maxAssetB, b.buyingAOffers, b.sellingAOffers)
	if e != nil {
		log.Info(e)
		b.deleteAllOffers()
		return
	}

	// reset cached xlm exposure here so we only compute it once per update
	// TODO 2 - calculate this here and pass it in
	b.txButler.ResetCachedXlmExposure()
	ops, e := b.strat.UpdateWithOps(b.buyingAOffers, b.sellingAOffers)
	if e != nil {
		log.Info(e)
		b.deleteAllOffers()
		return
	}
	if len(ops) > 0 {
		b.txButler.SubmitOps(ops)
	}

	e = b.strat.PostUpdate()
	if e != nil {
		log.Info(e)
		b.deleteAllOffers()
		return
	}
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
		log.Info(e)
		return
	}

	b.sellingAOffers = []horizon.Offer{}
	b.buyingAOffers = []horizon.Offer{}
	for _, offer := range offers {
		if offer.Selling == b.assetA {
			if offer.Buying == b.assetB {
				//log.Info("Found selling offer p:", offer.Price, " a:", offer.Amount)
				b.sellingAOffers = append(b.sellingAOffers, offer)
			}
		} else if offer.Selling == b.assetB {
			if offer.Buying == b.assetA {
				//log.Info("Found buying offer p:", offer.Price, " a:", offer.Amount)
				b.buyingAOffers = append(b.buyingAOffers, offer)
			}
		}
	}

	sort.Sort(kelp.ByPrice(b.buyingAOffers))
	sort.Sort(kelp.ByPrice(b.sellingAOffers)) // don't need to reverse since the prices are inverse
}
