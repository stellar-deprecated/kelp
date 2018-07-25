package api

import (
	"fmt"

	"github.com/lightyeario/kelp/model"
)

// PrepareDepositResult is the result of a PrepareDeposit call
type PrepareDepositResult struct {
	Fee      *model.Number // fee that will be deducted from your deposit, i.e. amount available is depositAmount - fee
	Address  string        // address you should send the funds to
	ExpireTs int64         // expire time as a unix timestamp, 0 if it does not expire
}

// DepositAPI is defined by anything where you can deposit funds.
type DepositAPI interface {
	/*
		Input:
			asset - asset you want to deposit
			amount - amount you want to deposit
		Output:
			PrepareDepositResult - contains the deposit instructions
			error - ErrDepositAmountAboveLimit, ErrTooManyDepositAddresses, or any other error
	*/
	PrepareDeposit(asset model.Asset, amount *model.Number) (*PrepareDepositResult, error)
}

// ErrDepositAmountAboveLimit error type
type ErrDepositAmountAboveLimit error

// MakeErrDepositAmountAboveLimit is a factory method
func MakeErrDepositAmountAboveLimit(amount *model.Number, limit *model.Number) ErrDepositAmountAboveLimit {
	return fmt.Errorf("deposit amount (%s) is greater than limit (%s)", amount.AsString(), limit.AsString())
}

// ErrTooManyDepositAddresses error type
type ErrTooManyDepositAddresses error

// MakeErrTooManyDepositAddresses is a factory method
func MakeErrTooManyDepositAddresses() ErrTooManyDepositAddresses {
	return fmt.Errorf("too many deposit addresses, try reusing one of them")
}

// WithdrawInfo is the result of a GetWithdrawInfo call
type WithdrawInfo struct {
	AmountToReceive *model.Number // amount that you will receive after any fees is taken (excludes fees charged on the deposit side)
}

// WithdrawFunds is the result of a WithdrawFunds call
type WithdrawFunds struct {
	WithdrawalID string
}

// WithdrawAPI is defined by anything where you can withdraw funds.
type WithdrawAPI interface {
	/*
		Input:
			asset - asset you want to withdraw
			amountToWithdraw - amount you want deducted from your account
			address - address you want to withdraw to
		Output:
			WithdrawInfo - details on how to perform the withdrawal
			error - ErrWithdrawAmountAboveLimit, ErrWithdrawAmountInvalid, or any other error
	*/
	GetWithdrawInfo(asset model.Asset, amountToWithdraw *model.Number, address string) (*WithdrawInfo, error)

	/*
		Input:
			asset - asset you want to withdraw
			amountToWithdraw - amount you want deducted from your account (fees will be deducted from here, use GetWithdrawInfo for fee estimate)
			address - address you want to withdraw to
		Output:
		    WithdrawFunds - result of the withdrawal
			error - any error
	*/
	WithdrawFunds(
		asset model.Asset,
		amountToWithdraw *model.Number,
		address string,
	) (*WithdrawFunds, error)
}

// ErrWithdrawAmountAboveLimit error type
type ErrWithdrawAmountAboveLimit error

// MakeErrWithdrawAmountAboveLimit is a factory method
func MakeErrWithdrawAmountAboveLimit(amount *model.Number, limit *model.Number) ErrWithdrawAmountAboveLimit {
	return fmt.Errorf("withdraw amount (%s) is greater than limit (%s)", amount.AsString(), limit.AsString())
}

// ErrWithdrawAmountInvalid error type
type ErrWithdrawAmountInvalid error

// MakeErrWithdrawAmountInvalid is a factory method
func MakeErrWithdrawAmountInvalid(amountToWithdraw *model.Number, fee *model.Number) ErrWithdrawAmountInvalid {
	return fmt.Errorf("amountToWithdraw is invalid: %s, fee: %s", amountToWithdraw.AsString(), fee.AsString())
}
