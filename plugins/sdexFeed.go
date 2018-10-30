package plugins

import (
	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
)

func GetSDEXPrice(sdex *SDEX, assetBase *horizon.Asset, assetQuote *horizon.Asset) (float64, error) {
	orderBook, e := utils.GetOrderBook(sdex.API, assetBase, assetQuote)
	if e != nil {
		return 0, e
	}
	bids := orderBook.Bids
	topBidPrice := utils.PriceAsFloat(bids[0].Price)
	asks := orderBook.Asks
	lowAskPrice := utils.PriceAsFloat(asks[0].Price)
	centerPrice := (topBidPrice + lowAskPrice) / 2
	return centerPrice, nil
}
