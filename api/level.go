package api

import "github.com/stellar/kelp/model"

// Level represents a layer in the orderbook
type Level struct {
	Price  model.Number
	Amount model.Number
}

// LevelProvider returns the levels for the given mid price, which controls the spread and number of levels
type LevelProvider interface {
	GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]Level, error)
	GetFillHandlers() ([]FillHandler, error)
}
