package api

import (
	"fmt"

	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/kelp/model"
)

// ExchangeAPIKey specifies API credentials for an exchange
type ExchangeAPIKey struct {
	Key    string
	Secret string
}

// Account allows you to access key account functions
type Account interface {
	GetAccountBalances(assetList []interface{}) (map[interface{}]model.Number, error)
}

// Ticker encapsulates all the data for a given Trading Pair
type Ticker struct {
	AskPrice *model.Number
	BidPrice *model.Number
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
}

// TradeAPI is the interface we use as a generic API for trading on any crypto exchange
type TradeAPI interface {
	GetAssetConverter() *model.AssetConverter

	Constrainable

	GetOrderBook(pair *model.TradingPair, maxCount int32) (*model.OrderBook, error)

	GetTrades(pair *model.TradingPair, maybeCursor interface{}) (*TradesResult, error)

	TradeFetcher

	GetOpenOrders(pairs []*model.TradingPair) (map[model.TradingPair][]model.OpenOrder, error)

	AddOrder(order *model.Order) (*model.TransactionID, error)

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
	SubmitOps(ops []build.TransactionMutator, asyncCallback func(hash string, e error)) error
	GetBalanceHack(asset horizon.Asset) (*Balance, error)
	LoadOffersHack() ([]horizon.Offer, error)
}
