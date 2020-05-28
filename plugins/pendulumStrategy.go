package plugins

import (
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

// pendulumConfig contains the configuration params for this Strategy
type pendulumConfig struct {
	PriceTolerance     float64 `valid:"-" toml:"PRICE_TOLERANCE"`
	AmountTolerance    float64 `valid:"-" toml:"AMOUNT_TOLERANCE"`
	AmountBaseBuy      float64 `valid:"-" toml:"AMOUNT_BASE_BUY"`
	AmountBaseSell     float64 `valid:"-" toml:"AMOUNT_BASE_SELL"`
	Spread             float64 `valid:"-" toml:"SPREAD"`                // this is the bid-ask spread (i.e. it is not the spread from the center price)
	MaxLevels          int16   `valid:"-" toml:"MAX_LEVELS"`            // max number of levels to have on either side
	SeedLastTradePrice float64 `valid:"-" toml:"SEED_LAST_TRADE_PRICE"` // price with which to start off as the last trade price (i.e. initial center price)
	MaxPrice           float64 `valid:"-" toml:"MAX_PRICE"`             // max price for which to place an order
	MinPrice           float64 `valid:"-" toml:"MIN_PRICE"`             // min price for which to place an order
	MinBase            float64 `valid:"-" toml:"MIN_BASE"`
	MinQuote           float64 `valid:"-" toml:"MIN_QUOTE"`
	LastTradeCursor    string  `valid:"-" toml:"LAST_TRADE_CURSOR"`
}

/*
Note on Spread vs. OffsetSpread (hardcoded for now)
# define the bid/ask spread that you are willing to provide. spread is a percentage specified as a decimal number (0 < spread < 1.00) - here it is 0.1%
# How to set these spread levels:
#     - The spread between an oscillating buy and sell is spread + offset_spread - spread/2 = offset_spread + spread/2
#         (spread/2 subtracted because price shifts by spread/2 when trade made!)
#     - You need to account for 2x fee because a buy and a sell will take the fee on both buy and sell
#     - You need to set both values so that you are not buying and selling at the same price levels
# As an example:
#     if the fees on the exchange is 0.10% and you want to break even on every trade then set spread to 0.0020 and offset_spread to 0.0010
#     this will ensure that the oscillating spread os 0.20%, so there is no net loss for every trade
#     (0.20% + 0.10% - 0.20%/2 = 0.20% + 0.10% - 0.10% = 0.20%)
# SPREAD - this is the difference between each level on the same side, a smaller value here means subsequent levels will be closer together
# OFFSET_SPREAD - this is the difference between the buy and sell at the same logical price level when they do overlap

For now we hardcode offsetSpread to be 0.5 * spread to keep it less confusing for users
*/

// String impl.
func (c pendulumConfig) String() string {
	return utils.StructString(c, 0, nil)
}

// makePendulumStrategy is a factory method for pendulumStrategy
func makePendulumStrategy(
	sdex *SDEX,
	exchangeShim api.ExchangeShim,
	ieif *IEIF,
	assetBase *hProtocol.Asset,
	assetQuote *hProtocol.Asset,
	config *pendulumConfig,
	tradeFetcher api.TradeFetcher,
	tradingPair *model.TradingPair,
	incrementTimestampCursor bool, // only do this if we are on ccxt
) api.Strategy {
	if config.AmountTolerance != 1.0 {
		panic("pendulum strategy needs to be configured with AMOUNT_TOLERANCE = 1.0")
	}

	orderConstraints := exchangeShim.GetOrderConstraints(tradingPair)
	sellLevelProvider := makePendulumLevelProvider(
		config.Spread,
		config.Spread/2,
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
		orderConstraints,
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
	buyLevelProvider := makePendulumLevelProvider(
		config.Spread,
		config.Spread/2,
		true, // real base is passed in as quote so pass in true
		config.AmountBaseBuy,
		config.MaxLevels,
		config.SeedLastTradePrice, // we don't invert seed last trade price for the buy side because it's handeld in the pendulumLevelProvider
		config.MinPrice,           // use minPrice for buy side
		config.MinQuote,           // use minQuote for buying side
		tradeFetcher,
		tradingPair,
		config.LastTradeCursor,
		incrementTimestampCursor,
		orderConstraints,
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
