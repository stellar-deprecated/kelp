package plugins

import (
	"fmt"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
)

// trackSDEXConfig contains the configuration params for this strategy
type trackSDEXConfig struct {
	Spread                 float64 `valid:"-" toml:"SPREAD"`
	PriceTolerance         float64 `valid:"-" toml:"PRICE_TOLERANCE"`
	AmountTolerance        float64 `valid:"-" toml:"AMOUNT_TOLERANCE"`
	BasePercentPerLevel    float64 `valid:"-" toml:"BASE_PERCENT_PER_LEVEL"`
	MaxLevels              int16   `valid:"-" toml:"MAX_LEVELS"`
	MaintainBalancePercent float64 `valid:"-" toml:"MAINTAIN_BALANCE_PERCENT"`
}

// String impl.
func (c trackSDEXConfig) String() string {
	return utils.StructString(c, nil)
}

// makeTrackSDEXStrategy is a factory method
func makeTrackSDEXStrategy(
	sdex *SDEX,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	config *trackSDEXConfig,
) (api.Strategy, error) {
	sdexMidPrice, e := GetSDEXPrice(
		sdex,
		assetBase,
		assetQuote,
	)
	if e != nil {
		return nil, fmt.Errorf("failed to get SDEX orderbook price", e)
	}
	sdexInversePrice := 1.0 / sdexMidPrice

	sellSideStrategy := makeSellSideStrategy(
		sdex,
		assetBase,
		assetQuote,
		makeSDEXLevelProvider(
			config.Spread,
			config.BasePercentPerLevel,
			config.MaxLevels,
			config.MaintainBalancePercent,
			sdexMidPrice,
			false,
		),
		config.PriceTolerance,
		config.AmountTolerance,
		false,
	)
	// switch sides of base/quote here for buy side
	buySideStrategy := makeSellSideStrategy(
		sdex,
		assetQuote,
		assetBase,
		makeSDEXLevelProvider(
			config.Spread,
			config.BasePercentPerLevel,
			config.MaxLevels,
			config.MaintainBalancePercent,
			sdexInversePrice,
			true,
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
