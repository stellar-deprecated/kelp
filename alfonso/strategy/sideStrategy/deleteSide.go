package sideStrategy

import (
	kelp "github.com/lightyeario/kelp/support"
	"github.com/lightyeario/kelp/support/exchange/number"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

// DeleteSideStrategy is a sideStrategy to delete the orders for a given currency pair on one side of the orderbook
type DeleteSideStrategy struct {
	txButler   *kelp.TxButler
	assetBase  *horizon.Asset
	assetQuote *horizon.Asset
}

// ensure it implements SideStrategy
var _ SideStrategy = &DeleteSideStrategy{}

// MakeDeleteSideStrategy is a factory method for DeleteSideStrategy
func MakeDeleteSideStrategy(
	txButler *kelp.TxButler,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
) SideStrategy {
	return &DeleteSideStrategy{
		txButler:   txButler,
		assetBase:  assetBase,
		assetQuote: assetQuote,
	}
}

// PruneExistingOffers impl
func (s *DeleteSideStrategy) PruneExistingOffers(offers []horizon.Offer) []horizon.Offer {
	for i := 0; i < len(offers); i++ {
		s.txButler.DeleteOffer(offers[i])
	}
	return []horizon.Offer{}
}

// PreUpdate impl
func (s *DeleteSideStrategy) PreUpdate(
	maxAssetBase float64,
	maxAssetQuote float64,
	offers []horizon.Offer,
) error {
	return nil
}

// UpdateWithOps impl
func (s *DeleteSideStrategy) UpdateWithOps(offers []horizon.Offer) (ops []build.TransactionMutator, newTopOffer *number.Number, e error) {
	return []build.TransactionMutator{}, nil, nil
}

// PostUpdate impl
func (s *DeleteSideStrategy) PostUpdate() error {
	return nil
}
