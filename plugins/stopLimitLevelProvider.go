package plugins

import (
	"fmt"
	"log"

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
	amountOfA        float64
	stopPrice        float64
	limitPrice       float64
	orderFilled      bool // tracks whether the order was filled (at any amount)
	orderConstraints *model.OrderConstraints
}

// ensure it implements LevelProvider
var _ api.LevelProvider = &stopLimitLevelProvider{}

// GetLevels impl.
func (p *stopLimitLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]api.Level, error) {

	if p.orderFilled {
		log.Fatal("the order was filled (at least partially), exiting")
	}

	if p.amountOfA > maxAssetBase {
		return nil, fmt.Errorf("account balance is less than specified amount order")
	}

	levels := []api.Level{}
	topBidPrice, e := getTopBid(p.sdex, p.assetBase, p.assetQuote)

	if topBidPrice <= p.stopPrice {
		level, e := p.getLevel()
		levels = append(levels, level)
		return levels, nil
	}
	return nil, nil
}

func (p *stopLimitLevelProvider) getLevel() (api.Level, error) {
	targetPrice := p.limitPrice
	targetAmount := p.amountOfA
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
