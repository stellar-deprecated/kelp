package plugins

import (
	"fmt"
	"strings"

	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/model"
	"github.com/interstellar/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
)

// sdexFeed represents a pricefeed from the SDEX
type sdexFeed struct {
	sdex       *SDEX
	assetBase  *horizon.Asset
	assetQuote *horizon.Asset
}

// ensure that it implements PriceFeed
var _ api.PriceFeed = &sdexFeed{}

// makeSDEXFeed creates a price feed from buysell's url fields
func makeSDEXFeed(url string) (*sdexFeed, error) {
	urlParts := strings.Split(url, "/")

	baseParts := strings.Split(urlParts[0], ":")
	baseCode := baseParts[0]
	baseIssuer := baseParts[1]
	baseConvert, e := utils.ParseAsset(baseCode, baseIssuer)
	if e != nil {
		return nil, fmt.Errorf("unable to convert base asset url to sdex asset: %s", e)
	}

	quoteParts := strings.Split(urlParts[1], ":")
	quoteCode := quoteParts[0]
	quoteIssuer := quoteParts[1]
	quoteConvert, e := utils.ParseAsset(quoteCode, quoteIssuer)
	if e != nil {
		return nil, fmt.Errorf("unable to convert quote asset url to sdex asset: %s", e)
	}

	tradingPair := &model.TradingPair{
		Base:  model.Asset(utils.Asset2CodeString(*baseConvert)),
		Quote: model.Asset(utils.Asset2CodeString(*quoteConvert)),
	}

	sdexAssetMap := map[model.Asset]horizon.Asset{
		tradingPair.Base:  *baseConvert,
		tradingPair.Quote: *quoteConvert,
	}

	feedSDEX := MakeSDEX(
		PrivateSdexHack.API,
		"",
		"",
		"",
		"",
		PrivateSdexHack.Network,
		nil,
		0,
		0,
		true,
		tradingPair,
		sdexAssetMap,
	)

	return &sdexFeed{
		sdex:       feedSDEX,
		assetBase:  baseConvert,
		assetQuote: quoteConvert,
	}, nil
}

// GetPrice returns the SDEX mid price for the trading pair
func (s *sdexFeed) GetPrice() (float64, error) {
	orderBook, e := s.sdex.GetOrderBook(s.sdex.pair)
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
