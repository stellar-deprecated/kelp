package plugins

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/queries"
	"github.com/stellar/kelp/support/utils"
)

var testBaseAsset = txnbuild.NativeAsset{}
var testQuoteAsset txnbuild.CreditAsset = txnbuild.CreditAsset{Code: "QUOTE", Issuer: "GBGQAGAMK6W6FH6AGGZ2BI2MY5TA5VJEHU2DQRFXACMAZHNRD3SXEV6Z"}

func makeWantVolumeFilter(config *VolumeFilterConfig, marketIDs []string, accountIDs []string, action queries.DailyVolumeAction) *volumeFilter {
	query, e := queries.MakeDailyVolumeByDateForMarketIdsAction(&sql.DB{}, marketIDs, action, accountIDs)
	if e != nil {
		panic(e)
	}

	return &volumeFilter{
		name:                   "volumeFilter",
		configValue:            "",
		baseAsset:              utils.NativeAsset,
		quoteAsset:             utils.NativeAsset,
		config:                 config,
		dailyVolumeByDateQuery: query,
	}
}

func TestMakeFilterVolume(t *testing.T) {
	testAssetDisplayFn := model.MakeSdexMappedAssetDisplayFn(map[model.Asset]hProtocol.Asset{model.Asset("XLM"): utils.NativeAsset})
	configValue := ""
	tradingPair := &model.TradingPair{Base: "XLM", Quote: "XLM"}
	testCases := []struct {
		name          string
		exchangeName  string
		marketIDs     []string
		accountIDs    []string
		wantMarketIDs []string
		wantFilter    *volumeFilter
	}{
		{
			name:          "0 market id or account id",
			exchangeName:  "exchange 2",
			marketIDs:     []string{},
			accountIDs:    []string{},
			wantMarketIDs: []string{"9db20cdd56"},
		},
		{
			name:          "1 market id",
			exchangeName:  "exchange 1",
			marketIDs:     []string{"marketID"},
			accountIDs:    []string{},
			wantMarketIDs: []string{"6d9862b0e2", "marketID"},
		},
		{
			name:          "2 market ids",
			exchangeName:  "exchange 2",
			marketIDs:     []string{"marketID1", "marketID2"},
			accountIDs:    []string{},
			wantMarketIDs: []string{"9db20cdd56", "marketID1", "marketID2"},
		},
		{
			name:          "2 dupe market ids, 1 distinct",
			exchangeName:  "exchange 1",
			marketIDs:     []string{"marketID1", "marketID1", "marketID2"},
			accountIDs:    []string{},
			wantMarketIDs: []string{"6d9862b0e2", "marketID1", "marketID2"},
		},
		{
			name:          "1 account id",
			exchangeName:  "exchange 2",
			marketIDs:     []string{},
			accountIDs:    []string{"accountID"},
			wantMarketIDs: []string{"9db20cdd56"},
		},
		{
			name:          "2 account ids",
			exchangeName:  "exchange 1",
			marketIDs:     []string{},
			accountIDs:    []string{"accountID1", "accountID2"},
			wantMarketIDs: []string{"6d9862b0e2"},
		},
		{
			name:          "account and market ids",
			exchangeName:  "exchange 2",
			marketIDs:     []string{"marketID"},
			accountIDs:    []string{"accountID"},
			wantMarketIDs: []string{"9db20cdd56", "marketID"},
		},
	}

	caseNo := 1
	for _, k := range testCases {
		// this lets us test both types of modes when varying the market and account ids
		for _, m := range []volumeFilterMode{volumeFilterModeExact, volumeFilterModeIgnore} {
			// this lets us test both buy and sell
			for _, action := range []queries.DailyVolumeAction{queries.DailyVolumeActionSell, queries.DailyVolumeActionBuy} {
				// this lets us run the for-loop below for both base and quote units within the config
				baseCapInBaseConfig := makeRawVolumeFilterConfig(
					pointy.Float64(1.0),
					nil,
					action,
					m,
					k.marketIDs,
					k.accountIDs,
				)
				baseCapInQuoteConfig := makeRawVolumeFilterConfig(
					nil,
					pointy.Float64(1.0),
					action,
					m,
					k.marketIDs,
					k.accountIDs,
				)
				for _, config := range []*VolumeFilterConfig{baseCapInBaseConfig, baseCapInQuoteConfig} {
					// configType is used to represent the type of config when printing test name
					configType := "quote"
					if config.BaseAssetCapInBaseUnits != nil {
						configType = "base"
					}

					wantFilter := makeWantVolumeFilter(config, k.wantMarketIDs, k.accountIDs, action)
					testCaseInstanceName := fmt.Sprintf("%d. %s,%s,%s,%s", caseNo, k.name, configType, m, action.String())
					caseNo++
					t.Run(testCaseInstanceName, func(t *testing.T) {
						actual, e := makeFilterVolume(
							configValue,
							k.exchangeName,
							tradingPair,
							testAssetDisplayFn,
							utils.NativeAsset,
							utils.NativeAsset,
							&sql.DB{},
							config,
						)

						if !assert.Nil(t, e) {
							return
						}

						assert.Equal(t, wantFilter, actual)
					})
				}
			}
		}
	}
}

func TestMakeFilterVolume_FailsOnEmptyConfig(t *testing.T) {
	testAssetDisplayFn := model.MakeSdexMappedAssetDisplayFn(map[model.Asset]hProtocol.Asset{model.Asset("XLM"): utils.NativeAsset})
	tradingPair := &model.TradingPair{Base: "XLM", Quote: "XLM"}

	configUnderTest := &VolumeFilterConfig{}
	_, e := makeFilterVolume(
		"someConfigValue",
		"someExchangeName",
		tradingPair,
		testAssetDisplayFn,
		utils.NativeAsset,
		utils.NativeAsset,
		&sql.DB{},
		configUnderTest,
	)
	if !assert.Error(t, e) {
		return
	}

	assert.True(t, strings.HasPrefix(e.Error(), "invalid config"), e.Error())
}

// volumeFilterFnTestCase is the input that will be reused across all tests of type TestVolumeFilterFn*
type volumeFilterFnTestCase struct {
	name         string
	cap          float64
	otb          float64
	tbb          float64
	inputPrice   float64
	inputAmount  float64
	wantPrice    *float64
	wantAmount   *float64
	wantTbbBase  float64
	wantTbbQuote float64
}

func TestVolumeFilterFn_BaseCap_Exact(t *testing.T) {
	// We want to test the following 4 valid combinations of OTB and TBB values:
	// otb = 0
	// tbb = 0
	// otb = 0 && tbb = 0
	// otb > 0 && tbb > 0
	// We also want to test 3 combinations of cap relationship to projected (<, =, >)
	// Finally, if projected > cap, we want to test 3 possible values of newAmount (+, 0, -)
	// The above gives 4 * (2 + 1*3) = 20.
	// One generated case is impossible in the code: otb = 0 && tbb = 0, newAmount < 0
	// so we have 19 cases
	testCases := []volumeFilterFnTestCase{
		{
			name:         "1. otb = 0; projected < cap",
			cap:          10.0,
			otb:          0,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  4.99,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(4.99),
			wantTbbBase:  9.99,
			wantTbbQuote: 9.98,
		},
		{
			name:         "2. otb = 0; projected = cap",
			cap:          10.0,
			otb:          0,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  5.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(5),
			wantTbbBase:  10,
			wantTbbQuote: 10,
		},
		{
			name:         "3. otb = 0; projected > cap, newAmount > 0",
			cap:          10.0,
			otb:          0,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  5.01,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(5.0),
			wantTbbBase:  10,
			wantTbbQuote: 10,
		},
		{
			name:         "4. otb = 0; projected > cap, newAmount = 0",
			cap:          10.0,
			otb:          0,
			tbb:          10,
			inputPrice:   2.0,
			inputAmount:  5.01,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  10,
			wantTbbQuote: 0,
		},
		{
			name:         "5. otb = 0; projected > cap, newAmount < 0",
			cap:          10.0,
			otb:          0,
			tbb:          11,
			inputPrice:   2.0,
			inputAmount:  5.01,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  11,
			wantTbbQuote: 0,
		},
		{
			name:         "6. tbb = 0; projected < cap",
			cap:          10.0,
			otb:          5,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  4.99,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(4.99),
			wantTbbBase:  4.99,
			wantTbbQuote: 9.98,
		},
		{
			name:         "7. tbb = 0; projected = cap",
			cap:          10.0,
			otb:          5,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  5.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(5.0),
			wantTbbBase:  5,
			wantTbbQuote: 10,
		},
		{
			name:         "8. tbb = 0; projected > cap, newAmount > 0",
			cap:          10.0,
			otb:          5,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  6.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(5.0),
			wantTbbBase:  5,
			wantTbbQuote: 10,
		},
		{
			name:         "9. tbb = 0; projected > cap, newAmount = 0",
			cap:          10.0,
			otb:          10,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  6.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 0,
		},
		{
			name:         "10. tbb = 0; projected > cap, newAmount < 0",
			cap:          10.0,
			otb:          11,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  6.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 0,
		},
		{
			name:         "11. otb = 0 && tbb = 0; projected < cap",
			cap:          10.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  5.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(5.0),
			wantTbbBase:  5,
			wantTbbQuote: 10,
		},
		{
			name:         "12. otb = 0 && tbb = 0; projected = cap",
			cap:          10.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  10.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(10.0),
			wantTbbBase:  10,
			wantTbbQuote: 20,
		},
		{
			// note that in this case, newAmount >= 0, since newAmount = cap - otb - tbb
			name:         "13. otb = 0 && tbb = 0; projected > cap, newAmount > 0",
			cap:          10.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  15.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(10.0),
			wantTbbBase:  10,
			wantTbbQuote: 20,
		},
		{
			name:         "14. otb = 0 && tbb = 0; projected > cap, newAmount = 0",
			cap:          0.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  15.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 0,
		},
		// it is not possible for otb = 0 && tbb = 0 and newAmount < 0, so skipping that case
		{
			name:         "15. otb > 0 && tbb > 0; projected < cap",
			cap:          10.0,
			otb:          1,
			tbb:          1,
			inputPrice:   2.0,
			inputAmount:  5.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(5.0),
			wantTbbBase:  6,
			wantTbbQuote: 10,
		},
		{
			name:         "16. otb > 0 && tbb > 0; projected = cap",
			cap:          10.0,
			otb:          2,
			tbb:          2,
			inputPrice:   2.0,
			inputAmount:  6.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(6.0),
			wantTbbBase:  8,
			wantTbbQuote: 12,
		},
		{
			name:         "17. otb > 0 && tbb > 0; projected > cap, newAmount > 0",
			cap:          10.0,
			otb:          2,
			tbb:          2,
			inputPrice:   2.0,
			inputAmount:  7.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(6.0),
			wantTbbBase:  8,
			wantTbbQuote: 12,
		},
		{
			name:         "18. otb > 0 && tbb > 0; projected > cap, newAmount = 0",
			cap:          10.0,
			otb:          5,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  7.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  5,
			wantTbbQuote: 0,
		},
		{
			name:         "19. otb > 0 && tbb > 0; projected > cap, newAmount < 0",
			cap:          10.0,
			otb:          5,
			tbb:          5.1,
			inputPrice:   2.0,
			inputAmount:  7.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  5.1,
			wantTbbQuote: 0,
		},
	}

	for _, k := range testCases {
		// convert to common format accepted by runTestVolumeFilterFn
		// test sell action
		sellName := fmt.Sprintf("%s, sell", k.name)
		inputSellOp := makeSellOpAmtPrice(k.inputAmount, k.inputPrice)
		var wantSellOp *txnbuild.ManageSellOffer
		if k.wantPrice != nil && k.wantAmount != nil {
			wantSellOp = makeSellOpAmtPrice(*k.wantAmount, *k.wantPrice)
		}

		runTestVolumeFilterFn(
			t,
			sellName,
			volumeFilterModeExact,
			queries.DailyVolumeActionSell,
			pointy.Float64(k.cap), // base cap
			nil,                   // quote cap nil because this test is for the BaseCap
			pointy.Float64(k.otb), // baseOTB
			nil,                   // quoteOTB nil because this test is for the BaseCap
			pointy.Float64(k.tbb), // baseTBB
			pointy.Float64(0),     // quoteTBB (non-nil since it accumulates)
			inputSellOp,
			wantSellOp,
			pointy.Float64(k.wantTbbBase),
			pointy.Float64(k.wantTbbQuote),
		)

		// test buy action
		buyName := fmt.Sprintf("%s, buy", k.name)
		inputBuyOp := makeBuyOpAmtPrice(k.inputAmount, k.inputPrice)
		var wantBuyOp *txnbuild.ManageSellOffer
		if k.wantPrice != nil && k.wantAmount != nil {
			wantBuyOp = makeBuyOpAmtPrice(*k.wantAmount, *k.wantPrice)
		}

		runTestVolumeFilterFn(
			t,
			buyName,
			volumeFilterModeExact,
			queries.DailyVolumeActionBuy,
			pointy.Float64(k.cap), // base cap
			nil,                   // quote cap nil because this test is for the BaseCap
			pointy.Float64(k.otb), // baseOTB
			nil,                   // quoteOTB nil because this test is for the BaseCap
			pointy.Float64(k.tbb), // baseTBB
			pointy.Float64(0),     // quoteTBB (non-nil since it accumulates)
			inputBuyOp,
			wantBuyOp,
			pointy.Float64(k.wantTbbBase),
			pointy.Float64(k.wantTbbQuote),
		)
	}
}

func TestVolumeFilterFn_BaseCap_Ignore(t *testing.T) {
	// We want to test the following 4 valid combinations of OTB and TBB values:
	// otb = 0
	// tbb = 0
	// otb = 0 && tbb = 0
	// otb > 0 && tbb > 0
	// 12 cases here; 4 combinations of tbb/otb values from bullet points above x 3 combinations of cap relationship to projected (<, =, >)
	testCases := []volumeFilterFnTestCase{
		{
			name:         "1. otb = 0; projected < cap",
			cap:          10.0,
			otb:          0,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  4.99,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(4.99),
			wantTbbBase:  9.99,
			wantTbbQuote: 9.98,
		},
		{
			name:         "2. otb = 0; projected = cap",
			cap:          10.0,
			otb:          0,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  5.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(5),
			wantTbbBase:  10,
			wantTbbQuote: 10,
		},
		{
			name:         "3. otb = 0; projected > cap",
			cap:          10.0,
			otb:          0,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  5.01,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  5,
			wantTbbQuote: 0,
		},
		{
			name:         "4. tbb = 0; projected < cap",
			cap:          10.0,
			otb:          5,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  4.99,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(4.99),
			wantTbbBase:  4.99,
			wantTbbQuote: 9.98,
		},
		{
			name:         "5. tbb = 0; projected = cap",
			cap:          10.0,
			otb:          5,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  5.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(5.0),
			wantTbbBase:  5,
			wantTbbQuote: 10,
		},
		{
			name:         "6. tbb = 0; projected > cap",
			cap:          10.0,
			otb:          5,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  6.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 0,
		},
		{
			name:         "7. otb = 0 && tbb = 0; projected < cap",
			cap:          10.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  5.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(5.0),
			wantTbbBase:  5,
			wantTbbQuote: 10,
		},
		{
			name:         "8. otb = 0 && tbb = 0; projected = cap",
			cap:          10.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  10.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(10.0),
			wantTbbBase:  10,
			wantTbbQuote: 20,
		},
		{
			name:         "9. otb = 0 && tbb = 0; projected > cap",
			cap:          10.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  15.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 0,
		},
		{
			name:         "10. otb > 0 && tbb > 0; projected < cap",
			cap:          10.0,
			otb:          1,
			tbb:          1,
			inputPrice:   2.0,
			inputAmount:  5.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(5.0),
			wantTbbBase:  6,
			wantTbbQuote: 10,
		},
		{
			name:         "11. otb > 0 && tbb > 0; projected = cap",
			cap:          10.0,
			otb:          2,
			tbb:          2,
			inputPrice:   2.0,
			inputAmount:  6.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(6.0),
			wantTbbBase:  8,
			wantTbbQuote: 12,
		},
		{
			name:         "12. otb > 0 && tbb > 0; projected > cap",
			cap:          10.0,
			otb:          2,
			tbb:          2,
			inputPrice:   2.0,
			inputAmount:  7.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  2,
			wantTbbQuote: 0,
		},
	}
	for _, k := range testCases {
		// convert to common format accepted by runTestVolumeFilterFn
		// test sell action
		sellName := fmt.Sprintf("%s, sell", k.name)
		inputSellOp := makeSellOpAmtPrice(k.inputAmount, k.inputPrice)
		var wantSellOp *txnbuild.ManageSellOffer
		if k.wantPrice != nil && k.wantAmount != nil {
			wantSellOp = makeSellOpAmtPrice(*k.wantAmount, *k.wantPrice)
		}

		runTestVolumeFilterFn(
			t,
			sellName,
			volumeFilterModeIgnore,
			queries.DailyVolumeActionSell,
			pointy.Float64(k.cap), // base cap
			nil,                   // quote cap nil because this test is for the BaseCap
			pointy.Float64(k.otb), // baseOTB
			nil,                   // quoteOTB nil because this test is for the BaseCap
			pointy.Float64(k.tbb), // baseTBB
			pointy.Float64(0),     // quoteTBB (non-nil since it accumulates)
			inputSellOp,
			wantSellOp,
			pointy.Float64(k.wantTbbBase),
			pointy.Float64(k.wantTbbQuote),
		)

		// test buy action
		buyName := fmt.Sprintf("%s, buy", k.name)
		inputBuyOp := makeBuyOpAmtPrice(k.inputAmount, k.inputPrice)
		var wantBuyOp *txnbuild.ManageSellOffer
		if k.wantPrice != nil && k.wantAmount != nil {
			wantBuyOp = makeBuyOpAmtPrice(*k.wantAmount, *k.wantPrice)
		}

		runTestVolumeFilterFn(
			t,
			buyName,
			volumeFilterModeIgnore,
			queries.DailyVolumeActionBuy,
			pointy.Float64(k.cap), // base cap
			nil,                   // quote cap nil because this test is for the BaseCap
			pointy.Float64(k.otb), // baseOTB
			nil,                   // quoteOTB nil because this test is for the BaseCap
			pointy.Float64(k.tbb), // baseTBB
			pointy.Float64(0),     // quoteTBB (non-nil since it accumulates)
			inputBuyOp,
			wantBuyOp,
			pointy.Float64(k.wantTbbBase),
			pointy.Float64(k.wantTbbQuote),
		)
	}
}

func TestVolumeFilterFn_QuoteCap_Ignore(t *testing.T) {
	// We want to test the following 4 valid combinations of OTB and TBB values:
	// otb = 0
	// tbb = 0
	// otb = 0 && tbb = 0
	// otb > 0 && tbb > 0
	// 12 cases here; 4 combinations of tbb/otb values from bullet points above x 3 combinations of cap relationship to projected (<, =, >)
	testCases := []volumeFilterFnTestCase{
		{
			name:         "1. otb = 0; projected < cap",
			cap:          10.0,
			otb:          0,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  2.49,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(2.49),
			wantTbbBase:  2.49,
			wantTbbQuote: 9.98,
		},
		{
			name:         "2. otb = 0; projected = cap",
			cap:          10.0,
			otb:          0,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  2.5,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(2.5),
			wantTbbBase:  2.5,
			wantTbbQuote: 10,
		},
		{
			name:         "3. otb = 0; projected > cap",
			cap:          10.0,
			otb:          0,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  5.01,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 5,
		},
		{
			name:         "4. tbb = 0; projected < cap",
			cap:          10.0,
			otb:          5,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  2.49,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(2.49),
			wantTbbBase:  2.49,
			wantTbbQuote: 4.98,
		},
		{
			name:         "5. tbb = 0; projected = cap",
			cap:          10.0,
			otb:          5,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  2.5,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(2.5),
			wantTbbBase:  2.5,
			wantTbbQuote: 5,
		},
		{
			name:         "6. tbb = 0; projected > cap",
			cap:          10.0,
			otb:          5,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  6.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 0,
		},
		{
			name:         "7. otb = 0 && tbb = 0; projected < cap",
			cap:          10.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  2.5,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(2.5),
			wantTbbBase:  2.5,
			wantTbbQuote: 5,
		},
		{
			name:         "8. otb = 0 && tbb = 0; projected = cap",
			cap:          10.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  5.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(5.0),
			wantTbbBase:  5,
			wantTbbQuote: 10,
		},
		{
			name:         "9. otb = 0 && tbb = 0; projected > cap",
			cap:          10.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  15.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 0,
		},
		{
			name:         "10. otb > 0 && tbb > 0; projected < cap",
			cap:          10.0,
			otb:          1,
			tbb:          1,
			inputPrice:   2.0,
			inputAmount:  2.5,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(2.5),
			wantTbbBase:  2.5,
			wantTbbQuote: 6,
		},
		{
			name:         "11. otb > 0 && tbb > 0; projected = cap",
			cap:          10.0,
			otb:          2,
			tbb:          2,
			inputPrice:   2.0,
			inputAmount:  3.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(3.0),
			wantTbbBase:  3,
			wantTbbQuote: 8,
		},
		{
			name:         "12. otb > 0 && tbb > 0; projected > cap",
			cap:          10.0,
			otb:          2,
			tbb:          2,
			inputPrice:   2.0,
			inputAmount:  7.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 2,
		},
	}
	for _, k := range testCases {
		// convert to common format accepted by runTestVolumeFilterFn
		// test sell action
		sellName := fmt.Sprintf("%s, sell", k.name)
		inputSellOp := makeSellOpAmtPrice(k.inputAmount, k.inputPrice)
		var wantSellOp *txnbuild.ManageSellOffer
		if k.wantPrice != nil && k.wantAmount != nil {
			wantSellOp = makeSellOpAmtPrice(*k.wantAmount, *k.wantPrice)
		}

		runTestVolumeFilterFn(
			t,
			sellName,
			volumeFilterModeIgnore,
			queries.DailyVolumeActionSell,
			nil,                   // base cap nil because this test is for the QuoteCap
			pointy.Float64(k.cap), // quote cap
			nil,                   // baseOTB nil because this test is for the QuoteCap
			pointy.Float64(k.otb), // quoteOTB
			pointy.Float64(0),     // baseTBB (non-nil since it accumulates)
			pointy.Float64(k.tbb), // quoteTBB
			inputSellOp,
			wantSellOp,
			pointy.Float64(k.wantTbbBase),
			pointy.Float64(k.wantTbbQuote),
		)

		// test buy action
		buyName := fmt.Sprintf("%s, buy", k.name)
		inputBuyOp := makeBuyOpAmtPrice(k.inputAmount, k.inputPrice)
		var wantBuyOp *txnbuild.ManageSellOffer
		if k.wantPrice != nil && k.wantAmount != nil {
			wantBuyOp = makeBuyOpAmtPrice(*k.wantAmount, *k.wantPrice)
		}

		runTestVolumeFilterFn(
			t,
			buyName,
			volumeFilterModeIgnore,
			queries.DailyVolumeActionBuy,
			nil,                   // base cap nil because this test is for the QuoteCap
			pointy.Float64(k.cap), // quote cap
			nil,                   // baseOTB nil because this test is for the QuoteCap
			pointy.Float64(k.otb), // quoteOTB
			pointy.Float64(0),     // baseTBB (non-nil since it accumulates)
			pointy.Float64(k.tbb), // quoteTBB
			inputBuyOp,
			wantBuyOp,
			pointy.Float64(k.wantTbbBase),
			pointy.Float64(k.wantTbbQuote),
		)
	}
}

func TestVolumeFilterFn_QuoteCap_Exact(t *testing.T) {
	// We want to test the following 4 valid combinations of OTB and TBB values:
	// otb = 0
	// tbb = 0
	// otb = 0 && tbb = 0
	// otb > 0 && tbb > 0
	// We also want to test 3 combinations of cap relationship to projected (<, =, >)
	// Finally, if projected > cap, we want to test 3 possible values of newAmount (+, 0, -)
	// The above gives 4 * (2 + 1*3) = 20.
	// One generated case is impossible in the code: otb = 0 && tbb = 0, newAmount < 0
	// so we have 19 cases
	testCases := []volumeFilterFnTestCase{
		{
			name:         "1. otb = 0; projected < cap",
			cap:          10.0,
			otb:          0,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  2.49,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(2.49),
			wantTbbBase:  2.49,
			wantTbbQuote: 9.98,
		},
		{
			name:         "2. otb = 0; projected = cap",
			cap:          10.0,
			otb:          0,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  2.5,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(2.5),
			wantTbbBase:  2.5,
			wantTbbQuote: 10,
		},
		{
			name:         "3. otb = 0; projected > cap, newAmount > 0",
			cap:          10.0,
			otb:          0,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  5.01,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(2.5),
			wantTbbBase:  2.5,
			wantTbbQuote: 10,
		},
		{
			name:         "4. otb = 0; projected > cap, newAmount = 0",
			cap:          10.0,
			otb:          0,
			tbb:          10,
			inputPrice:   2.0,
			inputAmount:  5.01,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 10,
		},
		{
			name:         "5. otb = 0; projected > cap, newAmount < 0",
			cap:          10.0,
			otb:          0,
			tbb:          11,
			inputPrice:   2.0,
			inputAmount:  5.01,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 11,
		},
		{
			name:         "6. tbb = 0; projected < cap",
			cap:          10.0,
			otb:          5,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  2.49,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(2.49),
			wantTbbBase:  2.49,
			wantTbbQuote: 4.98,
		},
		{
			name:         "7. tbb = 0; projected = cap",
			cap:          10.0,
			otb:          5,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  2.5,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(2.5),
			wantTbbBase:  2.5,
			wantTbbQuote: 5,
		},
		{
			name:         "8. tbb = 0; projected > cap, newAmount > 0",
			cap:          10.0,
			otb:          5,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  3.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(2.5),
			wantTbbBase:  2.5,
			wantTbbQuote: 5,
		},
		{
			name:         "9. tbb = 0; projected > cap, newAmount = 0",
			cap:          10.0,
			otb:          10,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  6.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 0,
		},
		{
			name:         "10. tbb = 0; projected > cap, newAmount < 0",
			cap:          10.0,
			otb:          11,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  6.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 0,
		},
		{
			name:         "11. otb = 0 && tbb = 0; projected < cap",
			cap:          10.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  4.99,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(4.99),
			wantTbbBase:  4.99,
			wantTbbQuote: 9.98,
		},
		{
			name:         "12. otb = 0 && tbb = 0; projected = cap",
			cap:          10.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  5.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(5.0),
			wantTbbBase:  5,
			wantTbbQuote: 10,
		},
		{
			// note that in this case, newAmount >= 0, since newAmount = cap - otb - tbb
			name:         "13. otb = 0 && tbb = 0; projected > cap, newAmount > 0",
			cap:          10.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  7.5,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(5.0),
			wantTbbBase:  5,
			wantTbbQuote: 10,
		},
		{
			name:         "14. otb = 0 && tbb = 0; projected > cap, newAmount = 0",
			cap:          0.0,
			otb:          0,
			tbb:          0,
			inputPrice:   2.0,
			inputAmount:  15.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 0,
		},
		// it is not possible for otb = 0 && tbb = 0 and newAmount < 0, so skipping that case
		{
			name:         "15. otb > 0 && tbb > 0; projected < cap",
			cap:          10.0,
			otb:          1,
			tbb:          1,
			inputPrice:   2.0,
			inputAmount:  2.5,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(2.5),
			wantTbbBase:  2.5,
			wantTbbQuote: 6,
		},
		{
			name:         "16. otb > 0 && tbb > 0; projected = cap",
			cap:          10.0,
			otb:          2,
			tbb:          2,
			inputPrice:   2.0,
			inputAmount:  3.0,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(3.0),
			wantTbbBase:  3,
			wantTbbQuote: 8,
		},
		{
			name:         "17. otb > 0 && tbb > 0; projected > cap, newAmount > 0",
			cap:          10.0,
			otb:          2,
			tbb:          2,
			inputPrice:   2.0,
			inputAmount:  3.5,
			wantPrice:    pointy.Float64(2.0),
			wantAmount:   pointy.Float64(3.0),
			wantTbbBase:  3,
			wantTbbQuote: 8,
		},
		{
			name:         "18. otb > 0 && tbb > 0; projected > cap, newAmount = 0",
			cap:          10.0,
			otb:          5,
			tbb:          5,
			inputPrice:   2.0,
			inputAmount:  7.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 5,
		},
		{
			name:         "19. otb > 0 && tbb > 0; projected > cap, newAmount < 0",
			cap:          10.0,
			otb:          5,
			tbb:          5.1,
			inputPrice:   2.0,
			inputAmount:  7.0,
			wantPrice:    nil,
			wantAmount:   nil,
			wantTbbBase:  0,
			wantTbbQuote: 5.1,
		},
	}
	for _, k := range testCases {
		// convert to common format accepted by runTestVolumeFilterFn
		// test sell action
		sellName := fmt.Sprintf("%s, sell", k.name)
		inputSellOp := makeSellOpAmtPrice(k.inputAmount, k.inputPrice)
		var wantSellOp *txnbuild.ManageSellOffer
		if k.wantPrice != nil && k.wantAmount != nil {
			wantSellOp = makeSellOpAmtPrice(*k.wantAmount, *k.wantPrice)
		}

		runTestVolumeFilterFn(
			t,
			sellName,
			volumeFilterModeExact,
			queries.DailyVolumeActionSell,
			nil,                   // base cap nil because this test is for the QuoteCap
			pointy.Float64(k.cap), // quote cap
			nil,                   // baseOTB nil because this test is for the QuoteCap
			pointy.Float64(k.otb), // quoteOTB
			pointy.Float64(0),     // baseTBB (non-nil since it accumulates)
			pointy.Float64(k.tbb), // quoteTBB
			inputSellOp,
			wantSellOp,
			pointy.Float64(k.wantTbbBase),
			pointy.Float64(k.wantTbbQuote),
		)

		// test buy action
		buyName := fmt.Sprintf("%s, buy", k.name)
		inputBuyOp := makeBuyOpAmtPrice(k.inputAmount, k.inputPrice)
		var wantBuyOp *txnbuild.ManageSellOffer
		if k.wantPrice != nil && k.wantAmount != nil {
			wantBuyOp = makeBuyOpAmtPrice(*k.wantAmount, *k.wantPrice)
		}

		runTestVolumeFilterFn(
			t,
			buyName,
			volumeFilterModeExact,
			queries.DailyVolumeActionBuy,
			nil,                   // base cap nil because this test is for the QuoteCap
			pointy.Float64(k.cap), // quote cap
			nil,                   // baseOTB nil because this test is for the QuoteCap
			pointy.Float64(k.otb), // quoteOTB
			pointy.Float64(0),     // baseTBB (non-nil since it accumulates)
			pointy.Float64(k.tbb), // quoteTBB
			inputBuyOp,
			wantBuyOp,
			pointy.Float64(k.wantTbbBase),
			pointy.Float64(k.wantTbbQuote),
		)
	}
}

func runTestVolumeFilterFn(
	t *testing.T,
	name string,
	mode volumeFilterMode,
	action queries.DailyVolumeAction,
	baseCap *float64,
	quoteCap *float64,
	baseOTB *float64,
	quoteOTB *float64,
	baseTBB *float64,
	quoteTBB *float64,
	inputOp *txnbuild.ManageSellOffer,
	wantOp *txnbuild.ManageSellOffer,
	wantBase *float64,
	wantQuote *float64,
) {
	t.Run(name, func(t *testing.T) {
		// exactly one of the two cap values must be set
		if baseCap == nil && quoteCap == nil {
			assert.Fail(t, "either one of the two cap values must be set")
			return
		}

		if baseCap != nil && quoteCap != nil {
			assert.Fail(t, "both of the cap values cannot be set")
			return
		}

		// we pass in nil market IDs and account IDs, as they don't affect correctness
		dailyOTB := makeIntermediateVolumeFilterConfig(baseOTB, quoteOTB)
		dailyTBBAccumulator := makeIntermediateVolumeFilterConfig(baseTBB, quoteTBB)
		lp := limitParameters{
			baseAssetCapInBaseUnits:  baseCap,
			baseAssetCapInQuoteUnits: quoteCap,
			mode:                     mode,
		}

		base := utils.Asset2Asset2(testBaseAsset)
		quote := utils.Asset2Asset2(testQuoteAsset)
		actual, e := volumeFilterFn(action, dailyOTB, dailyTBBAccumulator, inputOp, base, quote, lp)
		if !assert.Nil(t, e) {
			return
		}
		if !assert.Equal(t, wantOp, actual) {
			return
		}

		wantTBBAccumulator := makeIntermediateVolumeFilterConfig(wantBase, wantQuote)
		assert.Equal(t, wantTBBAccumulator, dailyTBBAccumulator)
	})
}

func makeSellOpAmtPrice(amount float64, price float64) *txnbuild.ManageSellOffer {
	return &txnbuild.ManageSellOffer{
		Buying:  testQuoteAsset,
		Selling: testBaseAsset,
		Amount:  fmt.Sprintf("%.7f", amount),
		Price:   fmt.Sprintf("%.7f", price),
	}
}

func makeBuyOpAmtPrice(amount float64, price float64) *txnbuild.ManageSellOffer {
	// in Kelp, all actions are performed in context of the base asset
	// a sell op sells base, and a buy op buys base/sells quote
	// I ran this using the buysell strategy with the amount set to 1000.0 units (of base on either side), with the buy op being displayed first:
	// 2020/12/23 02:35:59 submitting the following ops
	// 2020/12/23 02:35:59 &{Selling:{Code:COUPON Issuer:GBMMZMK2DC4FFP4CAI6KCVNCQ7WLO5A7DQU7EC7WGHRDQBZB763X4OQI} Buying:{} Amount:163.4735274 Price:6.1171984 OfferID:0 SourceAccount:0xc00047e5e0}
	// 2020/12/23 02:35:59 &{Selling:{} Buying:{Code:COUPON Issuer:GBMMZMK2DC4FFP4CAI6KCVNCQ7WLO5A7DQU7EC7WGHRDQBZB763X4OQI} Amount:1000.0000000 Price:0.1638006 OfferID:0 SourceAccount:0xc00047e6a0}
	// above, we can see that the selling operation (second op) has amount = 1000.0 and price = 0.16 which is correct
	// the buying op (first op) has amount set to 163.xx and price set to 6.xx, and we need to replicate this.
	// To do so, we need to:
	// set our Amount field on the manageSellOffer here equal to amount * price (so we get 1000 * 0.16 ~= 163.xx like displayed above)
	// set the Price field to 1/price (so we get 1/0.16 ~= 6.xx like displayed above)
	return &txnbuild.ManageSellOffer{
		Buying:  testBaseAsset,
		Selling: testQuoteAsset,
		Amount:  fmt.Sprintf("%.7f", amount*price),
		Price:   fmt.Sprintf("%.7f", 1/price),
	}
}

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name         string
		baseCapBase  *float64
		baseCapQuote *float64
		mode         volumeFilterMode
		action       queries.DailyVolumeAction
		marketIDs    []string
		accountIDs   []string
		wantErr      error
	}{
		{
			name:         "success - base + sell",
			baseCapBase:  pointy.Float64(1.0),
			baseCapQuote: nil,
			mode:         volumeFilterModeExact,
			action:       queries.DailyVolumeActionSell,
			marketIDs:    nil,
			accountIDs:   nil,
			wantErr:      nil,
		},
		{
			name:         "success - quote + buy",
			baseCapBase:  nil,
			baseCapQuote: pointy.Float64(1.0),
			mode:         volumeFilterModeExact,
			action:       queries.DailyVolumeActionBuy,
			marketIDs:    nil,
			accountIDs:   nil,
			wantErr:      nil,
		},
		{
			name:         "failure - 2 nil caps",
			baseCapBase:  nil,
			baseCapQuote: nil,
			mode:         volumeFilterModeExact,
			marketIDs:    nil,
			accountIDs:   nil,
			wantErr:      fmt.Errorf("invalid asset caps: only one asset cap can be non-nil, but both are nil"),
		},
		{
			name:         "failure - 2 non-nil caps",
			baseCapBase:  pointy.Float64(1.0),
			baseCapQuote: pointy.Float64(1.0),
			mode:         volumeFilterModeExact,
			marketIDs:    nil,
			accountIDs:   nil,
			wantErr:      fmt.Errorf("invalid asset caps: only one asset cap can be non-nil, but both are non-nil"),
		},
		{
			name:         "failure - invalid mode",
			baseCapBase:  pointy.Float64(1.0),
			baseCapQuote: nil,
			mode:         volumeFilterMode("hello"),
			action:       queries.DailyVolumeActionSell,
			marketIDs:    nil,
			accountIDs:   nil,
			wantErr:      fmt.Errorf("could not parse mode: invalid input mode 'hello'"),
		},
		{
			name:         "failure - invalid action",
			baseCapBase:  pointy.Float64(1.0),
			baseCapQuote: nil,
			mode:         volumeFilterModeExact,
			action:       queries.DailyVolumeAction("hello"),
			marketIDs:    nil,
			accountIDs:   nil,
			wantErr:      fmt.Errorf("could not parse action: invalid action value 'hello'"),
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			c := makeRawVolumeFilterConfig(k.baseCapBase, k.baseCapQuote, k.action, k.mode, k.marketIDs, k.accountIDs)
			gotErr := c.Validate()
			assert.Equal(t, k.wantErr, gotErr)
		})
	}
}
