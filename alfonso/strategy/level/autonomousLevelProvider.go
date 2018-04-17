package level

// autonomousLevelProvider provides levels based on an exponential curve wrt. the number of assets held in the account.
// This strategy does not allow using the balance of a single asset for more strategies other than this one because
// that would require building in some trade tracking along with asset balance tracking for this strategy. The support
// for this can always be added later.
type autonomousLevelProvider struct {
	// TODO 2 - need to add a price feed to peg the relative value of assets that are not relatively equal
	// TODO 2 - spread should be calculated to be inversely proportional to total sum of balances on either side (after taking the multiplicative effect of the priceFeed into account)
	spread                     float64
	plateauThresholdPercentage float64 // flattens price if any asset has this ratio of the total number of tokens
}

// ensure it implements Provider
var _ Provider = &autonomousLevelProvider{}

// MakeAutonomousLevelProvider is the factory method
func MakeAutonomousLevelProvider(spread float64, plateauThresholdPercentage float64) Provider {
	return &autonomousLevelProvider{
		spread: spread,
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

	targetPrice := centerPrice * (1 + p.spread)
	targetAmount := (2 * maxAssetBase * p.spread) / (4 + p.spread)
	// since targetAmount needs to be less then what we've set above based on the inequality formula, let's reduce it by 5%
	targetAmount *= 0.95
	level := Level{
		targetPrice:  targetPrice,
		targetAmount: targetAmount,
	}
	return []Level{level}, nil
}
