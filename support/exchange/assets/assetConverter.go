package assets

import (
	"errors"
)

// AssetConverter converts to and from the asset type, it is specific to an exchange
type AssetConverter struct {
	asset2String map[Asset]string
	string2Asset map[string]Asset
}

// makeAssetConverter is a factory method for AssetConverter
func makeAssetConverter(asset2String map[Asset]string) *AssetConverter {
	string2Asset := map[string]Asset{}
	for a, s := range asset2String {
		string2Asset[s] = a
	}

	return &AssetConverter{
		asset2String: asset2String,
		string2Asset: string2Asset,
	}
}

// ToString converts an asset to a string
func (c AssetConverter) ToString(a Asset) (string, error) {
	s, ok := c.asset2String[a]
	if !ok {
		return "", errors.New("could not recognize Asset: " + string(a))
	}
	return s, nil
}

// FromString converts from a string to an asset
func (c AssetConverter) FromString(s string) (Asset, error) {
	a, ok := c.string2Asset[s]
	if !ok {
		return "", errors.New("could not recognize string: " + s)
	}
	return a, nil
}

// Display is a basic converter for display purposes
var Display = makeAssetConverter(map[Asset]string{
	XLM: string(XLM),
	BTC: string(BTC),
	USD: string(USD),
})
