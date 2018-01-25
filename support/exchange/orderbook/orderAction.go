package orderbook

// OrderAction is the action of buy / sell
type OrderAction bool

// ActionBuy and ActionSell are the two actions
const (
	ActionBuy  OrderAction = false
	ActionSell OrderAction = true
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

var orderActionMap = map[string]OrderAction{
	"buy":  ActionBuy,
	"sell": ActionSell,
}

// OrderActionFromString is a convenience to convert from common strings to the corresponding OrderAction
func OrderActionFromString(s string) OrderAction {
	return orderActionMap[s]
}
