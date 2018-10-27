package plugins

import (
	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
)
func GetSDEXPrice(sdex *SDEX, assetBase *horizon.Asset, assetQuote *horizon.Asset, minTotalOrder float64) (float64, error) {
	orderBook, e := utils.GetOrderBook(sdex.API, assetBase, assetQuote)
	if e != nil {
		return 0, e
	}

	bids := orderBook.Bids
	asks := orderBook.Asks

	topBidPrice := utils.PriceAsFloat(bids[0].Price)
	lowAskPrice := utils.PriceAsFloat(asks[0].Price)

	total := 0.0
	refBidPrice := topBidPrice
	found := false
	for i := 0; i < len(bids) && found == false; i++ {
		total += utils.AmountStringAsFloat(bids[i].Amount)
		refBidPrice = utils.PriceAsFloat(bids[i].Price)
		if total >= minTotalOrder {
			found = true
		}
	}

	total = 0.0
	refAskPrice := lowAskPrice
	found = false
	for i := 0; i < len(asks) && found == false; i++ {
		total += utils.AmountStringAsFloat(asks[i].Amount)
		refAskPrice = utils.PriceAsFloat(asks[i].Price)
		if total >= minTotalOrder {
			found = true
		}
	}

	CenterPrice := (refBidPrice + refAskPrice) / 2

	if CenterPrice < topBidPrice {
		CenterPrice = topBidPrice * 1.0001
	}

	if CenterPrice > lowAskPrice {
		CenterPrice = lowAskPrice * 0.9999
	}

	return CenterPrice, nil
}
