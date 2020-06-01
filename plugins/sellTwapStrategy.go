package plugins

import (
	"fmt"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

// sellTwapConfig contains the configuration params for this Strategy
type sellTwapConfig struct {
	DataTypeA              string        `valid:"-" toml:"DATA_TYPE_A"`
	DataFeedAURL           string        `valid:"-" toml:"DATA_FEED_A_URL"`
	DataTypeB              string        `valid:"-" toml:"DATA_TYPE_B"`
	DataFeedBURL           string        `valid:"-" toml:"DATA_FEED_B_URL"`
	PriceTolerance         float64       `valid:"-" toml:"PRICE_TOLERANCE"`
	AmountTolerance        float64       `valid:"-" toml:"AMOUNT_TOLERANCE"`
	AmountOfABase          float64       `valid:"-" toml:"AMOUNT_OF_A_BASE"` // the size of order
	RateOffsetPercent      float64       `valid:"-" toml:"RATE_OFFSET_PERCENT"`
	RateOffset             float64       `valid:"-" toml:"RATE_OFFSET"`
	RateOffsetPercentFirst bool          `valid:"-" toml:"RATE_OFFSET_PERCENT_FIRST"`
	Levels                 []StaticLevel `valid:"-" toml:"LEVELS"`
}

// String impl.
func (c sellTwapConfig) String() string {
	return utils.StructString(c, 0, nil)
}

// makeSellTwapStrategy is a factory method for SellStrategy
func makeSellTwapStrategy(
	sdex *SDEX,
	pair *model.TradingPair,
	ieif *IEIF,
	assetBase *hProtocol.Asset,
	assetQuote *hProtocol.Asset,
	config *sellTwapConfig,
) (api.Strategy, error) {
	pf, e := MakeFeedPair(
		config.DataTypeA,
		config.DataFeedAURL,
		config.DataTypeB,
		config.DataFeedBURL,
	)
	if e != nil {
		return nil, fmt.Errorf("cannot make the sell strategy because we could not make the feed pair: %s", e)
	}

	orderConstraints := sdex.GetOrderConstraints(pair)
	offset := rateOffset{
		percent:      config.RateOffsetPercent,
		absolute:     config.RateOffset,
		percentFirst: config.RateOffsetPercentFirst,
	}
	sellSideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		ieif,
		assetBase,
		assetQuote,
		makeSellTwapLevelProvider(config.Levels, config.AmountOfABase, offset, pf, orderConstraints),
		config.PriceTolerance,
		config.AmountTolerance,
		false,
	)
	// switch sides of base/quote here for the delete side
	deleteSideStrategy := makeDeleteSideStrategy(sdex, assetQuote, assetBase)

	return makeComposeStrategy(
		assetBase,
		assetQuote,
		deleteSideStrategy,
		sellSideStrategy,
	), nil
}
