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

func makeWantVolumeFilter(config *VolumeFilterConfig, marketIDs []string, accountIDs []string, action string) *volumeFilter {
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
			// this lets us run the for-loop below for both base and quote units within the config
			baseCapInBaseConfig := makeRawVolumeFilterConfig(
				pointy.Float64(1.0),
				nil,
				m,
				k.marketIDs,
				k.accountIDs,
			)
			baseCapInQuoteConfig := makeRawVolumeFilterConfig(
				nil,
				pointy.Float64(1.0),
				m,
				k.marketIDs,
				k.accountIDs,
			)
			for _, config := range []*VolumeFilterConfig{baseCapInBaseConfig, baseCapInQuoteConfig} {
				// configType is used to represent the type of config when printing test name
				configType := "quote"
				if config.SellBaseAssetCapInBaseUnits != nil {
					configType = "base"
				}

				// TODO DS Vary filter action between buy and sell, once buy logic is implemented.
				wantFilter := makeWantVolumeFilter(config, k.wantMarketIDs, k.accountIDs, "sell")
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

func TestVolumeFilterFn(t *testing.T) {
	testCases := []struct {
		name               string
		mode               volumeFilterMode
		sellBaseCapInBase  *float64
		sellBaseCapInQuote *float64
		otbBaseCap         *float64
		otbQuoteCap        *float64
		tbbBaseCap         *float64
		tbbQuoteCap        *float64
		price              string
		inputAmount        string
		wantAmount         string
		wantTbbBaseCap     *float64
		wantTbbQuoteCap    *float64
	}{
		{
			name:               "selling, base units sell cap, don't keep selling base, exact mode",
			mode:               volumeFilterModeExact,
			sellBaseCapInBase:  pointy.Float64(0.0),
			sellBaseCapInQuote: nil,
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "",
			wantTbbBaseCap:     pointy.Float64(0.0),
			wantTbbQuoteCap:    pointy.Float64(0.0),
		},
		{
			name:               "selling, base units sell cap, don't keep selling base, ignore mode",
			mode:               volumeFilterModeIgnore,
			sellBaseCapInBase:  pointy.Float64(0.0),
			sellBaseCapInQuote: nil,
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "",
			wantTbbBaseCap:     pointy.Float64(0.0),
			wantTbbQuoteCap:    pointy.Float64(0.0),
		},
		{
			name:               "selling, base units sell cap, keep selling base, new amount, exact mode",
			mode:               volumeFilterModeExact,
			sellBaseCapInBase:  pointy.Float64(1.0),
			sellBaseCapInQuote: nil,
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "1.0000000",
			wantTbbBaseCap:     pointy.Float64(1.0),
			wantTbbQuoteCap:    pointy.Float64(2.0),
		},
		{
			name:               "selling, base units sell cap, keep selling base, new amount, ignore mode",
			mode:               volumeFilterModeIgnore,
			sellBaseCapInBase:  pointy.Float64(1.0),
			sellBaseCapInQuote: nil,
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "",
			wantTbbBaseCap:     pointy.Float64(0.0),
			wantTbbQuoteCap:    pointy.Float64(0.0),
		},
		{
			name:               "selling, base units sell cap, keep selling base, exact mode",
			mode:               volumeFilterModeExact,
			sellBaseCapInBase:  pointy.Float64(1.0),
			sellBaseCapInQuote: nil,
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "1.0000000",
			wantTbbBaseCap:     pointy.Float64(1.0),
			wantTbbQuoteCap:    pointy.Float64(2.0),
		},
		{
			name:               "selling, base units sell cap, keep selling base, ignore mode",
			mode:               volumeFilterModeIgnore,
			sellBaseCapInBase:  pointy.Float64(1.0),
			sellBaseCapInQuote: nil,
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "",
			wantTbbBaseCap:     pointy.Float64(0.0),
			wantTbbQuoteCap:    pointy.Float64(0.0),
		},
		{
			name:               "selling, quote units sell cap, don't keep selling quote, exact mode",
			mode:               volumeFilterModeExact,
			sellBaseCapInBase:  nil,
			sellBaseCapInQuote: pointy.Float64(0),
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "",
			wantTbbBaseCap:     pointy.Float64(0.0),
			wantTbbQuoteCap:    pointy.Float64(0.0),
		},
		{
			name:               "selling, quote units sell cap, don't keep selling quote, ignore mode",
			mode:               volumeFilterModeIgnore,
			sellBaseCapInBase:  nil,
			sellBaseCapInQuote: pointy.Float64(0),
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "",
			wantTbbBaseCap:     pointy.Float64(0.0),
			wantTbbQuoteCap:    pointy.Float64(0.0),
		},
		{
			name:               "selling, quote units sell cap, keep selling quote, new amount, exact mode",
			mode:               volumeFilterModeExact,
			sellBaseCapInBase:  nil,
			sellBaseCapInQuote: pointy.Float64(1.0),
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "0.5000000",
			wantTbbBaseCap:     pointy.Float64(0.5),
			wantTbbQuoteCap:    pointy.Float64(1.0),
		},
		{
			name:               "selling, quote units sell cap, keep selling quote, new amount, ignore mode",
			mode:               volumeFilterModeIgnore,
			sellBaseCapInBase:  nil,
			sellBaseCapInQuote: pointy.Float64(1.0),
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "",
			wantTbbBaseCap:     pointy.Float64(0.0),
			wantTbbQuoteCap:    pointy.Float64(0.0),
		},
		{
			name:               "selling, quote units sell cap, keep selling quote, exact mode",
			mode:               volumeFilterModeExact,
			sellBaseCapInBase:  nil,
			sellBaseCapInQuote: pointy.Float64(1.0),
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "0.5000000",
			wantTbbBaseCap:     pointy.Float64(0.5),
			wantTbbQuoteCap:    pointy.Float64(1.0),
		},
		{
			name:               "selling, quote units sell cap, keep selling quote, ignore mode",
			mode:               volumeFilterModeIgnore,
			sellBaseCapInBase:  nil,
			sellBaseCapInQuote: pointy.Float64(1.0),
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "",
			wantTbbBaseCap:     pointy.Float64(0.0),
			wantTbbQuoteCap:    pointy.Float64(0.0),
		},
		{
			name:               "selling, base and quote units sell cap, keep selling base and quote, exact mode",
			mode:               volumeFilterModeExact,
			sellBaseCapInBase:  pointy.Float64(1.0),
			sellBaseCapInQuote: pointy.Float64(1.0),
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "0.5000000",
			wantTbbBaseCap:     pointy.Float64(0.5),
			wantTbbQuoteCap:    pointy.Float64(1.0),
		},
		{
			name:               "selling, base and quote units sell cap, keep selling base and quote, ignore mode",
			mode:               volumeFilterModeIgnore,
			sellBaseCapInBase:  pointy.Float64(1.0),
			sellBaseCapInQuote: pointy.Float64(1.0),
			otbBaseCap:         pointy.Float64(0.0),
			otbQuoteCap:        pointy.Float64(0.0),
			tbbBaseCap:         pointy.Float64(0.0),
			tbbQuoteCap:        pointy.Float64(0.0),
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "",
			wantTbbBaseCap:     pointy.Float64(0.0),
			wantTbbQuoteCap:    pointy.Float64(0.0),
		},
	}

	// we fix the marketIDs and accountIDs, since volumeFilterFn output does not depend on them
	marketIDs := []string{}
	accountIDs := []string{}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			dailyOTB := makeRawVolumeFilterConfig(k.otbBaseCap, k.otbQuoteCap, k.mode, marketIDs, accountIDs)
			dailyTBBAccumulator := makeRawVolumeFilterConfig(k.tbbBaseCap, k.tbbQuoteCap, k.mode, marketIDs, accountIDs)
			wantTBBAccumulator := makeRawVolumeFilterConfig(k.wantTbbBaseCap, k.wantTbbQuoteCap, k.mode, marketIDs, accountIDs)

			op := makeManageSellOffer(k.price, k.inputAmount)
			wantOp := makeManageSellOffer(k.price, k.wantAmount)

			lp := limitParameters{
				sellBaseAssetCapInBaseUnits:  k.sellBaseCapInBase,
				sellBaseAssetCapInQuoteUnits: k.sellBaseCapInQuote,
				mode:                         k.mode,
			}

			actual, e := volumeFilterFn(dailyOTB, dailyTBBAccumulator, op, utils.NativeAsset, utils.NativeAsset, lp)

			if !assert.Nil(t, e) {
				return
			}

			assert.Equal(t, wantOp, actual)
			assert.Equal(t, wantTBBAccumulator, dailyTBBAccumulator)
		})

	}
}

func makeManageSellOffer(price, amount string) *txnbuild.ManageSellOffer {
	if amount == "" {
		return nil
	}

	return &txnbuild.ManageSellOffer{
		Buying:  txnbuild.NativeAsset{},
		Selling: txnbuild.NativeAsset{},
		Price:   price,
		Amount:  amount,
	}
}
