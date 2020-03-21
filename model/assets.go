package model

import (
	"errors"
	"fmt"
	"log"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/support/utils"
)

// Asset is typed and enlists the allowed assets that are understood by the bot
type Asset string

// this is the list of assets understood by the bot.
// This string can be converted by the specific exchange adapter as is needed by the exchange's API
const (
	XLM     Asset = "XLM"
	BTC     Asset = "BTC"
	USD     Asset = "USD"
	ETH     Asset = "ETH"
	LTC     Asset = "LTC"
	REP     Asset = "REP"
	ADA     Asset = "ADA"
	BCH     Asset = "BCH"
	DASH    Asset = "DASH"
	EOS     Asset = "EOS"
	GNO     Asset = "GNO"
	GRIN    Asset = "GRIN"
	FEE     Asset = "FEE"
	QTUM    Asset = "QTUM"
	USDT    Asset = "USDT"
	TUSD    Asset = "TUSD"
	USDC    Asset = "USDC"
	USDS    Asset = "USDS"
	PAX     Asset = "PAX"
	BUSD    Asset = "BUSD"
	DAI     Asset = "DAI"
	DAO     Asset = "DAO"
	ETC     Asset = "ETC"
	ICN     Asset = "ICN"
	MLN     Asset = "MLN"
	NMC     Asset = "NMC"
	XDG     Asset = "XDG"
	XMR     Asset = "XMR"
	XRP     Asset = "XRP"
	XVN     Asset = "XVN"
	ZEC     Asset = "ZEC"
	CAD     Asset = "CAD"
	EUR     Asset = "EUR"
	GBP     Asset = "GBP"
	JPY     Asset = "JPY"
	KRW     Asset = "KRW"
	OMG     Asset = "OMG"
	MANA    Asset = "MANA"
	BULL    Asset = "BULL"
	ETHBULL Asset = "ETHBULL"
)

// AssetConverterInterface is the interface which allows the creation of asset converters with logic instead of static bindings
type AssetConverterInterface interface {
	ToString(a Asset) (string, error)
	FromString(s string) (Asset, error)
	MustFromString(s string) Asset
}

// AssetConverter converts to and from the asset type, it is specific to an exchange
type AssetConverter struct {
	asset2String map[Asset]string
	string2Asset map[string]Asset
}

// ensure AssetConverter implements AssetConverterInterface
var _ AssetConverterInterface = AssetConverter{}

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
		return fmt.Sprintf("missing[%s]", string(a)), nil
	}
	return s, nil
}

// FromString converts from a string to an asset
func (c AssetConverter) FromString(s string) (Asset, error) {
	a, ok := c.string2Asset[s]
	if !ok {
		return "", errors.New("asset converter could not recognize string: " + s)
	}
	return a, nil
}

// MustFromString converts from a string to an asset, failing on errors
func (c AssetConverter) MustFromString(s string) Asset {
	a, e := c.FromString(s)
	if e != nil {
		log.Fatal(fmt.Errorf("exiting on an error-enforced asset conversion: %s", e))
	}
	return a
}

type displayAssetConverter struct{}

// ToString converts an asset to a string
func (c displayAssetConverter) ToString(a Asset) (string, error) {
	return string(a), nil
}

// FromString converts from a string to an asset
func (c displayAssetConverter) FromString(s string) (Asset, error) {
	return Asset(s), nil
}

// MustFromString converts from a string to an asset, failing on errors
func (c displayAssetConverter) MustFromString(s string) Asset {
	a, e := c.FromString(s)
	if e != nil {
		log.Fatal(fmt.Errorf("exiting on an error-enforced asset conversion: %s", e))
	}
	return a
}

// ensure that displayAssetConverter implements AssetConverterInterface
var _ AssetConverterInterface = displayAssetConverter{}

// Display is a basic string-mapping converter for display purposes
var Display = displayAssetConverter{}

// CcxtAssetConverter is the asset converter for the CCXT exchange interface
var CcxtAssetConverter = Display

// KrakenAssetConverter is the asset converter for the Kraken exchange
var KrakenAssetConverter = makeAssetConverter(map[Asset]string{
	XLM:  "XXLM",
	BTC:  "XXBT",
	USD:  "ZUSD",
	ETH:  "XETH",
	LTC:  "XLTC",
	REP:  "XREP",
	ADA:  "ADA",
	BCH:  "BCH",
	DASH: "DASH",
	EOS:  "EOS",
	GNO:  "GNO",
	FEE:  "KFEE",
	QTUM: "QTUM",
	USDT: "USDT",
	DAO:  "XDAO",
	ETC:  "XETC",
	ICN:  "XICN",
	MLN:  "XMLN",
	NMC:  "XNMC",
	XDG:  "XXDG",
	XMR:  "XXMR",
	XRP:  "XXRP",
	XVN:  "XXVN",
	ZEC:  "XZEC",
	CAD:  "ZCAD",
	EUR:  "ZEUR",
	GBP:  "ZGBP",
	JPY:  "ZJPY",
	KRW:  "ZKRW",
})

// KrakenAssetConverterOpenOrders is the asset converter for the Kraken exchange's GetOpenOrders API
var KrakenAssetConverterOpenOrders = makeAssetConverter(map[Asset]string{
	XLM:  "XLM",
	BTC:  "XBT",
	USD:  "USD",
	USDT: "USDT",
	REP:  "REP",
	ETH:  "ETH",
})

// FromHorizonAsset is a factory method
func FromHorizonAsset(hAsset hProtocol.Asset) Asset {
	if hAsset.Type == utils.Native {
		return XLM
	}
	return Asset(hAsset.Code)
}

// AssetDisplayFn is a convenient way to encapsulate the logic to display an Asset
type AssetDisplayFn func(Asset) (string, error)

// MakeSdexMappedAssetDisplayFn is a factory method for a commonly used AssetDisplayFn
func MakeSdexMappedAssetDisplayFn(sdexAssetMap map[Asset]hProtocol.Asset) AssetDisplayFn {
	return AssetDisplayFn(func(asset Asset) (string, error) {
		assetString, ok := sdexAssetMap[asset]
		if !ok {
			return "", fmt.Errorf("cannot recognize the asset %s", string(asset))
		}
		return utils.Asset2String(assetString), nil
	})
}

// MakePassthroughAssetDisplayFn is a factory method for a commonly used AssetDisplayFn
func MakePassthroughAssetDisplayFn() AssetDisplayFn {
	return AssetDisplayFn(func(asset Asset) (string, error) {
		return string(asset), nil
	})
}
