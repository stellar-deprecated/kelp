package plugins

import (
	"fmt"
	"log"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

// StaticLevel represents a layer in the orderbook defined statically
// extracted here because it's shared by strategy and sideStrategy where strategy depeneds on sideStrategy
type StaticLevel struct {
	SPREAD float64 `valid:"-" json:"spread"`
	AMOUNT float64 `valid:"-" json:"amount"`
}

// how much to offset your rates by. Can use percent and offset together.
// A positive value indicates that your base asset (ASSET_A) has a higher rate than the rate received from your price feed
// A negative value indicates that your base asset (ASSET_A) has a lower rate than the rate received from your price feed
type rateOffset struct {
	// specified as a decimal (ex: 0.05 = 5%).
	percent float64
	// specified as a decimal.
	absolute float64

	// specifies the order in which to offset the rates. If true then we apply the RATE_OFFSET_PERCENT first otherwise we apply the RATE_OFFSET first.
	// example rate calculation when set to true: ((rate_from_price_feed_a/rate_from_price_feed_b) * (1 + rate_offset_percent)) + rate_offset
	// example rate calculation when set to false: ((rate_from_price_feed_a/rate_from_price_feed_b) + rate_offset) * (1 + rate_offset_percent)
	percentFirst bool

	// set this to true if buying
	invert bool
}

// staticSpreadLevelProvider provides a fixed number of levels using a static percentage spread
type staticSpreadLevelProvider struct {
	staticLevels     []StaticLevel
	amountOfBase     float64
	offset           rateOffset
	pf               *api.FeedPair
	orderConstraints *model.OrderConstraints
}

// ensure it implements the LevelProvider interface
var _ api.LevelProvider = &staticSpreadLevelProvider{}

// makeStaticSpreadLevelProvider is a factory method
func makeStaticSpreadLevelProvider(staticLevels []StaticLevel, amountOfBase float64, offset rateOffset, pf *api.FeedPair, orderConstraints *model.OrderConstraints) api.LevelProvider {
	return &staticSpreadLevelProvider{
		staticLevels:     staticLevels,
		amountOfBase:     amountOfBase,
		offset:           offset,
		pf:               pf,
		orderConstraints: orderConstraints,
	}
}

// GetLevels impl.
func (p *staticSpreadLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]api.Level, error) {
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
func (p *staticSpreadLevelProvider) GetFillHandlers() ([]api.FillHandler, error) {
	return nil, nil
}
