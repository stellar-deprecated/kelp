package plugins

import (
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/lightyeario/kelp/support/utils"
	"github.com/nikhilsaraf/go-tools/multithreading"
	"github.com/pkg/errors"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

const baseReserve = 0.5
const baseFee = 0.0000100
const maxLumenTrust = math.MaxFloat64

// SDEX helps with building and submitting transactions to the Stellar network
type SDEX struct {
	API               *horizon.Client
	SourceAccount     string
	TradingAccount    string
	SourceSeed        string
	TradingSeed       string
	Network           build.Network
	threadTracker     *multithreading.ThreadTracker
	operationalBuffer float64
	simMode           bool

	// uninitialized
	seqNum       uint64
	reloadSeqNum bool
	// explicitly calculate liabilities here for now, we can switch over to using the values from Horizon once the protocol change has taken effect
	cachedLiabilities map[horizon.Asset]Liabilities

	// TODO 2 streamline requests instead of caching
	// cache balances to avoid redundant requests
	cachedBalances map[horizon.Asset]Balance
}

// Liabilities represents the "committed" units of an asset on both the buy and sell sides
type Liabilities struct {
	Buying  float64 // affects how much more can be bought
	Selling float64 // affects how much more can be sold
}

// Balance repesents an asset's balance response from the assetBalance method below
type Balance struct {
	Balance float64
	Trust   float64
	Reserve float64
}

// MakeSDEX is a factory method for SDEX
func MakeSDEX(
	api *horizon.Client,
	sourceSeed string,
	tradingSeed string,
	sourceAccount string,
	tradingAccount string,
	network build.Network,
	threadTracker *multithreading.ThreadTracker,
	operationalBuffer float64,
	simMode bool,
) *SDEX {
	sdex := &SDEX{
		API:               api,
		SourceSeed:        sourceSeed,
		TradingSeed:       tradingSeed,
		SourceAccount:     sourceAccount,
		TradingAccount:    tradingAccount,
		Network:           network,
		threadTracker:     threadTracker,
		operationalBuffer: operationalBuffer,
		simMode:           simMode,
	}

	log.Printf("Using network passphrase: %s\n", sdex.Network.Passphrase)

	if sdex.SourceAccount == "" {
		sdex.SourceAccount = sdex.TradingAccount
		sdex.SourceSeed = sdex.TradingSeed
		log.Println("No Source Account Set")
	}
	sdex.reloadSeqNum = true

	return sdex
}

func (sdex *SDEX) incrementSeqNum() {
	if sdex.reloadSeqNum {
		log.Println("reloading sequence number")
		seqNum, err := sdex.API.SequenceForAccount(sdex.SourceAccount)
		if err != nil {
			log.Printf("error getting seq num: %s\n", err)
			return
		}
		sdex.seqNum = uint64(seqNum)
		sdex.reloadSeqNum = false
	}
	sdex.seqNum++
}

// DeleteAllOffers is a helper that accumulates delete operations for the passed in offers
func (sdex *SDEX) DeleteAllOffers(offers []horizon.Offer) []build.TransactionMutator {
	ops := []build.TransactionMutator{}
	for _, offer := range offers {
		op := sdex.DeleteOffer(offer)
		ops = append(ops, &op)
	}
	return ops
}

// DeleteOffer returns the op that needs to be submitted to the network in order to delete the passed in offer
func (sdex *SDEX) DeleteOffer(offer horizon.Offer) build.ManageOfferBuilder {
	rate := build.Rate{
		Selling: utils.Asset2Asset(offer.Selling),
		Buying:  utils.Asset2Asset(offer.Buying),
		Price:   build.Price(offer.Price),
	}

	if sdex.SourceAccount == sdex.TradingAccount {
		return build.ManageOffer(false, build.Amount("0"), rate, build.OfferID(offer.ID))
	}
	return build.ManageOffer(false, build.Amount("0"), rate, build.OfferID(offer.ID), build.SourceAccount{AddressOrSeed: sdex.TradingAccount})
}

// ModifyBuyOffer modifies a buy offer
func (sdex *SDEX) ModifyBuyOffer(offer horizon.Offer, price float64, amount float64, incrementalNativeAmountRaw float64) (*build.ManageOfferBuilder, error) {
	return sdex.ModifySellOffer(offer, 1/price, amount*price, incrementalNativeAmountRaw)
}

// ModifySellOffer modifies a sell offer
func (sdex *SDEX) ModifySellOffer(offer horizon.Offer, price float64, amount float64, incrementalNativeAmountRaw float64) (*build.ManageOfferBuilder, error) {
	return sdex.createModifySellOffer(&offer, offer.Selling, offer.Buying, price, amount, incrementalNativeAmountRaw)
}

// CreateSellOffer creates a sell offer
func (sdex *SDEX) CreateSellOffer(base horizon.Asset, counter horizon.Asset, price float64, amount float64, incrementalNativeAmountRaw float64) (*build.ManageOfferBuilder, error) {
	return sdex.createModifySellOffer(nil, base, counter, price, amount, incrementalNativeAmountRaw)
}

// ParseOfferAmount is a convenience method to parse an offer amount
func (sdex *SDEX) ParseOfferAmount(amt string) (float64, error) {
	offerAmt, err := strconv.ParseFloat(amt, 64)
	if err != nil {
		log.Printf("error parsing offer amount: %s\n", err)
		return -1, err
	}
	return offerAmt, nil
}

func (sdex *SDEX) minReserve(subentries int32) float64 {
	return float64(2+subentries) * baseReserve
}

// ResetCachedBalances resets the cached balances map
func (sdex *SDEX) ResetCachedBalances() {
	sdex.cachedBalances = map[horizon.Asset]Balance{}
}

// assetBalance is a memoized version of _assetBalance
func (sdex *SDEX) assetBalance(asset horizon.Asset) (float64, float64, float64, error) {
	if v, ok := sdex.cachedBalances[asset]; ok {
		return v.Balance, v.Trust, v.Reserve, nil
	}

	b, t, r, e := sdex._assetBalance(asset)
	if e == nil {
		sdex.cachedBalances[asset] = Balance{
			Balance: b,
			Trust:   t,
			Reserve: r,
		}
	}

	return b, t, r, e
}

// assetBalance returns asset balance, asset trust limit, reserve balance (zero for non-XLM), error
func (sdex *SDEX) _assetBalance(asset horizon.Asset) (float64, float64, float64, error) {
	account, err := sdex.API.LoadAccount(sdex.TradingAccount)
	if err != nil {
		return -1, -1, -1, fmt.Errorf("error: unable to load account to fetch balance: %s", err)
	}

	for _, balance := range account.Balances {
		if utils.AssetsEqual(balance.Asset, asset) {
			b, e := strconv.ParseFloat(balance.Balance, 64)
			if e != nil {
				return -1, -1, -1, fmt.Errorf("error: cannot parse balance: %s", e)
			}
			if balance.Asset.Type == utils.Native {
				return b, maxLumenTrust, sdex.minReserve(account.SubentryCount) + sdex.operationalBuffer, e
			}

			t, e := strconv.ParseFloat(balance.Limit, 64)
			if e != nil {
				return -1, -1, -1, fmt.Errorf("error: cannot parse trust limit: %s", e)
			}
			return b, t, 0, nil
		}
	}
	return -1, -1, -1, errors.New("could not find a balance for the asset passed in")
}

// ComputeIncrementalNativeAmountRaw returns the native amount that will be added to liabilities because of fee and min-reserve additions
func (sdex *SDEX) ComputeIncrementalNativeAmountRaw(isNewOffer bool) float64 {
	incrementalNativeAmountRaw := 0.0
	if sdex.TradingAccount == sdex.SourceAccount {
		// at the minimum it will cost us a unit of base fee for this operation
		incrementalNativeAmountRaw += baseFee
	}
	if isNewOffer {
		// new offers will increase the min reserve
		incrementalNativeAmountRaw += baseReserve
	}
	return incrementalNativeAmountRaw
}

// createModifySellOffer is the main method that handles the logic of creating or modifying an offer, note that all offers are treated as sell offers in Stellar
func (sdex *SDEX) createModifySellOffer(offer *horizon.Offer, selling horizon.Asset, buying horizon.Asset, price float64, amount float64, incrementalNativeAmountRaw float64) (*build.ManageOfferBuilder, error) {
	if price <= 0 {
		return nil, fmt.Errorf("error: cannot create or modify offer, invalid price: %.7f", price)
	}
	if amount <= 0 {
		return nil, fmt.Errorf("error: cannot create or modify offer, invalid amount: %.7f", amount)
	}

	// check liability limits on the asset being sold
	incrementalSell := amount
	willOversell, e := sdex.willOversell(selling, amount)
	if e != nil {
		return nil, e
	}
	if willOversell {
		return nil, nil
	}

	// check trust limits on asset being bought
	incrementalBuy := price * amount
	willOverbuy, e := sdex.willOverbuy(buying, incrementalBuy)
	if e != nil {
		return nil, e
	}
	if willOverbuy {
		return nil, nil
	}

	// explicitly check that we will not oversell XLM because of fee and min reserves
	incrementalNativeAmountTotal := incrementalNativeAmountRaw
	if selling.Type == utils.Native {
		incrementalNativeAmountTotal += incrementalSell
	}
	willOversellNative, e := sdex.willOversellNative(incrementalNativeAmountTotal)
	if e != nil {
		return nil, e
	}
	if willOversellNative {
		return nil, nil
	}

	stringPrice := strconv.FormatFloat(price, 'f', int(utils.SdexPrecision), 64)
	rate := build.Rate{
		Selling: utils.Asset2Asset(selling),
		Buying:  utils.Asset2Asset(buying),
		Price:   build.Price(stringPrice),
	}

	mutators := []interface{}{
		rate,
		build.Amount(strconv.FormatFloat(amount, 'f', int(utils.SdexPrecision), 64)),
	}
	if offer != nil {
		mutators = append(mutators, build.OfferID(offer.ID))
	}
	if sdex.SourceAccount != sdex.TradingAccount {
		mutators = append(mutators, build.SourceAccount{AddressOrSeed: sdex.TradingAccount})
	}
	result := build.ManageOffer(false, mutators...)
	return &result, nil
}

// AddLiabilities updates the cached liabilities, units are in their respective assets
func (sdex *SDEX) AddLiabilities(selling horizon.Asset, buying horizon.Asset, incrementalSell float64, incrementalBuy float64, incrementalNativeAmountRaw float64) {
	sdex.cachedLiabilities[selling] = Liabilities{
		Selling: sdex.cachedLiabilities[selling].Selling + incrementalSell,
		Buying:  sdex.cachedLiabilities[selling].Buying,
	}
	sdex.cachedLiabilities[buying] = Liabilities{
		Selling: sdex.cachedLiabilities[buying].Selling,
		Buying:  sdex.cachedLiabilities[buying].Buying + incrementalBuy,
	}
	sdex.cachedLiabilities[utils.NativeAsset] = Liabilities{
		Selling: sdex.cachedLiabilities[utils.NativeAsset].Selling + incrementalNativeAmountRaw,
		Buying:  sdex.cachedLiabilities[utils.NativeAsset].Buying,
	}
}

// willOversellNative returns willOversellNative, error
func (sdex *SDEX) willOversellNative(incrementalNativeAmount float64) (bool, error) {
	nativeBal, _, minAccountBal, e := sdex.assetBalance(utils.NativeAsset)
	if e != nil {
		return false, e
	}
	nativeLiabilities, e := sdex.assetLiabilities(utils.NativeAsset)
	if e != nil {
		return false, e
	}

	willOversellNative := incrementalNativeAmount > (nativeBal - minAccountBal - nativeLiabilities.Selling)
	if willOversellNative {
		log.Printf("we will oversell the native asset after considering fee and min reserves, incrementalNativeAmount = %.7f, nativeBal = %.7f, minAccountBal = %.7f, nativeLiabilities.Selling = %.7f\n",
			incrementalNativeAmount, nativeBal, minAccountBal, nativeLiabilities.Selling)
	}
	return willOversellNative, nil
}

// willOversell returns willOversell, error
func (sdex *SDEX) willOversell(asset horizon.Asset, amountSelling float64) (bool, error) {
	bal, _, minAccountBal, e := sdex.assetBalance(asset)
	if e != nil {
		return false, e
	}
	liabilities, e := sdex.assetLiabilities(asset)
	if e != nil {
		return false, e
	}

	willOversell := amountSelling > (bal - minAccountBal - liabilities.Selling)
	if willOversell {
		log.Printf("we will oversell the asset '%s', amountSelling = %.7f, bal = %.7f, minAccountBal = %.7f, liabilities.Selling = %.7f\n",
			utils.Asset2String(asset), amountSelling, bal, minAccountBal, liabilities.Selling)
	}
	return willOversell, nil
}

// willOverbuy returns willOverbuy, error
func (sdex *SDEX) willOverbuy(asset horizon.Asset, amountBuying float64) (bool, error) {
	if asset.Type == utils.Native {
		// you can never overbuy the native asset
		return false, nil
	}

	_, trust, _, e := sdex.assetBalance(asset)
	if e != nil {
		return false, e
	}
	liabilities, e := sdex.assetLiabilities(asset)
	if e != nil {
		return false, e
	}

	willOverbuy := amountBuying > (trust - liabilities.Buying)
	return willOverbuy, nil
}

// SubmitOps submits the passed in operations to the network asynchronously in a single transaction
func (sdex *SDEX) SubmitOps(ops []build.TransactionMutator, asyncCallback func(hash string, e error)) error {
	sdex.incrementSeqNum()
	muts := []build.TransactionMutator{
		build.Sequence{Sequence: sdex.seqNum},
		sdex.Network,
		build.SourceAccount{AddressOrSeed: sdex.SourceAccount},
	}
	muts = append(muts, ops...)
	tx, e := build.Transaction(muts...)
	if e != nil {
		return errors.Wrap(e, "SubmitOps error: ")
	}

	// convert to xdr string
	txeB64, e := sdex.sign(tx)
	if e != nil {
		return e
	}
	log.Printf("tx XDR: %s\n", txeB64)

	// submit
	if !sdex.simMode {
		log.Println("submitting tx XDR to network (async)")
		sdex.threadTracker.TriggerGoroutine(func() {
			sdex.submit(txeB64, asyncCallback)
		})
	} else {
		log.Println("not submitting tx XDR to network in simulation mode, calling asyncCallback with empty hash value")
		sdex.invokeAsyncCallback(asyncCallback, "", nil)
	}
	return nil
}

// CreateBuyOffer creates a buy offer
func (sdex *SDEX) CreateBuyOffer(base horizon.Asset, counter horizon.Asset, price float64, amount float64, incrementalNativeAmountRaw float64) (*build.ManageOfferBuilder, error) {
	return sdex.CreateSellOffer(counter, base, 1/price, amount*price, incrementalNativeAmountRaw)
}

func (sdex *SDEX) sign(tx *build.TransactionBuilder) (string, error) {
	var txe build.TransactionEnvelopeBuilder
	var e error

	if sdex.SourceSeed != sdex.TradingSeed {
		txe, e = tx.Sign(sdex.SourceSeed, sdex.TradingSeed)
	} else {
		txe, e = tx.Sign(sdex.SourceSeed)
	}
	if e != nil {
		return "", e
	}

	return txe.Base64()
}

func (sdex *SDEX) submit(txeB64 string, asyncCallback func(hash string, e error)) {
	resp, err := sdex.API.SubmitTransaction(txeB64)
	if err != nil {
		if herr, ok := errors.Cause(err).(*horizon.Error); ok {
			var rcs *horizon.TransactionResultCodes
			rcs, err = herr.ResultCodes()
			if err != nil {
				log.Printf("(async) error: no result codes from horizon: %s\n", err)
				sdex.invokeAsyncCallback(asyncCallback, "", err)
				return
			}
			if rcs.TransactionCode == "tx_bad_seq" {
				log.Println("(async) error: tx_bad_seq, setting flag to reload seq number")
				sdex.reloadSeqNum = true
			}
			log.Println("(async) error: result code details: tx code =", rcs.TransactionCode, ", opcodes =", rcs.OperationCodes)
		} else {
			log.Printf("(async) error: tx failed for unknown reason, error message: %s\n", err)
		}
		sdex.invokeAsyncCallback(asyncCallback, "", err)
		return
	}

	log.Printf("(async) tx confirmation hash: %s\n", resp.Hash)
	sdex.invokeAsyncCallback(asyncCallback, resp.Hash, nil)
}

func (sdex *SDEX) invokeAsyncCallback(asyncCallback func(hash string, e error), hash string, e error) {
	if asyncCallback == nil {
		return
	}

	sdex.threadTracker.TriggerGoroutine(func() {
		asyncCallback(hash, e)
	})
}

func (sdex *SDEX) logLiabilities(asset horizon.Asset, assetStr string) {
	l, e := sdex.assetLiabilities(asset)
	if e != nil {
		log.Printf("could not fetch liability for asset '%s', error = %s\n", assetStr, e)
		return
	}

	bal, trust, minAccountBal, e := sdex.assetBalance(asset)
	if e != nil {
		log.Printf("cannot fetch balance for asset '%s', error = %s\n", assetStr, e)
		return
	}

	trustString := "math.MaxFloat64"
	if trust != maxLumenTrust {
		trustString = fmt.Sprintf("%.7f", trust)
	}
	log.Printf("asset=%s, balance=%.7f, trust=%s, minAccountBal=%.7f, buyingLiabilities=%.7f, sellingLiabilities=%.7f\n",
		assetStr, bal, trustString, minAccountBal, l.Buying, l.Selling)
}

// LogAllLiabilities logs the liabilities for the two assets along with the native asset
func (sdex *SDEX) LogAllLiabilities(assetBase horizon.Asset, assetQuote horizon.Asset) {
	sdex.logLiabilities(assetBase, "base  ")
	sdex.logLiabilities(assetQuote, "quote ")

	if assetBase != utils.NativeAsset && assetQuote != utils.NativeAsset {
		sdex.logLiabilities(utils.NativeAsset, "native")
	}
}

// RecomputeAndLogCachedLiabilities clears the cached liabilities and recomputes from the network before logging
func (sdex *SDEX) RecomputeAndLogCachedLiabilities(assetBase horizon.Asset, assetQuote horizon.Asset) {
	sdex.cachedLiabilities = map[horizon.Asset]Liabilities{}
	// reset cached balances too so we fetch fresh balances
	sdex.ResetCachedBalances()
	sdex.LogAllLiabilities(assetBase, assetQuote)
}

// ResetCachedLiabilities resets the cache to include only the two assets passed in
func (sdex *SDEX) ResetCachedLiabilities(assetBase horizon.Asset, assetQuote horizon.Asset) error {
	// re-compute the liabilities
	sdex.cachedLiabilities = map[horizon.Asset]Liabilities{}
	baseLiabilities, basePairLiabilities, e := sdex.pairLiabilities(assetBase, assetQuote)
	if e != nil {
		return e
	}
	quoteLiabilities, quotePairLiabilities, e := sdex.pairLiabilities(assetQuote, assetBase)
	if e != nil {
		return e
	}

	// delete liability amounts related to all offers (filter on only those offers involving **both** assets in case the account is used by multiple bots)
	sdex.cachedLiabilities[assetBase] = Liabilities{
		Buying:  baseLiabilities.Buying - basePairLiabilities.Buying,
		Selling: baseLiabilities.Selling - basePairLiabilities.Selling,
	}
	sdex.cachedLiabilities[assetQuote] = Liabilities{
		Buying:  quoteLiabilities.Buying - quotePairLiabilities.Buying,
		Selling: quoteLiabilities.Selling - quotePairLiabilities.Selling,
	}
	return nil
}

// AvailableCapacity returns the buying and selling amounts available for a given asset
func (sdex *SDEX) AvailableCapacity(asset horizon.Asset, incrementalNativeAmountRaw float64) (*Liabilities, error) {
	l, e := sdex.assetLiabilities(asset)
	if e != nil {
		return nil, e
	}

	bal, trust, minAccountBal, e := sdex.assetBalance(asset)
	if e != nil {
		return nil, e
	}

	// factor in cost of increase in minReserve and fee when calculating selling capacity of native asset
	incrementalSellingLiability := 0.0
	if asset == utils.NativeAsset {
		incrementalSellingLiability = incrementalNativeAmountRaw
	}

	return &Liabilities{
		Buying:  trust - l.Buying,
		Selling: bal - minAccountBal - l.Selling - incrementalSellingLiability,
	}, nil
}

// assetLiabilities returns the liabilities for the asset
func (sdex *SDEX) assetLiabilities(asset horizon.Asset) (*Liabilities, error) {
	if v, ok := sdex.cachedLiabilities[asset]; ok {
		return &v, nil
	}

	assetLiabilities, _, e := sdex._liabilities(asset, asset) // pass in the same asset, we ignore the returned object anyway
	return assetLiabilities, e
}

// pairLiabilities returns the liabilities for the asset along with the pairLiabilities
func (sdex *SDEX) pairLiabilities(asset horizon.Asset, otherAsset horizon.Asset) (*Liabilities, *Liabilities, error) {
	assetLiabilities, pairLiabilities, e := sdex._liabilities(asset, otherAsset)
	return assetLiabilities, pairLiabilities, e
}

// liabilities returns the asset liabilities and pairLiabilities (non-nil only if the other asset is specified)
func (sdex *SDEX) _liabilities(asset horizon.Asset, otherAsset horizon.Asset) (*Liabilities, *Liabilities, error) {
	// uses all offers for this trading account to accommodate sharing by other bots
	offers, err := utils.LoadAllOffers(sdex.TradingAccount, sdex.API)
	if err != nil {
		assetString := utils.Asset2String(asset)
		log.Printf("error: cannot load offers to compute liabilities for asset (%s): %s\n", assetString, err)
		return nil, nil, err
	}

	// liabilities for the asset
	liabilities := Liabilities{}
	// liabilities for the asset w.r.t. the trading pair
	pairLiabilities := Liabilities{}
	for _, offer := range offers {
		if offer.Selling == asset {
			offerAmt, err := sdex.ParseOfferAmount(offer.Amount)
			if err != nil {
				return nil, nil, err
			}
			liabilities.Selling += offerAmt

			if offer.Buying == otherAsset {
				pairLiabilities.Selling += offerAmt
			}
		} else if offer.Buying == asset {
			offerAmt, err := sdex.ParseOfferAmount(offer.Amount)
			if err != nil {
				return nil, nil, err
			}
			offerPrice, err := sdex.ParseOfferAmount(offer.Price)
			if err != nil {
				return nil, nil, err
			}
			buyingAmount := offerAmt * offerPrice
			liabilities.Buying += buyingAmount

			if offer.Selling == otherAsset {
				pairLiabilities.Buying += buyingAmount
			}
		}
	}

	sdex.cachedLiabilities[asset] = liabilities
	return &liabilities, &pairLiabilities, nil
}
