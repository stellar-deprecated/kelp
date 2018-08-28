package api

import (
	"github.com/lightyeario/kelp/model"
	"github.com/stellar/go/clients/horizon"
)

// Level represents a layer in the orderbook
type Level struct {
	Price  model.Number
	Amount model.Number
}

// LevelProvider returns the levels for the given center price, which controls the spread and number of levels
type LevelProvider interface {
	GetLevels(maxAssetBase float64, maxAssetQuote float64, buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]Level, error)
}
