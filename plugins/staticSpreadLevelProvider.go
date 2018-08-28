package plugins

import (
	"log"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
)

// staticLevel represents a layer in the orderbook defined statically
// extracted here because it's shared by strategy and sideStrategy where strategy depeneds on sideStrategy
type staticLevel struct {
	SPREAD float64 `valid:"-"`
	AMOUNT float64 `valid:"-"`
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
	staticLevels []staticLevel
	amountOfBase float64
	offset       rateOffset
	pf           *api.FeedPair
}

// ensure it implements the LevelProvider interface
var _ api.LevelProvider = &staticSpreadLevelProvider{}

// makeStaticSpreadLevelProvider is a factory method
func makeStaticSpreadLevelProvider(staticLevels []staticLevel, amountOfBase float64, offset rateOffset, pf *api.FeedPair) api.LevelProvider {
	return &staticSpreadLevelProvider{
		staticLevels: staticLevels,
		amountOfBase: amountOfBase,
		offset:       offset,
		pf:           pf,
	}
}

// GetLevels impl.
func (p *staticSpreadLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64, buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]api.Level, error) {
	centerPrice, e := p.pf.GetCenterPrice()
	if e != nil {
		log.Printf("error: center price couldn't be loaded! | %s\n", e)
		return nil, e
	}
	log.Printf("center price: %.7f\n", centerPrice)
	if p.offset.percent != 0.0 || p.offset.absolute != 0 {
		// if inverted, we want to invert before we compute the adjusted price, and then invert back
		if p.offset.invert {
			centerPrice = 1 / centerPrice
		}
		scaleFactor := 1 + p.offset.percent
		if p.offset.percentFirst {
			centerPrice = (centerPrice * scaleFactor) + p.offset.absolute
		} else {
			centerPrice = (centerPrice + p.offset.absolute) * scaleFactor
		}
		if p.offset.invert {
			centerPrice = 1 / centerPrice
		}
		log.Printf("center price (adjusted): %.7f\n", centerPrice)
	}

	levels := []api.Level{}
	for _, sl := range p.staticLevels {
		absoluteSpread := centerPrice * sl.SPREAD
		levels = append(levels, api.Level{
			// we always add here because it is only used in the context of selling so we always charge a higher price to include a spread
			Price:  *model.NumberFromFloat(centerPrice+absoluteSpread, utils.SdexPrecision),
			Amount: *model.NumberFromFloat(sl.AMOUNT*p.amountOfBase, utils.SdexPrecision),
		})
	}
	return levels, nil
}
