package plugins

import (
	"fmt"
	"log"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

// sellTwapLevelProvider provides a fixed number of levels using a static percentage spread
type sellTwapLevelProvider struct {
	staticLevels     []StaticLevel
	amountOfBase     float64
	offset           rateOffset
	pf               *api.FeedPair
	orderConstraints *model.OrderConstraints
}

// ensure it implements the LevelProvider interface
var _ api.LevelProvider = &sellTwapLevelProvider{}

// makeSellTwapLevelProvider is a factory method
func makeSellTwapLevelProvider(
	staticLevels []StaticLevel,
	amountOfBase float64,
	offset rateOffset,
	pf *api.FeedPair,
	orderConstraints *model.OrderConstraints,
) api.LevelProvider {
	return &sellTwapLevelProvider{
		staticLevels:     staticLevels,
		amountOfBase:     amountOfBase,
		offset:           offset,
		pf:               pf,
		orderConstraints: orderConstraints,
	}
}

// GetLevels impl.
func (p *sellTwapLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]api.Level, error) {
	midPrice, e := p.pf.GetFeedPairPrice()
	if e != nil {
		return nil, fmt.Errorf("mid price couldn't be loaded: %s", e)
	}
	if p.offset.percent != 0.0 || p.offset.absolute != 0 {
		// if inverted, we want to invert before we compute the adjusted price, and then invert back
		if p.offset.invert {
			midPrice = 1 / midPrice
		}
		scaleFactor := 1 + p.offset.percent
		if p.offset.percentFirst {
			midPrice = (midPrice * scaleFactor) + p.offset.absolute
		} else {
			midPrice = (midPrice + p.offset.absolute) * scaleFactor
		}
		if p.offset.invert {
			midPrice = 1 / midPrice
		}
		log.Printf("mid price (adjusted): %.7f\n", midPrice)
	}

	levels := []api.Level{}
	for _, sl := range p.staticLevels {
		absoluteSpread := midPrice * sl.SPREAD
		levels = append(levels, api.Level{
			// we always add here because it is only used in the context of selling so we always charge a higher price to include a spread
			Price:  *model.NumberFromFloat(midPrice+absoluteSpread, p.orderConstraints.PricePrecision),
			Amount: *model.NumberFromFloat(sl.AMOUNT*p.amountOfBase, p.orderConstraints.VolumePrecision),
		})
	}
	return levels, nil
}

// GetFillHandlers impl
func (p *sellTwapLevelProvider) GetFillHandlers() ([]api.FillHandler, error) {
	return nil, nil
}
