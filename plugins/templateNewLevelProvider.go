package plugins

import (
	"fmt"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

// templateNewLevelProvider is a template that you can use to create your own trading strategies
type templateNewLevelProvider struct {
	priceFeedA       api.PriceFeed
	priceFeedB       api.PriceFeed
	priceFeedC       api.PriceFeed
	offset           rateOffset
	orderConstraints *model.OrderConstraints
}

// ensure it implements the LevelProvider interface
var _ api.LevelProvider = &templateNewLevelProvider{}

// makeTemplateNewLevelProvider is a factory method
func makeTemplateNewLevelProvider(
	priceFeedA api.PriceFeed,
	priceFeedB api.PriceFeed,
	priceFeedC api.PriceFeed,
	offset rateOffset,
	orderConstraints *model.OrderConstraints,
) (api.LevelProvider, error) {
	return &templateNewLevelProvider{
		priceFeedA:       priceFeedA,
		priceFeedB:       priceFeedB,
		priceFeedC:       priceFeedC,
		offset:           offset,
		orderConstraints: orderConstraints,
	}, nil
}

// GetLevels impl.
func (p *templateNewLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]api.Level, error) {
	// maxAssetBase is the current base asset balance
	// maxAssetQuote is the current quote asset balance

	/* ------- TODO add whatever logic you want for your trading strategy here ------- */
	// fetch three prices for three feeds
	priceA, e := p.priceFeedA.GetPrice()
	if e != nil {
		return nil, fmt.Errorf("could not fetch price from feed A")
	}
	priceB, e := p.priceFeedB.GetPrice()
	if e != nil {
		return nil, fmt.Errorf("could not fetch price from feed B")
	}
	priceC, e := p.priceFeedC.GetPrice()
	if e != nil {
		return nil, fmt.Errorf("could not fetch price from feed C")
	}

	// take max price of the three retrieved
	var price float64
	if priceA > priceB {
		price = priceA
	} else {
		price = priceB
	}
	if priceC > price {
		price = priceC
	}

	// use an amount in terms of the base asset
	amount := 100.0
	/* ------- /TODO add whatever logic you want for your trading strategy here ------- */

	// convert the outcome of your trading strategy calculations into levels here
	// a level is a price point on the orderbook. These should typically be higher than the mid price for maker orders
	// this is the outcome of all the calculations in your trading strategy. Once you return this, the framework will
	// handle how to transform your orders to these levels.
	return []api.Level{
		api.Level{
			Price:  *model.NumberFromFloat(price, p.orderConstraints.PricePrecision),
			Amount: *model.NumberFromFloat(amount, p.orderConstraints.PricePrecision),
		},
		api.Level{ // example to create a second level at a 10% higher price and twice the amount
			Price:  *model.NumberFromFloat(price*1.10, p.orderConstraints.PricePrecision),
			Amount: *model.NumberFromFloat(amount*2, p.orderConstraints.PricePrecision),
		},
		// ... create as many levels as you want here for more depth
	}, nil
}

// GetFillHandlers impl
func (p *templateNewLevelProvider) GetFillHandlers() ([]api.FillHandler, error) {
	return nil, nil
}
