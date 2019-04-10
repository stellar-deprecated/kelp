package plugins

import (
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/nikhilsaraf/go-tools/multithreading"
	"github.com/pkg/errors"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

const baseReserve = 0.5
const baseFee = 0.0000100
const maxLumenTrust = math.MaxFloat64
const maxPageLimit = 200

var sdexOrderConstraints = model.MakeOrderConstraints(7, 7, 0.0000001)

// TODO we need a reasonable value for the resolution here (currently arbitrary 300000 from a test in horizon)
const fetchTradesResolution = 300000

// SDEX helps with building and submitting transactions to the Stellar network
type SDEX struct {
	API                           *horizon.Client
	SourceAccount                 string
	TradingAccount                string
	SourceSeed                    string
	TradingSeed                   string
	Network                       build.Network
	threadTracker                 *multithreading.ThreadTracker
	operationalBuffer             float64
	operationalBufferNonNativePct float64
	simMode                       bool
	pair                          *model.TradingPair
	assetMap                      map[model.Asset]horizon.Asset // this is needed until we fully address putting SDEX behind the Exchange interface
	opFeeStroopsFn                OpFeeStroops
	tradingOnSdex                 bool

	// uninitialized
	seqNum             uint64
	reloadSeqNum       bool
	ieif               *IEIF
	ocOverridesHandler *OrderConstraintsOverridesHandler
}

// enforce SDEX implements api.Constrainable
var _ api.Constrainable = &SDEX{}

// enforce SDEX implements api.ExchangeShim
var _ api.ExchangeShim = &SDEX{}

// Balance repesents an asset's balance response from the assetBalance method below
type Balance struct {
	Balance float64
	Trust   float64
	Reserve float64
}

// MakeSDEX is a factory method for SDEX
func MakeSDEX(
	api *horizon.Client,
	ieif *IEIF,
	exchangeShim api.ExchangeShim,
	sourceSeed string,
	tradingSeed string,
	sourceAccount string,
	tradingAccount string,
	network build.Network,
	threadTracker *multithreading.ThreadTracker,
	operationalBuffer float64,
	operationalBufferNonNativePct float64,
	simMode bool,
	pair *model.TradingPair,
	assetMap map[model.Asset]horizon.Asset,
	opFeeStroopsFn OpFeeStroops,
) *SDEX {
	sdex := &SDEX{
		API:                           api,
		ieif:                          ieif,
		SourceSeed:                    sourceSeed,
		TradingSeed:                   tradingSeed,
		SourceAccount:                 sourceAccount,
		TradingAccount:                tradingAccount,
		Network:                       network,
		threadTracker:                 threadTracker,
		operationalBuffer:             operationalBuffer,
		operationalBufferNonNativePct: operationalBufferNonNativePct,
		simMode:            simMode,
		pair:               pair,
		assetMap:           assetMap,
		opFeeStroopsFn:     opFeeStroopsFn,
		tradingOnSdex:      exchangeShim == nil,
		ocOverridesHandler: MakeEmptyOrderConstraintsOverridesHandler(),
	}

	if exchangeShim == nil {
		exchangeShim = sdex
	}
	// TODO 2 remove this hack, we need to find a way of having ieif get a handle to compute balances or always compute and pass balances in?
	ieif.SetExchangeShim(exchangeShim)

	log.Printf("Using network passphrase: %s\n", sdex.Network.Passphrase)

	if sdex.SourceAccount == "" {
		sdex.SourceAccount = sdex.TradingAccount
		sdex.SourceSeed = sdex.TradingSeed
		log.Println("No Source Account Set")
	}
	sdex.reloadSeqNum = true

	return sdex
}

// IEIF exoses the ieif var
func (sdex *SDEX) IEIF() *IEIF {
	return sdex.ieif
}

// GetAccountBalances impl
func (sdex *SDEX) GetAccountBalances(assetList []interface{}) (map[interface{}]model.Number, error) {
	m := map[interface{}]model.Number{}
	for _, elem := range assetList {
		var a horizon.Asset
		if v, ok := elem.(horizon.Asset); ok {
			a = v
		} else {
			return nil, fmt.Errorf("invalid type of asset passed in, only horizon.Asset accepted")
		}

		balance, e := sdex.ieif.assetBalance(a)
		if e != nil {
			return nil, fmt.Errorf("could not fetch asset balance: %s", e)
		}

		m[elem] = *model.NumberFromFloat(balance.Balance, utils.SdexPrecision)
	}
	return m, nil
}

// GetAssetConverter impl
func (sdex *SDEX) GetAssetConverter() *model.AssetConverter {
	return model.Display
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

// GetOrderConstraints impl
func (sdex *SDEX) GetOrderConstraints(pair *model.TradingPair) *model.OrderConstraints {
	return sdex.ocOverridesHandler.Apply(pair, sdexOrderConstraints)
}

// OverrideOrderConstraints impl, can partially override values for specific pairs
func (sdex *SDEX) OverrideOrderConstraints(pair *model.TradingPair, override *model.OrderConstraintsOverride) {
	sdex.ocOverridesHandler.Upsert(pair, override)
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

func (sdex *SDEX) minReserve(subentries int32) float64 {
	return float64(2+subentries) * baseReserve
}

// assetBalance returns asset balance, asset trust limit, reserve balance (zero for non-XLM), error
func (sdex *SDEX) _assetBalance(asset horizon.Asset) (*api.Balance, error) {
	account, err := sdex.API.LoadAccount(sdex.TradingAccount)
	if err != nil {
		return nil, fmt.Errorf("error: unable to load account to fetch balance: %s", err)
	}

	for _, balance := range account.Balances {
		if utils.AssetsEqual(balance.Asset, asset) {
			b, e := strconv.ParseFloat(balance.Balance, 64)
			if e != nil {
				return nil, fmt.Errorf("error: cannot parse balance: %s", e)
			}
			if balance.Asset.Type == utils.Native {
				return &api.Balance{
					Balance: b,
					Trust:   maxLumenTrust,
					Reserve: sdex.minReserve(account.SubentryCount) + sdex.operationalBuffer,
				}, nil
			}

			t, e := strconv.ParseFloat(balance.Limit, 64)
			if e != nil {
				return nil, fmt.Errorf("error: cannot parse trust limit: %s", e)
			}

			return &api.Balance{
				Balance: b,
				Trust:   t,
				Reserve: b * sdex.operationalBufferNonNativePct,
			}, nil
		}
	}
	return nil, errors.New("could not find a balance for the asset passed in")
}

// GetBalanceHack impl
func (sdex *SDEX) GetBalanceHack(asset horizon.Asset) (*api.Balance, error) {
	b, e := sdex._assetBalance(asset)
	return b, e
}

// LoadOffersHack impl
func (sdex *SDEX) LoadOffersHack() ([]horizon.Offer, error) {
	return sdex._loadOffers()
}

func (sdex *SDEX) _loadOffers() ([]horizon.Offer, error) {
	return utils.LoadAllOffers(sdex.TradingAccount, sdex.API)
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
		return nil, fmt.Errorf("error: cannot create or modify offer, invalid price: %.8f", price)
	}
	if amount <= 0 {
		return nil, fmt.Errorf("error: cannot create or modify offer, invalid amount: %.8f", amount)
	}

	// check liability limits on the asset being sold
	incrementalSell := amount
	willOversell, e := sdex.ieif.willOversell(selling, amount)
	if e != nil {
		return nil, e
	}
	if willOversell {
		return nil, nil
	}

	// check trust limits on asset being bought
	incrementalBuy := price * amount
	willOverbuy, e := sdex.ieif.willOverbuy(buying, incrementalBuy)
	if e != nil {
		return nil, e
	}
	if willOverbuy {
		return nil, nil
	}

	// explicitly check that we will not oversell XLM because of fee and min reserves
	if sdex.tradingOnSdex {
		incrementalNativeAmountTotal := incrementalNativeAmountRaw
		if selling.Type == utils.Native {
			incrementalNativeAmountTotal += incrementalSell
		}
		willOversellNative, e := sdex.ieif.willOversellNative(incrementalNativeAmountTotal)
		if e != nil {
			return nil, e
		}
		if willOversellNative {
			return nil, nil
		}
	}

	stringPrice := strconv.FormatFloat(price, 'f', int(sdexOrderConstraints.PricePrecision), 64)
	rate := build.Rate{
		Selling: utils.Asset2Asset(selling),
		Buying:  utils.Asset2Asset(buying),
		Price:   build.Price(stringPrice),
	}

	mutators := []interface{}{
		rate,
		build.Amount(strconv.FormatFloat(amount, 'f', int(sdexOrderConstraints.VolumePrecision), 64)),
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

// SubmitOpsSynch is the forced synchronous version of SubmitOps below
func (sdex *SDEX) SubmitOpsSynch(ops []build.TransactionMutator, asyncCallback func(hash string, e error)) error {
	return sdex.submitOps(ops, asyncCallback, false)
}

// SubmitOps submits the passed in operations to the network asynchronously in a single transaction
func (sdex *SDEX) SubmitOps(ops []build.TransactionMutator, asyncCallback func(hash string, e error)) error {
	return sdex.submitOps(ops, asyncCallback, true)
}

// submitOps submits the passed in operations to the network in a single transaction. Asynchronous or not based on flag.
func (sdex *SDEX) submitOps(ops []build.TransactionMutator, asyncCallback func(hash string, e error), asyncMode bool) error {
	sdex.incrementSeqNum()
	muts := []build.TransactionMutator{
		build.Sequence{Sequence: sdex.seqNum},
		sdex.Network,
		build.SourceAccount{AddressOrSeed: sdex.SourceAccount},
	}
	// compute fee per operation
	opFee, e := sdex.opFeeStroopsFn()
	if e != nil {
		return fmt.Errorf("SubmitOps error when computing op fee: %s", e)
	}
	muts = append(muts, build.BaseFee{Amount: opFee})
	// add transaction mutators
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
		if asyncMode {
			log.Println("submitting tx XDR to network (async)")
			e = sdex.threadTracker.TriggerGoroutine(func(inputs []interface{}) {
				sdex.submit(txeB64, asyncCallback, true)
			}, nil)
			if e != nil {
				return fmt.Errorf("unable to trigger goroutine to submit tx XDR to network asynchronously: %s", e)
			}
		} else {
			log.Println("submitting tx XDR to network (synch)")
			sdex.submit(txeB64, asyncCallback, false)
		}
	} else {
		log.Println("not submitting tx XDR to network in simulation mode, calling asyncCallback with empty hash value")
		sdex.invokeAsyncCallback(asyncCallback, "", nil, asyncMode)
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

func (sdex *SDEX) submit(txeB64 string, asyncCallback func(hash string, e error), asyncMode bool) {
	resp, err := sdex.API.SubmitTransaction(txeB64)
	if err != nil {
		if herr, ok := errors.Cause(err).(*horizon.Error); ok {
			var rcs *horizon.TransactionResultCodes
			rcs, err = herr.ResultCodes()
			if err != nil {
				log.Printf("(async) error: no result codes from horizon: %s\n", err)
				sdex.invokeAsyncCallback(asyncCallback, "", err, asyncMode)
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
		sdex.invokeAsyncCallback(asyncCallback, "", err, asyncMode)
		return
	}

	modeString := "(synch)"
	if asyncMode {
		modeString = "(async)"
	}
	log.Printf("%s tx confirmation hash: %s\n", modeString, resp.Hash)
	sdex.invokeAsyncCallback(asyncCallback, resp.Hash, nil, asyncMode)
}

func (sdex *SDEX) invokeAsyncCallback(asyncCallback func(hash string, err error), hash string, err error, asyncMode bool) {
	if asyncCallback == nil {
		return
	}

	if asyncMode {
		e := sdex.threadTracker.TriggerGoroutine(func(inputs []interface{}) {
			asyncCallback(hash, err)
		}, nil)
		if e != nil {
			log.Printf("unable to trigger goroutine for invokeAsyncCallback: %s", e)
			return
		}
	} else {
		asyncCallback(hash, err)
	}
}

// Assets returns the base and quote asset used by sdex
func (sdex *SDEX) Assets() (baseAsset horizon.Asset, quoteAsset horizon.Asset, e error) {
	var ok bool
	baseAsset, ok = sdex.assetMap[sdex.pair.Base]
	if !ok {
		return horizon.Asset{}, horizon.Asset{}, fmt.Errorf("unexpected error, base asset was not found in sdex.assetMap")
	}

	quoteAsset, ok = sdex.assetMap[sdex.pair.Quote]
	if !ok {
		return horizon.Asset{}, horizon.Asset{}, fmt.Errorf("unexpected error, quote asset was not found in sdex.assetMap")
	}

	return baseAsset, quoteAsset, nil
}

// enforce SDEX implementing api.FillTrackable
var _ api.FillTrackable = &SDEX{}

// GetTradeHistory fetches trades for the trading account bound to this instance of SDEX
func (sdex *SDEX) GetTradeHistory(pair model.TradingPair, maybeCursorStart interface{}, maybeCursorEnd interface{}) (*api.TradeHistoryResult, error) {
	if pair != *sdex.pair {
		return nil, fmt.Errorf("passed in pair (%s) did not match sdex.pair (%s)", pair.String(), sdex.pair.String())
	}

	baseAsset, quoteAsset, e := sdex.Assets()
	if e != nil {
		return nil, fmt.Errorf("error while converting pair to base and quote asset: %s", e)
	}

	var cursorStart string
	if maybeCursorStart != nil {
		var ok bool
		cursorStart, ok = maybeCursorStart.(string)
		if !ok {
			return nil, fmt.Errorf("could not convert maybeCursorStart to string, type=%s, maybeCursorStart=%v", reflect.TypeOf(maybeCursorStart), maybeCursorStart)
		}
	}
	var cursorEnd string
	if maybeCursorEnd != nil {
		var ok bool
		cursorEnd, ok = maybeCursorEnd.(string)
		if !ok {
			return nil, fmt.Errorf("could not convert maybeCursorEnd to string, type=%s, maybeCursorEnd=%v", reflect.TypeOf(maybeCursorEnd), maybeCursorEnd)
		}
	}

	trades := []model.Trade{}
	for {
		tradesPage, e := sdex.API.LoadTrades(baseAsset, quoteAsset, 0, fetchTradesResolution, horizon.Cursor(cursorStart), horizon.Order(horizon.OrderAsc), horizon.Limit(maxPageLimit))
		if e != nil {
			if strings.Contains(e.Error(), "Rate limit exceeded") {
				// return normally, we will continue loading trades in the next call from where we left off
				return &api.TradeHistoryResult{
					Cursor: cursorStart,
					Trades: trades,
				}, nil
			}
			return nil, fmt.Errorf("error while fetching trades in SDEX (cursor=%s): %s", cursorStart, e)
		}

		if len(tradesPage.Embedded.Records) == 0 {
			return &api.TradeHistoryResult{
				Cursor: cursorStart,
				Trades: trades,
			}, nil
		}

		updatedResult, hitCursorEnd, e := sdex.tradesPage2TradeHistoryResult(baseAsset, quoteAsset, tradesPage, cursorEnd)
		if e != nil {
			return nil, fmt.Errorf("error converting tradesPage2TradesResult: %s", e)
		}
		cursorStart = updatedResult.Cursor.(string)
		trades = append(trades, updatedResult.Trades...)

		if hitCursorEnd {
			return &api.TradeHistoryResult{
				Cursor: cursorStart,
				Trades: trades,
			}, nil
		}
	}
}

func (sdex *SDEX) getOrderAction(baseAsset horizon.Asset, quoteAsset horizon.Asset, trade horizon.Trade) *model.OrderAction {
	if trade.BaseAccount != sdex.TradingAccount && trade.CounterAccount != sdex.TradingAccount {
		return nil
	}

	tradeBaseAsset := utils.Native
	if trade.BaseAssetType != utils.Native {
		tradeBaseAsset = trade.BaseAssetCode + ":" + trade.BaseAssetIssuer
	}
	tradeQuoteAsset := utils.Native
	if trade.CounterAssetType != utils.Native {
		tradeQuoteAsset = trade.CounterAssetCode + ":" + trade.CounterAssetIssuer
	}
	sdexBaseAsset := utils.Asset2String(baseAsset)
	sdexQuoteAsset := utils.Asset2String(quoteAsset)

	// compare the base and quote asset on the trade to what we are using as our base and quote
	// then compare whether it was the base or the quote that was the seller
	actionSell := model.OrderActionSell
	actionBuy := model.OrderActionBuy
	if sdexBaseAsset == tradeBaseAsset && sdexQuoteAsset == tradeQuoteAsset {
		if trade.BaseIsSeller {
			return &actionSell
		}
		return &actionBuy
	} else if sdexBaseAsset == tradeQuoteAsset && sdexQuoteAsset == tradeBaseAsset {
		if trade.BaseIsSeller {
			return &actionBuy
		}
		return &actionSell
	} else {
		return nil
	}
}

// returns tradeHistoryResult, hitCursorEnd, and any error
func (sdex *SDEX) tradesPage2TradeHistoryResult(baseAsset horizon.Asset, quoteAsset horizon.Asset, tradesPage horizon.TradesPage, cursorEnd string) (*api.TradeHistoryResult, bool, error) {
	var cursor string
	trades := []model.Trade{}

	for _, t := range tradesPage.Embedded.Records {
		orderAction := sdex.getOrderAction(baseAsset, quoteAsset, t)
		if orderAction == nil {
			// we have encountered a trade that is different from the base and quote asset for our trading account
			continue
		}

		vol, e := model.NumberFromString(t.BaseAmount, sdexOrderConstraints.VolumePrecision)
		if e != nil {
			return nil, false, fmt.Errorf("could not convert baseAmount to model.Number: %s", e)
		}
		floatPrice := float64(t.Price.N) / float64(t.Price.D)
		price := model.NumberFromFloat(floatPrice, sdexOrderConstraints.PricePrecision)

		trades = append(trades, model.Trade{
			Order: model.Order{
				Pair:        sdex.pair,
				OrderAction: *orderAction,
				OrderType:   model.OrderTypeLimit,
				Price:       price,
				Volume:      vol,
				Timestamp:   model.MakeTimestampFromTime(t.LedgerCloseTime),
			},
			TransactionID: model.MakeTransactionID(t.ID),
			Cost:          price.Multiply(*vol),
			Fee:           model.NumberFromFloat(baseFee, sdexOrderConstraints.PricePrecision),
		})

		cursor = t.PT
		if cursor == cursorEnd {
			return &api.TradeHistoryResult{
				Cursor: cursor,
				Trades: trades,
			}, true, nil
		}
	}

	return &api.TradeHistoryResult{
		Cursor: cursor,
		Trades: trades,
	}, false, nil
}

// GetLatestTradeCursor impl.
func (sdex *SDEX) GetLatestTradeCursor() (interface{}, error) {
	baseAsset, quoteAsset, e := sdex.Assets()
	if e != nil {
		return nil, fmt.Errorf("error while convertig pair to base and quote asset: %s", e)
	}

	tradesPage, e := sdex.API.LoadTrades(baseAsset, quoteAsset, 0, fetchTradesResolution, horizon.Order(horizon.OrderDesc), horizon.Limit(1))
	if e != nil {
		return nil, fmt.Errorf("error while fetching latest trade cursor in SDEX: %s", e)
	}

	records := tradesPage.Embedded.Records
	if len(records) == 0 {
		// we want to use nil as the latest trade cursor if there are no trades
		return nil, nil
	}

	return records[0].PT, nil
}

// GetOrderBook gets the SDEX orderbook
func (sdex *SDEX) GetOrderBook(pair *model.TradingPair, maxCount int32) (*model.OrderBook, error) {
	if pair != sdex.pair {
		return nil, fmt.Errorf("unregistered trading pair (%s) cannot be converted to horizon.Assets, instance's pair: %s", pair.String(), sdex.pair.String())
	}

	baseAsset, quoteAsset, e := sdex.Assets()
	if e != nil {
		return nil, fmt.Errorf("cannot get SDEX orderbook: %s", e)
	}

	ob, e := sdex.API.LoadOrderBook(baseAsset, quoteAsset)
	if e != nil {
		return nil, fmt.Errorf("cannot get SDEX orderbook: %s", e)
	}

	ts := model.MakeTimestamp(time.Now().UnixNano() / int64(time.Millisecond))
	transformedBids, e := sdex.transformHorizonOrders(pair, ob.Bids, model.OrderActionBuy, ts, maxCount)
	if e != nil {
		return nil, fmt.Errorf("could not transform bid side of SDEX orderbook: %s", e)
	}

	transformedAsks, e := sdex.transformHorizonOrders(pair, ob.Asks, model.OrderActionSell, ts, maxCount)
	if e != nil {
		return nil, fmt.Errorf("could not transform ask side of SDEX orderbook: %s", e)
	}

	return model.MakeOrderBook(
		pair,
		transformedAsks,
		transformedBids,
	), nil
}

func (sdex *SDEX) transformHorizonOrders(
	pair *model.TradingPair,
	side []horizon.PriceLevel,
	orderAction model.OrderAction,
	ts *model.Timestamp,
	maxCount int32,
) ([]model.Order, error) {
	transformed := []model.Order{}
	for i, o := range side {
		if i >= int(maxCount) {
			break
		}

		floatPrice := float64(o.PriceR.N) / float64(o.PriceR.D)
		price := model.NumberFromFloat(floatPrice, sdexOrderConstraints.PricePrecision)

		volume, e := model.NumberFromString(o.Amount, sdexOrderConstraints.VolumePrecision)
		if e != nil {
			return nil, fmt.Errorf("could not parse amount for horizon order: %s", e)
		}
		// special handling of amount for bids
		if orderAction.IsBuy() {
			// use floatPrice here for more accuracy since floatPrice is what will be used in stellar-core
			volume = volume.Scale(1.0 / floatPrice)
		}

		transformed = append(transformed, model.Order{
			Pair:        pair,
			OrderAction: orderAction,
			OrderType:   model.OrderTypeLimit,
			Price:       price,
			Volume:      volume,
			Timestamp:   ts,
		})
	}
	return transformed, nil
}
