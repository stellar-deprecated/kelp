package plugins

import (
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

// swingConfig contains the configuration params for this Strategy
type swingConfig struct {
	PriceTolerance     float64 `valid:"-" toml:"PRICE_TOLERANCE"`
	AmountTolerance    float64 `valid:"-" toml:"AMOUNT_TOLERANCE"`
	AmountBaseBuy      float64 `valid:"-" toml:"AMOUNT_BASE_BUY"`
	AmountBaseSell     float64 `valid:"-" toml:"AMOUNT_BASE_SELL"`
	Spread             float64 `valid:"-" toml:"SPREAD"`                // this is the bid-ask spread (i.e. it is not the spread from the center price)
	OffsetSpread       float64 `valid:"-" toml:"OFFSET_SPREAD"`         // this is the bid-ask spread at the same last price level (i.e. it is not the spread from the center price)
	MaxLevels          int16   `valid:"-" toml:"MAX_LEVELS"`            // max number of levels to have on either side
	SeedLastTradePrice float64 `valid:"-" toml:"SEED_LAST_TRADE_PRICE"` // price with which to start off as the last trade price (i.e. initial center price)
	MinPrice           float64 `valid:"-" toml:"MIN_PRICE"`             // min price for which to place an order
	MaxPrice           float64 `valid:"-" toml:"MAX_PRICE"`             // max price for which to place an order
	MinBase            float64 `valid:"-" toml:"MIN_BASE"`
	MinQuote           float64 `valid:"-" toml:"MIN_QUOTE"`
	LastTradeCursor    string  `valid:"-" toml:"LAST_TRADE_CURSOR"`
}

// String impl.
func (c swingConfig) String() string {
	return utils.StructString(c, 0, nil)
}

// makeSwingStrategy is a factory method for swingStrategy
func makeSwingStrategy(
	sdex *SDEX,
	ieif *IEIF,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	config *swingConfig,
	tradeFetcher api.TradeFetcher,
	tradingPair *model.TradingPair,
	incrementTimestampCursor bool, // only do this if we are on ccxt
) api.Strategy {
	if config.AmountTolerance != 1.0 {
		panic("swing strategy needs to be configured with AMOUNT_TOLERANCE = 1.0")
	}

	orderConstraints := sdex.GetOrderConstraints(tradingPair)
	sellLevelProvider := makeSwingLevelProvider(
		config.Spread,
		config.OffsetSpread,
		false,
		config.AmountBaseSell,
		config.MaxLevels,
		config.SeedLastTradePrice,
		config.MaxPrice,
		config.MinBase,
		tradeFetcher,
		tradingPair,
		config.LastTradeCursor,
		incrementTimestampCursor,
	)
	sellSideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		ieif,
		assetBase,
		assetQuote,
		sellLevelProvider,
		config.PriceTolerance,
		config.AmountTolerance,
		false,
	)
	buyLevelProvider := makeSwingLevelProvider(
		config.Spread,
		config.OffsetSpread,
		true, // real base is passed in as quote so pass in true
		config.AmountBaseBuy,
		config.MaxLevels,
		config.SeedLastTradePrice, // we don't invert seed last trade price for the buy side because it's handeld in the swingLevelProvider
		config.MinPrice,           // use minPrice for buy side
		config.MinQuote,           // use minQuote for buying side
		tradeFetcher,
		tradingPair,
		config.LastTradeCursor,
		incrementTimestampCursor,
	)
	// switch sides of base/quote here for buy side
	buySideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		ieif,
		assetQuote,
		assetBase,
		buyLevelProvider,
		config.PriceTolerance,
		config.AmountTolerance,
		true,
	)

	return makeComposeStrategy(
		assetBase,
		assetQuote,
		buySideStrategy,
		sellSideStrategy,
	)
}
