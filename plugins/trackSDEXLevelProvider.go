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
	balanceRatio := maxAssetBase / (maxAssetBase + (maxAssetQuote / p.sdexMidPrice))

	log.Printf("balanceRatio = %s", balanceRatio)

	// don't place any orders if below minimum balance parameter
	if balanceRatio < p.maintainBalancePercent {
		return levels, nil
	}

	levelCounter := 1.0
	for i := int16(0); i < p.maxLevels; i++ {
		level, e := p.getLevel(maxAssetBase, maxAssetQuote, levelCounter, balanceRatio)
		if e != nil {
			return nil, e
		}
		levelCounter += 1.0
		levels = append(levels, level)
	}
	return levels, nil
}

func (p *SDEXLevelProvider) getLevel(maxAssetBase float64, maxAssetQuote float64, levelCounter float64, balanceRatio float64) (api.Level, error) {
	//find balance mismatch for amount adjustment, helps keep assets from running out
	ratioGap := balanceRatio - 0.5

	targetPrice := p.sdexMidPrice * (1.0 + p.spread*levelCounter)
	targetAmount := maxAssetBase * p.basePercentPerLevel * (1.0 + ratioGap)
	if p.isBuy {
		targetAmount = maxAssetBase * p.basePercentPerLevel * p.sdexMidPrice * (1.0 + ratioGap)
	}

	level := api.Level{
		Price:  *model.NumberFromFloat(targetPrice, utils.SdexPrecision),
		Amount: *model.NumberFromFloat(targetAmount, utils.SdexPrecision),
	}
	return level, nil
}
