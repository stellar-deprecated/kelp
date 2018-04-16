package level

// autonomousLevelProvider provides levels based on an exponential curve wrt. the number of assets held in the account.
// This strategy does not allow using the balance of a single asset for more strategies other than this one because
// that would require building in some trade tracking along with asset balance tracking for this strategy. The support
// for this can always be added later.
type autonomousLevelProvider struct {
	// TODO 2 - need to add a price feed to peg the relative value of assets that are not relatively equal
	// TODO 2 - spread should be calculated to be inversely proportional to total sum of balances on either side (after taking the multiplicative effect of the priceFeed into account)
	spread                     float64
	maxLevels                  int8
	plateauThresholdPercentage float64 // flattens price if any asset has this ratio of the total number of tokens
}

// ensure it implements Provider
var _ Provider = &autonomousLevelProvider{}

// MakeAutonomousLevelProvider is the factory method
func MakeAutonomousLevelProvider(spread float64, maxLevels int8, plateauThresholdPercentage float64) Provider {
	return &autonomousLevelProvider{
		spread:                     spread,
		maxLevels:                  maxLevels,
		plateauThresholdPercentage: plateauThresholdPercentage,
	}
}

// GetLevels impl.
func (p *autonomousLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]Level, error) {
	sum := maxAssetQuote + maxAssetBase
	var centerPrice float64
	if maxAssetQuote/sum >= p.plateauThresholdPercentage {
		centerPrice = p.plateauThresholdPercentage / (1 - p.plateauThresholdPercentage)
	} else if maxAssetBase/sum >= p.plateauThresholdPercentage {
		centerPrice = (1 - p.plateauThresholdPercentage) / p.plateauThresholdPercentage
	} else {
		centerPrice = maxAssetQuote / maxAssetBase
	}

	levels := []Level{}
	for spreadMultiple := int8(1); spreadMultiple < p.maxLevels+1; spreadMultiple++ {
		levelSpread := p.spread * float64(spreadMultiple)
		targetPrice := centerPrice * (1 + levelSpread)

		// TODO 2 - need to add a better model for the targetAmount
		targetAmount := 4 * sum * p.spread
		levels = append(levels, Level{
			targetPrice:  targetPrice,
			targetAmount: targetAmount,
		})
	}
	return levels, nil
}
