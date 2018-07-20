package treasury

import (
	"github.com/lightyeario/kelp/model/assets"
	"github.com/lightyeario/kelp/support/exchange/api/number"
)

// Account allows you to access key account functions
type Account interface {
	GetAccountBalances(assetList []model.Asset) (map[model.Asset]number.Number, error)
}
