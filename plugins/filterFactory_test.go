package plugins

import (
	"fmt"
	"testing"

	"github.com/openlyinc/pointy"
	"github.com/stellar/kelp/queries"
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
		// the first %s represents the action (buy or sell), the second %s represents mode (exact or ignore)
		// we loop over the actions and modes below and inject them into the input and wantConfig
		{
			configInput: "volume/daily/%s/base/3500.0/%s",
			wantConfig: &VolumeFilterConfig{
				BaseAssetCapInBaseUnits:  pointy.Float64(3500.0),
				BaseAssetCapInQuoteUnits: nil,
				additionalMarketIDs:      nil,
				optionalAccountIDs:       nil,
			},
		}, {
			configInput: "volume/daily/%s/quote/4000.0/%s",
			wantConfig: &VolumeFilterConfig{
				BaseAssetCapInBaseUnits:  nil,
				BaseAssetCapInQuoteUnits: pointy.Float64(4000.0),
				additionalMarketIDs:      nil,
				optionalAccountIDs:       nil,
			},
		},
		{
			configInput: "volume/daily/%s/base/3500.0/%s",
			wantConfig: &VolumeFilterConfig{
				BaseAssetCapInBaseUnits:  pointy.Float64(3500.0),
				BaseAssetCapInQuoteUnits: nil,
				additionalMarketIDs:      nil,
				optionalAccountIDs:       nil,
			},
		}, {
			configInput: "volume/daily/%s/quote/1000.0/%s",
			wantConfig: &VolumeFilterConfig{
				BaseAssetCapInBaseUnits:  nil,
				BaseAssetCapInQuoteUnits: pointy.Float64(1000.0),
				additionalMarketIDs:      nil,
				optionalAccountIDs:       nil,
			},
		}, {
			configInput: "volume/daily:market_ids=[4c19915f47,db4531d586]/%s/base/3500.0/%s",
			wantConfig: &VolumeFilterConfig{
				BaseAssetCapInBaseUnits:  pointy.Float64(3500.0),
				BaseAssetCapInQuoteUnits: nil,
				additionalMarketIDs:      []string{"4c19915f47", "db4531d586"},
				optionalAccountIDs:       nil,
			},
		}, {
			configInput: "volume/daily:account_ids=[account1,account2]/%s/base/3500.0/%s",
			wantConfig: &VolumeFilterConfig{
				BaseAssetCapInBaseUnits:  pointy.Float64(3500.0),
				BaseAssetCapInQuoteUnits: nil,
				additionalMarketIDs:      nil,
				optionalAccountIDs:       []string{"account1", "account2"},
			},
		}, {
			configInput: "volume/daily:market_ids=[4c19915f47,db4531d586]:account_ids=[account1,account2]/%s/base/3500.0/%s",
			wantConfig: &VolumeFilterConfig{
				BaseAssetCapInBaseUnits:  pointy.Float64(3500.0),
				BaseAssetCapInQuoteUnits: nil,
				additionalMarketIDs:      []string{"4c19915f47", "db4531d586"},
				optionalAccountIDs:       []string{"account1", "account2"},
			},
		},
	}

	modes := []volumeFilterMode{volumeFilterModeExact, volumeFilterModeIgnore}
	actions := []queries.DailyVolumeAction{queries.DailyVolumeActionBuy, queries.DailyVolumeActionSell}
	for _, k := range testCases {
		// loop over both modes, and inject the desired mode in the config
		for _, m := range modes {
			wantConfig := k.wantConfig
			wantConfig.mode = m

			// loop over both actions, and inject the desired action in the config
			for _, a := range actions {
				wantConfig.action = a
				configInput := fmt.Sprintf(k.configInput, a, m)

				t.Run(configInput, func(t *testing.T) {
					actual, e := makeVolumeFilterConfig(configInput)
					if !assert.NoError(t, e) {
						return
					}
					assertVolumeFilterConfigEqual(t, wantConfig, actual)
				})
			}
		}
	}
}

func assertVolumeFilterConfigEqual(t *testing.T, want *VolumeFilterConfig, actual *VolumeFilterConfig) {
	if want == nil {
		assert.Nil(t, actual)
	} else if actual == nil {
		assert.Fail(t, fmt.Sprintf("actual was nil but expected %v", *want))
	} else {
		assert.Equal(t, want.BaseAssetCapInBaseUnits, actual.BaseAssetCapInBaseUnits)
		assert.Equal(t, want.BaseAssetCapInQuoteUnits, actual.BaseAssetCapInQuoteUnits)
		assert.Equal(t, want.action, actual.action)
		assert.Equal(t, want.mode, actual.mode)
		assert.Equal(t, want.additionalMarketIDs, actual.additionalMarketIDs)
		assert.Equal(t, want.optionalAccountIDs, actual.optionalAccountIDs)
	}
}
