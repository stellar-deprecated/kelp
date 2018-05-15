package treasury

import (
	"github.com/lightyeario/kelp/support/exchange/api/assets"
	"github.com/lightyeario/kelp/support/exchange/api/number"
)

// Account allows you to access key account functions
type Account interface {
	GetAccountBalances(assetList []assets.Asset) (map[assets.Asset]number.Number, error)
}
