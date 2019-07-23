package plugins

import (
	"github.com/stellar/go/build"
	hProtocol "github.com/stellar/go/protocols/horizon"
)

// SubmitFilter allows you to filter out operations before submitting to the network
type SubmitFilter interface {
	Apply(
		ops []build.TransactionMutator,
		sellingOffers []hProtocol.Offer, // quoted quote/base
		buyingOffers []hProtocol.Offer, // quoted base/quote
	) ([]build.TransactionMutator, error)
}
