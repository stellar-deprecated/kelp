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

func makeTestVolumeFilterConfig(baseCapInBase, baseCapInQuote float64, additionalMarketIDs, optionalAccountIDs []string, mode volumeFilterMode) *VolumeFilterConfig {
	var baseCapInBasePtr *float64
	if baseCapInBase >= 0 {
		baseCapInBasePtr = pointy.Float64(baseCapInBase)
	}

	var baseCapInQuotePtr *float64
	if baseCapInQuote >= 0 {
		baseCapInQuotePtr = pointy.Float64(baseCapInQuote)
	}

	return &VolumeFilterConfig{
		SellBaseAssetCapInBaseUnits:  baseCapInBasePtr,
		SellBaseAssetCapInQuoteUnits: baseCapInQuotePtr,
		mode:                         mode,
		additionalMarketIDs:          additionalMarketIDs,
		optionalAccountIDs:           optionalAccountIDs,
	}
}

func makeTestVolumeFilter(baseCapInBase, baseCapInQuote float64, additionalMarketIDs, optionalAccountIDs []string, mode volumeFilterMode) *volumeFilter {
	config := makeTestVolumeFilterConfig(baseCapInBase, baseCapInQuote, additionalMarketIDs, optionalAccountIDs, mode)
	query := mustMakeTestDailyVolumeQuery(optionalAccountIDs, additionalMarketIDs)

	return &volumeFilter{
		name:                   "volumeFilter",
		baseAsset:              testNativeAsset,
		quoteAsset:             testNativeAsset,
		config:                 config,
		dailyVolumeByDateQuery: query,
	}
}

func TestMakeFilterVolume(t *testing.T) {
	testAssetDisplayFn := model.MakeSdexMappedAssetDisplayFn(map[model.Asset]hProtocol.Asset{model.Asset("XLM"): testNativeAsset})

	testCases := []struct {
		name           string
		config         *VolumeFilterConfig
		baseCapInBase  float64
		baseCapInQuote float64
		marketIDs      []string
		accountIDs     []string
		mode           volumeFilterMode
	}{
		// TODO DS Confirm the empty config fails once validation is added to the constructor
		{
			name:           "empty config",
			baseCapInBase:  -1.,
			baseCapInQuote: -1.,
			marketIDs:      []string{},
			accountIDs:     []string{},
			mode:           volumeFilterModeExact,
		},
		{
			name:           "non nil cap in base, nil cap in quote",
			baseCapInBase:  1.0,
			baseCapInQuote: -1.0,
			marketIDs:      []string{},
			accountIDs:     []string{},
			mode:           volumeFilterModeExact,
		},
		{
			name:           "nil cap in base, non nil cap in quote",
			baseCapInBase:  -1.0,
			baseCapInQuote: 1.0,
			marketIDs:      []string{},
			accountIDs:     []string{},
			mode:           volumeFilterModeExact,
		},
		{
			name:           "exact mode",
			baseCapInBase:  1.0,
			baseCapInQuote: -1.0,
			marketIDs:      []string{},
			accountIDs:     []string{},
			mode:           volumeFilterModeExact,
		},
		{
			name:           "ignore mode",
			baseCapInBase:  1.0,
			baseCapInQuote: -1.0,
			marketIDs:      []string{},
			accountIDs:     []string{},
			mode:           volumeFilterModeIgnore,
		},
		{
			name:           "1 market id",
			baseCapInBase:  1.0,
			baseCapInQuote: -1.0,
			marketIDs:      []string{"marketID"},
			accountIDs:     []string{},
			mode:           volumeFilterModeExact,
		},
		{
			name:           "2 market ids",
			baseCapInBase:  1.0,
			baseCapInQuote: -1.0,
			marketIDs:      []string{"marketID1", "marketID2"},
			accountIDs:     []string{},
			mode:           volumeFilterModeExact,
		},
		{
			name:           "2 dupe market ids, 1 distinct",
			baseCapInBase:  1.0,
			baseCapInQuote: -1.0,
			marketIDs:      []string{"marketID1", "marketID1", "marketID2"},
			accountIDs:     []string{},
			mode:           volumeFilterModeExact,
		},
		{
			name:           "1 account id",
			baseCapInBase:  1.0,
			baseCapInQuote: -1.0,
			marketIDs:      []string{},
			accountIDs:     []string{"accountID"},
			mode:           volumeFilterModeExact,
		},
		{
			name:           "2 account ids",
			baseCapInBase:  1.0,
			baseCapInQuote: -1.0,
			marketIDs:      []string{},
			accountIDs:     []string{"accountID1", "accountID2"},
			mode:           volumeFilterModeExact,
		},
		{
			name:           "account and market ids",
			baseCapInBase:  1.0,
			baseCapInQuote: -1.0,
			marketIDs:      []string{"marketID"},
			accountIDs:     []string{"accountID"},
			mode:           volumeFilterModeExact,
		},
	}

	configValue := ""
	exchangeName := ""
	tradingPair := &model.TradingPair{Base: "XLM", Quote: "XLM"}

	for _, k := range testCases {
		config := makeTestVolumeFilterConfig(k.baseCapInBase, k.baseCapInQuote, k.marketIDs, k.accountIDs, k.mode)
		wantFilter := makeTestVolumeFilter(k.baseCapInBase, k.baseCapInQuote, k.marketIDs, k.accountIDs, k.mode)
		t.Run(k.name, func(t *testing.T) {
			actual, e := makeFilterVolume(
				configValue,
				exchangeName,
				tradingPair,
				testAssetDisplayFn,
				testNativeAsset,
				testNativeAsset,
				&sql.DB{},
				config,
			)

			assert.Nil(t, e)
			assert.Equal(t, wantFilter, actual)
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

func TestVolumeFilterFn(t *testing.T) {
	testCases := []struct {
		name               string
		filter             *volumeFilter
		sellBaseCapInBase  *float64
		sellBaseCapInQuote *float64
		otbBaseCap         float64
		otbQuoteCap        float64
		tbbBaseCap         float64
		tbbQuoteCap        float64
		price              string
		inputAmount        string
		wantAmount         string
		wantTbbBaseCap     float64
		wantTbbQuoteCap    float64
	}{
		{
			name:               "selling, base units sell cap, don't keep selling base",
			sellBaseCapInBase:  pointy.Float64(0.0),
			sellBaseCapInQuote: nil,
			otbBaseCap:         0.0,
			otbQuoteCap:        0.0,
			tbbBaseCap:         0.0,
			tbbQuoteCap:        0.0,
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "",
			wantTbbBaseCap:     0.0,
			wantTbbQuoteCap:    0.0,
		},
		{
			name:               "selling, base units sell cap, keep selling base",
			sellBaseCapInBase:  pointy.Float64(1.0),
			sellBaseCapInQuote: nil,
			otbBaseCap:         0.0,
			otbQuoteCap:        0.0,
			tbbBaseCap:         0.0,
			tbbQuoteCap:        0.0,
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "1.0000000",
			wantTbbBaseCap:     1.0,
			wantTbbQuoteCap:    2.0,
		},
		{
			name:               "selling, quote units sell cap, don't keep selling quote",
			sellBaseCapInBase:  nil,
			sellBaseCapInQuote: pointy.Float64(0),
			otbBaseCap:         0.0,
			otbQuoteCap:        0.0,
			tbbBaseCap:         0.0,
			tbbQuoteCap:        0.0,
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "",
			wantTbbBaseCap:     0.0,
			wantTbbQuoteCap:    0.0,
		},
		{
			name:               "selling, quote units sell cap, keep selling quote",
			sellBaseCapInBase:  nil,
			sellBaseCapInQuote: pointy.Float64(1.),
			otbBaseCap:         0.0,
			otbQuoteCap:        0.0,
			tbbBaseCap:         0.0,
			tbbQuoteCap:        0.0,
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "0.5000000",
			wantTbbBaseCap:     0.5,
			wantTbbQuoteCap:    1.0,
		},
		{
			name:               "selling, base and quote units sell cap, keep selling base and quote",
			sellBaseCapInBase:  pointy.Float64(1.),
			sellBaseCapInQuote: pointy.Float64(1.),
			otbBaseCap:         0.0,
			otbQuoteCap:        0.0,
			tbbBaseCap:         0.0,
			tbbQuoteCap:        0.0,
			price:              "2.0",
			inputAmount:        "100.0",
			wantAmount:         "0.5000000",
			wantTbbBaseCap:     0.5,
			wantTbbQuoteCap:    1.0,
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			marketIDs := []string{}
			accountIDs := []string{}
			mode := volumeFilterModeExact
			dailyOTB := makeTestVolumeFilterConfig(k.otbBaseCap, k.otbQuoteCap, marketIDs, accountIDs, mode)
			dailyTBB := makeTestVolumeFilterConfig(k.tbbBaseCap, k.tbbQuoteCap, marketIDs, accountIDs, mode)
			wantTBB := makeTestVolumeFilterConfig(k.wantTbbBaseCap, k.wantTbbQuoteCap, marketIDs, accountIDs, mode)
			op := makeManageSellOffer(k.price, k.inputAmount)
			wantOp := makeManageSellOffer(k.price, k.wantAmount)

			actual, e := volumeFilterFn(dailyOTB, dailyTBB, op, testNativeAsset, testNativeAsset, k.sellBaseCapInBase, k.sellBaseCapInQuote, volumeFilterModeExact)

			assert.Nil(t, e)
			assert.Equal(t, wantOp, actual)
			assert.Equal(t, wantTBB, dailyTBB)
		})
	}
}
