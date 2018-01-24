package orderbook

// OrderType represents a type of an order, example market, limit, etc.
type OrderType int8

// These are the available order types
const (
	TypeMarket OrderType = 0
	TypeLimit  OrderType = 1
)

// IsMarket returns true for market orders
func (o OrderType) IsMarket() bool {
	return o == TypeMarket
}

// IsLimit returns true for limit orders
func (o OrderType) IsLimit() bool {
	return o == TypeLimit
}

// String is the stringer function
func (o OrderType) String() string {
	if o == TypeMarket {
		return "market"
	} else if o == TypeLimit {
		return "limit"
	}
	return "error, unrecognized order type"
}
