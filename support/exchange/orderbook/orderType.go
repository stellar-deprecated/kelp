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

var orderTypeMap = map[string]OrderType{
	"market": TypeMarket,
	"limit":  TypeLimit,
}

// OrderTypeFromString is a convenience to convert from common strings to the corresponding OrderType
func OrderTypeFromString(s string) OrderType {
	return orderTypeMap[s]
}
