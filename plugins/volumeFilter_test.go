package plugins

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

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

func TestMakeFilterVolume(t *testing.T) {
	testAssetDisplayFn := model.AssetDisplayFn(func(asset model.Asset) (string, error) {
		sdexAssetMap := map[model.Asset]hProtocol.Asset{
			model.Asset("XLM"): hProtocol.Asset{Type: "native"},
		}
		assetString, ok := sdexAssetMap[asset]
		if !ok {
			return "", fmt.Errorf("cannot recognize the asset %s", string(asset))
		}
		return utils.Asset2String(assetString), nil
	})

	testNativeAsset := hProtocol.Asset{Type: "native"}
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
			wantError:        fmt.Errorf("could not make daily volume by date query: could not make daily volume by date action: the provided db should be non-nil"),
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
