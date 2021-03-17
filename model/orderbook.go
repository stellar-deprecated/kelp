package model

import (
	"fmt"
	"strconv"

	"github.com/stellar/kelp/support/utils"
)

// OrderAction is the action of buy / sell
type OrderAction bool

// OrderActionBuy and OrderActionSell are the two actions
const (
	OrderActionBuy  OrderAction = false
	OrderActionSell OrderAction = true
)

// minQuoteVolumePrecision allows for having precise enough minQuoteVolume values
const minQuoteVolumePrecision = 10
const nilString = "<nil>"

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
	tsString := nilString
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

// TopAsk returns the best ask in an orderbook
func (o OrderBook) TopAsk() *Order {
	if len(o.Asks()) > 0 {
		return &o.Asks()[0]
	}
	return nil
}

// TopBid returns the best bid in an orderbook
func (o OrderBook) TopBid() *Order {
	if len(o.Bids()) > 0 {
		return &o.Bids()[0]
	}
	return nil
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

// AsInt64 converts to an integer
func (t *TransactionID) AsInt64() (int64, error) {
	return strconv.ParseInt(t.String(), 10, 64)
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
	expireTimeString := nilString
	if o.ExpireTime != nil {
		expireTimeString = fmt.Sprintf("%d", o.ExpireTime.AsInt64())
	}
	return fmt.Sprintf("OpenOrder[order=%s, ID=%s, startTime=%d, expireTime=%s, volumeExecuted=%s]",
		o.Order.String(),
		o.ID,
		o.StartTime.AsInt64(),
		expireTimeString,
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
	OrderID       string
	Cost          *Number
	Fee           *Number
}

// TradesByTsID implements sort.Interface for []Trade based on Timestamp and TransactionID
type TradesByTsID []Trade

func (t TradesByTsID) Len() int {
	return len(t)
}
func (t TradesByTsID) Swap(i int, j int) {
	t[i], t[j] = t[j], t[i]
}
func (t TradesByTsID) Less(i int, j int) bool {
	if t[i].Order.Timestamp.AsInt64() < t[j].Order.Timestamp.AsInt64() {
		return true
	} else if t[i].Order.Timestamp.AsInt64() > t[j].Order.Timestamp.AsInt64() {
		return false
	}

	if t[i].TransactionID != nil && t[j].TransactionID != nil {
		return t[i].TransactionID.String() < t[j].TransactionID.String()
	}
	if t[i].TransactionID != nil {
		return false
	}
	return true
}

func (t Trade) String() string {
	return fmt.Sprintf("Trade[txid: %s, orderId: %s, ts: %s, pair: %s, action: %s, type: %s, counterPrice: %s, baseVolume: %s, counterCost: %s, fee: %s]",
		utils.CheckedString(t.TransactionID),
		t.OrderID,
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

// OrderConstraints describes constraints when placing orders on an excahnge
type OrderConstraints struct {
	PricePrecision  int8
	VolumePrecision int8
	MinBaseVolume   Number
	MinQuoteVolume  *Number
}

// MakeOrderConstraints is a factory method for OrderConstraints
func MakeOrderConstraints(pricePrecision int8, volumePrecision int8, minBaseVolume float64) *OrderConstraints {
	return &OrderConstraints{
		PricePrecision:  pricePrecision,
		VolumePrecision: volumePrecision,
		MinBaseVolume:   *NumberFromFloat(minBaseVolume, volumePrecision),
		MinQuoteVolume:  nil,
	}
}

// MakeOrderConstraintsWithCost is a factory method for OrderConstraints
func MakeOrderConstraintsWithCost(pricePrecision int8, volumePrecision int8, minBaseVolume float64, minQuoteVolume float64) *OrderConstraints {
	return &OrderConstraints{
		PricePrecision:  pricePrecision,
		VolumePrecision: volumePrecision,
		MinBaseVolume:   *NumberFromFloat(minBaseVolume, volumePrecision),
		MinQuoteVolume:  NumberFromFloat(minQuoteVolume, minQuoteVolumePrecision),
	}
}

// MakeOrderConstraintsWithOverride is a factory method for OrderConstraints, oc is not a pointer because we want a copy since we modify it
func MakeOrderConstraintsWithOverride(oc OrderConstraints, override *OrderConstraintsOverride) *OrderConstraints {
	if override.PricePrecision != nil {
		oc.PricePrecision = *override.PricePrecision
	}
	if override.VolumePrecision != nil {
		oc.VolumePrecision = *override.VolumePrecision
	}
	if override.MinBaseVolume != nil {
		oc.MinBaseVolume = *override.MinBaseVolume
	}
	if override.MinQuoteVolume != nil {
		oc.MinQuoteVolume = *override.MinQuoteVolume
	}
	return &oc
}

// MakeOrderConstraintsFromOverride is a factory method to convert an OrderConstraintsOverride to an OrderConstraints
func MakeOrderConstraintsFromOverride(override *OrderConstraintsOverride) *OrderConstraints {
	return MakeOrderConstraintsWithOverride(OrderConstraints{}, override)
}

// OrderConstraints describes constraints when placing orders on an excahnge
func (o *OrderConstraints) String() string {
	minQuoteVolumeStr := nilString
	if o.MinQuoteVolume != nil {
		minQuoteVolumeStr = o.MinQuoteVolume.AsString()
	}

	return fmt.Sprintf("OrderConstraints[PricePrecision: %d, VolumePrecision: %d, MinBaseVolume: %s, MinQuoteVolume: %s]",
		o.PricePrecision, o.VolumePrecision, o.MinBaseVolume.AsString(), minQuoteVolumeStr)
}

// OrderConstraintsOverride describes an override for an OrderConstraint
type OrderConstraintsOverride struct {
	PricePrecision  *int8
	VolumePrecision *int8
	MinBaseVolume   *Number
	MinQuoteVolume  **Number
}

// MakeOrderConstraintsOverride is a factory method
func MakeOrderConstraintsOverride(
	pricePrecision *int8,
	volumePrecision *int8,
	minBaseVolume *Number,
	minQuoteVolume **Number,
) *OrderConstraintsOverride {
	return &OrderConstraintsOverride{
		PricePrecision:  pricePrecision,
		VolumePrecision: volumePrecision,
		MinBaseVolume:   minBaseVolume,
		MinQuoteVolume:  minQuoteVolume,
	}
}

// MakeOrderConstraintsOverrideFromConstraints is a factory method for OrderConstraintsOverride
func MakeOrderConstraintsOverrideFromConstraints(oc *OrderConstraints) *OrderConstraintsOverride {
	return &OrderConstraintsOverride{
		PricePrecision:  &oc.PricePrecision,
		VolumePrecision: &oc.VolumePrecision,
		MinBaseVolume:   &oc.MinBaseVolume,
		MinQuoteVolume:  &oc.MinQuoteVolume,
	}
}

// IsComplete returns true if the override contains all values
func (override *OrderConstraintsOverride) IsComplete() bool {
	if override.PricePrecision == nil {
		return false
	}

	if override.VolumePrecision == nil {
		return false
	}

	if override.MinBaseVolume == nil {
		return false
	}

	if override.MinQuoteVolume == nil {
		return false
	}

	return true
}

// Augment only updates values if updates are non-nil
func (override *OrderConstraintsOverride) Augment(updates *OrderConstraintsOverride) {
	if updates.PricePrecision != nil {
		override.PricePrecision = updates.PricePrecision
	}

	if updates.VolumePrecision != nil {
		override.VolumePrecision = updates.VolumePrecision
	}

	if updates.MinBaseVolume != nil {
		override.MinBaseVolume = updates.MinBaseVolume
	}

	if updates.MinQuoteVolume != nil {
		override.MinQuoteVolume = updates.MinQuoteVolume
	}
}
