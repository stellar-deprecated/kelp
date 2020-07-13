package plugins

import (
	"fmt"
	"testing"

	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"
)

func TestParseIdsArray(t *testing.T) {
	testCases := []struct {
		arrayString string
		want        []string
	}{
		{
			arrayString: "[abcde1234Z,01234gFHij]",
			want:        []string{"abcde1234Z", "01234gFHij"},
		}, {
			arrayString: "[abcde1234Z, 01234gFHij]",
			want:        []string{"abcde1234Z", "01234gFHij"},
		}, {
			arrayString: "[abcde1234Z]",
			want:        []string{"abcde1234Z"},
		}, {
			arrayString: "[account1]",
			want:        []string{"account1"},
		}, {
			arrayString: "[]",
			want:        []string{},
		},
	}

	for _, kase := range testCases {
		t.Run(kase.arrayString, func(t *testing.T) {
			output, e := parseIdsArray(kase.arrayString)
			if !assert.NoError(t, e) {
				return
			}

			assert.Equal(t, kase.want, output)
		})
	}
}

func TestParseVolumeFilterModifier(t *testing.T) {
	testCases := []struct {
		modifierMapping  string
		wantIds          []string
		wantModifierType string
		wantError        error
	}{
		{
			modifierMapping:  "market_ids=[abcde1234Z,01234gFHij]",
			wantIds:          []string{"abcde1234Z", "01234gFHij"},
			wantModifierType: "market_ids",
			wantError:        nil,
		}, {
			modifierMapping:  "account_ids=[abcde1234Z,01234gFHij]",
			wantIds:          []string{"abcde1234Z", "01234gFHij"},
			wantModifierType: "account_ids",
			wantError:        nil,
		}, {
			modifierMapping:  "market_ids=[abcde,01234]",
			wantIds:          nil,
			wantModifierType: "market_ids",
			wantError:        fmt.Errorf("invalid id entry 'abcde'"),
		}, {
			modifierMapping:  "market_ids=[abcde1234Z,01234]",
			wantIds:          nil,
			wantModifierType: "market_ids",
			wantError:        fmt.Errorf("invalid id entry '01234'"),
		}, {
			modifierMapping:  "account_ids=[abcde1234Z,01234]",
			wantIds:          []string{"abcde1234Z", "01234"},
			wantModifierType: "account_ids",
			wantError:        nil,
		}, {
			modifierMapping:  "account_ids=[abcde,01234]",
			wantIds:          []string{"abcde", "01234"},
			wantModifierType: "account_ids",
			wantError:        nil,
		}, {
			modifierMapping:  "market_ids=[]",
			wantIds:          nil,
			wantModifierType: "market_ids",
			wantError:        fmt.Errorf("array length required to be greater than 0"),
		}, {
			modifierMapping:  "account_ids=[]",
			wantIds:          []string{},
			wantModifierType: "account_ids",
			wantError:        nil,
		},
	}

	for _, k := range testCases {
		t.Run(k.modifierMapping, func(t *testing.T) {
			ids, modifierType, e := parseVolumeFilterModifier(k.modifierMapping)
			assert.Equal(t, k.wantError, e)
			assert.Equal(t, k.wantIds, ids)
			assert.Equal(t, k.wantModifierType, modifierType)
		})
	}
}

func TestAddModifierToConfig(t *testing.T) {
	testCases := []struct {
		modifierMapping string
		wantConfig      *VolumeFilterConfig
	}{
		{
			modifierMapping: "market_ids=[abcde1234Z]",
			wantConfig:      &VolumeFilterConfig{additionalMarketIDs: []string{"abcde1234Z"}},
		}, {
			modifierMapping: "account_ids=[accountX]",
			wantConfig:      &VolumeFilterConfig{optionalAccountIDs: []string{"accountX"}},
		},
	}

	for _, k := range testCases {
		t.Run(k.modifierMapping, func(t *testing.T) {
			config := &VolumeFilterConfig{}
			e := addModifierToConfig(config, k.modifierMapping)
			if !assert.NoError(t, e) {
				return
			}
			assertVolumeFilterConfigEqual(t, k.wantConfig, config)
		})
	}
}

func TestMakeVolumeFilterConfig(t *testing.T) {
	testCases := []struct {
		configInput string
		wantError   error
		wantConfig  *VolumeFilterConfig
	}{
		{
			configInput: "volume/daily/sell/base/3500.0/exact",
			wantConfig: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  pointy.Float64(3500.0),
				SellBaseAssetCapInQuoteUnits: nil,
				mode:                         volumeFilterModeExact,
				additionalMarketIDs:          nil,
				optionalAccountIDs:           nil,
			},
		}, {
			configInput: "volume/daily/sell/quote/1000.0/ignore",
			wantConfig: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  nil,
				SellBaseAssetCapInQuoteUnits: pointy.Float64(1000.0),
				mode:                         volumeFilterModeIgnore,
				additionalMarketIDs:          nil,
				optionalAccountIDs:           nil,
			},
		}, {
			configInput: "volume/daily:market_ids=[4c19915f47,db4531d586]/sell/base/3500.0/exact",
			wantConfig: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  pointy.Float64(3500.0),
				SellBaseAssetCapInQuoteUnits: nil,
				mode:                         volumeFilterModeExact,
				additionalMarketIDs:          []string{"4c19915f47", "db4531d586"},
				optionalAccountIDs:           nil,
			},
		}, {
			configInput: "volume/daily:account_ids=[account1,account2]/sell/base/3500.0/exact",
			wantConfig: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  pointy.Float64(3500.0),
				SellBaseAssetCapInQuoteUnits: nil,
				mode:                         volumeFilterModeExact,
				additionalMarketIDs:          nil,
				optionalAccountIDs:           []string{"account1", "account2"},
			},
		}, {
			configInput: "volume/daily:market_ids=[4c19915f47,db4531d586]:account_ids=[account1,account2]/sell/base/3500.0/exact",
			wantConfig: &VolumeFilterConfig{
				SellBaseAssetCapInBaseUnits:  pointy.Float64(3500.0),
				SellBaseAssetCapInQuoteUnits: nil,
				mode:                         volumeFilterModeExact,
				additionalMarketIDs:          []string{"4c19915f47", "db4531d586"},
				optionalAccountIDs:           []string{"account1", "account2"},
			},
		},
	}

	for _, k := range testCases {
		t.Run(k.configInput, func(t *testing.T) {
			actual, e := makeVolumeFilterConfig(k.configInput)
			if !assert.NoError(t, e) {
				return
			}
			assertVolumeFilterConfigEqual(t, k.wantConfig, actual)
		})
	}
}

func assertVolumeFilterConfigEqual(t *testing.T, want *VolumeFilterConfig, actual *VolumeFilterConfig) {
	if want == nil {
		assert.Nil(t, actual)
	} else if actual == nil {
		assert.Fail(t, fmt.Sprintf("actual was nil but expected %v", *want))
	} else {
		assert.Equal(t, want.SellBaseAssetCapInBaseUnits, actual.SellBaseAssetCapInBaseUnits)
		assert.Equal(t, want.SellBaseAssetCapInQuoteUnits, actual.SellBaseAssetCapInQuoteUnits)
		assert.Equal(t, want.mode, actual.mode)
		assert.Equal(t, want.additionalMarketIDs, actual.additionalMarketIDs)
		assert.Equal(t, want.optionalAccountIDs, actual.optionalAccountIDs)
	}
}
