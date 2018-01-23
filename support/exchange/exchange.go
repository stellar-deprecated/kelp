package exchange

import (
	"github.com/lightyeario/kelp/support/exchange/assets"
	"github.com/lightyeario/kelp/support/exchange/number"
	"github.com/lightyeario/kelp/support/exchange/orderbook"
	"github.com/lightyeario/kelp/support/exchange/trades"
)

// Ticker encapsulates all the data for a given Trading Pair
type Ticker struct {
	AskPrice  *number.Number
	AskVolume *number.Number
	BidPrice  *number.Number
	BidVolume *number.Number
}

// TradesResult is the result of a GetTrades call
type TradesResult struct {
	Cursor interface{}
	Trades []trades.Trade
}

// TradeHistoryResult is the result of a GetTradeHistory call
type TradeHistoryResult struct {
	Trades []trades.Trade
}

// Exchange is the interface we use as a generic API to all crypto exchanges
type Exchange interface {
	// Public
	GetTickerPrice(pairs []assets.TradingPair) (map[assets.TradingPair]Ticker, error)

	// Private
	GetAccountBalances(assetList []assets.Asset) (map[assets.Asset]number.Number, error)

	// Public
	GetOrderBook(pair assets.TradingPair, maxCount int32) (*orderbook.OrderBook, error)

	// Public
	GetTrades(pair assets.TradingPair, maybeCursor interface{}) (*TradesResult, error)

	// Private
	GetTradeHistory(maybeCursorStart interface{}, maybeCursorEnd interface{}) (*TradeHistoryResult, error)
}
