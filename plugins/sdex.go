package plugins

import (
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/lightyeario/kelp/support/utils"
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
	operationalBuffer float64
	simMode           bool

	// uninitialized
	seqNum       uint64
	reloadSeqNum bool
	// explicitly calculate liabilities here for now, we can switch over to using the values from Horizon once the protocol change has taken effect
	cachedLiabilities map[horizon.Asset]Liabilities
}

// Liabilities represents the "committed" units of an asset on both the buy and sell sides
type Liabilities struct {
	Buying  float64 // affects how much more can be bought
	Selling float64 // affects how much more can be sold
}

// MakeSDEX is a factory method for SDEX
func MakeSDEX(
	api *horizon.Client,
	sourceSeed string,
	tradingSeed string,
	sourceAccount string,
	tradingAccount string,
	network build.Network,
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
func (sdex *SDEX) ModifyBuyOffer(offer horizon.Offer, price float64, amount float64) *build.ManageOfferBuilder {
	return sdex.ModifySellOffer(offer, 1/price, amount*price)
}

// ModifySellOffer modifies a sell offer
func (sdex *SDEX) ModifySellOffer(offer horizon.Offer, price float64, amount float64) *build.ManageOfferBuilder {
	return sdex.createModifySellOffer(&offer, offer.Selling, offer.Buying, price, amount)
}

// CreateSellOffer creates a sell offer
func (sdex *SDEX) CreateSellOffer(base horizon.Asset, counter horizon.Asset, price float64, amount float64) *build.ManageOfferBuilder {
	if amount > 0 {
		return sdex.createModifySellOffer(nil, base, counter, price, amount)
	}
	log.Println("error: cannot place sell order, zero amount")
	return nil
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

// assetBalance returns asset balance, asset trust limit, reserve balance (zero for non-XLM), error
func (sdex *SDEX) assetBalance(asset horizon.Asset) (float64, float64, float64, error) {
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
			return b, t, 0, e
		}
	}
	return -1, -1, -1, errors.New("could not find a balance for the asset passed in")
}

// createModifySellOffer is the main method that handles the logic of creating or modifying an offer, note that all offers are treated as sell offers in Stellar
func (sdex *SDEX) createModifySellOffer(offer *horizon.Offer, selling horizon.Asset, buying horizon.Asset, price float64, amount float64) *build.ManageOfferBuilder {
	// check liability limits on the asset being sold
	willOversell, incrementalSell, e := sdex.willOversell(offer, selling, amount)
	if e != nil {
		log.Println(e)
		return nil
	}
	if willOversell {
		log.Printf("not placing offer because we run the risk of overselling the asset '%s'\n", utils.Asset2String(selling))
		return nil
	}

	// check trust limits on asset being bought
	willOverbuy, incrementalBuy, e := sdex.willOverbuy(offer, buying, price*amount)
	if e != nil {
		log.Println(e)
		return nil
	}
	if willOverbuy {
		log.Printf("not placing offer because we run the risk of overbuying the asset '%s'\n", utils.Asset2String(buying))
		return nil
	}

	// explicitly check that we will not oversell XLM because of fee and min reserves
	incrementalNativeAmountRaw := 0.0
	if sdex.TradingAccount == sdex.SourceAccount {
		// at the minimum it will cost us a unit of base fee for this operation
		incrementalNativeAmountRaw += baseFee
	}
	if offer == nil {
		// new offers will increase the min reserve
		incrementalNativeAmountRaw += baseReserve
	}
	incrementalNativeAmountTotal := incrementalNativeAmountRaw
	if selling.Type == utils.Native {
		incrementalNativeAmountTotal += incrementalSell
	} else if buying.Type == utils.Native {
		incrementalNativeAmountTotal -= incrementalBuy
	}
	willOversellNative, e := sdex.willOversellNative(incrementalNativeAmountTotal)
	if e != nil {
		log.Println(e)
		return nil
	}
	if willOversellNative {
		log.Println("not placing offer because we run the risk of overselling the native asset after including fee and min reserves")
		return nil
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

	// update the cached liabilities
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
	return &result
}

// willOversellNative returns willOversellNative, error
func (sdex *SDEX) willOversellNative(incrementalNativeAmount float64) (bool, error) {
	nativeBal, _, minAccountBal, e := sdex.assetBalance(utils.NativeAsset)
	if e != nil {
		return false, e
	}
	nativeLiabilities, e := sdex.liabilities(utils.NativeAsset)
	if e != nil {
		return false, e
	}

	willOversellNative := incrementalNativeAmount > (nativeBal - minAccountBal - nativeLiabilities.Selling)
	return willOversellNative, nil
}

// willOversell returns willOversell, incrementalAmount, error
func (sdex *SDEX) willOversell(offer *horizon.Offer, asset horizon.Asset, amountSelling float64) (bool, float64, error) {
	offerSellingAmount, e := sdex.getSignedAmount(offer, asset)
	if e != nil {
		return false, 0, e
	}
	if offerSellingAmount > 0 {
		return false, 0, fmt.Errorf("error: offer.Selling (%s) does not match asset being sold (%s)", utils.Asset2String(offer.Selling), asset)
	}

	incrementalAmount := amountSelling - offerSellingAmount
	// reducing our selling amount can never exceed the threshold for that asset
	if incrementalAmount < 0 {
		return false, incrementalAmount, nil
	}

	bal, _, minAccountBal, e := sdex.assetBalance(asset)
	if e != nil {
		return false, 0, e
	}
	liabilities, e := sdex.liabilities(asset)
	if e != nil {
		return false, 0, e
	}

	willOversell := incrementalAmount > (bal - minAccountBal - liabilities.Selling)
	return willOversell, incrementalAmount, nil
}

// willOverbuy returns willOverbuy, incrementalAmount, error
func (sdex *SDEX) willOverbuy(offer *horizon.Offer, asset horizon.Asset, amountBuying float64) (bool, float64, error) {
	offerBuyingAmount, e := sdex.getSignedAmount(offer, asset)
	if e != nil {
		return false, 0, e
	}
	if offerBuyingAmount < 0 {
		return false, 0, fmt.Errorf("error: offer.Buying (%s) does not match asset being bought (%s)", utils.Asset2String(offer.Buying), asset)
	}

	incrementalAmount := amountBuying - offerBuyingAmount
	// if we are reducing our buying amount of an asset then it cannot exceed the threshold for that asset
	if incrementalAmount < 0 {
		return false, incrementalAmount, nil
	}
	if asset.Type == utils.Native {
		// you can never overbuy the native asset
		return false, incrementalAmount, nil
	}

	_, trust, _, e := sdex.assetBalance(asset)
	if e != nil {
		return false, 0, e
	}
	liabilities, e := sdex.liabilities(asset)
	if e != nil {
		return false, 0, e
	}

	willOverbuy := incrementalAmount > (trust - liabilities.Buying)
	return willOverbuy, incrementalAmount, nil
}

// getSignedAmount gets the amount in the unit of the asset passed in, + for buying, - for selling. 0 if asset is not in the offer
func (sdex *SDEX) getSignedAmount(offer *horizon.Offer, asset horizon.Asset) (float64, error) {
	if offer == nil {
		return 0, nil
	}

	if offer.Buying != asset && offer.Selling != asset {
		return 0, nil
	}

	offerAmt, e := sdex.ParseOfferAmount(offer.Amount)
	if e != nil {
		return 0, e
	}

	if offer.Selling == asset {
		// negative sign because offer is being sold
		return -offerAmt, nil
	}

	offerPrice, e := sdex.ParseOfferAmount(offer.Price)
	if e != nil {
		return 0, e
	}
	// positive sign because offer is being bought; multiply with offerPrice to keep the units in the asset being bought
	return offerAmt * offerPrice, nil
}

// SubmitOps submits the passed in operations to the network asynchronously in a single transaction
func (sdex *SDEX) SubmitOps(ops []build.TransactionMutator) error {
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
		go sdex.submit(txeB64)
	} else {
		log.Println("not submitting tx XDR to network in simulation mode")
	}

	return nil
}

// CreateBuyOffer creates a buy offer
func (sdex *SDEX) CreateBuyOffer(base horizon.Asset, counter horizon.Asset, price float64, amount float64) *build.ManageOfferBuilder {
	return sdex.CreateSellOffer(counter, base, 1/price, amount*price)
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

func (sdex *SDEX) submit(txeB64 string) {
	resp, err := sdex.API.SubmitTransaction(txeB64)
	if err != nil {
		if herr, ok := errors.Cause(err).(*horizon.Error); ok {
			var rcs *horizon.TransactionResultCodes
			rcs, err = herr.ResultCodes()
			if err != nil {
				log.Printf("(async) error: no result codes from horizon: %s\n", err)
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
		return
	}

	log.Printf("(async) tx confirmation hash: %s\n", resp.Hash)
}

// ResetCachedLiabilities resets the cache
func (sdex *SDEX) ResetCachedLiabilities() {
	sdex.cachedLiabilities = map[horizon.Asset]Liabilities{}
}

func (sdex *SDEX) liabilities(asset horizon.Asset) (*Liabilities, error) {
	if v, ok := sdex.cachedLiabilities[asset]; ok {
		return &v, nil
	}

	// uses all offers for this trading account to accommodate sharing by other bots
	offers, err := utils.LoadAllOffers(sdex.TradingAccount, sdex.API)
	if err != nil {
		assetString := utils.Asset2String(asset)
		log.Printf("error: cannot load offers to compute liabilities for asset (%s): %s\n", assetString, err)
		return nil, err
	}

	liabilities := Liabilities{}
	for _, offer := range offers {
		if offer.Selling == asset {
			offerAmt, err := sdex.ParseOfferAmount(offer.Amount)
			if err != nil {
				return nil, err
			}
			liabilities.Selling += offerAmt
		} else if offer.Buying == asset {
			offerAmt, err := sdex.ParseOfferAmount(offer.Amount)
			if err != nil {
				return nil, err
			}
			liabilities.Buying += offerAmt
		}
	}

	sdex.cachedLiabilities[asset] = liabilities
	return &liabilities, nil
}
