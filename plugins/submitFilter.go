package plugins

import (
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
)

// SubmitFilter allows you to filter out operations before submitting to the network
type SubmitFilter interface {
	Apply(
		ops []txnbuild.Operation,
		sellingOffers []hProtocol.Offer, // quoted quote/base
		buyingOffers []hProtocol.Offer, // quoted base/quote
	) ([]txnbuild.Operation, error)
}
