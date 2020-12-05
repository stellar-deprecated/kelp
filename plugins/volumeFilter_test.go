package plugins

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/openlyinc/pointy"
	"github.com/stellar/kelp/queries"
	"github.com/stellar/kelp/support/utils"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/model"
	"github.com/stretchr/testify/assert"
)

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
	modes := []volumeFilterMode{volumeFilterModeExact, volumeFilterModeIgnore}

	testCases := []struct {
		name          string
		exchangeName  string
		marketIDs     []string
		accountIDs    []string
		wantMarketIDs []string
		wantFilter    *volumeFilter
	}{
		// TODO DS Confirm the empty config fails once validation is added to the constructor
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

	for _, k := range testCases {
		// this lets us test both types of modes when varying the market and account ids
		for _, m := range modes {
			// this lets us test both buy and sell
			// TODO DS Add buy action
			for _, action := range []queries.DailyVolumeAction{queries.DailyVolumeActionSell} {
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

					// TODO DS Vary filter action between buy and sell, once buy logic is implemented.
					wantFilter := makeWantVolumeFilter(config, k.wantMarketIDs, k.accountIDs, action)
					t.Run(fmt.Sprintf("%s/%s/%s", k.name, configType, m), func(t *testing.T) {
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

func TestVolumeFilterFn(t *testing.T) {
	testCases := []struct {
		name           string
		mode           volumeFilterMode
		baseCapInBase  *float64
		baseCapInQuote *float64
		otbBase        *float64
		otbQuote       *float64
		tbbBase        *float64
		tbbQuote       *float64
		inputOp        *txnbuild.ManageSellOffer
		wantOp         *txnbuild.ManageSellOffer
		wantTbbBase    *float64
		wantTbbQuote   *float64
	}{
		{
			name:           "1. selling, base units sell cap, don't keep selling base, exact mode",
			mode:           volumeFilterModeExact,
			baseCapInBase:  pointy.Float64(0.0),
			baseCapInQuote: nil,
			otbBase:        pointy.Float64(0.0),
			otbQuote:       pointy.Float64(0.0),
			tbbBase:        pointy.Float64(0.0),
			tbbQuote:       pointy.Float64(0.0),
			inputOp:        makeManageSellOffer("2.0", "100.0"),
			wantOp:         nil,
			wantTbbBase:    pointy.Float64(0.0),
			wantTbbQuote:   pointy.Float64(0.0),
		},
		{
			name:           "2. selling, base units sell cap, don't keep selling base, ignore mode",
			mode:           volumeFilterModeIgnore,
			baseCapInBase:  pointy.Float64(0.0),
			baseCapInQuote: nil,
			otbBase:        pointy.Float64(0.0),
			otbQuote:       pointy.Float64(0.0),
			tbbBase:        pointy.Float64(0.0),
			tbbQuote:       pointy.Float64(0.0),
			inputOp:        makeManageSellOffer("2.0", "100.0"),
			wantOp:         nil,
			wantTbbBase:    pointy.Float64(0.0),
			wantTbbQuote:   pointy.Float64(0.0),
		},
		{
			name:           "3. selling, base units sell cap, keep selling base, exact mode",
			mode:           volumeFilterModeExact,
			baseCapInBase:  pointy.Float64(1.0),
			baseCapInQuote: nil,
			otbBase:        pointy.Float64(0.0),
			otbQuote:       pointy.Float64(0.0),
			tbbBase:        pointy.Float64(0.0),
			tbbQuote:       pointy.Float64(0.0),
			inputOp:        makeManageSellOffer("2.0", "100.0"),
			wantOp:         makeManageSellOffer("2.0", "1.0000000"),
			wantTbbBase:    pointy.Float64(1.0),
			wantTbbQuote:   pointy.Float64(2.0),
		},
		{
			name:           "4. selling, base units sell cap, keep selling base, ignore mode",
			mode:           volumeFilterModeIgnore,
			baseCapInBase:  pointy.Float64(1.0),
			baseCapInQuote: nil,
			otbBase:        pointy.Float64(0.0),
			otbQuote:       pointy.Float64(0.0),
			tbbBase:        pointy.Float64(0.0),
			tbbQuote:       pointy.Float64(0.0),
			inputOp:        makeManageSellOffer("2.0", "100.0"),
			wantOp:         nil,
			wantTbbBase:    pointy.Float64(0.0),
			wantTbbQuote:   pointy.Float64(0.0),
		},
		{
			name:           "7. selling, quote units sell cap, don't keep selling quote, exact mode",
			mode:           volumeFilterModeExact,
			baseCapInBase:  nil,
			baseCapInQuote: pointy.Float64(0),
			otbBase:        pointy.Float64(0.0),
			otbQuote:       pointy.Float64(0.0),
			tbbBase:        pointy.Float64(0.0),
			tbbQuote:       pointy.Float64(0.0),
			inputOp:        makeManageSellOffer("2.0", "100.0"),
			wantOp:         nil,
			wantTbbBase:    pointy.Float64(0.0),
			wantTbbQuote:   pointy.Float64(0.0),
		},
		{
			name:           "8. selling, quote units sell cap, don't keep selling quote, ignore mode",
			mode:           volumeFilterModeIgnore,
			baseCapInBase:  nil,
			baseCapInQuote: pointy.Float64(0),
			otbBase:        pointy.Float64(0.0),
			otbQuote:       pointy.Float64(0.0),
			tbbBase:        pointy.Float64(0.0),
			tbbQuote:       pointy.Float64(0.0),
			inputOp:        makeManageSellOffer("2.0", "100.0"),
			wantOp:         nil,
			wantTbbBase:    pointy.Float64(0.0),
			wantTbbQuote:   pointy.Float64(0.0),
		},
		{
			name:           "9. selling, quote units sell cap, keep selling quote, exact mode",
			mode:           volumeFilterModeExact,
			baseCapInBase:  nil,
			baseCapInQuote: pointy.Float64(1.0),
			otbBase:        pointy.Float64(0.0),
			otbQuote:       pointy.Float64(0.0),
			tbbBase:        pointy.Float64(0.0),
			tbbQuote:       pointy.Float64(0.0),
			inputOp:        makeManageSellOffer("2.0", "100.0"),
			wantOp:         makeManageSellOffer("2.0", "0.5000000"),
			wantTbbBase:    pointy.Float64(0.5),
			wantTbbQuote:   pointy.Float64(1.0),
		},
		{
			name:           "10. selling, quote units sell cap, keep selling quote, ignore mode",
			mode:           volumeFilterModeIgnore,
			baseCapInBase:  nil,
			baseCapInQuote: pointy.Float64(1.0),
			otbBase:        pointy.Float64(0.0),
			otbQuote:       pointy.Float64(0.0),
			tbbBase:        pointy.Float64(0.0),
			tbbQuote:       pointy.Float64(0.0),
			inputOp:        makeManageSellOffer("2.0", "100.0"),
			wantOp:         nil,
			wantTbbBase:    pointy.Float64(0.0),
			wantTbbQuote:   pointy.Float64(0.0),
		},
	}

	// we fix the marketIDs and accountIDs, since volumeFilterFn output does not depend on them
	marketIDs := []string{}
	accountIDs := []string{}

	for _, k := range testCases {
		for _, action := range []queries.DailyVolumeAction{queries.DailyVolumeActionSell} {
			t.Run(k.name, func(t *testing.T) {
				// exactly one of the two cap values must be set
				if k.baseCapInBase == nil && k.baseCapInQuote == nil {
					assert.Fail(t, "either one of the two cap values must be set")
					return
				}

				if k.baseCapInBase != nil && k.baseCapInQuote != nil {
					assert.Fail(t, "both of the cap values cannot be set")
					return
				}

				dailyOTB := makeRawVolumeFilterConfig(k.otbBase, k.otbQuote, action, k.mode, marketIDs, accountIDs)
				dailyTBBAccumulator := makeRawVolumeFilterConfig(k.tbbBase, k.tbbQuote, action, k.mode, marketIDs, accountIDs)
				lp := limitParameters{
					baseAssetCapInBaseUnits:  k.baseCapInBase,
					baseAssetCapInQuoteUnits: k.baseCapInQuote,
					mode:                     k.mode,
				}

				actual, e := volumeFilterFn(dailyOTB, dailyTBBAccumulator, k.inputOp, utils.NativeAsset, utils.NativeAsset, lp)
				if !assert.Nil(t, e) {
					return
				}
				assert.Equal(t, k.wantOp, actual)

				wantTBBAccumulator := makeRawVolumeFilterConfig(k.wantTbbBase, k.wantTbbQuote, action, k.mode, marketIDs, accountIDs)
				assert.Equal(t, wantTBBAccumulator, dailyTBBAccumulator)
			})
		}
	}
}

func makeManageSellOffer(price string, amount string) *txnbuild.ManageSellOffer {
	return &txnbuild.ManageSellOffer{
		Buying:  txnbuild.NativeAsset{},
		Selling: txnbuild.NativeAsset{},
		Price:   price,
		Amount:  amount,
	}
}
