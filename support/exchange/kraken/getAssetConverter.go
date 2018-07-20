package kraken

import (
	"github.com/lightyeario/kelp/model"
)

// GetAssetConverter impl.
func (k krakenExchange) GetAssetConverter() *model.AssetConverter {
	return k.assetConverter
}
