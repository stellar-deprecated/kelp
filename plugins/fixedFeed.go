package plugins

import (
	"fmt"
	"strconv"
)

// fixedFeed represents a fixed feed
type fixedFeed struct {
	price float64
}

func newFixedFeed(url string) (*fixedFeed, error) {
	m := new(fixedFeed)
	pA, e := strconv.ParseFloat(url, 64)
	if e != nil {
		return nil, fmt.Errorf("unable to parse float: %s", e)
	}

	m.price = pA
	return m, nil
}

// GetPrice impl
func (f *fixedFeed) GetPrice() (float64, error) {
	return f.price, nil
}
