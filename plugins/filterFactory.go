package plugins

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/queries"
)

var filterIDRegex *regexp.Regexp

func init() {
	rxp, e := regexp.Compile("^[a-zA-Z0-9]{10}$")
	if e != nil {
		panic("unable to compile filterID regexp")
	}
	filterIDRegex = rxp
}

var filterMap = map[string]func(f *FilterFactory, configInput string) (SubmitFilter, error){
	"volume":    filterVolume,
	"price":     filterPrice,
	"priceFeed": filterPriceFeed,
}

// FilterFactory is a struct that handles creating all the filters
type FilterFactory struct {
	ExchangeName   string
	TradingPair    *model.TradingPair
	AssetDisplayFn model.AssetDisplayFn
	BaseAsset      hProtocol.Asset
	QuoteAsset     hProtocol.Asset
	DB             *sql.DB
}

// MakeFilter is the function that makes the required filters
func (f *FilterFactory) MakeFilter(configInput string) (SubmitFilter, error) {
	parts := strings.Split(configInput, "/")
	if len(parts) <= 0 {
		return nil, fmt.Errorf("invalid input (%s), needs at least 1 delimiter (/)", configInput)
	}

	filterName := parts[0]
	factoryMethod, ok := filterMap[filterName]
	if !ok {
		return nil, fmt.Errorf("could not find filter of type '%s'", filterName)
	}

	return factoryMethod(f, configInput)
}

func filterVolume(f *FilterFactory, configInput string) (SubmitFilter, error) {
	config, e := makeVolumeFilterConfig(configInput)
	if e != nil {
		return nil, fmt.Errorf("could not make VolumeFilterConfig for configInput (%s): %s", configInput, e)
	}

	return makeFilterVolume(
		configInput,
		f.ExchangeName,
		f.TradingPair,
		f.AssetDisplayFn,
		f.BaseAsset,
		f.QuoteAsset,
		f.DB,
		config,
	)
}

func makeRawVolumeFilterConfig(
	baseAssetCapInBaseUnits *float64,
	baseAssetCapInQuoteUnits *float64,
	action queries.DailyVolumeAction,
	mode volumeFilterMode,
	additionalMarketIDs []string,
	optionalAccountIDs []string,
) *VolumeFilterConfig {
	return &VolumeFilterConfig{
		BaseAssetCapInBaseUnits:  baseAssetCapInBaseUnits,
		BaseAssetCapInQuoteUnits: baseAssetCapInQuoteUnits,
		action:                   action,
		mode:                     mode,
		additionalMarketIDs:      additionalMarketIDs,
		optionalAccountIDs:       optionalAccountIDs,
	}
}

func makeVolumeFilterConfig(configInput string) (*VolumeFilterConfig, error) {
	parts := strings.Split(configInput, "/")
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid input (%s), needs 6 parts separated by the delimiter (/)", configInput)
	}

	mode, e := parseVolumeFilterMode(parts[5])
	if e != nil {
		return nil, fmt.Errorf("could not parse volume filter mode from input (%s): %s", configInput, e)
	}
	config := &VolumeFilterConfig{mode: mode}

	limitWindowParts := strings.Split(parts[1], ":")
	if limitWindowParts[0] != "daily" {
		return nil, fmt.Errorf("invalid input (%s), the second part needs to equal or start with \"daily\"", configInput)
	}

	action, e := queries.ParseDailyVolumeAction(parts[2])
	if e != nil {
		return nil, fmt.Errorf("could not parse volume filter action from input (%s): %s", configInput, e)
	}
	config.action = action

	errInvalid := fmt.Errorf("invalid input (%s), the modifier for \"daily\" can be either \"market_ids\" or \"account_ids\" like so 'daily:market_ids=[4c19915f47,db4531d586]' or 'daily:account_ids=[account1,account2]' or 'daily:market_ids=[4c19915f47,db4531d586]:account_ids=[account1,account2]'", configInput)
	if len(limitWindowParts) == 2 {
		e = addModifierToConfig(config, limitWindowParts[1])
		if e != nil {
			return nil, fmt.Errorf("%s: could not addModifierToConfig for %s: %s", errInvalid, limitWindowParts[1], e)
		}
	} else if len(limitWindowParts) == 3 {
		e = addModifierToConfig(config, limitWindowParts[1])
		if e != nil {
			return nil, fmt.Errorf("%s: could not addModifierToConfig for %s: %s", errInvalid, limitWindowParts[1], e)
		}

		e = addModifierToConfig(config, limitWindowParts[2])
		if e != nil {
			return nil, fmt.Errorf("%s: could not addModifierToConfig for %s: %s", errInvalid, limitWindowParts[2], e)
		}
	} else if len(limitWindowParts) != 1 {
		return nil, fmt.Errorf("invalid input (%s), the second part needs to be \"daily\" and can have only one modifier \"market_ids\" like so 'daily:market_ids=[4c19915f47,db4531d586]'", configInput)
	}

	limit, e := strconv.ParseFloat(parts[4], 64)
	if e != nil {
		return nil, fmt.Errorf("could not parse the fourth part as a float value from config value (%s): %s", configInput, e)
	}
	if parts[3] == "base" {
		config.BaseAssetCapInBaseUnits = &limit
	} else if parts[3] == "quote" {
		config.BaseAssetCapInQuoteUnits = &limit
	} else {
		return nil, fmt.Errorf("invalid input (%s), the third part needs to be \"base\" or \"quote\"", configInput)
	}

	if e = config.Validate(); e != nil {
		return nil, fmt.Errorf("invalid input (%s), did not pass validation: %s", configInput, e)
	}
	return config, nil
}

func addModifierToConfig(config *VolumeFilterConfig, modifierMapping string) error {
	ids, modifierType, e := parseVolumeFilterModifier(modifierMapping)
	if e != nil {
		return fmt.Errorf("could not parseVolumeFilterModifier: %s", e)
	}

	if modifierType == "market_ids" {
		config.additionalMarketIDs = ids
		return nil
	} else if modifierType == "account_ids" {
		config.optionalAccountIDs = ids
		return nil
	}
	return fmt.Errorf("programmer error? invalid modifier type '%s', should have thrown an error above when calling parseVolumeFilterModifier", modifierType)
}

func parseVolumeFilterModifier(modifierMapping string) ([]string, string, error) {
	modifierParts := strings.Split(modifierMapping, "=")
	if len(modifierParts) != 2 {
		return nil, "", fmt.Errorf("invalid parts for modifier with length %d, should have been 2", len(modifierParts))
	}

	ids, e := parseIdsArray(modifierParts[1])
	if e != nil {
		return nil, "", fmt.Errorf("%s", e)
	}

	if strings.HasPrefix(modifierMapping, "market_ids=") {
		if len(ids) == 0 {
			return nil, "market_ids", fmt.Errorf("array length required to be greater than 0")
		}

		for _, id := range ids {
			if !filterIDRegex.MatchString(id) {
				return nil, "market_ids", fmt.Errorf("invalid id entry '%s'", id)
			}
		}

		return ids, "market_ids", nil
	} else if strings.HasPrefix(modifierMapping, "account_ids=") {
		return ids, "account_ids", nil
	}

	return nil, "", fmt.Errorf("invalid prefix for volume filter modifier '%s'", modifierMapping)
}

func parseIdsArray(arrayString string) ([]string, error) {
	if !strings.HasPrefix(arrayString, "[") {
		return nil, fmt.Errorf("arrayString should begin with '['")
	}

	if !strings.HasSuffix(arrayString, "]") {
		return nil, fmt.Errorf("arrayString should end with ']'")
	}

	arrayStringCleaned := arrayString[:len(arrayString)-1][1:]
	ids := strings.Split(arrayStringCleaned, ",")

	idsTrimmed := []string{}
	for _, id := range ids {
		trimmedID := strings.TrimSpace(id)

		// skip empty items
		if len(trimmedID) == 0 {
			continue
		}

		idsTrimmed = append(idsTrimmed, trimmedID)
	}
	return idsTrimmed, nil
}

func filterPrice(f *FilterFactory, configInput string) (SubmitFilter, error) {
	parts := strings.Split(configInput, "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid input (%s), needs 3 parts separated by the delimiter (/)", configInput)
	}

	limit, e := strconv.ParseFloat(parts[2], 64)
	if e != nil {
		return nil, fmt.Errorf("could not parse the third part as a float value from config value (%s): %s", configInput, e)
	}
	if parts[1] == "min" {
		config := MinPriceFilterConfig{MinPrice: &limit}
		return MakeFilterMinPrice(f.BaseAsset, f.QuoteAsset, &config)
	} else if parts[1] == "max" {
		config := MaxPriceFilterConfig{MaxPrice: &limit}
		return MakeFilterMaxPrice(f.BaseAsset, f.QuoteAsset, &config)
	}
	return nil, fmt.Errorf("invalid price filter type in second argument (%s)", configInput)
}

func filterPriceFeed(f *FilterFactory, configInput string) (SubmitFilter, error) {
	// parts[0] = "priceFeed", parts[1] = comparisonMode, parts[2] = feedDataType, parts[3] = feedURL which can have more "/" chars
	parts := strings.Split(configInput, "/")
	if len(parts) < 4 {
		return nil, fmt.Errorf("\"priceFeed\" filter needs at least 4 parts separated by the '/' delimiter (priceFeed/<comparisonMode>/<feedDataType>/<feedURL>) but we received %s", configInput)
	}

	cmString := parts[1]
	feedType := parts[2]
	feedURL := strings.Join(parts[3:len(parts)], "/")
	pf, e := MakePriceFeed(feedType, feedURL)
	if e != nil {
		return nil, fmt.Errorf("could not make price feed for config input string '%s': %s", configInput, e)
	}

	filter, e := MakeFilterPriceFeed(f.BaseAsset, f.QuoteAsset, cmString, pf)
	if e != nil {
		return nil, fmt.Errorf("could not make price feed filter for config input string '%s': %s", configInput, e)
	}

	return filter, nil
}
