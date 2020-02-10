package plugins

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/model"
)

var marketIDRegex *regexp.Regexp

func init() {
	midRxp, e := regexp.Compile("^[a-zA-Z0-9]{10}$")
	if e != nil {
		panic("unable to compile marketID regexp")
	}
	marketIDRegex = midRxp
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
	} else if len(limitWindowParts) == 2 {
		errInvalid := fmt.Errorf("invalid input (%s), the modifier for \"daily\" can only be \"market_ids\" like so 'daily:market_ids=[4c19915f47,db4531d586]'", configInput)
		if !strings.HasPrefix(limitWindowParts[1], "market_ids=") {
			return nil, fmt.Errorf("%s: invalid modifier prefix in '%s'", errInvalid, limitWindowParts[1])
		}

		modifierParts := strings.Split(limitWindowParts[1], "=")
		if len(modifierParts) != 2 {
			return nil, fmt.Errorf("%s: invalid parts for modifier with length %d, should have been 2", errInvalid, len(modifierParts))
		}

		marketIds, e := parseMarketIdsArray(modifierParts[1])
		if e != nil {
			return nil, fmt.Errorf("%s: %s", errInvalid, e)
		}

		config.additionalMarketIDs = marketIds
	} else if len(limitWindowParts) != 1 {
		return nil, fmt.Errorf("invalid input (%s), the second part needs to be \"daily\" and can have only one modifier \"market_ids\" like so 'daily:market_ids=[4c19915f47,db4531d586]'", configInput)
	}

	if parts[2] != "sell" {
		return nil, fmt.Errorf("invalid input (%s), the third part needs to be \"sell\"", configInput)
	}
	limit, e := strconv.ParseFloat(parts[4], 64)
	if e != nil {
		return nil, fmt.Errorf("could not parse the fourth part as a float value from config value (%s): %s", configInput, e)
	}
	if parts[3] == "base" {
		config.SellBaseAssetCapInBaseUnits = &limit
	} else if parts[3] == "quote" {
		config.SellBaseAssetCapInQuoteUnits = &limit
	} else {
		return nil, fmt.Errorf("invalid input (%s), the third part needs to be \"base\" or \"quote\"", configInput)
	}
	if e := config.Validate(); e != nil {
		return nil, fmt.Errorf("invalid input (%s), did not pass validation: %s", configInput, e)
	}

	return makeFilterVolume(
		f.ExchangeName,
		f.TradingPair,
		f.AssetDisplayFn,
		f.BaseAsset,
		f.QuoteAsset,
		f.DB,
		config,
	)
}

func parseMarketIdsArray(marketIdsArrayString string) ([]string, error) {
	if !strings.HasPrefix(marketIdsArrayString, "[") {
		return nil, fmt.Errorf("market_ids array should begin with '['")
	}

	if !strings.HasSuffix(marketIdsArrayString, "]") {
		return nil, fmt.Errorf("market_ids array should end with ']'")
	}

	arrayStringCleaned := marketIdsArrayString[:len(marketIdsArrayString)-1][1:]
	marketIds := strings.Split(arrayStringCleaned, ",")
	if len(marketIds) == 0 {
		return nil, fmt.Errorf("market_ids array length should be greater than 0")
	}

	marketIdsTrimmed := []string{}
	for _, mid := range marketIds {
		trimmedMid := strings.TrimSpace(mid)
		if !marketIDRegex.MatchString(trimmedMid) {
			return nil, fmt.Errorf("invalid market_id entry '%s'", trimmedMid)
		}
		marketIdsTrimmed = append(marketIdsTrimmed, trimmedMid)
	}
	return marketIdsTrimmed, nil
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
