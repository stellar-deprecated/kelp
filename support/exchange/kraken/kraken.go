package kraken

import (
	"fmt"

	"github.com/Beldur/kraken-go-api-client"
	"github.com/lightyeario/kelp/model/assets"
	"github.com/lightyeario/kelp/support/exchange/api"
	tApi "github.com/lightyeario/kelp/support/treasury/api"
)

// ensure that krakenExchange conforms to the Exchange interface
var _ api.Exchange = krakenExchange{}
var _ tApi.Account = krakenExchange{}
var _ tApi.DepositAPI = krakenExchange{}
var _ tApi.WithdrawAPI = krakenExchange{}

// krakenExchange is the implementation for the Kraken Exchange
type krakenExchange struct {
	assetConverter *assets.AssetConverter
	api            *krakenapi.KrakenApi
	delimiter      string
	precision      int8
	withdrawKeys   asset2Address2Key
	isSimulated    bool // will simulate add and cancel orders if this is true
}

type asset2Address2Key map[assets.Asset]map[string]string

func (m asset2Address2Key) getKey(asset assets.Asset, address string) (string, error) {
	address2Key, ok := m[asset]
	if !ok {
		return "", fmt.Errorf("asset is not registered in asset2Address2Key: %v", asset)
	}

	key, ok := address2Key[address]
	if !ok {
		return "", fmt.Errorf("address is not registered in asset2Address2Key: %v (asset = %v)", address, asset)
	}

	return key, nil
}

// MakeKrakenExchange is a factory method to make the kraken exchange
// TODO 2, should take in config file for kraken api keys + withdrawalKeys mapping
func MakeKrakenExchange() api.Exchange {
	return &krakenExchange{
		assetConverter: assets.KrakenAssetConverter,
		api:            krakenapi.New("", ""),
		delimiter:      "",
		withdrawKeys:   asset2Address2Key{},
		precision:      8,
	}
}
