package level

import (
	"log"
	"math/rand"
	"time"
)

// autonomousLevelProvider provides levels based on an exponential curve wrt. the number of assets held in the account.
// This strategy does not allow using the balance of a single asset for more strategies other than this one because
// that would require building in some trade tracking along with asset balance tracking for this strategy. The support
// for this can always be added later.
type autonomousLevelProvider struct {
	spread                        float64
	plateauThresholdPercentage    float64 // flattens price if any asset has this ratio of the total number of tokens
	useMaxQuoteInTargetAmountCalc bool    // else use maxBase
	minAmountSpread               float64 // % that we take off the top of each amount order size which effectively serves as our spread when multiple levels are consumed
	maxAmountSpread               float64 // % that we take off the top of each amount order size which effectively serves as our spread when multiple levels are consumed
	maxLevels                     int16
	levelDensity                  float64
	ensureFirstNLevels            int16   // always adds the first N levels, meaningless if levelDensity = 1.0
	minAmountCarryoverSpread      float64 // the minimum spread % we take off the amountCarryover before placing the orders
	maxAmountCarryoverSpread      float64 // the maximum spread % we take off the amountCarryover before placing the orders
	carryoverInclusionProbability float64 // probability of including the carryover at a level that will be added

	// precomputed before construction
	randGen          *rand.Rand
	diffAmountSpread float64 // maxAmountSpread - minAmountSpread
}

// ensure it implements Provider
var _ Provider = &autonomousLevelProvider{}

// MakeAutonomousLevelProvider is the factory method
func MakeAutonomousLevelProvider(
	spread float64,
	plateauThresholdPercentage float64,
	useMaxQuoteInTargetAmountCalc bool,
	minAmountSpread float64,
	maxAmountSpread float64,
	maxLevels int16,
	levelDensity float64,
	ensureFirstNLevels int16,
	minAmountCarryoverSpread float64,
	maxAmountCarryoverSpread float64,
	carryoverInclusionProbability float64,
) Provider {
	validateSpread(minAmountSpread)
	validateSpread(maxAmountSpread)
	if minAmountSpread > maxAmountSpread {
		log.Fatalf("minAmountSpread (%.7f) needs to be <= maxAmountSpread (%.7f)\n", minAmountSpread, maxAmountSpread)
	}
	validateSpread(minAmountCarryoverSpread)
	validateSpread(maxAmountCarryoverSpread)
	if minAmountCarryoverSpread > maxAmountCarryoverSpread {
		log.Fatalf("minAmountCarryoverSpread (%.7f) needs to be <= maxAmountCarryoverSpread (%.7f)\n", minAmountCarryoverSpread, maxAmountCarryoverSpread)
	}
	// carryoverInclusionProbability is a value between 0 and 1
	validateSpread(carryoverInclusionProbability)

	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	diffAmountSpread := maxAmountSpread - minAmountSpread
	return &autonomousLevelProvider{
		spread: spread,
		plateauThresholdPercentage:    plateauThresholdPercentage,
		useMaxQuoteInTargetAmountCalc: useMaxQuoteInTargetAmountCalc,
		minAmountSpread:               minAmountSpread,
		maxAmountSpread:               maxAmountSpread,
		maxLevels:                     maxLevels,
		levelDensity:                  levelDensity,
		ensureFirstNLevels:            ensureFirstNLevels,
		minAmountCarryoverSpread:      minAmountCarryoverSpread,
		maxAmountCarryoverSpread:      maxAmountCarryoverSpread,
		carryoverInclusionProbability: carryoverInclusionProbability,
		randGen:          randGen,
		diffAmountSpread: diffAmountSpread,
	}
}

func validateSpread(spread float64) {
	if spread > 1.0 || spread < 0.0 {
		log.Fatalf("spread values need to be inclusively between 0 and 1: %.7f\n", spread)
	}
}

// GetLevels impl.
func (p *autonomousLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]Level, error) {
	_maxAssetBase := maxAssetBase
	_maxAssetQuote := maxAssetQuote
	// represents the amount that was meant to be included in a previous level that we excluded because we skipped that level
	amountCarryover := 0.0
	levels := []Level{}
	for i := int16(0); i < p.maxLevels; i++ {
		level, e := p.getLevel(_maxAssetBase, _maxAssetQuote)
		if e != nil {
			return nil, e
		}

		// always update _maxAssetBase and _maxAssetQuote to account for the level we just calculated, ensures price moves across levels regardless of inclusion of prior levels
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

		// decide whether to add this level
		includeLevelUsingProbability := p.randGen.Float64() < p.levelDensity
		includeLevelUsingConstraint := i < p.ensureFirstNLevels
		includeLevel := includeLevelUsingConstraint || includeLevelUsingProbability
		if !includeLevel {
			// accummulate targetAmount into amountCarryover
			amountCarryover += level.TargetAmount()
			continue
		}

		includeCarryover := p.randGen.Float64() < p.carryoverInclusionProbability
		if includeCarryover {
			// take a spread off the amountCarryover, clamping to min/max values
			amountCarryoverSpread := p.randGen.Float64()
			if amountCarryoverSpread < p.minAmountCarryoverSpread {
				amountCarryoverSpread = p.minAmountCarryoverSpread
			} else if amountCarryoverSpread > p.maxAmountCarryoverSpread {
				amountCarryoverSpread = p.maxAmountCarryoverSpread
			}
			amountCarryover = (1 - amountCarryoverSpread) * amountCarryover

			// include a partial amount of the carryover after we took the spread
			amountCarryoverToInclude := p.randGen.Float64() * amountCarryover
			// update amountCarryover to reflect inclusion in the level
			amountCarryover -= amountCarryoverToInclude
			// include the amountCarryover we computed, price of the level remains unchanged
			level = Level{
				targetPrice:  level.TargetPrice(),
				targetAmount: level.TargetAmount() + amountCarryoverToInclude,
			}
		}

		// always append the level here
		levels = append(levels, level)
	}
	return levels, nil
}

// getAmountSpread returns a random value between minAmountSpread (inclusive) and maxAmountSpread (exclusive)
func (p *autonomousLevelProvider) getAmountSpread() float64 {
	// generates a float between 0 and 1
	randFloat := p.randGen.Float64()
	// reduce to a float between 0 and diffAmountSpread
	spreadAboveMin := p.diffAmountSpread * randFloat
	// convert to a float between minAmountSpread and maxAmountSpread
	return p.minAmountSpread + spreadAboveMin
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
	targetAmount *= (1 - p.getAmountSpread())
	level := Level{
		targetPrice:  targetPrice,
		targetAmount: targetAmount,
	}
	return level, nil
}
