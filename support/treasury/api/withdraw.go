package treasury

import (
	"github.com/lightyeario/kelp/support/exchange/api/assets"
	"github.com/lightyeario/kelp/support/exchange/api/number"
)

// WithdrawAPI is defined by anything where you can withdraw assets. Examples of this are Exchange and Anchor
type WithdrawAPI interface {
	/*
		Input:
			asset - asset you want to withdraw
			amountToWithdraw - amount you want deducted from your account
			address - address you want to withdraw to
		Output:
			amountToReceive - amount that you will receive after any fees is taken (excludes fees charged on the deposit side)
			e - any error
	*/
	GetWithdrawInfo(
		asset assets.Asset,
		amountToWithdraw number.Number,
		address string,
	) (amountToReceive number.Number, e error)

	/*
		Input:
			asset - asset you want to withdraw
			amountToWithdraw - amount you want deducted from your account (fees will be deducted from here, use GetWithdrawInfo for fee estimate)
			address - address you want to withdraw to
		Output:
			e - any error
	*/
	WithdrawFunds(
		asset assets.Asset,
		amountToWithdraw number.Number,
		address string,
	) error
}
