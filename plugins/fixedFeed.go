package plugins

import (
	"strconv"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/support/logger"
)

// fixedFeed represents a fixed feed
type fixedFeed struct {
	price float64
	l     logger.Logger
}

// ensure that it implements PriceFeed
var _ api.PriceFeed = &fixedFeed{}

func newFixedFeed(url string, l logger.Logger) *fixedFeed {
	m := new(fixedFeed)
	m.l = l
	pA, err := strconv.ParseFloat(url, 64)
	if err != nil {
		return nil
	}

	m.price = pA
	return m
}

// GetPrice impl
func (f *fixedFeed) GetPrice() (float64, error) {
	return f.price, nil
}
