package plugins

import (
	"fmt"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

// BuySellConfig contains the configuration params for this strategy
type BuySellConfig struct {
	PriceTolerance         float64       `valid:"-" toml:"PRICE_TOLERANCE" json:"price_tolerance"`
	AmountTolerance        float64       `valid:"-" toml:"AMOUNT_TOLERANCE" json:"amount_tolerance"`
	RateOffsetPercent      float64       `valid:"-" toml:"RATE_OFFSET_PERCENT" json:"rate_offset_percent"`
	RateOffset             float64       `valid:"-" toml:"RATE_OFFSET" json:"rate_offset"`
	RateOffsetPercentFirst bool          `valid:"-" toml:"RATE_OFFSET_PERCENT_FIRST" json:"rate_offset_percent_first"`
	AmountOfABase          float64       `valid:"-" toml:"AMOUNT_OF_A_BASE" json:"amount_of_a_base"` // the size of order to keep on either side
	DataTypeA              string        `valid:"-" toml:"DATA_TYPE_A" json:"data_type_a"`
	DataFeedAURL           string        `valid:"-" toml:"DATA_FEED_A_URL" json:"data_feed_a_url"`
	DataTypeB              string        `valid:"-" toml:"DATA_TYPE_B" json:"data_type_b"`
	DataFeedBURL           string        `valid:"-" toml:"DATA_FEED_B_URL" json:"data_feed_b_url"`
	Levels                 []StaticLevel `valid:"-" toml:"LEVELS" json:"levels"`
}

// MakeBuysellConfig factory method
func MakeBuysellConfig(
	priceTolerance float64,
	amountTolerance float64,
	rateOffsetPercent float64,
	rateOffset float64,
	rateOffsetPercentFirst bool,
	amountOfABase float64,
	dataTypeA string,
	dataFeedAURL string,
	dataTypeB string,
	dataFeedBURL string,
	levels []StaticLevel,
) *BuySellConfig {
	return &BuySellConfig{
		PriceTolerance:         priceTolerance,
		AmountTolerance:        amountTolerance,
		RateOffsetPercent:      rateOffsetPercent,
		RateOffset:             rateOffset,
		RateOffsetPercentFirst: rateOffsetPercentFirst,
		AmountOfABase:          amountOfABase,
		DataTypeA:              dataTypeA,
		DataFeedAURL:           dataFeedAURL,
		DataTypeB:              dataTypeB,
		DataFeedBURL:           dataFeedBURL,
		Levels:                 levels,
	}
}

// String impl.
func (c BuySellConfig) String() string {
	return utils.StructString(c, 0, nil)
}

// makeBuySellStrategy is a factory method
func makeBuySellStrategy(
	sdex *SDEX,
	pair *model.TradingPair,
	ieif *IEIF,
	assetBase *hProtocol.Asset,
	assetQuote *hProtocol.Asset,
	config *BuySellConfig,
) (api.Strategy, error) {
	offsetSell := rateOffset{
		percent:      config.RateOffsetPercent,
		absolute:     config.RateOffset,
		percentFirst: config.RateOffsetPercentFirst,
	}
	sellSideFeedPair, e := MakeFeedPair(
		config.DataTypeA,
		config.DataFeedAURL,
		config.DataTypeB,
		config.DataFeedBURL,
	)
	if e != nil {
		return nil, fmt.Errorf("cannot make the buysell strategy because we could not make the sell side feed pair: %s", e)
	}
	orderConstraints := sdex.GetOrderConstraints(pair)
	sellSideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		ieif,
		assetBase,
		assetQuote,
		makeStaticSpreadLevelProvider(
			config.Levels,
			config.AmountOfABase,
			offsetSell,
			sellSideFeedPair,
			orderConstraints,
		),
		config.PriceTolerance,
		config.AmountTolerance,
		false,
	)

	offsetBuy := rateOffset{
		percent:      config.RateOffsetPercent,
		absolute:     config.RateOffset,
		percentFirst: config.RateOffsetPercentFirst,
		invert:       true,
	}
	buySideFeedPair, e := MakeFeedPair(
		config.DataTypeB,
		config.DataFeedBURL,
		config.DataTypeA,
		config.DataFeedAURL,
	)
	if e != nil {
		return nil, fmt.Errorf("cannot make the buysell strategy because we could not make the buy side feed pair: %s", e)
	}
	// switch sides of base/quote here for buy side
	buySideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		ieif,
		assetQuote,
		assetBase,
		makeStaticSpreadLevelProvider(
			config.Levels,
			config.AmountOfABase,
			offsetBuy,
			buySideFeedPair,
			orderConstraints,
		),
		config.PriceTolerance,
		config.AmountTolerance,
		true,
	)

	return makeComposeStrategy(
		assetBase,
		assetQuote,
		buySideStrategy,
		sellSideStrategy,
	), nil
}
