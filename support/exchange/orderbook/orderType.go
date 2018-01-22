package orderbook

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
