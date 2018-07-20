package api

import (
	"github.com/lightyeario/kelp/model/assets"
	"github.com/lightyeario/kelp/support/exchange/api/number"
	"github.com/lightyeario/kelp/support/exchange/api/orderbook"
	"github.com/lightyeario/kelp/support/exchange/api/trades"
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
	GetPrecision() int8

	GetAssetConverter() *assets.AssetConverter

	GetTickerPrice(pairs []assets.TradingPair) (map[assets.TradingPair]Ticker, error)

	GetOrderBook(pair *assets.TradingPair, maxCount int32) (*orderbook.OrderBook, error)

	GetTrades(pair *assets.TradingPair, maybeCursor interface{}) (*TradesResult, error)

	GetTradeHistory(maybeCursorStart interface{}, maybeCursorEnd interface{}) (*TradeHistoryResult, error)

	GetOpenOrders() (map[assets.TradingPair][]orderbook.OpenOrder, error)

	AddOrder(order *orderbook.Order) (*orderbook.TransactionID, error)

	CancelOrder(txID *orderbook.TransactionID) (trades.CancelOrderResult, error)
}
