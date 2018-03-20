package level

// Level represents a layer in the orderbook
// extracted here because it's shared by strategy and sideStrategy and strategy depeneds on sideStrategy
type Level struct {
	SPREAD float64 `valid:"-"`
	AMOUNT float64 `valid:"-"`
}
