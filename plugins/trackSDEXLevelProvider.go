package plugins

import (
	"log"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/support/utils"
)

type SDEXLevelProvider struct {
	spread                 float64
	basePercentPerLevel    float64
	maxLevels              int16
	maintainBalancePercent float64
	sdexMidPrice           float64
	isBuy                  bool
}

func makeSDEXLevelProvider(
	spread float64,
	basePercentPerLevel float64,
	maxLevels int16,
	maintainBalancePercent float64,
	sdexMidPrice float64,
	isBuy bool,
) api.LevelProvider {
	validateTotalAmount(maxLevels, basePercentPerLevel)
	return &SDEXLevelProvider{
		spread:                 spread,
		basePercentPerLevel:    basePercentPerLevel,
		maxLevels:              maxLevels,
		maintainBalancePercent: maintainBalancePercent,
		sdexMidPrice:           sdexMidPrice,
		isBuy:                  isBuy,
	}
}

func validateTotalAmount(maxLevels int16, basePercentPerLevel float64) {
	l := float64(maxLevels)
	totalAmount := l * basePercentPerLevel
	if totalAmount > 1 {
		log.Fatalf("Number of levels * percent per level must be < 1.0\n")
	}

}

func (p *SDEXLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]api.Level, error) {
	levels := []api.Level{}
	totalAssetValue := maxAssetBase + (maxAssetQuote / p.sdexMidPrice)
	balanceRatio := maxAssetBase / totalAssetValue
	ratioGap := balanceRatio - 0.5

	log.Printf("balanceRatio = %v", balanceRatio)

	allowedSpend := maxAssetBase
	if ratioGap < 0 {
		allowedSpend = maxAssetBase - totalAssetValue*p.maintainBalancePercent
	}

	spent := 0.0
	levelCounter := 1.0
	for i := int16(0); i < p.maxLevels && spent < allowedSpend; i++ {
		level, amountSpent, e := p.getLevel(maxAssetBase, maxAssetQuote, levelCounter, ratioGap)
		if e != nil {
			return nil, e
		}
		levelCounter += 1.0
		spent += amountSpent
		levels = append(levels, level)
	}
	return levels, nil
}

func (p *SDEXLevelProvider) getLevel(maxAssetBase float64, maxAssetQuote, levelCounter float64, ratioGap float64) (api.Level, float64, error) {
	//find balance mismatch for amount adjustment, helps keep assets from running out

	targetPrice := p.sdexMidPrice * (1.0 + p.spread*levelCounter)
	targetAmount := maxAssetBase * p.basePercentPerLevel * (1.0 + ratioGap)
	if p.isBuy {
		targetAmount = maxAssetBase * p.basePercentPerLevel * (1.0 + ratioGap) * targetPrice
	}

	level := api.Level{
		Price:  *model.NumberFromFloat(targetPrice, utils.SdexPrecision),
		Amount: *model.NumberFromFloat(targetAmount, utils.SdexPrecision),
	}
	return level, targetAmount, nil
}
