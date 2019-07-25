package api

import (
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/model"
)

// Strategy represents some logic for a bot to follow while doing market making
type Strategy interface {
	PruneExistingOffers(buyingAOffers []hProtocol.Offer, sellingAOffers []hProtocol.Offer) ([]txnbuild.Operation, []hProtocol.Offer, []hProtocol.Offer)
	PreUpdate(maxAssetA float64, maxAssetB float64, trustA float64, trustB float64) error
	UpdateWithOps(buyingAOffers []hProtocol.Offer, sellingAOffers []hProtocol.Offer) ([]txnbuild.Operation, error)
	PostUpdate() error
	GetFillHandlers() ([]FillHandler, error)
}

// SideStrategy represents a strategy on a single side of the orderbook
type SideStrategy interface {
	PruneExistingOffers(offers []hProtocol.Offer) ([]txnbuild.Operation, []hProtocol.Offer)
	PreUpdate(maxAssetA float64, maxAssetB float64, trustA float64, trustB float64) error
	UpdateWithOps(offers []hProtocol.Offer) (ops []txnbuild.Operation, newTopOffer *model.Number, e error)
	PostUpdate() error
	GetFillHandlers() ([]FillHandler, error)
}
