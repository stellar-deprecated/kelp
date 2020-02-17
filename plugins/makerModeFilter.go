package plugins

import (
	"fmt"
	"log"
	"strconv"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

type makerModeFilter struct {
	name         string
	tradingPair  *model.TradingPair
	exchangeShim api.ExchangeShim
	sdex         *SDEX
}

// MakeFilterMakerMode makes a submit filter based on the passed in submitMode
func MakeFilterMakerMode(exchangeShim api.ExchangeShim, sdex *SDEX, tradingPair *model.TradingPair) SubmitFilter {
	return &makerModeFilter{
		name:         "makeModeFilter",
		tradingPair:  tradingPair,
		exchangeShim: exchangeShim,
		sdex:         sdex,
	}
}

var _ SubmitFilter = &makerModeFilter{}

func (f *makerModeFilter) Apply(ops []txnbuild.Operation, sellingOffers []hProtocol.Offer, buyingOffers []hProtocol.Offer) ([]txnbuild.Operation, error) {
	ob, e := f.exchangeShim.GetOrderBook(f.tradingPair, 50)
	if e != nil {
		return nil, fmt.Errorf("could not fetch orderbook: %s", e)
	}

	baseAsset, quoteAsset, e := f.sdex.Assets()
	if e != nil {
		return nil, fmt.Errorf("could not get assets: %s", e)
	}

	innerFn := func(op *txnbuild.ManageSellOffer) (*txnbuild.ManageSellOffer, error) {
		topBidPrice, e := f.topOrderPriceExcludingTrader(ob.Bids(), buyingOffers, false)
		if e != nil {
			return nil, fmt.Errorf("could not get topOrderPriceExcludingTrader for bids: %s", e)
		}
		topAskPrice, e := f.topOrderPriceExcludingTrader(ob.Asks(), sellingOffers, true)
		if e != nil {
			return nil, fmt.Errorf("could not get topOrderPriceExcludingTrader for asks: %s", e)
		}

		return f.transformOfferMakerMode(baseAsset, quoteAsset, topBidPrice, topAskPrice, op)
	}
	ops, e = filterOps(f.name, baseAsset, quoteAsset, sellingOffers, buyingOffers, ops, innerFn)
	if e != nil {
		return nil, fmt.Errorf("could not apply filter: %s", e)
	}
	return ops, nil
}

func isNewLevel(lastPrice *model.Number, priceNumber *model.Number, isSell bool) bool {
	if lastPrice == nil {
		return true
	} else if isSell && priceNumber.AsFloat() > lastPrice.AsFloat() {
		return true
	} else if !isSell && priceNumber.AsFloat() < lastPrice.AsFloat() {
		return true
	}
	return false
}

func (f *makerModeFilter) collateOffers(traderOffers []hProtocol.Offer, isSell bool) ([]api.Level, error) {
	oc := f.exchangeShim.GetOrderConstraints(f.tradingPair)

	levels := []api.Level{}
	var lastPrice *model.Number
	for _, tOffer := range traderOffers {
		price := float64(tOffer.PriceR.N) / float64(tOffer.PriceR.D)
		amount := utils.AmountStringAsFloat(tOffer.Amount)
		if !isSell {
			price = 1 / price
			amount = amount / price
		}
		priceNumber := model.NumberFromFloat(price, oc.PricePrecision)
		amountNumber := model.NumberFromFloat(amount, oc.VolumePrecision)

		if isNewLevel(lastPrice, priceNumber, isSell) {
			lastPrice = priceNumber
			levels = append(levels, api.Level{
				Price:  *priceNumber,
				Amount: *amountNumber,
			})
		} else if priceNumber.AsFloat() == lastPrice.AsFloat() {
			levels[len(levels)-1].Amount = *levels[len(levels)-1].Amount.Add(*amountNumber)
		} else {
			return nil, fmt.Errorf("invalid ordering of prices (isSell=%v), lastPrice=%s, priceNumber=%s", isSell, lastPrice.AsString(), priceNumber.AsString())
		}
	}
	return levels, nil
}

func (f *makerModeFilter) topOrderPriceExcludingTrader(obSide []model.Order, traderOffers []hProtocol.Offer, isSell bool) (*model.Number, error) {
	traderLevels, e := f.collateOffers(traderOffers, isSell)
	if e != nil {
		return nil, fmt.Errorf("unable to collate offers: %s", e)
	}

	for i, obOrder := range obSide {
		if i >= len(traderLevels) {
			return obOrder.Price, nil
		}

		traderLevel := traderLevels[i]
		if isNewLevel(obOrder.Price, &traderLevel.Price, isSell) {
			return obOrder.Price, nil
		} else if traderLevel.Amount.AsFloat() < obOrder.Volume.AsFloat() {
			return obOrder.Price, nil
		}
	}

	// orderbook only had trader's orders
	return nil, nil
}

func (f *makerModeFilter) transformOfferMakerMode(
	baseAsset hProtocol.Asset,
	quoteAsset hProtocol.Asset,
	topBidPrice *model.Number,
	topAskPrice *model.Number,
	op *txnbuild.ManageSellOffer,
) (*txnbuild.ManageSellOffer, error) {
	// delete operations should never be dropped
	if op.Amount == "0" {
		return op, nil
	}

	isSell, e := utils.IsSelling(baseAsset, quoteAsset, op.Selling, op.Buying)
	if e != nil {
		return nil, fmt.Errorf("error when running the isSelling check for offer '%+v': %s", *op, e)
	}

	sellPrice, e := strconv.ParseFloat(op.Price, 64)
	if e != nil {
		return nil, fmt.Errorf("could not convert price (%s) to float: %s", op.Price, e)
	}

	var keep bool
	if !isSell && topAskPrice != nil {
		// invert price when buying
		keep = 1/sellPrice < topAskPrice.AsFloat()
		log.Printf("makerModeFilter:  buying, keep = (op price) %.7f < %.7f (topAskPrice): keep = %v", 1/sellPrice, topAskPrice.AsFloat(), keep)
	} else if isSell && topBidPrice != nil {
		keep = sellPrice > topBidPrice.AsFloat()
		log.Printf("makerModeFilter: selling, keep = (op price) %.7f > %.7f (topBidPrice): keep = %v", sellPrice, topBidPrice.AsFloat(), keep)
	} else {
		price := sellPrice
		action := "selling"
		if !isSell {
			price = 1 / price
			action = " buying"
		}
		keep = true
		log.Printf("makerModeFilter: %s, no market (op price = %.7f): keep = %v", action, price, keep)
	}

	if keep {
		return op, nil
	}

	// we don't want to keep it so return the dropped command
	return nil, nil
}
