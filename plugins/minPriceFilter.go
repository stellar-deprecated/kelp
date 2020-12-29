package plugins

import (
	"fmt"
	"log"
	"strconv"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/support/utils"
)

// MinPriceFilterConfig ensures that any one constraint that is hit will result in deleting all offers and pausing until limits are no longer constrained
type MinPriceFilterConfig struct {
	MinPrice *float64
}

type minPriceFilter struct {
	name       string
	config     *MinPriceFilterConfig
	baseAsset  hProtocol.Asset
	quoteAsset hProtocol.Asset
}

// MakeFilterMinPrice makes a submit filter that limits orders placed based on the price
func MakeFilterMinPrice(baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset, config *MinPriceFilterConfig) (SubmitFilter, error) {
	return &minPriceFilter{
		name:       "minPriceFilter",
		config:     config,
		baseAsset:  baseAsset,
		quoteAsset: quoteAsset,
	}, nil
}

var _ SubmitFilter = &minPriceFilter{}

// Validate ensures validity
func (c *MinPriceFilterConfig) Validate() error {
	if c.MinPrice == nil {
		return fmt.Errorf("needs a minPrice config value")
	}
	return nil
}

// String is the stringer method
func (c *MinPriceFilterConfig) String() string {
	return fmt.Sprintf("MinPriceFilterConfig[MinPrice=%s]", utils.CheckedFloatPtr(c.MinPrice))
}

func (f *minPriceFilter) Apply(ops []txnbuild.Operation, sellingOffers []hProtocol.Offer, buyingOffers []hProtocol.Offer) ([]txnbuild.Operation, error) {
	ops, e := filterOps(f.name, f.baseAsset, f.quoteAsset, sellingOffers, buyingOffers, ops, f.minPriceFilterFn)
	if e != nil {
		return nil, fmt.Errorf("could not apply filter: %s", e)
	}
	return ops, nil
}

func (f *minPriceFilter) minPriceFilterFn(op *txnbuild.ManageSellOffer) (*txnbuild.ManageSellOffer, error) {
	isSell, e := utils.IsSelling(f.baseAsset, f.quoteAsset, op.Selling, op.Buying)
	if e != nil {
		return nil, fmt.Errorf("error when running the isSelling check for offer '%+v': %s", *op, e)
	}

	sellPrice, e := strconv.ParseFloat(op.Price, 64)
	if e != nil {
		return nil, fmt.Errorf("could not convert price (%s) to float: %s", op.Price, e)
	}

	// reorient price to be in the context of the bot's base and quote asset, in quote units
	price := sellPrice
	if !isSell {
		// invert price for buy side
		price = 1 / sellPrice
	}

	// keep only those ops that meet the comparison mode using the value from the price feed as the threshold
	opRet := op
	if price < *f.config.MinPrice {
		opRet = nil
	}

	log.Printf("minPriceFilter: isSell=%v, price=%.10f, thresholdMinPrice=%.10f, keep=%v", isSell, price, *f.config.MinPrice, opRet != nil)
	return opRet, nil
}
