package treasury

import (
	"github.com/lightyeario/kelp/support/exchange/api/assets"
	"github.com/lightyeario/kelp/support/exchange/api/number"
)

// WithdrawInfo is the result of a GetWithdrawInfo call
type WithdrawInfo struct {
	amountToReceive *number.Number // amount that you will receive after any fees is taken (excludes fees charged on the deposit side)
}

// WithdrawAPI is defined by anything where you can withdraw assets. Examples of this are Exchange and Anchor
type WithdrawAPI interface {
	/*
		Input:
			asset - asset you want to withdraw
			amountToWithdraw - amount you want deducted from your account
			address - address you want to withdraw to
		Output:
			WithdrawInfo - details on how to perform the withdrawal
			error - any error
	*/
	GetWithdrawInfo(asset assets.Asset, amountToWithdraw *number.Number, address string) (*WithdrawInfo, error)

	/*
		Input:
			asset - asset you want to withdraw
			amountToWithdraw - amount you want deducted from your account (fees will be deducted from here, use GetWithdrawInfo for fee estimate)
			address - address you want to withdraw to
		Output:
			error - any error
	*/
	WithdrawFunds(
		asset assets.Asset,
		amountToWithdraw *number.Number,
		address string,
	) error
}
