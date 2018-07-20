package api

import (
	"github.com/lightyeario/kelp/model"
)

// Account allows you to access key account functions
type Account interface {
	GetAccountBalances(assetList []model.Asset) (map[model.Asset]model.Number, error)
}
