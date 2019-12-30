package plugins

import (
	"fmt"
	"strconv"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/support/utils"
)

// MaxPriceFilterConfig ensures that any one constraint that is hit will result in deleting all offers and pausing until limits are no longer constrained
type MaxPriceFilterConfig struct {
	MaxPrice *float64
}

type maxPriceFilter struct {
	name       string
	config     *MaxPriceFilterConfig
	baseAsset  hProtocol.Asset
	quoteAsset hProtocol.Asset
}

// MakeFilterMaxPrice makes a submit filter that limits orders placed based on the price
func MakeFilterMaxPrice(baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset, config *MaxPriceFilterConfig) (SubmitFilter, error) {
	return &maxPriceFilter{
		name:       "maxPriceFilter",
		config:     config,
		baseAsset:  baseAsset,
		quoteAsset: quoteAsset,
	}, nil
}

var _ SubmitFilter = &maxPriceFilter{}

// Validate ensures validity
func (c *MaxPriceFilterConfig) Validate() error {
	if c.MaxPrice == nil {
		return fmt.Errorf("needs a maxPrice config value")
	}
	return nil
}

// String is the stringer method
func (c *MaxPriceFilterConfig) String() string {
	return fmt.Sprintf("MaxPriceFilterConfig[MaxPrice=%s]", utils.CheckedFloatPtr(c.MaxPrice))
}

func (f *maxPriceFilter) Apply(ops []txnbuild.Operation, sellingOffers []hProtocol.Offer, buyingOffers []hProtocol.Offer) ([]txnbuild.Operation, error) {
	ops, e := filterOps(f.name, f.baseAsset, f.quoteAsset, sellingOffers, buyingOffers, ops, f.maxPriceFilterFn)
	if e != nil {
		return nil, fmt.Errorf("could not apply filter: %s", e)
	}
	return ops, nil
}

func (f *maxPriceFilter) maxPriceFilterFn(op *txnbuild.ManageSellOffer) (*txnbuild.ManageSellOffer, bool, error) {
	isSell, e := utils.IsSelling(f.baseAsset, f.quoteAsset, op.Selling, op.Buying)
	if e != nil {
		return nil, false, fmt.Errorf("error when running the isSelling check: %s", e)
	}

	sellPrice, e := strconv.ParseFloat(op.Price, 64)
	if e != nil {
		return nil, false, fmt.Errorf("could not convert price (%s) to float: %s", op.Price, e)
	}

	if isSell {
		if sellPrice > *f.config.MaxPrice {
			return nil, false, nil
		}
		return op, true, nil
	}

	// TODO for buy side
	return op, true, nil
}
