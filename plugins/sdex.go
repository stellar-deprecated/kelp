package plugins

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/nikhilsaraf/go-tools/multithreading"
	"github.com/pkg/errors"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/networking"
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
	API                           *horizonclient.Client
	SourceAccount                 string
	TradingAccount                string
	SourceSeed                    string
	TradingSeed                   string
	Network                       string
	threadTracker                 *multithreading.ThreadTracker
	operationalBuffer             float64
	operationalBufferNonNativePct float64
	simMode                       bool
	pair                          *model.TradingPair
	assetMap                      map[model.Asset]hProtocol.Asset // this is needed until we fully address putting SDEX behind the Exchange interface
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
	api *horizonclient.Client,
	ieif *IEIF,
	exchangeShim api.ExchangeShim,
	sourceSeed string,
	tradingSeed string,
	sourceAccount string,
	tradingAccount string,
	network string,
	threadTracker *multithreading.ThreadTracker,
	operationalBuffer float64,
	operationalBufferNonNativePct float64,
	simMode bool,
	pair *model.TradingPair,
	assetMap map[model.Asset]hProtocol.Asset,
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

	log.Printf("Using network passphrase: %s\n", sdex.Network)

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
		var a hProtocol.Asset
		if v, ok := elem.(hProtocol.Asset); ok {
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
func (sdex *SDEX) GetAssetConverter() model.AssetConverterInterface {
	return model.Display
}

func (sdex *SDEX) incrementSeqNum() {
	if sdex.reloadSeqNum {
		log.Println("reloading sequence number")
		acctReq := horizonclient.AccountRequest{AccountID: sdex.SourceAccount}
		accountDetail, err := sdex.API.AccountDetail(acctReq)
		if err != nil {
			log.Printf("error loading account detail: %s\n", err)
			return
		}
		seqNum, err := accountDetail.GetSequenceNumber()
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
func (sdex *SDEX) DeleteAllOffers(offers []hProtocol.Offer) []txnbuild.Operation {
	ops := []txnbuild.Operation{}
	for _, offer := range offers {
		op := sdex.DeleteOffer(offer)
		ops = append(ops, &op)
	}
	return ops
}

// DeleteOffer returns the op that needs to be submitted to the network in order to delete the passed in offer
func (sdex *SDEX) DeleteOffer(offer hProtocol.Offer) txnbuild.ManageSellOffer {
	var result txnbuild.ManageSellOffer
	var e error
	if sdex.SourceAccount == sdex.TradingAccount {
		result, e = txnbuild.DeleteOfferOp(offer.ID)
	} else {
		result, e = txnbuild.DeleteOfferOp(offer.ID, &txnbuild.SimpleAccount{AccountID: sdex.TradingAccount})
	}

	if e != nil {
		panic(fmt.Sprintf("unexpected error while creating delete offer op: %s", e))
	}
	return result
}

// ModifyBuyOffer modifies a buy offer
func (sdex *SDEX) ModifyBuyOffer(offer hProtocol.Offer, price float64, amount float64, incrementalNativeAmountRaw float64) (*txnbuild.ManageSellOffer, error) {
	return sdex.ModifySellOffer(offer, 1/price, amount*price, incrementalNativeAmountRaw)
}

// ModifySellOffer modifies a sell offer
func (sdex *SDEX) ModifySellOffer(offer hProtocol.Offer, price float64, amount float64, incrementalNativeAmountRaw float64) (*txnbuild.ManageSellOffer, error) {
	return sdex.createModifySellOffer(&offer, offer.Selling, offer.Buying, price, amount, incrementalNativeAmountRaw)
}

// CreateSellOffer creates a sell offer
func (sdex *SDEX) CreateSellOffer(base hProtocol.Asset, counter hProtocol.Asset, price float64, amount float64, incrementalNativeAmountRaw float64) (*txnbuild.ManageSellOffer, error) {
	return sdex.createModifySellOffer(nil, base, counter, price, amount, incrementalNativeAmountRaw)
}

func (sdex *SDEX) minReserve(subentries int32) float64 {
	return float64(2+subentries) * baseReserve
}

// assetBalance returns asset balance, asset trust limit, reserve balance (zero for non-XLM), error
func (sdex *SDEX) _assetBalance(asset hProtocol.Asset) (*api.Balance, error) {
	acctReq := horizonclient.AccountRequest{AccountID: sdex.TradingAccount}
	account, err := sdex.API.AccountDetail(acctReq)
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
func (sdex *SDEX) GetBalanceHack(asset hProtocol.Asset) (*api.Balance, error) {
	b, e := sdex._assetBalance(asset)
	return b, e
}

// LoadOffersHack impl
func (sdex *SDEX) LoadOffersHack() ([]hProtocol.Offer, error) {
	return sdex._loadOffers()
}

func (sdex *SDEX) _loadOffers() ([]hProtocol.Offer, error) {
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
func (sdex *SDEX) createModifySellOffer(offer *hProtocol.Offer, selling hProtocol.Asset, buying hProtocol.Asset, price float64, amount float64, incrementalNativeAmountRaw float64) (*txnbuild.ManageSellOffer, error) {
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
	stringAmount := strconv.FormatFloat(amount, 'f', int(sdexOrderConstraints.VolumePrecision), 64)

	result, err := txnbuild.CreateOfferOp(utils.Asset2Asset(selling), utils.Asset2Asset(buying), stringAmount, stringPrice)
	if err != nil {
		return nil, err
	}

	if offer != nil {
		result.OfferID = offer.ID
	}
	if sdex.SourceAccount != sdex.TradingAccount {
		result.SourceAccount = &txnbuild.SimpleAccount{AccountID: sdex.TradingAccount}
	}

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
func (sdex *SDEX) submitOps(opsOld []build.TransactionMutator, asyncCallback func(hash string, e error), asyncMode bool) error {
	ops := api.ConvertTM2Operation(opsOld)

	sdex.incrementSeqNum()
	tx := txnbuild.Transaction{
		// sequence number is decremented here because Transaction.Build auto increments sequence number
		SourceAccount: &txnbuild.SimpleAccount{
			AccountID: sdex.SourceAccount,
			Sequence:  int64(sdex.seqNum - 1),
		},
		Operations: ops,
		Timebounds: txnbuild.NewInfiniteTimeout(),
		Network:    sdex.Network,
	}

	// compute fee per operation
	opFee, e := sdex.opFeeStroopsFn()
	if e != nil {
		return fmt.Errorf("SubmitOps error when computing op fee: %s", e)
	}
	tx.BaseFee = uint32(opFee)
	e = tx.Build()
	if e != nil {
		return errors.Wrap(e, "SubmitOps error: ")
	}

	// convert to xdr string
	txeB64, e := sdex.sign(&tx)
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
func (sdex *SDEX) CreateBuyOffer(base hProtocol.Asset, counter hProtocol.Asset, price float64, amount float64, incrementalNativeAmountRaw float64) (*txnbuild.ManageSellOffer, error) {
	return sdex.CreateSellOffer(counter, base, 1/price, amount*price, incrementalNativeAmountRaw)
}

func (sdex *SDEX) sign(tx *txnbuild.Transaction) (string, error) {
	var e error
	if sdex.SourceSeed != sdex.TradingSeed {
		e = utils.SignWithSeed(tx, sdex.SourceSeed, sdex.TradingSeed)
	} else {
		e = utils.SignWithSeed(tx, sdex.SourceSeed)
	}
	if e != nil {
		return "", fmt.Errorf("error signing transaction: %s", e)
	}

	return tx.Base64()
}

func (sdex *SDEX) submit(txeB64 string, asyncCallback func(hash string, e error), asyncMode bool) {
	resp, err := sdex.API.SubmitTransactionXDR(txeB64)
	if err != nil {
		if herr, ok := errors.Cause(err).(*horizonclient.Error); ok {
			var rcs *hProtocol.TransactionResultCodes
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
func (sdex *SDEX) Assets() (baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset, e error) {
	var ok bool
	baseAsset, ok = sdex.assetMap[sdex.pair.Base]
	if !ok {
		return hProtocol.Asset{}, hProtocol.Asset{}, fmt.Errorf("unexpected error, base asset was not found in sdex.assetMap")
	}

	quoteAsset, ok = sdex.assetMap[sdex.pair.Quote]
	if !ok {
		return hProtocol.Asset{}, hProtocol.Asset{}, fmt.Errorf("unexpected error, quote asset was not found in sdex.assetMap")
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
		tradeReq := horizonclient.TradeRequest{
			BaseAssetType:      horizonclient.AssetType(baseAsset.Type),
			BaseAssetCode:      baseAsset.Code,
			BaseAssetIssuer:    baseAsset.Issuer,
			CounterAssetType:   horizonclient.AssetType(quoteAsset.Type),
			CounterAssetCode:   quoteAsset.Code,
			CounterAssetIssuer: quoteAsset.Issuer,
			Order:              horizonclient.OrderAsc,
			Cursor:             cursorStart,
			Limit:              uint(maxPageLimit),
		}

		tradesPage, e := sdex.API.Trades(tradeReq)
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

func makeEffectsLink(trade hProtocol.Trade) string {
	effectsLink := trade.Links.Operation.Href
	if !strings.HasSuffix(effectsLink, "/") {
		effectsLink = effectsLink + "/"
	}
	return effectsLink + "effects?limit=200"
}

func (sdex *SDEX) getOrderAction(baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset, trade hProtocol.Trade) (*model.OrderAction, error) {
	if trade.BaseAccount != sdex.TradingAccount && trade.CounterAccount != sdex.TradingAccount {
		// if the trade is different from what we expect for this bot instance then return empty values so we ignore this trade
		return nil, nil
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
	if !(sdexBaseAsset == tradeBaseAsset && sdexQuoteAsset == tradeQuoteAsset) && !(sdexBaseAsset == tradeQuoteAsset && sdexQuoteAsset == tradeBaseAsset) {
		return nil, nil
	}

	effectsLink := makeEffectsLink(trade)
	for {
		var output map[string]interface{}
		e := networking.JSONRequest(http.DefaultClient, "GET", effectsLink, "", map[string]string{}, &output, "error")
		if e != nil {
			return nil, fmt.Errorf("could not get effect related to trade to fetch orderAction (URL=%s): %s", effectsLink, e)
		}
		embedded, ok := output["_embedded"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("could not cast _embedded field in effect from URL when trying to fetch orderAction (URL=%s): type=%T and json=%v", effectsLink, output["_embedded"], output["_embedded"])
		}
		effectRecords, ok := embedded["records"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("could not cast records to a []interface{} from URL when trying to fetch orderAction (URL=%s): type=%T and json=%v", effectsLink, embedded["records"], embedded["records"])
		}
		if len(effectRecords) == 0 {
			break
		}

		for i, effect := range effectRecords {
			effectMap, ok := effect.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("could not cast record to a map[string]iterface{} for effect at index %d from URL when trying to fetch orderAction (URL=%s): type=%T and json=%v", i, effectsLink, effect, effect)
			}

			effectType, ok := effectMap["type"].(string)
			if !ok {
				return nil, fmt.Errorf("could not cast 'type' for effect record at index %d from URL when trying to fetch orderAction (URL=%s): type=%T and json=%v", i, effectsLink, effectMap["type"], effectMap["type"])
			}
			if effectType != "trade" {
				continue
			}

			accountString, ok := effectMap["account"].(string)
			if !ok {
				return nil, fmt.Errorf("could not cast 'account' for effect record at index %d from URL when trying to fetch orderAction (URL=%s): type=%T and json=%v", i, effectsLink, effectMap["account"], effectMap["account"])
			}
			if accountString != sdex.TradingAccount {
				continue
			}

			soldAsset, e := sdex.parseAssetFromEffect(effectMap, "sold")
			if e != nil {
				return nil, fmt.Errorf("could not parse asset with prefix '%s' from effect at index %d from URL when trying to fetch orderAction (URL=%s): %s", "sold", i, effectsLink, e)
			}

			boughtAsset, e := sdex.parseAssetFromEffect(effectMap, "bought")
			if e != nil {
				return nil, fmt.Errorf("could not parse asset with prefix '%s' from effect at index %d from URL when trying to fetch orderAction (URL=%s): %s", "bought", i, effectsLink, e)
			}

			if !(sdexBaseAsset == soldAsset && sdexQuoteAsset == boughtAsset) && !(sdexBaseAsset == boughtAsset && sdexQuoteAsset == soldAsset) {
				// continue here because it could be another trade in this list of assets
				// i.e. we could have multiple trades in the same path payment but we want to consider these trades individually
				continue
			}

			// compare the base and quote asset on the trade to what we are using as our base and quote
			actionSell := model.OrderActionSell
			actionBuy := model.OrderActionBuy
			if sdexBaseAsset == soldAsset {
				return &actionSell, nil
			}
			return &actionBuy, nil
		}

		links, ok := output["_links"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("could not cast _links field in effect from URL when trying to fetch orderAction (URL=%s): type=%T and json=%v", effectsLink, output["_links"], output["_links"])
		}
		next, ok := links["next"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("could not cast 'next' field in effect's _links from URL when trying to fetch orderAction (URL=%s): type=%T and json=%v", effectsLink, links["next"], links["next"])
		}

		nextHref, ok := next["href"].(string)
		if !ok {
			return nil, fmt.Errorf("could not cast 'href' field in next object of effect's _links from URL when trying to fetch orderAction (URL=%s): type=%T and json=%v", effectsLink, next["href"], next["href"])
		}

		effectsLink = nextHref
	}

	// if we found nothing for this bot instance then return empty values so it ignores this trade
	return nil, nil
}

func (sdex *SDEX) parseAssetFromEffect(effectMap map[string]interface{}, prefix string) (string, error) {
	assetType := effectMap[prefix+"_asset_type"]
	assetTypeString, ok := assetType.(string)
	if !ok {
		return "", fmt.Errorf("could not cast '%s_asset_type' to a string in the operation json result: %s (type=%T)", prefix, assetType, assetType)
	}
	if assetTypeString == utils.Native {
		return utils.Native, nil
	}

	assetCode := effectMap[prefix+"_asset_code"]
	assetCodeString, ok := assetCode.(string)
	if !ok {
		return "", fmt.Errorf("could not cast '%s_asset_code' to a string in the operation json result: %s (type=%T)", prefix, assetCode, assetCode)
	}

	assetIssuer := effectMap[prefix+"_asset_issuer"]
	assetIssuerString, ok := assetIssuer.(string)
	if !ok {
		return "", fmt.Errorf("could not cast '%s_asset_issuer' to a string in the operation json result: %s (type=%T)", prefix, assetIssuer, assetIssuer)
	}

	return assetCodeString + ":" + assetIssuerString, nil
}

// returns tradeHistoryResult, hitCursorEnd, and any error
func (sdex *SDEX) tradesPage2TradeHistoryResult(baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset, tradesPage hProtocol.TradesPage, cursorEnd string) (*api.TradeHistoryResult, bool, error) {
	var cursor string
	trades := []model.Trade{}

	for _, t := range tradesPage.Embedded.Records {
		orderAction, e := sdex.getOrderAction(baseAsset, quoteAsset, t)
		if e != nil {
			return nil, false, fmt.Errorf("could not load orderAction: %s", e)
		}
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

	tradeReq := horizonclient.TradeRequest{
		BaseAssetType:      horizonclient.AssetType(baseAsset.Type),
		BaseAssetCode:      baseAsset.Code,
		BaseAssetIssuer:    baseAsset.Issuer,
		CounterAssetType:   horizonclient.AssetType(quoteAsset.Type),
		CounterAssetCode:   quoteAsset.Code,
		CounterAssetIssuer: quoteAsset.Issuer,
		Order:              horizonclient.OrderDesc,
		Limit:              uint(1),
	}

	tradesPage, e := sdex.API.Trades(tradeReq)
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

	obReq := horizonclient.OrderBookRequest{
		SellingAssetType:   horizonclient.AssetType(baseAsset.Type),
		SellingAssetCode:   baseAsset.Code,
		SellingAssetIssuer: baseAsset.Issuer,
		BuyingAssetType:    horizonclient.AssetType(quoteAsset.Type),
		BuyingAssetCode:    quoteAsset.Code,
		BuyingAssetIssuer:  quoteAsset.Issuer,
	}

	ob, e := sdex.API.OrderBook(obReq)
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
	side []hProtocol.PriceLevel,
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
