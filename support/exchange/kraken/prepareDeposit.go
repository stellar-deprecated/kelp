package kraken

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/lightyeario/kelp/model/assets"
	"github.com/lightyeario/kelp/support/exchange/api/number"
	"github.com/lightyeario/kelp/support/treasury/api"
)

// PrepareDeposit impl.
func (k krakenExchange) PrepareDeposit(asset model.Asset, amount *number.Number) (*treasury.PrepareDepositResult, error) {
	krakenAsset, e := k.assetConverter.ToString(asset)
	if e != nil {
		return nil, e
	}

	dm, e := k.getDepositMethods(krakenAsset)
	if e != nil {
		return nil, e
	}

	if dm.limit != nil && dm.limit.AsFloat() < amount.AsFloat() {
		return nil, treasury.MakeErrDepositAmountAboveLimit(amount, dm.limit)
	}

	// get any unused address on the account or generate a new address if no existing unused address
	generateNewAddress := false
	for {
		addressList, e := k.getDepositAddress(krakenAsset, dm.method, generateNewAddress)
		if e != nil {
			if strings.Contains(e.Error(), "EFunding:Too many addresses") {
				return nil, treasury.MakeErrTooManyDepositAddresses()
			}
			return nil, e
		}
		// TODO 2 - filter addresses that may be "in progress" - save suggested address on account before using and filter using that list
		// discard addresses that have been used up
		addressList = keepOnlyNew(addressList)

		if len(addressList) > 0 {
			earliestAddress := addressList[len(addressList)-1]
			return &treasury.PrepareDepositResult{
				Fee:      dm.fee,
				Address:  earliestAddress.address,
				ExpireTs: earliestAddress.expireTs,
			}, nil
		}

		// error if we just tried to generate a new address which failed
		if generateNewAddress {
			return nil, fmt.Errorf("attempt to generate a new address failed")
		}

		// retry the loop by attempting to generate a new address
		generateNewAddress = true
	}
}

func keepOnlyNew(addressList []depositAddress) []depositAddress {
	ret := []depositAddress{}
	for _, a := range addressList {
		if a.isNew {
			ret = append(ret, a)
		}
	}
	return ret
}

type depositMethod struct {
	method     string
	limit      *number.Number
	fee        *number.Number
	genAddress bool
}

func (k krakenExchange) getDepositMethods(asset string) (*depositMethod, error) {
	resp, e := k.api.Query(
		"DepositMethods",
		map[string]string{"asset": asset},
	)
	if e != nil {
		return nil, e
	}

	switch arr := resp.(type) {
	case []interface{}:
		switch m := arr[0].(type) {
		case map[string]interface{}:
			return parseDepositMethods(m)
		default:
			return nil, fmt.Errorf("could not parse inner response type of returned []interface{} from DepositMethods: %s", reflect.TypeOf(m))
		}
	default:
		return nil, fmt.Errorf("could not parse response type from DepositMethods: %s", reflect.TypeOf(arr))
	}
}

type depositAddress struct {
	address  string
	expireTs int64
	isNew    bool
}

func (k krakenExchange) getDepositAddress(asset string, method string, genAddress bool) ([]depositAddress, error) {
	input := map[string]string{
		"asset":  asset,
		"method": method,
	}
	if genAddress {
		// only set "new" if it's supposed to be 'true'. If you set it to 'false' then it will be treated as true by Kraken :(
		input["new"] = "true"
	}
	resp, e := k.api.Query("DepositAddresses", input)
	if e != nil {
		return []depositAddress{}, e
	}

	addressList := []depositAddress{}
	switch arr := resp.(type) {
	case []interface{}:
		for _, elem := range arr {
			switch m := elem.(type) {
			case map[string]interface{}:
				da, e := parseDepositAddress(m)
				if e != nil {
					return []depositAddress{}, e
				}
				addressList = append(addressList, *da)
			default:
				return []depositAddress{}, fmt.Errorf("could not parse inner response type of returned []interface{} from DepositAddresses: %s", reflect.TypeOf(m))
			}
		}
	default:
		return []depositAddress{}, fmt.Errorf("could not parse response type from DepositAddresses: %s", reflect.TypeOf(arr))
	}
	return addressList, nil
}

func parseDepositAddress(m map[string]interface{}) (*depositAddress, error) {
	// address
	address, e := parseString(m, "address", "DepositAddresses")
	if e != nil {
		return nil, e
	}

	// expiretm
	expireN, e := parseNumber(m, "expiretm", "DepositAddresses")
	if e != nil {
		return nil, e
	}
	expireTs := int64(expireN.AsFloat())

	// new
	isNew, e := parseBool(m, "new", "DepositAddresses")
	if e != nil {
		if !strings.HasPrefix(e.Error(), prefixFieldNotFound) {
			return nil, e
		}
		// new may be missing in which case it's false
		isNew = false
	}

	return &depositAddress{
		address:  address,
		expireTs: expireTs,
		isNew:    isNew,
	}, nil
}

func parseDepositMethods(m map[string]interface{}) (*depositMethod, error) {
	// method
	method, e := parseString(m, "method", "DepositMethods")
	if e != nil {
		return nil, e
	}

	// limit
	var limit *number.Number
	limB, e := parseBool(m, "limit", "DepositMethods")
	if e != nil {
		// limit is special as it can be a boolean or a number
		limit, e = parseNumber(m, "limit", "DepositMethods")
		if e != nil {
			return nil, e
		}
	} else {
		if limB {
			return nil, fmt.Errorf("invalid value for 'limit' as a response from DepositMethods: boolean value of 'limit' should never be 'true' as it should be a number in that case")
		}
		limit = nil
	}

	// fee
	fee, e := parseNumber(m, "fee", "DepositMethods")
	if e != nil {
		if !strings.HasPrefix(e.Error(), prefixFieldNotFound) {
			return nil, e
		}
		// fee may be missing in which case it's null
		fee = nil
	}

	// gen-address
	genAddress, e := parseBool(m, "gen-address", "DepositMethods")
	if e != nil {
		return nil, e
	}

	return &depositMethod{
		method:     method,
		limit:      limit,
		fee:        fee,
		genAddress: genAddress,
	}, nil
}
