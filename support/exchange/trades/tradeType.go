package trades

// TradeType is the type of Trade (buy/sell and market/limit)
type TradeType struct {
	_name      string
	_direction _direction
	_type      _type
}

// The types of trades as enums
var (
	BuyMarket  = &TradeType{"buy_market", buy, market}
	BuyLimit   = &TradeType{"buy_limit", buy, limit}
	SellMarket = &TradeType{"sell_market", sell, market}
	SellLimit  = &TradeType{"sell_limit", sell, limit}
)

func (t TradeType) String() string {
	return t._name
}

// IsBuy method
func (t TradeType) IsBuy() bool {
	return t._direction == buy
}

// IsSell method
func (t TradeType) IsSell() bool {
	return t._direction == sell
}

// IsMarket method
func (t TradeType) IsMarket() bool {
	return t._type == market
}

// IsLimit method
func (t TradeType) IsLimit() bool {
	return t._type == limit
}

type _type int8

// type of Market and Limit
const (
	market _type = 0
	limit  _type = 1
)

type _direction int8

// direction Buy and Sell
const (
	buy  _direction = 0
	sell _direction = 1
)
