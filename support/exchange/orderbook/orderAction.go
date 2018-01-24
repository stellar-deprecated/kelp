package orderbook

// OrderAction is the action of buy / sell
type OrderAction bool

// ActionBuy and ActionSell are the two actions
const (
	ActionBuy  OrderAction = true
	ActionSell OrderAction = false
)

// IsBuy returns true for buy actions
func (a OrderAction) IsBuy() bool {
	return a == ActionBuy
}

// IsSell returns true for sell actions
func (a OrderAction) IsSell() bool {
	return a == ActionSell
}

// String is the stringer function
func (a OrderAction) String() string {
	if a == ActionBuy {
		return "buy"
	} else if a == ActionSell {
		return "sell"
	}
	return "error, unrecognized order action"
}
