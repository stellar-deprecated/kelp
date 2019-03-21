package plugins

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

// balancedLevelProvider provides levels based on an exponential curve wrt. the number of assets held in the account.
// This strategy does not allow using the balance of a single asset for more strategies other than this one because
// that would require building in some trade tracking along with asset balance tracking for this strategy. The support
// for this can always be added later.
type balancedLevelProvider struct {
	spread                        float64
	useMaxQuoteInTargetAmountCalc bool    // else use maxBase
	minAmountSpread               float64 // % that we take off the top of each amount order size which effectively serves as our spread when multiple levels are consumed
	maxAmountSpread               float64 // % that we take off the top of each amount order size which effectively serves as our spread when multiple levels are consumed
	maxLevels                     int16
	levelDensity                  float64
	ensureFirstNLevels            int16   // always adds the first N levels, meaningless if levelDensity = 1.0
	minAmountCarryoverSpread      float64 // the minimum spread % we take off the amountCarryover before placing the orders
	maxAmountCarryoverSpread      float64 // the maximum spread % we take off the amountCarryover before placing the orders
	carryoverInclusionProbability float64 // probability of including the carryover at a level that will be added
	virtualBalanceBase            float64 // virtual balance to use so we can smoothen out the curve
	virtualBalanceQuote           float64 // virtual balance to use so we can smoothen out the curve
	orderConstraints              *model.OrderConstraints
	shouldRefresh                 bool // boolean for whether to generate levels, starts true

	// precomputed before construction
	randGen *rand.Rand

	// uninitialized
	lastLevels []api.Level // keeps the levels generated on the previous run to use if no offers were taken
}

// ensure it implements LevelProvider
var _ api.LevelProvider = &balancedLevelProvider{}

// ensure this implements api.FillHandler
var _ api.FillHandler = &balancedLevelProvider{}

// makeBalancedLevelProvider is the factory method
func makeBalancedLevelProvider(
	spread float64,
	useMaxQuoteInTargetAmountCalc bool,
	minAmountSpread float64,
	maxAmountSpread float64,
	maxLevels int16,
	levelDensity float64,
	ensureFirstNLevels int16,
	minAmountCarryoverSpread float64,
	maxAmountCarryoverSpread float64,
	carryoverInclusionProbability float64,
	virtualBalanceBase float64,
	virtualBalanceQuote float64,
	orderConstraints *model.OrderConstraints,
) api.LevelProvider {
	if minAmountSpread <= 0 {
		log.Fatalf("minAmountSpread (%.7f) needs to be > 0 for the algorithm to work sustainably\n", minAmountSpread)
	}

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
	shouldRefresh := true

	return &balancedLevelProvider{
		spread: spread,
		useMaxQuoteInTargetAmountCalc: useMaxQuoteInTargetAmountCalc,
		minAmountSpread:               minAmountSpread,
		maxAmountSpread:               maxAmountSpread,
		maxLevels:                     maxLevels,
		levelDensity:                  levelDensity,
		ensureFirstNLevels:            ensureFirstNLevels,
		minAmountCarryoverSpread:      minAmountCarryoverSpread,
		maxAmountCarryoverSpread:      maxAmountCarryoverSpread,
		carryoverInclusionProbability: carryoverInclusionProbability,
		virtualBalanceBase:            virtualBalanceBase,
		virtualBalanceQuote:           virtualBalanceQuote,
		orderConstraints:              orderConstraints,
		randGen:                       randGen,
		shouldRefresh:                 shouldRefresh,
	}
}

func validateSpread(spread float64) {
	if spread > 1.0 || spread < 0.0 {
		log.Fatalf("spread values need to be inclusively between 0 and 1: %.7f\n", spread)
	}
}

// GetLevels impl.
func (p *balancedLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]api.Level, error) {
	if !p.shouldRefresh {
		log.Println("no offers were taken, leave levels as they are")
		return p.lastLevels, nil
	}

	levels, e := p.recomputeLevels(maxAssetBase, maxAssetQuote)
	if e != nil {
		return nil, fmt.Errorf("unable to generate new levels: %s", e)
	}

	p.lastLevels = levels
	p.shouldRefresh = false

	return levels, nil
}

func (p *balancedLevelProvider) computeNewLevelWithCarryover(level api.Level, amountCarryover float64) (api.Level, float64) {
	// include a partial amount of the carryover
	amountCarryoverToInclude := p.randGen.Float64() * amountCarryover
	// update amountCarryover to reflect inclusion in the level
	amountCarryover -= amountCarryoverToInclude
	// include the amountCarryover we computed, price of the level remains unchanged
	level = api.Level{
		Price:  level.Price,
		Amount: *level.Amount.Add(*model.NumberFromFloat(amountCarryoverToInclude, level.Amount.Precision())),
	}

	return level, amountCarryover
}

func updateAssetBalances(level api.Level, useMaxQuoteInTargetAmountCalc bool, maxAssetBase float64, maxAssetQuote float64) (float64, float64) {
	// targetPrice is always quote/base
	var baseDecreased float64
	var quoteIncreased float64
	if useMaxQuoteInTargetAmountCalc {
		// targetAmount is in quote so divide by price (quote/base) to give base
		baseDecreased = level.Amount.AsFloat() / level.Price.AsFloat()
		// targetAmount is in quote so use directly
		quoteIncreased = level.Amount.AsFloat()
	} else {
		// targetAmount is in base so use directly
		baseDecreased = level.Amount.AsFloat()
		// targetAmount is in base so multiply by price (quote/base) to give quote
		quoteIncreased = level.Amount.AsFloat() * level.Price.AsFloat()
	}
	// subtract because we had to sell that many units to reach the next level
	newMaxAssetBase := maxAssetBase - baseDecreased
	// add because we had to buy these many units to reach the next level
	newMaxAssetQuote := maxAssetQuote + quoteIncreased

	return newMaxAssetBase, newMaxAssetQuote
}

func (p *balancedLevelProvider) shouldIncludeLevel(levelIndex int16) bool {
	includeLevelUsingProbability := p.randGen.Float64() < p.levelDensity
	includeLevelUsingConstraint := levelIndex < p.ensureFirstNLevels
	return includeLevelUsingConstraint || includeLevelUsingProbability
}

func (p *balancedLevelProvider) shouldIncludeCarryover() bool {
	return p.randGen.Float64() < p.carryoverInclusionProbability
}

// getRandomSpread returns a random value between the two params (inclusive)
func (p *balancedLevelProvider) getRandomSpread(minSpread float64, maxSpread float64) float64 {
	// generates a float between 0 and 1
	randFloat := p.randGen.Float64()

	// reduce to a float between 0 and diffSpread
	diffSpread := maxSpread - minSpread
	spreadAboveMin := diffSpread * randFloat

	// convert to a float between minSpread and maxSpread
	return minSpread + spreadAboveMin
}

func (p *balancedLevelProvider) getLevel(maxAssetBase float64, maxAssetQuote float64) (api.Level, error) {
	centerPrice := maxAssetQuote / maxAssetBase
	// price always adds the spread
	targetPrice := centerPrice * (1 + p.spread/2)

	targetAmount := (2 * maxAssetBase * p.spread) / (4 + p.spread)
	if p.useMaxQuoteInTargetAmountCalc {
		targetAmount = (2 * maxAssetQuote * p.spread) / (4 + p.spread)
	}
	// since targetAmount needs to be less then what we've set above based on the inequality formula, let's reduce it by 5%
	targetAmount *= (1 - p.getRandomSpread(p.minAmountSpread, p.maxAmountSpread))
	level := api.Level{
		Price:  *model.NumberFromFloat(targetPrice, p.orderConstraints.PricePrecision),
		Amount: *model.NumberFromFloat(targetAmount, p.orderConstraints.VolumePrecision),
	}
	return level, nil
}

// GetFillHandlers impl
func (p *balancedLevelProvider) GetFillHandlers() ([]api.FillHandler, error) {
	return []api.FillHandler{p}, nil
}

// HandleFill impl
func (p *balancedLevelProvider) HandleFill(trade model.Trade) error {
	log.Println("an offer was taken, levels will be recomputed")
	p.shouldRefresh = true
	return nil
}

func (p *balancedLevelProvider) recomputeLevels(maxAssetBase float64, maxAssetQuote float64) ([]api.Level, error) {
	_maxAssetBase := maxAssetBase + p.virtualBalanceBase
	_maxAssetQuote := maxAssetQuote + p.virtualBalanceQuote
	// represents the amount that was meant to be included in a previous level that we excluded because we skipped that level
	amountCarryover := 0.0
	levels := []api.Level{}
	for i := int16(0); i < p.maxLevels; i++ {
		level, e := p.getLevel(_maxAssetBase, _maxAssetQuote)
		if e != nil {
			return nil, e
		}

		// always update _maxAssetBase and _maxAssetQuote to account for the level we just calculated, ensures price moves across levels regardless of inclusion of prior levels
		_maxAssetBase, _maxAssetQuote = updateAssetBalances(level, p.useMaxQuoteInTargetAmountCalc, _maxAssetBase, _maxAssetQuote)

		// always take a spread off the amountCarryover
		amountCarryoverSpread := p.getRandomSpread(p.minAmountCarryoverSpread, p.maxAmountCarryoverSpread)
		amountCarryover *= (1 - amountCarryoverSpread)

		if !p.shouldIncludeLevel(i) {
			// accummulate targetAmount into amountCarryover
			amountCarryover += level.Amount.AsFloat()
			continue
		}

		if p.shouldIncludeCarryover() {
			level, amountCarryover = p.computeNewLevelWithCarryover(level, amountCarryover)
		}
		levels = append(levels, level)
	}
	return levels, nil
}
