package orderbook

import (
	"strconv"
)

// OrderType represents a type of an order, either an ask or a bid
type OrderType int8

// TypeBid and TypeAsk are the two types of orders
const (
	TypeBid OrderType = 1
	TypeAsk OrderType = 0
)

// IsAsk returns true of the order is of type ask
func (o OrderType) IsAsk() bool {
	return o == TypeAsk
}

// IsBid returns true of the order is of type bid
func (o OrderType) IsBid() bool {
	return o == TypeBid
}

// String is the stringer function
func (o OrderType) String() string {
	if o == TypeBid {
		return "bid"
	} else if o == TypeAsk {
		return "ask"
	}
	return "error, unrecognized order type: " + strconv.FormatInt(int64(o), 10)
}
