package treasury

import (
	"fmt"

	"github.com/lightyeario/kelp/support/exchange/api/assets"
	"github.com/lightyeario/kelp/support/exchange/api/number"
)

// PrepareDepositResult is the result of a PrepareDeposit call
type PrepareDepositResult struct {
	Fee      *number.Number // fee that will be deducted from your deposit, i.e. amount available is depositAmount - fee
	Address  string         // address you should send the funds to
	ExpireTs int64          // expire time as a unix timestamp, 0 if it does not expire
}

// DepositAPI is defined by anything where you can deposit assets. Examples of this are Exchange and Anchor
type DepositAPI interface {
	/*
		Input:
			asset - asset you want to deposit
			amount - amount you want to deposit
		Output:
			PrepareDepositResult - contains the deposit instructions
			error - ErrAmountAboveLimit, ErrTooManyDepositAddresses, or any other error
	*/
	PrepareDeposit(asset assets.Asset, amount *number.Number) (*PrepareDepositResult, error)
}

// ErrAmountAboveLimit error type
type ErrAmountAboveLimit error

// MakeErrAmountAboveLimit is a factory method
func MakeErrAmountAboveLimit(amount *number.Number, limit *number.Number) ErrAmountAboveLimit {
	return fmt.Errorf("amount (%s) is greater than limit (%s)", amount.AsString(), limit.AsString())
}

// ErrTooManyDepositAddresses error type
type ErrTooManyDepositAddresses error

// MakeErrTooManyDepositAddresses is a factory method
func MakeErrTooManyDepositAddresses() ErrTooManyDepositAddresses {
	return fmt.Errorf("too many deposit addresses, try reusing one of them")
}
