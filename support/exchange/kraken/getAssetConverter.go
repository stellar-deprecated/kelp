package kraken

import (
	"github.com/lightyeario/kelp/support/exchange/api/assets"
)

// GetAssetConverter impl.
func (k krakenExchange) GetAssetConverter() *assets.AssetConverter {
	return k.assetConverter
}
