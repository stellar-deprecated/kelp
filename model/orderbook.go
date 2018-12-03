package model

import (
	"fmt"

	"github.com/interstellar/kelp/support/utils"
)

// OrderAction is the action of buy / sell
type OrderAction bool

// OrderActionBuy and OrderActionSell are the two actions
const (
	OrderActionBuy  OrderAction = false
	OrderActionSell OrderAction = true
)

// IsBuy returns true for buy actions
func (a OrderAction) IsBuy() bool {
	return a == OrderActionBuy
}

// IsSell returns true for sell actions
func (a OrderAction) IsSell() bool {
	return a == OrderActionSell
}

// Reverse returns the opposite action
func (a OrderAction) Reverse() OrderAction {
	if a.IsSell() {
		return OrderActionBuy
	}
	return OrderActionSell
}

// String is the stringer function
func (a OrderAction) String() string {
	if a == OrderActionBuy {
		return "buy"
	} else if a == OrderActionSell {
		return "sell"
	}
	return "error, unrecognized order action"
}

var orderActionMap = map[string]OrderAction{
	"buy":  OrderActionBuy,
	"sell": OrderActionSell,
}

// OrderActionFromString is a convenience to convert from common strings to the corresponding OrderAction
func OrderActionFromString(s string) OrderAction {
	return orderActionMap[s]
}

// OrderType represents a type of an order, example market, limit, etc.
type OrderType int8

// These are the available order types
const (
	OrderTypeMarket OrderType = 0
	OrderTypeLimit  OrderType = 1
)

// IsMarket returns true for market orders
func (o OrderType) IsMarket() bool {
	return o == OrderTypeMarket
}

// IsLimit returns true for limit orders
func (o OrderType) IsLimit() bool {
	return o == OrderTypeLimit
}

// String is the stringer function
func (o OrderType) String() string {
	if o == OrderTypeMarket {
		return "market"
	} else if o == OrderTypeLimit {
		return "limit"
	}
	return "error, unrecognized order type"
}

var orderTypeMap = map[string]OrderType{
	"market": OrderTypeMarket,
	"limit":  OrderTypeLimit,
}

// OrderTypeFromString is a convenience to convert from common strings to the corresponding OrderType
func OrderTypeFromString(s string) OrderType {
	return orderTypeMap[s]
}

// Order represents an order in the orderbook
type Order struct {
	Pair        *TradingPair
	OrderAction OrderAction
	OrderType   OrderType
	Price       *Number
	Volume      *Number
	Timestamp   *Timestamp
}

// String is the stringer function
func (o Order) String() string {
	tsString := "<nil>"
	if o.Timestamp != nil {
		tsString = fmt.Sprintf("%d", o.Timestamp.AsInt64())
	}

	return fmt.Sprintf("Order[pair=%s, action=%s, type=%s, price=%s, vol=%s, ts=%s]",
		o.Pair,
		o.OrderAction,
		o.OrderType,
		o.Price.AsString(),
		o.Volume.AsString(),
		tsString,
	)
}

// OrderBook encapsulates the concept of an orderbook on a market
type OrderBook struct {
	pair *TradingPair
	asks []Order
	bids []Order
}

// Pair returns trading pair
func (o OrderBook) Pair() *TradingPair {
	return o.pair
}

// Asks returns the asks in an orderbook
func (o OrderBook) Asks() []Order {
	return o.asks
}

// Bids returns the bids in an orderbook
func (o OrderBook) Bids() []Order {
	return o.bids
}

// MakeOrderBook creates a new OrderBook from the asks and the bids
func MakeOrderBook(pair *TradingPair, asks []Order, bids []Order) *OrderBook {
	return &OrderBook{
		pair: pair,
		asks: asks,
		bids: bids,
	}
}

// TransactionID is typed for the concept of a transaction ID of an order
type TransactionID string

// String is the stringer function
func (t *TransactionID) String() string {
	return string(*t)
}

// MakeTransactionID is a factory method for convenience
func MakeTransactionID(s string) *TransactionID {
	t := TransactionID(s)
	return &t
}

// OpenOrder represents an open order for a trading account
type OpenOrder struct {
	Order
	ID             string
	StartTime      *Timestamp
	ExpireTime     *Timestamp
	VolumeExecuted *Number
}

// String is the stringer function
func (o OpenOrder) String() string {
	return fmt.Sprintf("OpenOrder[order=%s, ID=%s, startTime=%d, expireTime=%d, volumeExecuted=%s]",
		o.Order.String(),
		o.ID,
		o.StartTime.AsInt64(),
		o.ExpireTime.AsInt64(),
		o.VolumeExecuted.AsString(),
	)
}

// CancelOrderResult is the result of a CancelOrder call
type CancelOrderResult int8

// These are the available types
const (
	CancelResultCancelSuccessful CancelOrderResult = 0
	CancelResultPending          CancelOrderResult = 1
	CancelResultFailed           CancelOrderResult = 2
)

// String is the stringer function
func (r CancelOrderResult) String() string {
	if r == CancelResultCancelSuccessful {
		return "cancelled"
	} else if r == CancelResultPending {
		return "pending"
	} else if r == CancelResultFailed {
		return "failed"
	}
	return "error, unrecognized CancelOrderResult"
}

// Trade represents a trade on an exchange
type Trade struct {
	Order
	TransactionID *TransactionID
	Cost          *Number
	Fee           *Number
}

func (t Trade) String() string {
	return fmt.Sprintf("Trade[txid: %s, ts: %s, pair: %s, action: %s, type: %s, counterPrice: %s, baseVolume: %s, counterCost: %s, fee: %s]",
		utils.CheckedString(t.TransactionID),
		utils.CheckedString(t.Timestamp),
		*t.Pair,
		t.OrderAction,
		t.OrderType,
		utils.CheckedString(t.Price),
		utils.CheckedString(t.Volume),
		utils.CheckedString(t.Cost),
		utils.CheckedString(t.Fee),
	)
}
