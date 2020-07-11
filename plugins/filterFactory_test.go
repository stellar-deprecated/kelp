package plugins

import (
	"fmt"
	"testing"

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
			assert.Equal(t, k.wantConfig.SellBaseAssetCapInBaseUnits, config.SellBaseAssetCapInBaseUnits)
			assert.Equal(t, k.wantConfig.SellBaseAssetCapInQuoteUnits, config.SellBaseAssetCapInQuoteUnits)
			assert.Equal(t, k.wantConfig.mode, config.mode)
			assert.Equal(t, k.wantConfig.additionalMarketIDs, config.additionalMarketIDs)
			assert.Equal(t, k.wantConfig.optionalAccountIDs, config.optionalAccountIDs)
		})
	}

}
