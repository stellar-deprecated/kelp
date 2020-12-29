package plugins

import (
	"fmt"
	"log"
	"strconv"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/support/utils"
)

type comparisonMode int8

const (
	comparisonModeOutsideExclude comparisonMode = iota // gt for sell, lt for buy
	comparisonModeOutsideInclude                       // gte for sell, lte for buy
)

func (c comparisonMode) keepSellOp(threshold float64, price float64) bool {
	if c == comparisonModeOutsideExclude {
		return price > threshold
	} else if c == comparisonModeOutsideInclude {
		return price >= threshold
	}
	panic("unidentified comparisonMode")
}

func (c comparisonMode) keepBuyOp(threshold float64, price float64) bool {
	if c == comparisonModeOutsideExclude {
		return price < threshold
	} else if c == comparisonModeOutsideInclude {
		return price <= threshold
	}
	panic("unidentified comparisonMode")
}

type priceFeedFilter struct {
	name       string
	baseAsset  hProtocol.Asset
	quoteAsset hProtocol.Asset
	pf         api.PriceFeed
	cm         comparisonMode
}

// MakeFilterPriceFeed makes a submit filter that limits orders placed based on the value of the price feed
func MakeFilterPriceFeed(baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset, comparisonModeString string, pf api.PriceFeed) (SubmitFilter, error) {
	var cm comparisonMode
	if comparisonModeString == "outside-exclude" {
		cm = comparisonModeOutsideExclude
	} else if comparisonModeString == "outside-include" {
		cm = comparisonModeOutsideInclude
	} else {
		return nil, fmt.Errorf("invalid comparisonMode ('%s') used for priceFeedFilter", comparisonModeString)
	}

	return &priceFeedFilter{
		name:       "priceFeedFilter",
		baseAsset:  baseAsset,
		quoteAsset: quoteAsset,
		cm:         cm,
		pf:         pf,
	}, nil
}

var _ SubmitFilter = &priceFeedFilter{}

func (f *priceFeedFilter) Apply(ops []txnbuild.Operation, sellingOffers []hProtocol.Offer, buyingOffers []hProtocol.Offer) ([]txnbuild.Operation, error) {
	ops, e := filterOps(f.name, f.baseAsset, f.quoteAsset, sellingOffers, buyingOffers, ops, f.priceFeedFilterFn)
	if e != nil {
		return nil, fmt.Errorf("could not apply filter: %s", e)
	}
	return ops, nil
}

func (f *priceFeedFilter) priceFeedFilterFn(op *txnbuild.ManageSellOffer) (*txnbuild.ManageSellOffer, error) {
	isSell, e := utils.IsSelling(f.baseAsset, f.quoteAsset, op.Selling, op.Buying)
	if e != nil {
		return nil, fmt.Errorf("error when running the isSelling check for offer '%+v': %s", *op, e)
	}

	sellPrice, e := strconv.ParseFloat(op.Price, 64)
	if e != nil {
		return nil, fmt.Errorf("could not convert price (%s) to float: %s", op.Price, e)
	}

	thresholdFeedPrice, e := f.pf.GetPrice()
	if e != nil {
		return nil, fmt.Errorf("could not get price from priceFeed: %s", e)
	}

	// reorient price to be in the context of the bot's base and quote asset, in quote units
	price := sellPrice
	if !isSell {
		// invert price for buy side
		price = 1 / sellPrice
	}

	// keep only those ops that meet the comparison mode using the value from the price feed as the threshold
	// the "outside" comparison mode on the priceFeed means different things for bids and asks
	opRet := op
	if isSell && !f.cm.keepSellOp(thresholdFeedPrice, price) {
		opRet = nil
	} else if !isSell && !f.cm.keepBuyOp(thresholdFeedPrice, price) {
		opRet = nil
	}

	log.Printf("priceFeedFilter: isSell=%v, price=%.10f, thresholdFeedPrice=%.10f, keep=%v", isSell, price, thresholdFeedPrice, opRet != nil)
	return opRet, nil
}
