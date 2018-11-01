package plugins

import (
	"fmt"
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
	if maxAssetBase <= 0.0 {
		return nil, fmt.Errorf("Had none of the base asset, unable to generate levels")
	}

	levels := []api.Level{}
	totalAssetValue := maxAssetBase + (maxAssetQuote / p.sdexMidPrice)
	balanceRatio := maxAssetBase / totalAssetValue
	ratioGap := balanceRatio - 0.5

	log.Printf("balanceRatio = %v", balanceRatio)

	allowedSpend := maxAssetBase - totalAssetValue*p.maintainBalancePercent
	if allowedSpend < 0.0 {
		allowedSpend = 0.0
	}
	backupBalance := maxAssetBase - allowedSpend

	spent := 0.0
	levelCounter := 1.0
	overSpent := false
	for i := int16(0); i < p.maxLevels; i++ {
		if spent >= maxAssetBase {
			return nil, fmt.Errorf("Level provider spent more than asset balance")
		}
		if spent >= allowedSpend {
			overSpent = true
		}

		level, amountSpent, e := p.getLevel(maxAssetBase, levelCounter, ratioGap, overSpent, backupBalance)
		if e != nil {
			return nil, e
		}

		spent += amountSpent
		if overSpent {
			backupBalance -= amountSpent
		}
		if spent < maxAssetBase {
			levels = append(levels, level)
		}
		levelCounter += 1.0
	}
	return levels, nil
}

func (p *SDEXLevelProvider) getLevel(maxAssetBase float64, levelCounter float64, ratioGap float64, overSpent bool, backupBalance float64) (api.Level, float64, error) {

	targetPrice := p.sdexMidPrice * (1.0 + p.spread*levelCounter)
	targetAmount := maxAssetBase * p.basePercentPerLevel * (1.0 + ratioGap*2)
	if p.isBuy {
		targetAmount = maxAssetBase * p.basePercentPerLevel * (1.0 + ratioGap*2) * targetPrice
	}

	if overSpent {
		targetAmount = backupBalance * 0.001
	}

	level := api.Level{
		Price:  *model.NumberFromFloat(targetPrice, utils.SdexPrecision),
		Amount: *model.NumberFromFloat(targetAmount, utils.SdexPrecision),
	}
	return level, targetAmount, nil
}
