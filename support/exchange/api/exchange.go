package api

import (
	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/support/exchange/api/trades"
)

// Ticker encapsulates all the data for a given Trading Pair
type Ticker struct {
	AskPrice  *model.Number
	AskVolume *model.Number
	BidPrice  *model.Number
	BidVolume *model.Number
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

	GetAssetConverter() *model.AssetConverter

	GetTickerPrice(pairs []model.TradingPair) (map[model.TradingPair]Ticker, error)

	GetOrderBook(pair *model.TradingPair, maxCount int32) (*model.OrderBook, error)

	GetTrades(pair *model.TradingPair, maybeCursor interface{}) (*TradesResult, error)

	GetTradeHistory(maybeCursorStart interface{}, maybeCursorEnd interface{}) (*TradeHistoryResult, error)

	GetOpenOrders() (map[model.TradingPair][]model.OpenOrder, error)

	AddOrder(order *model.Order) (*model.TransactionID, error)

	CancelOrder(txID *model.TransactionID) (trades.CancelOrderResult, error)
}
