package sideStrategy

import (
	"fmt"

	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/log"
)

// DeleteSideStrategy is a sideStrategy to delete the orders for a given currency pair on one side of the orderbook
type DeleteSideStrategy struct {
	txButler   *utils.TxButler
	assetBase  *horizon.Asset
	assetQuote *horizon.Asset
}

// ensure it implements SideStrategy
var _ SideStrategy = &DeleteSideStrategy{}

// MakeDeleteSideStrategy is a factory method for DeleteSideStrategy
func MakeDeleteSideStrategy(
	txButler *utils.TxButler,
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
func (s *DeleteSideStrategy) PruneExistingOffers(offers []horizon.Offer) ([]build.TransactionMutator, []horizon.Offer) {
	log.Info(fmt.Sprintf("deleteSideStrategy: deleting %d offers", len(offers)))
	pruneOps := []build.TransactionMutator{}
	for i := 0; i < len(offers); i++ {
		pOp := s.txButler.DeleteOffer(offers[i])
		pruneOps = append(pruneOps, &pOp)
	}
	return pruneOps, []horizon.Offer{}
}

// PreUpdate impl
func (s *DeleteSideStrategy) PreUpdate(maxAssetBase float64, maxAssetQuote float64, trustBase float64, trustQuote float64) error {
	return nil
}

// UpdateWithOps impl
func (s *DeleteSideStrategy) UpdateWithOps(offers []horizon.Offer) (ops []build.TransactionMutator, newTopOffer *model.Number, e error) {
	return []build.TransactionMutator{}, nil, nil
}

// PostUpdate impl
func (s *DeleteSideStrategy) PostUpdate() error {
	return nil
}
