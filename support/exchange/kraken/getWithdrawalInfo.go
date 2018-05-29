package kraken

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/lightyeario/kelp/support/exchange/api/assets"
	"github.com/lightyeario/kelp/support/exchange/api/number"
	"github.com/lightyeario/kelp/support/treasury/api"
)

// GetWithdrawInfo impl.
func (k krakenExchange) GetWithdrawInfo(
	asset assets.Asset,
	amountToWithdraw *number.Number,
	address string,
) (*treasury.WithdrawInfo, error) {
	krakenAsset, e := k.assetConverter.ToString(asset)
	if e != nil {
		return nil, e
	}

	withdrawKey, e := k.withdrawKeys.getKey(asset, address)
	if e != nil {
		return nil, e
	}
	resp, e := k.api.Query(
		"WithdrawInfo",
		map[string]string{
			"asset":  krakenAsset,
			"key":    withdrawKey,
			"amount": amountToWithdraw.AsString(),
		},
	)
	if e != nil {
		return nil, e
	}

	return parseWithdrawInfoResponse(resp, amountToWithdraw)
}

func parseWithdrawInfoResponse(resp interface{}, amountToWithdraw *number.Number) (*treasury.WithdrawInfo, error) {
	switch m := resp.(type) {
	case map[string]interface{}:
		info, e := parseWithdrawInfo(m)
		if e != nil {
			return nil, e
		}
		if info.limit != nil && info.limit.AsFloat() < amountToWithdraw.AsFloat() {
			return nil, treasury.MakeErrWithdrawAmountAboveLimit(amountToWithdraw, info.limit)
		}
		if info.fee != nil && info.fee.AsFloat() >= amountToWithdraw.AsFloat() {
			return nil, treasury.MakeErrWithdrawAmountInvalid(amountToWithdraw, info.fee)
		}

		return &treasury.WithdrawInfo{AmountToReceive: info.amount}, nil
	default:
		return nil, fmt.Errorf("could not parse response type from WithdrawInfo: %s", reflect.TypeOf(m))
	}
}

type withdrawInfo struct {
	limit  *number.Number
	fee    *number.Number
	amount *number.Number
}

func parseWithdrawInfo(m map[string]interface{}) (*withdrawInfo, error) {
	// limit
	limit, e := parseNumber(m, "limit", "WithdrawInfo")
	if e != nil {
		return nil, e
	}

	// fee
	fee, e := parseNumber(m, "fee", "WithdrawInfo")
	if e != nil {
		if !strings.HasPrefix(e.Error(), prefixFieldNotFound) {
			return nil, e
		}
		// fee may be missing in which case it's null
		fee = nil
	}

	// amount
	amount, e := parseNumber(m, "amount", "WithdrawInfo")
	if e != nil {
		return nil, e
	}

	return &withdrawInfo{
		limit:  limit,
		fee:    fee,
		amount: amount,
	}, nil
}
