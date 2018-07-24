package api

// Level represents a layer in the orderbook
type Level struct {
	Price  float64
	Amount float64
}

// LevelProvider returns the levels for the given center price, which controls the spread and number of levels
type LevelProvider interface {
	GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]Level, error)
}
