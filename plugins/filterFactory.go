package plugins

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/model"
)

var filterMap = map[string]func(f *FilterFactory, configInput string) (SubmitFilter, error){
	"volume": filterVolume,
	"price":  filterPrice,
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
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid input (%s), needs 4 parts separated by the delimiter (/)", configInput)
	}

	config := &VolumeFilterConfig{}
	if parts[1] != "sell" {
		return nil, fmt.Errorf("invalid input (%s), the second part needs to be \"sell\"", configInput)
	}
	limit, e := strconv.ParseFloat(parts[3], 64)
	if e != nil {
		return nil, fmt.Errorf("could not parse the fourth part as a float value from config value (%s): %s", configInput, e)
	}
	if parts[2] == "base" {
		config.SellBaseAssetCapInBaseUnits = &limit
	} else if parts[2] == "quote" {
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
