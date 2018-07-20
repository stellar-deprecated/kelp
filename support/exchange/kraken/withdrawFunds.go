package kraken

import (
	"fmt"
	"reflect"

	"github.com/lightyeario/kelp/model/assets"
	"github.com/lightyeario/kelp/support/exchange/api/number"
	"github.com/lightyeario/kelp/support/treasury/api"
)

// WithdrawFunds impl.
func (k krakenExchange) WithdrawFunds(
	asset model.Asset,
	amountToWithdraw *number.Number,
	address string,
) (*treasury.WithdrawFunds, error) {
	krakenAsset, e := k.assetConverter.ToString(asset)
	if e != nil {
		return nil, e
	}

	withdrawKey, e := k.withdrawKeys.getKey(asset, address)
	if e != nil {
		return nil, e
	}
	resp, e := k.api.Query(
		"Withdraw",
		map[string]string{
			"asset":  krakenAsset,
			"key":    withdrawKey,
			"amount": amountToWithdraw.AsString(),
		},
	)
	if e != nil {
		return nil, e
	}

	return parseWithdrawResponse(resp)
}

func parseWithdrawResponse(resp interface{}) (*treasury.WithdrawFunds, error) {
	switch m := resp.(type) {
	case map[string]interface{}:
		refid, e := parseString(m, "refid", "Withdraw")
		if e != nil {
			return nil, e
		}
		return &treasury.WithdrawFunds{
			WithdrawalID: refid,
		}, nil
	default:
		return nil, fmt.Errorf("could not parse response type from Withdraw: %s", reflect.TypeOf(m))
	}
}
