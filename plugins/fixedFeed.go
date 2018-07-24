package plugins

import (
	"strconv"
)

// fixedFeed represents a fixed feed
type fixedFeed struct {
	price float64
}

func newFixedFeed(url string) *fixedFeed {
	m := new(fixedFeed)
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
