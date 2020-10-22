package plugins

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/stellar/go/txnbuild"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/postgresdb"
	"github.com/stellar/kelp/support/utils"
	"github.com/stretchr/testify/assert"
)

func connectTestDb() *sql.DB {
	postgresDbConfig := &postgresdb.Config{
		Host:      "localhost",
		Port:      5432,
		DbName:    "test_database",
		User:      os.Getenv("POSTGRES_USER"),
		SSLEnable: false,
	}

	_, e := postgresdb.CreateDatabaseIfNotExists(postgresDbConfig)
	if e != nil {
		panic(e)
	}

	db, e := sql.Open("postgres", postgresDbConfig.MakeConnectString())
	if e != nil {
		panic(e)
	}
	return db
}

var testNativeAsset hProtocol.Asset = hProtocol.Asset{Type: "native"}

func TestMakeFilterVolume(t *testing.T) {
	testAssetDisplayFn := model.AssetDisplayFn(func(asset model.Asset) (string, error) {
		sdexAssetMap := map[model.Asset]hProtocol.Asset{
			model.Asset("XLM"): testNativeAsset,
		}
		assetString, ok := sdexAssetMap[asset]
		if !ok {
			return "", fmt.Errorf("cannot recognize the asset %s", string(asset))
		}
		return utils.Asset2String(assetString), nil
	})

	testEmptyVolumeConfig := &VolumeFilterConfig{}
	db := connectTestDb()

	emptyConfigVolumeQuery, e := makeDailyVolumeByDateQuery(db, "", "native", "native", []string{}, []string{})
	if e != nil {
		t.Log("could not make empty config volume query")
		return
	}

	marketIDs := []string{"marketID"}
	marketConfigVolumeQuery, e := makeDailyVolumeByDateQuery(db, "", "native", "native", []string{}, marketIDs)
	if e != nil {
		t.Log("could not make market config volume query")
		return
	}

	accountIDs := []string{"accountID"}
	accountConfigVolumeQuery, e := makeDailyVolumeByDateQuery(db, "", "native", "native", accountIDs, []string{})
	if e != nil {
		t.Log("could not make account config volume query")
		return
	}

	accountMarketConfigVolumeQuery, e := makeDailyVolumeByDateQuery(db, "", "native", "native", accountIDs, marketIDs)
	if e != nil {
		t.Log("could not make account and market config volume query")
		return
	}

	testCases := []struct {
		name           string
		configValue    string
		exchangeName   string
		tradingPair    *model.TradingPair
		assetDisplayFn model.AssetDisplayFn
		baseAsset      hProtocol.Asset
		quoteAsset     hProtocol.Asset
		db             *sql.DB
		config         *VolumeFilterConfig

		wantError        error
		wantSubmitFilter SubmitFilter
	}{
		// Note: config can be additional market IDs, make sure to test additionalMarketIds and optionalAccountIDs
		{
			name:             "failure - display base",
			configValue:      "",
			exchangeName:     "",
			tradingPair:      &model.TradingPair{Base: "FAIL", Quote: ""},
			assetDisplayFn:   testAssetDisplayFn,
			baseAsset:        hProtocol.Asset{},
			quoteAsset:       hProtocol.Asset{},
			db:               db,
			config:           &VolumeFilterConfig{},
			wantError:        fmt.Errorf("could not convert base asset (FAIL) from trading pair via the passed in assetDisplayFn: cannot recognize the asset FAIL"),
			wantSubmitFilter: nil,
		},
		{
			name:             "failure - display quote",
			configValue:      "",
			exchangeName:     "",
			tradingPair:      &model.TradingPair{Base: "XLM", Quote: "FAIL"},
			assetDisplayFn:   testAssetDisplayFn,
			baseAsset:        testNativeAsset,
			quoteAsset:       hProtocol.Asset{},
			config:           &VolumeFilterConfig{},
			wantError:        fmt.Errorf("could not convert quote asset (FAIL) from trading pair via the passed in assetDisplayFn: cannot recognize the asset FAIL"),
			wantSubmitFilter: nil,
		},
		{
			name:             "failure - query",
			configValue:      "",
			exchangeName:     "",
			tradingPair:      &model.TradingPair{Base: "XLM", Quote: "XLM"},
			assetDisplayFn:   testAssetDisplayFn,
			baseAsset:        testNativeAsset,
			quoteAsset:       testNativeAsset,
			config:           &VolumeFilterConfig{},
			db:               nil,
			wantError:        fmt.Errorf("could not make daily volume by date Query: could not make daily volume by date action: the provided db should be non-nil"),
			wantSubmitFilter: nil,
		},
		{
			name:           "success - empty config",
			configValue:    "",
			exchangeName:   "",
			tradingPair:    &model.TradingPair{Base: "XLM", Quote: "XLM"},
			assetDisplayFn: testAssetDisplayFn,
			baseAsset:      testNativeAsset,
			quoteAsset:     testNativeAsset,
			config:         testEmptyVolumeConfig,
			db:             db,
			wantError:      nil,
			wantSubmitFilter: &volumeFilter{
				name:                   "volumeFilter",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 testEmptyVolumeConfig,
				dailyVolumeByDateQuery: emptyConfigVolumeQuery,
			},
		},
		{
			name:           "success - market ids",
			configValue:    "",
			exchangeName:   "",
			tradingPair:    &model.TradingPair{Base: "XLM", Quote: "XLM"},
			assetDisplayFn: testAssetDisplayFn,
			baseAsset:      testNativeAsset,
			quoteAsset:     testNativeAsset,
			config:         &VolumeFilterConfig{additionalMarketIDs: marketIDs},
			db:             db,
			wantError:      nil,
			wantSubmitFilter: &volumeFilter{
				name:                   "volumeFilter",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{additionalMarketIDs: marketIDs},
				dailyVolumeByDateQuery: marketConfigVolumeQuery,
			},
		},
		{
			name:           "success - account ids",
			configValue:    "",
			exchangeName:   "",
			tradingPair:    &model.TradingPair{Base: "XLM", Quote: "XLM"},
			assetDisplayFn: testAssetDisplayFn,
			baseAsset:      testNativeAsset,
			quoteAsset:     testNativeAsset,
			config:         &VolumeFilterConfig{optionalAccountIDs: accountIDs},
			db:             db,
			wantError:      nil,
			wantSubmitFilter: &volumeFilter{
				name:                   "volumeFilter",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{optionalAccountIDs: accountIDs},
				dailyVolumeByDateQuery: accountConfigVolumeQuery,
			},
		},
		{
			name:           "success - account and market ids",
			configValue:    "",
			exchangeName:   "",
			tradingPair:    &model.TradingPair{Base: "XLM", Quote: "XLM"},
			assetDisplayFn: testAssetDisplayFn,
			baseAsset:      testNativeAsset,
			quoteAsset:     testNativeAsset,
			config:         &VolumeFilterConfig{optionalAccountIDs: accountIDs, additionalMarketIDs: marketIDs},
			db:             db,
			wantError:      nil,
			wantSubmitFilter: &volumeFilter{
				name:                   "volumeFilter",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{optionalAccountIDs: accountIDs, additionalMarketIDs: marketIDs},
				dailyVolumeByDateQuery: accountMarketConfigVolumeQuery,
			},
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			actual, e := makeFilterVolume(
				k.configValue,
				k.exchangeName,
				k.tradingPair,
				k.assetDisplayFn,
				k.baseAsset,
				k.quoteAsset,
				k.db,
				k.config,
			)

			assert.Equal(t, k.wantError, e)
			assert.Equal(t, k.wantSubmitFilter, actual)
		})
	}
}

func TestVolumeFilterFn(t *testing.T) {
	db := connectTestDb()
	dailyVolumeByDateQuery, e := makeDailyVolumeByDateQuery(db, "", "native", "native", []string{}, []string{})
	if e != nil {
		t.Log("could not make empty config volume query")
		return
	}

	emptyFilter := &volumeFilter{
		name:                   "volumeFilter",
		configValue:            "",
		baseAsset:              testNativeAsset,
		quoteAsset:             testNativeAsset,
		config:                 &VolumeFilterConfig{},
		dailyVolumeByDateQuery: dailyVolumeByDateQuery,
	}

	testCases := []struct {
		name      string
		filter    *volumeFilter
		dailyOTB  *VolumeFilterConfig
		dailyTBB  *VolumeFilterConfig
		op        *txnbuild.ManageSellOffer
		wantOp    *txnbuild.ManageSellOffer
		wantError error
		wantTBB   *VolumeFilterConfig
	}{
		{
			name:      "failure - is selling",
			filter:    emptyFilter,
			dailyOTB:  nil,
			dailyTBB:  nil,
			op:        &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.CreditAsset{}},
			wantOp:    nil,
			wantError: fmt.Errorf("error when running the isSelling check for offer '{Selling:{Code: Issuer:} Buying:{} Amount: Price: OfferID:0 SourceAccount:<nil>}': invalid assets, there are more than 2 distinct assets: sdexBase={native  }, sdexQuote={native  }, selling={ }, buying={}"),
			wantTBB:   nil,
		},
		{
			name:      "failure - invalid sell price in op",
			filter:    emptyFilter,
			dailyOTB:  nil,
			dailyTBB:  nil,
			op:        &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.NativeAsset{}},
			wantOp:    nil,
			wantError: fmt.Errorf("could not convert price () to float: strconv.ParseFloat: parsing \"\": invalid syntax"),
			wantTBB:   nil,
		},
		{
			name:      "failure - invalid amount in op",
			filter:    emptyFilter,
			dailyOTB:  nil,
			dailyTBB:  nil,
			op:        &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.NativeAsset{}, Price: "0.0"},
			wantOp:    nil,
			wantError: fmt.Errorf("could not convert amount () to float: strconv.ParseFloat: parsing \"\": invalid syntax"),
			wantTBB:   nil,
		},
		{
			name:      "failure - invalid amount in op",
			filter:    emptyFilter,
			dailyOTB:  nil,
			dailyTBB:  nil,
			op:        &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.NativeAsset{}, Price: "0.0"},
			wantOp:    nil,
			wantError: fmt.Errorf("could not convert amount () to float: strconv.ParseFloat: parsing \"\": invalid syntax"),
			wantTBB:   nil,
		},
		{
			name:      "success - selling, no filter sell caps",
			filter:    emptyFilter,
			dailyOTB:  &VolumeFilterConfig{},
			dailyTBB:  &VolumeFilterConfig{SellBaseAssetCapInBaseUnits: createFloatPtr(0.0), SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0)},
			op:        &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.NativeAsset{}, Price: "1.0", Amount: "100.0"},
			wantOp:    &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.NativeAsset{}, Price: "1.0", Amount: "100.0"},
			wantError: nil,
			wantTBB:   &VolumeFilterConfig{SellBaseAssetCapInBaseUnits: createFloatPtr(100.0), SellBaseAssetCapInQuoteUnits: createFloatPtr(100.0)},
		},
		{
			name: "success - selling, base units sell cap, don't keep selling base",
			filter: &volumeFilter{
				name:                   "volumeFilter",
				configValue:            "",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{SellBaseAssetCapInBaseUnits: createFloatPtr(0.0)},
				dailyVolumeByDateQuery: dailyVolumeByDateQuery,
			},
			dailyOTB: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  createFloatPtr(0.0),
				SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0),
			},
			dailyTBB: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  createFloatPtr(0.0),
				SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0),
			},
			op:        &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.NativeAsset{}, Price: "1.0", Amount: "100.0"},
			wantOp:    nil,
			wantError: nil,
			wantTBB:   &VolumeFilterConfig{SellBaseAssetCapInBaseUnits: createFloatPtr(0.0), SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0)},
		},
		{
			name: "success - selling, base units sell cap, keep selling base",
			filter: &volumeFilter{
				name:                   "volumeFilter",
				configValue:            "",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{SellBaseAssetCapInBaseUnits: createFloatPtr(1.0), mode: volumeFilterModeExact},
				dailyVolumeByDateQuery: dailyVolumeByDateQuery,
			},
			dailyOTB: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  createFloatPtr(0.0),
				SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0),
			},
			dailyTBB: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  createFloatPtr(0.0),
				SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0),
			},
			op:        &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.NativeAsset{}, Price: "1.0", Amount: "100.0"},
			wantOp:    &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.NativeAsset{}, Price: "1.0", Amount: "1.0000000"},
			wantError: nil,
			wantTBB:   &VolumeFilterConfig{SellBaseAssetCapInBaseUnits: createFloatPtr(1.0), SellBaseAssetCapInQuoteUnits: createFloatPtr(1.0)},
		},
		{
			name: "success - selling, quote units sell cap, don't keep selling quote",
			filter: &volumeFilter{
				name:                   "volumeFilter",
				configValue:            "",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0)},
				dailyVolumeByDateQuery: dailyVolumeByDateQuery,
			},
			dailyOTB: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  createFloatPtr(0.0),
				SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0),
			},
			dailyTBB: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  createFloatPtr(0.0),
				SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0),
			},
			op:        &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.NativeAsset{}, Price: "1.0", Amount: "100.0"},
			wantOp:    nil,
			wantError: nil,
			wantTBB:   &VolumeFilterConfig{SellBaseAssetCapInBaseUnits: createFloatPtr(0.0), SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0)},
		},
		{
			name: "success - selling, quote units sell cap, keep selling quote",
			filter: &volumeFilter{
				name:                   "volumeFilter",
				configValue:            "",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{SellBaseAssetCapInQuoteUnits: createFloatPtr(1.0), mode: volumeFilterModeExact},
				dailyVolumeByDateQuery: dailyVolumeByDateQuery,
			},
			dailyOTB: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  createFloatPtr(0.0),
				SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0),
			},
			dailyTBB: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  createFloatPtr(0.0),
				SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0),
			},
			op:        &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.NativeAsset{}, Price: "1.0", Amount: "100.0"},
			wantOp:    &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.NativeAsset{}, Price: "1.0", Amount: "1.0000000"},
			wantError: nil,
			wantTBB:   &VolumeFilterConfig{SellBaseAssetCapInBaseUnits: createFloatPtr(1.0), SellBaseAssetCapInQuoteUnits: createFloatPtr(1.0)},
		},
		{
			name: "success - selling, base and quote units sell cap, keep selling base and quote",
			filter: &volumeFilter{
				name:                   "volumeFilter",
				configValue:            "",
				baseAsset:              testNativeAsset,
				quoteAsset:             testNativeAsset,
				config:                 &VolumeFilterConfig{SellBaseAssetCapInBaseUnits: createFloatPtr(1.0), SellBaseAssetCapInQuoteUnits: createFloatPtr(1.0), mode: volumeFilterModeExact},
				dailyVolumeByDateQuery: dailyVolumeByDateQuery,
			},
			dailyOTB: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  createFloatPtr(0.0),
				SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0),
			},
			dailyTBB: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  createFloatPtr(0.0),
				SellBaseAssetCapInQuoteUnits: createFloatPtr(0.0),
			},
			op:        &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.NativeAsset{}, Price: "1.0", Amount: "100.0"},
			wantOp:    &txnbuild.ManageSellOffer{Buying: txnbuild.NativeAsset{}, Selling: txnbuild.NativeAsset{}, Price: "1.0", Amount: "1.0000000"},
			wantError: nil,
			wantTBB:   &VolumeFilterConfig{SellBaseAssetCapInBaseUnits: createFloatPtr(1.0), SellBaseAssetCapInQuoteUnits: createFloatPtr(1.0)},
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			tbb := k.dailyTBB
			actual, e := k.filter.volumeFilterFn(k.dailyOTB, tbb, k.op)
			assert.Equal(t, k.wantError, e)
			assert.Equal(t, k.wantOp, actual)
			assert.Equal(t, k.wantTBB, tbb)
		})
	}
}

func createFloatPtr(x float64) *float64 {
	return &x
}
