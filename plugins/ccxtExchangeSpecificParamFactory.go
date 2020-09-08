package plugins

import (
	"fmt"
	"strconv"

	"github.com/stellar/kelp/api"
)

type ccxtExchangeSpecificParamFactoryCoinbasepro struct{}

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

type ccxtExchangeSpecificParamFactoryBinance struct{}

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
