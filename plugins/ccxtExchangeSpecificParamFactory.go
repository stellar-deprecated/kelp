package plugins

import (
	"fmt"
	"strconv"

	"github.com/stellar/kelp/api"
)

type ccxtExchangeSpecificParamFactoryCoinbasepro struct{}

func (f *ccxtExchangeSpecificParamFactoryCoinbasepro) getParamsForGetOrderBook() map[string]interface{} {
	return nil
}

func (f *ccxtExchangeSpecificParamFactoryCoinbasepro) getParamsForAddOrder(submitMode api.SubmitMode) interface{} {
	if submitMode == api.SubmitModeMakerOnly {
		return map[string]interface{}{
			"post_only": true,
		}
	}
	return nil
}

func (f *ccxtExchangeSpecificParamFactoryCoinbasepro) getParamsForGetTradeHistory() interface{} {
	return nil
}

var _ ccxtExchangeSpecificParamFactory = &ccxtExchangeSpecificParamFactoryCoinbasepro{}

type ccxtExchangeSpecificParamFactoryBinance struct {
	validOrderBookLevels []int
	lastValidLimit       int
	cachedResults        map[int]int
}

func makeCcxtExchangeSpecificParamFactoryBinance() *ccxtExchangeSpecificParamFactoryBinance {
	validOrderBookLevels := []int{5, 10, 20, 50, 100, 500, 1000, 5000}
	return &ccxtExchangeSpecificParamFactoryBinance{
		validOrderBookLevels: validOrderBookLevels,
		lastValidLimit:       validOrderBookLevels[len(validOrderBookLevels)-1],
		cachedResults:        map[int]int{},
	}
}

func (f *ccxtExchangeSpecificParamFactoryBinance) getParamsForGetOrderBook() map[string]interface{} {
	return map[string]interface{}{
		"transform_limit": f.transformLimit,
	}
}

func (f *ccxtExchangeSpecificParamFactoryBinance) transformLimit(limit int) (int /*newLimit*/, error) {
	if newLimit, ok := f.cachedResults[limit]; ok {
		if newLimit == -1 {
			return -1, fmt.Errorf("limit requested (%d) is higher than the maximum limit allowed (%d)", limit, f.lastValidLimit)
		}
		return newLimit, nil
	}

	for _, validLimit := range f.validOrderBookLevels {
		if limit <= validLimit {
			f.cachedResults[limit] = validLimit
			return validLimit, nil
		}
	}

	f.cachedResults[limit] = -1
	return -1, fmt.Errorf("limit requested (%d) is higher than the maximum limit allowed (%d)", limit, f.lastValidLimit)
}

func (f *ccxtExchangeSpecificParamFactoryBinance) getParamsForAddOrder(submitMode api.SubmitMode) interface{} {
	return nil
}

func (f *ccxtExchangeSpecificParamFactoryBinance) getParamsForGetTradeHistory() interface{} {
	return map[string]interface{}{
		"order_id": func(info interface{}) (string, error) {
			rawInfo, ok := info.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("unable to convert input 'info' to a map[string]interface{}: %+v (type=%T)", rawInfo, rawInfo)
			}

			orderIDFloat64, ok := rawInfo["orderId"].(float64)
			if !ok {
				return "", fmt.Errorf("unable to parse info[\"orderId\"] as a float64: %+v (type=%T)", rawInfo["orderId"], rawInfo["orderId"])
			}

			orderIDInt64 := int64(orderIDFloat64)
			orderID := strconv.FormatInt(orderIDInt64, 10)
			return orderID, nil
		},
	}
}

var _ ccxtExchangeSpecificParamFactory = &ccxtExchangeSpecificParamFactoryBinance{}
