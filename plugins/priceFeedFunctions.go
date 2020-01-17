package plugins

import (
	"fmt"

	"github.com/stellar/kelp/api"
)

type fnFactory func(feeds []api.PriceFeed) (api.PriceFeed, error)

var fnFactoryMap = map[string]fnFactory{
	"max":    max,
	"invert": invert,
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
				return 0.0, fmt.Errorf("error fetching price from feed (index=%d) in 'max' function feed: %s", i, e)
			}

			if innerPrice <= 0.0 {
				return 0.0, fmt.Errorf("inner price of feed at index %d was <= 0.0 (%.10f)", i, innerPrice)
			}

			if innerPrice > max {
				max = innerPrice
			}
		}
		return max, nil
	}), nil
}

func invert(feeds []api.PriceFeed) (api.PriceFeed, error) {
	if len(feeds) != 1 {
		return nil, fmt.Errorf("need to provide exactly 1 price feed to the 'invert' function but found %d price feeds", len(feeds))
	}

	return makeFunctionFeed(func() (float64, error) {
		innerPrice, e := feeds[0].GetPrice()
		if e != nil {
			return 0.0, fmt.Errorf("error fetching price from feed in 'invert' function feed: %s", e)
		}

		if innerPrice <= 0.0 {
			return 0.0, fmt.Errorf("inner price of feed was <= 0.0 (%.10f)", innerPrice)
		}

		return 1 / innerPrice, nil
	}), nil
}
