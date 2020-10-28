package plugins

import (
	"database/sql"
	"testing"

	"github.com/openlyinc/pointy"
	"github.com/stellar/kelp/queries"
	"github.com/stellar/kelp/support/utils"

	"github.com/stellar/go/txnbuild"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/model"
	"github.com/stretchr/testify/assert"
)

var testNativeAsset hProtocol.Asset = hProtocol.Asset{Type: "native"}

func mustMakeTestDailyVolumeQuery(optionalAccountIDs, additionalMarketIDs []string) *queries.DailyVolumeByDate {
	marketID := MakeMarketID("", "native", "native")
	marketIDs := utils.Dedupe(append([]string{marketID}, additionalMarketIDs...))

	query, e := queries.MakeDailyVolumeByDateForMarketIdsAction(&sql.DB{}, marketIDs, "sell", optionalAccountIDs)
	if e != nil {
		panic(e)
	}

	return query
}

func TestMakeFilterVolume(t *testing.T) {
	testAssetDisplayFn := model.MakeSdexMappedAssetDisplayFn(map[model.Asset]hProtocol.Asset{model.Asset("XLM"): testNativeAsset})

	testCases := []struct {
		name           string
		assetDisplayFn model.AssetDisplayFn
		config         *VolumeFilterConfig

		wantError        error
		wantSubmitFilter SubmitFilter
	}{
		// TODO DS Confirm the empty config fails once validation is added to the constructor
		{
			name:      "empty config",
			config:    &VolumeFilterConfig{},
			wantError: nil,
			wantSubmitFilter: &volumeFilter{
				name:                   "volumeFilter",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{},
				dailyVolumeByDateQuery: mustMakeTestDailyVolumeQuery([]string{}, []string{}),
			},
		},
		{
			name: "non nil cap in base, nil cap in quote",
			config: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
				mode:                        volumeFilterModeExact,
				additionalMarketIDs:         []string{},
				optionalAccountIDs:          []string{},
			},
			wantError: nil,
			wantSubmitFilter: &volumeFilter{
				name:       "volumeFilter",
				baseAsset:  testNativeAsset,
				quoteAsset: testNativeAsset,
				config: &VolumeFilterConfig{
					SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
					mode:                        volumeFilterModeExact,
					additionalMarketIDs:         []string{},
					optionalAccountIDs:          []string{},
				},
				dailyVolumeByDateQuery: mustMakeTestDailyVolumeQuery([]string{}, []string{}),
			},
		},
		{
			name: "nil cap in base, non nil cap in quote",
			config: &VolumeFilterConfig{
				SellBaseAssetCapInQuoteUnits: pointy.Float64(1.0),
				mode:                         volumeFilterModeExact,
				additionalMarketIDs:          []string{},
				optionalAccountIDs:           []string{},
			},
			wantError: nil,
			wantSubmitFilter: &volumeFilter{
				name:       "volumeFilter",
				baseAsset:  testNativeAsset,
				quoteAsset: testNativeAsset,
				config: &VolumeFilterConfig{
					SellBaseAssetCapInQuoteUnits: pointy.Float64(1.0),
					mode:                         volumeFilterModeExact,
					additionalMarketIDs:          []string{},
					optionalAccountIDs:           []string{},
				},
				dailyVolumeByDateQuery: mustMakeTestDailyVolumeQuery([]string{}, []string{}),
			},
		},
		{
			name: "exact mode",
			config: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
				mode:                        volumeFilterModeExact,
				additionalMarketIDs:         []string{},
				optionalAccountIDs:          []string{},
			},
			wantError: nil,
			wantSubmitFilter: &volumeFilter{
				name:       "volumeFilter",
				baseAsset:  testNativeAsset,
				quoteAsset: testNativeAsset,
				config: &VolumeFilterConfig{
					SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
					mode:                        volumeFilterModeExact,
					additionalMarketIDs:         []string{},
					optionalAccountIDs:          []string{},
				},
				dailyVolumeByDateQuery: mustMakeTestDailyVolumeQuery([]string{}, []string{}),
			},
		},
		{
			name: "ignore mode",
			config: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
				mode:                        volumeFilterModeIgnore,
				additionalMarketIDs:         []string{},
				optionalAccountIDs:          []string{},
			},
			wantError: nil,
			wantSubmitFilter: &volumeFilter{
				name:       "volumeFilter",
				baseAsset:  testNativeAsset,
				quoteAsset: testNativeAsset,
				config: &VolumeFilterConfig{
					SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
					mode:                        volumeFilterModeIgnore,
					additionalMarketIDs:         []string{},
					optionalAccountIDs:          []string{},
				},
				dailyVolumeByDateQuery: mustMakeTestDailyVolumeQuery([]string{}, []string{}),
			},
		},
		{
			name: "1 market id",
			config: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
				mode:                        volumeFilterModeIgnore,
				additionalMarketIDs:         []string{"marketID"},
				optionalAccountIDs:          []string{},
			},
			wantError: nil,
			wantSubmitFilter: &volumeFilter{
				name:       "volumeFilter",
				baseAsset:  testNativeAsset,
				quoteAsset: testNativeAsset,
				config: &VolumeFilterConfig{
					SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
					mode:                        volumeFilterModeIgnore,
					additionalMarketIDs:         []string{"marketID"},
					optionalAccountIDs:          []string{},
				},
				dailyVolumeByDateQuery: mustMakeTestDailyVolumeQuery([]string{}, []string{"marketID"}),
			},
		},
		{
			name: "2 market ids",
			config: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
				mode:                        volumeFilterModeIgnore,
				additionalMarketIDs:         []string{"marketID1", "marketID2"},
				optionalAccountIDs:          []string{},
			},
			wantError: nil,
			wantSubmitFilter: &volumeFilter{
				name:       "volumeFilter",
				baseAsset:  testNativeAsset,
				quoteAsset: testNativeAsset,
				config: &VolumeFilterConfig{
					SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
					mode:                        volumeFilterModeIgnore,
					additionalMarketIDs:         []string{"marketID1", "marketID2"},
					optionalAccountIDs:          []string{},
				},
				dailyVolumeByDateQuery: mustMakeTestDailyVolumeQuery([]string{}, []string{"marketID1", "marketID2"}),
			},
		},
		{
			name: "2 dupe market ids, 1 distinct",
			config: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
				mode:                        volumeFilterModeIgnore,
				additionalMarketIDs:         []string{"marketID1", "marketID1", "marketID2"},
				optionalAccountIDs:          []string{},
			},
			wantError: nil,
			wantSubmitFilter: &volumeFilter{
				name:       "volumeFilter",
				baseAsset:  testNativeAsset,
				quoteAsset: testNativeAsset,
				config: &VolumeFilterConfig{
					SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
					mode:                        volumeFilterModeIgnore,
					additionalMarketIDs:         []string{"marketID1", "marketID1", "marketID2"},
					optionalAccountIDs:          []string{},
				},
				dailyVolumeByDateQuery: mustMakeTestDailyVolumeQuery([]string{}, []string{"marketID1", "marketID1", "marketID2"}),
			},
		},
		{
			name: "1 account id",
			config: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
				mode:                        volumeFilterModeIgnore,
				additionalMarketIDs:         []string{},
				optionalAccountIDs:          []string{"accountID"},
			},
			wantError: nil,
			wantSubmitFilter: &volumeFilter{
				name:       "volumeFilter",
				baseAsset:  testNativeAsset,
				quoteAsset: testNativeAsset,
				config: &VolumeFilterConfig{
					SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
					mode:                        volumeFilterModeIgnore,
					additionalMarketIDs:         []string{},
					optionalAccountIDs:          []string{"accountID"},
				},
				dailyVolumeByDateQuery: mustMakeTestDailyVolumeQuery([]string{"accountID"}, []string{}),
			},
		},
		{
			name: "2 account ids",
			config: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
				mode:                        volumeFilterModeIgnore,
				additionalMarketIDs:         []string{},
				optionalAccountIDs:          []string{"accountID1", "accountID2"},
			},
			wantError: nil,
			wantSubmitFilter: &volumeFilter{
				name:       "volumeFilter",
				baseAsset:  testNativeAsset,
				quoteAsset: testNativeAsset,
				config: &VolumeFilterConfig{
					SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
					mode:                        volumeFilterModeIgnore,
					additionalMarketIDs:         []string{},
					optionalAccountIDs:          []string{"accountID1", "accountID2"},
				},
				dailyVolumeByDateQuery: mustMakeTestDailyVolumeQuery([]string{"accountID1", "accountID2"}, []string{}),
			},
		},
		{
			name: "account and market ids",
			config: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
				mode:                        volumeFilterModeIgnore,
				additionalMarketIDs:         []string{"marketID"},
				optionalAccountIDs:          []string{"accountID"},
			},
			wantError: nil,
			wantSubmitFilter: &volumeFilter{
				name:       "volumeFilter",
				baseAsset:  testNativeAsset,
				quoteAsset: testNativeAsset,
				config: &VolumeFilterConfig{
					SellBaseAssetCapInBaseUnits: pointy.Float64(1.0),
					mode:                        volumeFilterModeIgnore,
					additionalMarketIDs:         []string{"marketID"},
					optionalAccountIDs:          []string{"accountID"},
				},
				dailyVolumeByDateQuery: mustMakeTestDailyVolumeQuery([]string{"accountID"}, []string{"marketID"}),
			},
		},
	}

	configValue := ""
	exchangeName := ""
	tradingPair := &model.TradingPair{Base: "XLM", Quote: "XLM"}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			actual, e := makeFilterVolume(
				configValue,
				exchangeName,
				tradingPair,
				testAssetDisplayFn,
				testNativeAsset,
				testNativeAsset,
				&sql.DB{},
				k.config,
			)

			assert.Equal(t, k.wantError, e)
			assert.Equal(t, k.wantSubmitFilter, actual)
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

func makeTestVolumeFilterConfig(baseCap, quoteCap float64) *VolumeFilterConfig {
	return &VolumeFilterConfig{
		SellBaseAssetCapInBaseUnits:  pointy.Float64(baseCap),
		SellBaseAssetCapInQuoteUnits: pointy.Float64(quoteCap),
		mode:                         volumeFilterModeExact,
		additionalMarketIDs:          []string{},
		optionalAccountIDs:           []string{},
	}
}

func TestVolumeFilterFn(t *testing.T) {
	dailyVolumeByDateQuery := mustMakeTestDailyVolumeQuery([]string{}, []string{})

	emptyFilter := &volumeFilter{
		name:                   "volumeFilter",
		configValue:            "",
		baseAsset:              testNativeAsset,
		quoteAsset:             testNativeAsset,
		config:                 &VolumeFilterConfig{},
		dailyVolumeByDateQuery: dailyVolumeByDateQuery,
	}

	testCases := []struct {
		name            string
		filter          *volumeFilter
		otbBaseCap      float64
		otbQuoteCap     float64
		tbbBaseCap      float64
		tbbQuoteCap     float64
		price           string
		inputAmount     string
		wantAmount      string
		wantError       error
		wantTbbBaseCap  float64
		wantTbbQuoteCap float64
	}{
		{
			name:            "selling, no filter sell caps",
			filter:          emptyFilter,
			otbBaseCap:      0.0,
			otbQuoteCap:     0.0,
			tbbBaseCap:      0.0,
			tbbQuoteCap:     0.0,
			price:           "2.0",
			inputAmount:     "100.0",
			wantAmount:      "100.0",
			wantError:       nil,
			wantTbbBaseCap:  100.0,
			wantTbbQuoteCap: 200.0,
		},
		{
			name: "selling, base units sell cap, don't keep selling base",
			filter: &volumeFilter{
				name:                   "volumeFilter",
				configValue:            "",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{SellBaseAssetCapInBaseUnits: pointy.Float64(0.0)},
				dailyVolumeByDateQuery: dailyVolumeByDateQuery,
			},
			otbBaseCap:      0.0,
			otbQuoteCap:     0.0,
			tbbBaseCap:      0.0,
			tbbQuoteCap:     0.0,
			price:           "2.0",
			inputAmount:     "100.0",
			wantAmount:      "",
			wantError:       nil,
			wantTbbBaseCap:  0.0,
			wantTbbQuoteCap: 0.0,
		},
		{
			name: "selling, base units sell cap, keep selling base",
			filter: &volumeFilter{
				name:                   "volumeFilter",
				configValue:            "",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{SellBaseAssetCapInBaseUnits: pointy.Float64(1.0), mode: volumeFilterModeExact},
				dailyVolumeByDateQuery: dailyVolumeByDateQuery,
			},
			otbBaseCap:      0.0,
			otbQuoteCap:     0.0,
			tbbBaseCap:      0.0,
			tbbQuoteCap:     0.0,
			price:           "2.0",
			inputAmount:     "100.0",
			wantAmount:      "1.0000000",
			wantError:       nil,
			wantTbbBaseCap:  1.0,
			wantTbbQuoteCap: 2.0,
		},
		{
			name: "selling, quote units sell cap, don't keep selling quote",
			filter: &volumeFilter{
				name:                   "volumeFilter",
				configValue:            "",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{SellBaseAssetCapInQuoteUnits: pointy.Float64(0.0)},
				dailyVolumeByDateQuery: dailyVolumeByDateQuery,
			},
			otbBaseCap:      0.0,
			otbQuoteCap:     0.0,
			tbbBaseCap:      0.0,
			tbbQuoteCap:     0.0,
			price:           "2.0",
			inputAmount:     "100.0",
			wantAmount:      "",
			wantError:       nil,
			wantTbbBaseCap:  0.0,
			wantTbbQuoteCap: 0.0,
		},
		{
			name: "selling, quote units sell cap, keep selling quote",
			filter: &volumeFilter{
				name:                   "volumeFilter",
				configValue:            "",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{SellBaseAssetCapInQuoteUnits: pointy.Float64(1.0), mode: volumeFilterModeExact},
				dailyVolumeByDateQuery: dailyVolumeByDateQuery,
			},
			otbBaseCap:      0.0,
			otbQuoteCap:     0.0,
			tbbBaseCap:      0.0,
			tbbQuoteCap:     0.0,
			price:           "2.0",
			inputAmount:     "100.0",
			wantAmount:      "0.5000000",
			wantError:       nil,
			wantTbbBaseCap:  0.5,
			wantTbbQuoteCap: 1.0,
		},
		{
			name: "selling, base and quote units sell cap, keep selling base and quote",
			filter: &volumeFilter{
				name:                   "volumeFilter",
				configValue:            "",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{SellBaseAssetCapInBaseUnits: pointy.Float64(1.0), SellBaseAssetCapInQuoteUnits: pointy.Float64(1.0), mode: volumeFilterModeExact},
				dailyVolumeByDateQuery: dailyVolumeByDateQuery,
			},
			otbBaseCap:      0.0,
			otbQuoteCap:     0.0,
			tbbBaseCap:      0.0,
			tbbQuoteCap:     0.0,
			price:           "2.0",
			inputAmount:     "100.0",
			wantAmount:      "0.5000000",
			wantError:       nil,
			wantTbbBaseCap:  0.5,
			wantTbbQuoteCap: 1.0,
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			dailyOTB := makeTestVolumeFilterConfig(k.otbBaseCap, k.otbQuoteCap)
			dailyTBB := makeTestVolumeFilterConfig(k.tbbBaseCap, k.tbbQuoteCap)
			wantTBB := makeTestVolumeFilterConfig(k.wantTbbBaseCap, k.wantTbbQuoteCap)
			op := makeManageSellOffer(k.price, k.inputAmount)
			wantOp := makeManageSellOffer(k.price, k.wantAmount)

			actual, e := k.filter.volumeFilterFn(dailyOTB, dailyTBB, op)
			assert.Equal(t, k.wantError, e)
			assert.Equal(t, wantOp, actual)
			assert.Equal(t, wantTBB, dailyTBB)
		})
	}
}
