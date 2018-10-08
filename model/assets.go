package model

import (
	"errors"
	"fmt"
	"log"
)

// Asset is typed and enlists the allowed assets that are understood by the bot
type Asset string

// this is the list of assets understood by the bot.
// This string can be converted by the specific exchange adapter as is needed by the exchange's API
const (
	XLM  Asset = "XLM"
	BTC  Asset = "BTC"
	USD  Asset = "USD"
	ETH  Asset = "ETH"
	LTC  Asset = "LTC"
	REP  Asset = "REP"
	ADA  Asset = "ADA"
	BCH  Asset = "BCH"
	DASH Asset = "DASH"
	EOS  Asset = "EOS"
	GNO  Asset = "GNO"
	FEE  Asset = "FEE"
	QTUM Asset = "QTUM"
	USDT Asset = "USDT"
	DAO  Asset = "DAO"
	ETC  Asset = "ETC"
	ICN  Asset = "ICN"
	MLN  Asset = "MLN"
	NMC  Asset = "NMC"
	XDG  Asset = "XDG"
	XMR  Asset = "XMR"
	XRP  Asset = "XRP"
	XVN  Asset = "XVN"
	ZEC  Asset = "ZEC"
	CAD  Asset = "CAD"
	EUR  Asset = "EUR"
	GBP  Asset = "GBP"
	JPY  Asset = "JPY"
	KRW  Asset = "KRW"
	OMG  Asset = "OMG"
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

// Display is a basic converter for display purposes
var Display = makeAssetConverter(map[Asset]string{
	XLM:  string(XLM),
	BTC:  string(BTC),
	USD:  string(USD),
	ETH:  string(ETH),
	LTC:  string(LTC),
	REP:  string(REP),
	ADA:  string(ADA),
	BCH:  string(BCH),
	DASH: string(DASH),
	EOS:  string(EOS),
	GNO:  string(GNO),
	FEE:  string(FEE),
	QTUM: string(QTUM),
	USDT: string(USDT),
	DAO:  string(DAO),
	ETC:  string(ETC),
	ICN:  string(ICN),
	MLN:  string(MLN),
	NMC:  string(NMC),
	XDG:  string(XDG),
	XMR:  string(XMR),
	XRP:  string(XRP),
	XVN:  string(XVN),
	ZEC:  string(ZEC),
	CAD:  string(CAD),
	EUR:  string(EUR),
	GBP:  string(GBP),
	JPY:  string(JPY),
	KRW:  string(KRW),
	OMG:  string(OMG),
})

// CcxtAssetConverter is the asset converter for the CCXT exchange interface
// TODO define a scalable approach to referencing assets using CCXT w.r.t. dev cost
var CcxtAssetConverter = makeAssetConverter(map[Asset]string{
	XLM:  string(XLM),
	BTC:  string(BTC),
	USD:  string(USD),
	ETH:  string(ETH),
	LTC:  string(LTC),
	REP:  string(REP),
	BCH:  string(BCH),
	USDT: string(USDT),
	ICN:  string(ICN),
	OMG:  string(OMG),
})

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
