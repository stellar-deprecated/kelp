package api

import (
	"fmt"
	"math"

	"github.com/stellar/go/build"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/go/xdr"
	"github.com/stellar/kelp/model"
)

// ExchangeAPIKey specifies API credentials for an exchange
type ExchangeAPIKey struct {
	Key    string
	Secret string
}

// ExchangeParam specifies an additional parameter to be sent when initializing the exchange
type ExchangeParam struct {
	Param string
	Value interface{}
}

// ExchangeHeader specifies additional HTTP headers
type ExchangeHeader struct {
	Header string
	Value  string
}

// Account allows you to access key account functions
type Account interface {
	GetAccountBalances(assetList []interface{}) (map[interface{}]model.Number, error)
}

// Ticker encapsulates all the data for a given Trading Pair
type Ticker struct {
	AskPrice  *model.Number
	BidPrice  *model.Number
	LastPrice *model.Number
}

// TradesResult is the result of a GetTrades call
type TradesResult struct {
	Cursor interface{}
	Trades []model.Trade
}

// TradeHistoryResult is the result of a GetTradeHistory call
// this should be the same object as TradesResult but it's a separate object for backwards compatibility
type TradeHistoryResult struct {
	Cursor interface{}
	Trades []model.Trade
}

// TickerAPI is the interface we use as a generic API for getting ticker data from any crypto exchange
type TickerAPI interface {
	GetTickerPrice(pairs []model.TradingPair) (map[model.TradingPair]Ticker, error)
}

// FillTracker knows how to track fills against open orders
type FillTracker interface {
	GetPair() (pair *model.TradingPair)
	// TrackFills should be executed in a new thread
	TrackFills() error
	IsRunningInBackground() bool
	FillTrackSingleIteration() ([]model.Trade, error)
	RegisterHandler(handler FillHandler)
	NumHandlers() uint8
}

// FillHandler is invoked by the FillTracker (once registered) anytime an order is filled
type FillHandler interface {
	HandleFill(trade model.Trade) error
}

// TradeFetcher is the common method between FillTrackable and exchange
// temporarily extracted out from TradeAPI so SDEX has the flexibility to only implement this rather than exchange and FillTrackable
type TradeFetcher interface {
	GetTradeHistory(pair model.TradingPair, maybeCursorStart interface{}, maybeCursorEnd interface{}) (*TradeHistoryResult, error)
}

// FillTrackable enables any implementing exchange to support fill tracking
type FillTrackable interface {
	TradeFetcher
	GetLatestTradeCursor() (interface{}, error)
}

// Constrainable extracts out the method that SDEX can implement for now
type Constrainable interface {
	// return nil if the constraint does not exist for the exchange
	GetOrderConstraints(pair *model.TradingPair) *model.OrderConstraints

	OverrideOrderConstraints(pair *model.TradingPair, override *model.OrderConstraintsOverride)
}

// OrderbookFetcher extracts out the method that should go into ExchangeShim for now
type OrderbookFetcher interface {
	GetOrderBook(pair *model.TradingPair, maxCount int32) (*model.OrderBook, error)
}

// TradeAPI is the interface we use as a generic API for trading on any crypto exchange
type TradeAPI interface {
	GetAssetConverter() model.AssetConverterInterface

	Constrainable

	OrderbookFetcher

	GetTrades(pair *model.TradingPair, maybeCursor interface{}) (*TradesResult, error)

	FillTrackable

	GetOpenOrders(pairs []*model.TradingPair) (map[model.TradingPair][]model.OpenOrder, error)

	AddOrder(order *model.Order, submitMode SubmitMode) (*model.TransactionID, error)

	CancelOrder(txID *model.TransactionID, pair model.TradingPair) (model.CancelOrderResult, error)
}

// PrepareDepositResult is the result of a PrepareDeposit call
type PrepareDepositResult struct {
	Fee      *model.Number // fee that will be deducted from your deposit, i.e. amount available is depositAmount - fee
	Address  string        // address you should send the funds to
	ExpireTs int64         // expire time as a unix timestamp, 0 if it does not expire
}

// DepositAPI is defined by anything where you can deposit funds.
type DepositAPI interface {
	/*
		Input:
			asset - asset you want to deposit
			amount - amount you want to deposit
		Output:
			PrepareDepositResult - contains the deposit instructions
			error - ErrDepositAmountAboveLimit, ErrTooManyDepositAddresses, or any other error
	*/
	PrepareDeposit(asset model.Asset, amount *model.Number) (*PrepareDepositResult, error)
}

// ErrDepositAmountAboveLimit error type
type ErrDepositAmountAboveLimit error

// MakeErrDepositAmountAboveLimit is a factory method
func MakeErrDepositAmountAboveLimit(amount *model.Number, limit *model.Number) ErrDepositAmountAboveLimit {
	return fmt.Errorf("deposit amount (%s) is greater than limit (%s)", amount.AsString(), limit.AsString())
}

// ErrTooManyDepositAddresses error type
type ErrTooManyDepositAddresses error

// MakeErrTooManyDepositAddresses is a factory method
func MakeErrTooManyDepositAddresses() ErrTooManyDepositAddresses {
	return fmt.Errorf("too many deposit addresses, try reusing one of them")
}

// WithdrawInfo is the result of a GetWithdrawInfo call
type WithdrawInfo struct {
	AmountToReceive *model.Number // amount that you will receive after any fees is taken (excludes fees charged on the deposit side)
}

// WithdrawFunds is the result of a WithdrawFunds call
type WithdrawFunds struct {
	WithdrawalID string
}

// WithdrawAPI is defined by anything where you can withdraw funds.
type WithdrawAPI interface {
	/*
		Input:
			asset - asset you want to withdraw
			amountToWithdraw - amount you want deducted from your account
			address - address you want to withdraw to
		Output:
			WithdrawInfo - details on how to perform the withdrawal
			error - ErrWithdrawAmountAboveLimit, ErrWithdrawAmountInvalid, or any other error
	*/
	GetWithdrawInfo(asset model.Asset, amountToWithdraw *model.Number, address string) (*WithdrawInfo, error)

	/*
		Input:
			asset - asset you want to withdraw
			amountToWithdraw - amount you want deducted from your account (fees will be deducted from here, use GetWithdrawInfo for fee estimate)
			address - address you want to withdraw to
		Output:
		    WithdrawFunds - result of the withdrawal
			error - any error
	*/
	WithdrawFunds(
		asset model.Asset,
		amountToWithdraw *model.Number,
		address string,
	) (*WithdrawFunds, error)
}

// ErrWithdrawAmountAboveLimit error type
type ErrWithdrawAmountAboveLimit error

// MakeErrWithdrawAmountAboveLimit is a factory method
func MakeErrWithdrawAmountAboveLimit(amount *model.Number, limit *model.Number) ErrWithdrawAmountAboveLimit {
	return fmt.Errorf("withdraw amount (%s) is greater than limit (%s)", amount.AsString(), limit.AsString())
}

// ErrWithdrawAmountInvalid error type
type ErrWithdrawAmountInvalid error

// MakeErrWithdrawAmountInvalid is a factory method
func MakeErrWithdrawAmountInvalid(amountToWithdraw *model.Number, fee *model.Number) ErrWithdrawAmountInvalid {
	return fmt.Errorf("amountToWithdraw is invalid: %s, fee: %s", amountToWithdraw.AsString(), fee.AsString())
}

// Exchange is the interface we use as a generic API for all crypto exchanges
type Exchange interface {
	Account
	TickerAPI
	TradeAPI
	DepositAPI
	WithdrawAPI
}

// Balance repesents various aspects of an asset's balance
type Balance struct {
	Balance float64
	Trust   float64
	Reserve float64
}

// ExchangeShim is the interface we use as a generic API for all crypto exchanges
type ExchangeShim interface {
	SubmitOps(ops []build.TransactionMutator, submitMode SubmitMode, asyncCallback func(hash string, e error)) error
	SubmitOpsSynch(ops []build.TransactionMutator, submitMode SubmitMode, asyncCallback func(hash string, e error)) error // forced synchronous version of SubmitOps
	GetBalanceHack(asset hProtocol.Asset) (*Balance, error)
	LoadOffersHack() ([]hProtocol.Offer, error)
	Constrainable
	OrderbookFetcher
	FillTrackable
}

// ConvertOperation2TM is a temporary adapter to support transitioning from the old Go SDK to the new SDK without having to bump the major version
func ConvertOperation2TM(ops []txnbuild.Operation) []build.TransactionMutator {
	muts := []build.TransactionMutator{}
	for _, o := range ops {
		var mob build.ManageOfferBuilder
		if mso, ok := o.(*txnbuild.ManageSellOffer); ok {
			mob = build.ManageOffer(
				false,
				build.Amount(mso.Amount),
				build.Rate{
					Selling: build.Asset{Code: mso.Selling.GetCode(), Issuer: mso.Selling.GetIssuer(), Native: mso.Selling.IsNative()},
					Buying:  build.Asset{Code: mso.Buying.GetCode(), Issuer: mso.Buying.GetIssuer(), Native: mso.Buying.IsNative()},
					Price:   build.Price(mso.Price),
				},
				build.OfferID(mso.OfferID),
			)
			if mso.SourceAccount != nil {
				mob.Mutate(build.SourceAccount{AddressOrSeed: mso.SourceAccount.GetAccountID()})
			}
		} else {
			panic(fmt.Sprintf("could not convert txnbuild.Operation to build.TransactionMutator: %v\n", o))
		}
		muts = append(muts, mob)
	}
	return muts
}

// ConvertTM2Operation is a temporary adapter to support transitioning from the old Go SDK to the new SDK without having to bump the major version
func ConvertTM2Operation(muts []build.TransactionMutator) []txnbuild.Operation {
	msos := ConvertTM2MSO(muts)
	return ConvertMSO2Ops(msos)
}

// ConvertTM2MSO converts mutators from the old SDK to ManageSellOffer ops in the new one.
func ConvertTM2MSO(muts []build.TransactionMutator) []*txnbuild.ManageSellOffer {
	msos := []*txnbuild.ManageSellOffer{}
	for _, m := range muts {
		var mso *txnbuild.ManageSellOffer
		if mob, ok := m.(build.ManageOfferBuilder); ok {
			mso = convertMOB2MSO(mob)
		} else if mob, ok := m.(*build.ManageOfferBuilder); ok {
			mso = convertMOB2MSO(*mob)
		} else {
			panic(fmt.Sprintf("could not convert build.TransactionMutator to txnbuild.Operation: %v (type=%T)\n", m, m))
		}
		msos = append(msos, mso)
	}
	return msos
}

// ConvertMSO2Ops converts manage sell offers into Operations.
func ConvertMSO2Ops(msos []*txnbuild.ManageSellOffer) []txnbuild.Operation {
	ops := []txnbuild.Operation{}
	for _, mso := range msos {
		ops = append(ops, mso)
	}
	return ops
}

func convertMOB2MSO(mob build.ManageOfferBuilder) *txnbuild.ManageSellOffer {
	mso := &txnbuild.ManageSellOffer{
		Amount:  fmt.Sprintf("%.7f", float64(mob.MO.Amount)/math.Pow(10, 7)),
		OfferID: int64(mob.MO.OfferId),
		Price:   fmt.Sprintf("%.7f", float64(mob.MO.Price.N)/float64(mob.MO.Price.D)),
	}
	if mob.O.SourceAccount != nil {
		mso.SourceAccount = &txnbuild.SimpleAccount{
			AccountID: mob.O.SourceAccount.Address(),
		}
	}

	if mob.MO.Buying.Type == xdr.AssetTypeAssetTypeNative {
		mso.Buying = txnbuild.NativeAsset{}
	} else {
		var tipe, code, issuer string
		mob.MO.Buying.MustExtract(&tipe, &code, &issuer)
		mso.Buying = txnbuild.CreditAsset{
			Code:   code,
			Issuer: issuer,
		}
	}

	if mob.MO.Selling.Type == xdr.AssetTypeAssetTypeNative {
		mso.Selling = txnbuild.NativeAsset{}
	} else {
		var tipe, code, issuer string
		mob.MO.Selling.MustExtract(&tipe, &code, &issuer)
		mso.Selling = txnbuild.CreditAsset{
			Code:   code,
			Issuer: issuer,
		}
	}

	return mso
}
