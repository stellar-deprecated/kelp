package plugins

import (
	"fmt"
	"log"
	"os"

	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/model"
	"github.com/interstellar/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
)

// stopLimitLevelProvider generates a level if the stop condition is met
type stopLimitLevelProvider struct {
	sdex             *SDEX
	assetBase        *horizon.Asset
	assetQuote       *horizon.Asset
	amountOfBase     float64
	stopPrice        float64
	limitPrice       float64
	orderFilled      bool // tracks whether the order was filled (at any amount)
	orderConstraints *model.OrderConstraints
}

// ensure it implements LevelProvider
var _ api.LevelProvider = &stopLimitLevelProvider{}

// ensure this implements api.FillHandler
var _ api.FillHandler = &balancedLevelProvider{}

// makeStopLimitLevelProvider is the factory method
func makeStopLimitLevelProvider(
	sdex *SDEX,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	amountOfBase float64,
	stopPrice float64,
	limitPrice float64,
	orderConstraints *model.OrderConstraints,
) api.LevelProvider {
	orderFilled := false
	return &stopLimitLevelProvider{
		sdex:             sdex,
		assetBase:        assetBase,
		assetQuote:       assetQuote,
		amountOfBase:     amountOfBase,
		stopPrice:        stopPrice,
		limitPrice:       limitPrice,
		orderFilled:      orderFilled,
		orderConstraints: orderConstraints,
	}
}

// GetLevels impl.
func (p *stopLimitLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]api.Level, error) {
	if p.orderFilled {
		log.Println("the order was placed and filled, exiting")
		os.Exit(0)
	}

	if p.amountOfBase > maxAssetBase {
		return nil, fmt.Errorf("account balance is less than specified amount order")
	}

	levels := []api.Level{}
	topBidPrice, e := getTopBid(p.sdex, p.assetBase, p.assetQuote)
	if e != nil {
		return nil, fmt.Errorf("unable to get top bid from SDEX")
	}

	if topBidPrice <= p.stopPrice {
		level, e := p.getLevel()
		if e != nil {
			return nil, fmt.Errorf("unable to generate the order level")
		}
		levels = append(levels, level)
		log.Println("stop was triggered, placing order")
		return levels, nil
	}
	log.Println("stop was not triggered")
	return nil, nil
}

func (p *stopLimitLevelProvider) getLevel() (api.Level, error) {
	targetPrice := p.limitPrice
	targetAmount := p.amountOfBase
	level := api.Level{
		Price:  *model.NumberFromFloat(targetPrice, p.orderConstraints.PricePrecision),
		Amount: *model.NumberFromFloat(targetAmount, p.orderConstraints.VolumePrecision),
	}
	return level, nil
}

// GetFillHandlers impl
func (p *stopLimitLevelProvider) GetFillHandlers() ([]api.FillHandler, error) {
	return []api.FillHandler{p}, nil
}

// HandleFill impl
func (p *stopLimitLevelProvider) HandleFill(trade model.Trade) error {
	log.Println("the order was taken, will exit next cycle")
	p.orderFilled = true
	return nil
}

func getTopBid(sdex *SDEX, assetBase *horizon.Asset, assetQuote *horizon.Asset) (float64, error) {
	orderBook, e := utils.GetOrderBook(sdex.API, assetBase, assetQuote)
	if e != nil {
		return 0, e
	}
	bids := orderBook.Bids
	topBidPrice := utils.PriceAsFloat(bids[0].Price)
	return topBidPrice, nil
}
