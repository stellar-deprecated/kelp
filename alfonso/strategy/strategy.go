package strategy

import (
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

// Strategy represents some logic for a bot to follow while doing market making
type Strategy interface {
	PruneExistingOffers(buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]horizon.Offer, []horizon.Offer)
	PreUpdate(maxAssetA float64, maxAssetB float64, buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) error
	UpdateWithOps(buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]build.TransactionMutator, error)
	PostUpdate() error
}
