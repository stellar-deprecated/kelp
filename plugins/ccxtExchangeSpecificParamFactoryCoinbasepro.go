package plugins

import (
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

var _ ccxtExchangeSpecificParamFactory = &ccxtExchangeSpecificParamFactoryCoinbasepro{}
