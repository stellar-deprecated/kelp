package treasury

import (
	"github.com/lightyeario/kelp/support/exchange/api/assets"
	"github.com/lightyeario/kelp/support/exchange/api/number"
)

// DepositAPI is defined by anything where you can deposit assets. Examples of this are Exchange and Anchor
type DepositAPI interface {
	/*
		Input:
			asset - asset you want to deposit
			amount - amount you want to deposit
		Output:
			fee - fee deducted from your amount, i.e. amount available is amount - fee
			address - address you should send the funds to
			e - any error
	*/
	PrepareDeposit(
		asset assets.Asset,
		amount number.Number,
	) (fee number.Number, address string, e error)
}
