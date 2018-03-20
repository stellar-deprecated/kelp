package strategy

import (
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/log"
)

// Strategy represents some logic for a bot to follow while doing market making
type Strategy interface {
	PruneExistingOffers(buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]horizon.Offer, []horizon.Offer)
	PreUpdate(maxAssetA float64, maxAssetB float64, buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) error
	UpdateWithOps(buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]build.TransactionMutator, error)
	PostUpdate() error
}

// selectOfferSide is a helper method to select a single non-nil array of offers
func selectOfferSide(
	buyingAOffers []horizon.Offer,
	sellingAOffers []horizon.Offer,
) []horizon.Offer {
	hasBothSides := buyingAOffers != nil && sellingAOffers != nil
	hasNoSides := buyingAOffers == nil && sellingAOffers == nil
	if hasBothSides || hasNoSides {
		log.Panic("invalid arguments to PreUpdate")
	}

	if buyingAOffers != nil {
		return buyingAOffers
	}
	return sellingAOffers
}
