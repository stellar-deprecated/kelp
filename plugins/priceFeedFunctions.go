package plugins

import (
	"fmt"

	"github.com/stellar/kelp/api"
)

type fnFactory func(feeds []api.PriceFeed) (api.PriceFeed, error)

var fnFactoryMap = map[string]fnFactory{
	"max": max,
}

func max(feeds []api.PriceFeed) (api.PriceFeed, error) {
	if len(feeds) < 2 {
		return nil, fmt.Errorf("need to provide at least 2 price feeds to the 'max' price feed function but found only %d price feeds", len(feeds))
	}

	return makeFunctionFeed(func() (float64, error) {
		max := -1.0
		for i, f := range feeds {
			innerPrice, e := f.GetPrice()
			if e != nil {
				return 0.0, fmt.Errorf("error fetching price from feed in 'max' function feed: %s", e)
			}

			if innerPrice <= 0.0 {
				return 0.0, fmt.Errorf("inner price of feed at index %d was <= 0.0", i)
			}

			if innerPrice > max {
				max = innerPrice
			}
		}
		return max, nil
	}), nil
}
