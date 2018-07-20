package kraken

import (
	"github.com/lightyeario/kelp/model/assets"
)

// GetAssetConverter impl.
func (k krakenExchange) GetAssetConverter() *assets.AssetConverter {
	return k.assetConverter
}
