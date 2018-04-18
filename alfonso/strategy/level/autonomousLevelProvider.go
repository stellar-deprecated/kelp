package level

import "log"

// autonomousLevelProvider provides levels based on an exponential curve wrt. the number of assets held in the account.
// This strategy does not allow using the balance of a single asset for more strategies other than this one because
// that would require building in some trade tracking along with asset balance tracking for this strategy. The support
// for this can always be added later.
type autonomousLevelProvider struct {
	spread                        float64
	plateauThresholdPercentage    float64 // flattens price if any asset has this ratio of the total number of tokens
	useMaxQuoteInTargetAmountCalc bool    // else use maxBase
	amountSpread                  float64 // % that we take off the top of each amount order size which effectively serves as our spread when multiple levels are consumed
	levels                        int8
}

// ensure it implements Provider
var _ Provider = &autonomousLevelProvider{}

// MakeAutonomousLevelProvider is the factory method
func MakeAutonomousLevelProvider(spread float64, plateauThresholdPercentage float64, useMaxQuoteInTargetAmountCalc bool, amountSpread float64, levels int8) Provider {
	if amountSpread >= 1.0 || amountSpread <= 0.0 {
		log.Fatal("amountSpread needs to be between 0 and 1 (exclusive): ", amountSpread)
	}

	return &autonomousLevelProvider{
		spread: spread,
		plateauThresholdPercentage:    plateauThresholdPercentage,
		useMaxQuoteInTargetAmountCalc: useMaxQuoteInTargetAmountCalc,
		amountSpread:                  amountSpread,
		levels:                        levels,
	}
}

// GetLevels impl.
func (p *autonomousLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]Level, error) {
	_maxAssetBase := maxAssetBase
	_maxAssetQuote := maxAssetQuote
	levels := []Level{}
	for i := int8(0); i < p.levels; i++ {
		level, e := p.getLevel(_maxAssetBase, _maxAssetQuote)
		if e != nil {
			return nil, e
		}
		levels = append(levels, level)

		// targetPrice is always quote/base
		var baseDecreased float64
		var quoteIncreased float64
		if p.useMaxQuoteInTargetAmountCalc {
			// targetAmount is in quote so divide by price (quote/base) to give base
			baseDecreased = level.TargetAmount() / level.TargetPrice()
			// targetAmount is in quote so use directly
			quoteIncreased = level.TargetAmount()
		} else {
			// targetAmount is in base so use directly
			baseDecreased = level.TargetAmount()
			// targetAmount is in base so multiply by price (quote/base) to give quote
			quoteIncreased = level.TargetAmount() * level.TargetPrice()
		}

		// subtract because we had to sell that many units to reach the next level
		_maxAssetBase -= baseDecreased
		// add because we had to buy these many units to reach the next level
		_maxAssetQuote += quoteIncreased
	}
	return levels, nil
}

func (p *autonomousLevelProvider) getLevel(maxAssetBase float64, maxAssetQuote float64) (Level, error) {
	sum := maxAssetQuote + maxAssetBase
	var centerPrice float64
	if maxAssetQuote/sum >= p.plateauThresholdPercentage {
		centerPrice = p.plateauThresholdPercentage / (1 - p.plateauThresholdPercentage)
	} else if maxAssetBase/sum >= p.plateauThresholdPercentage {
		centerPrice = (1 - p.plateauThresholdPercentage) / p.plateauThresholdPercentage
	} else {
		centerPrice = maxAssetQuote / maxAssetBase
	}

	// price always adds the spread
	targetPrice := centerPrice * (1 + p.spread/2)

	targetAmount := (2 * maxAssetBase * p.spread) / (4 + p.spread)
	if p.useMaxQuoteInTargetAmountCalc {
		targetAmount = (2 * maxAssetQuote * p.spread) / (4 + p.spread)
	}
	// since targetAmount needs to be less then what we've set above based on the inequality formula, let's reduce it by 5%
	targetAmount *= (1 - p.amountSpread)
	level := Level{
		targetPrice:  targetPrice,
		targetAmount: targetAmount,
	}
	return level, nil
}
