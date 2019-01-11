package plugins

import (
	"fmt"
	"strings"

	"github.com/interstellar/kelp/support/utils"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

type sdexFeed struct {
	sdex       *SDEX
	assetBase  *horizon.Asset
	assetQuote *horizon.Asset
}

// newSDEXFeed creates a price feed from buysell's url fields
func newSDEXFeed(sdex *SDEX, url string) (*sdexFeed, error) {
	s := new(sdexFeed)
	s.sdex = sdex
	urlParts := strings.Split(url, "/")

	baseURL := strings.Split(urlParts[0], ":")
	baseCode := baseURL[0]
	baseIssuer := baseURL[1]
	baseConvert, e := parseAsset(baseCode, baseIssuer)
	if e != nil {
		return nil, fmt.Errorf("unable to convert base asset url to sdex asset")
	}
	s.assetBase = baseConvert

	quoteURL := strings.Split(urlParts[1], ":")
	quoteCode := quoteURL[0]
	quoteIssuer := quoteURL[1]
	quoteConvert, e := parseAsset(quoteCode, quoteIssuer)
	if e != nil {
		return nil, fmt.Errorf("unable to convert quote asset url to sdex asset")
	}
	s.assetQuote = quoteConvert
	return s, nil
}

func (s *sdexFeed) GetPrice() (float64, error) {
	orderBook, e := utils.GetOrderBook(s.sdex.API, s.assetBase, s.assetQuote)
	if e != nil {
		return 0, fmt.Errorf("unable to get sdex price: %s", e)
	}
	bids := orderBook.Bids
	topBidPrice := utils.PriceAsFloat(bids[0].Price)
	asks := orderBook.Asks
	lowAskPrice := utils.PriceAsFloat(asks[0].Price)
	centerPrice := (topBidPrice + lowAskPrice) / 2
	return centerPrice, nil
}

func parseAsset(code string, issuer string) (*horizon.Asset, error) {
	if code != "XLM" && issuer == "" {
		return nil, fmt.Errorf("error: issuer can only be empty if asset is XLM")
	}

	if code == "XLM" && issuer != "" {
		return nil, fmt.Errorf("error: issuer needs to be empty if asset is XLM")
	}

	if code == "XLM" {
		asset := utils.Asset2Asset2(build.NativeAsset())
		return &asset, nil
	}

	asset := utils.Asset2Asset2(build.CreditAsset(code, issuer))
	return &asset, nil
}
