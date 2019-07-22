package plugins

import (
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/api"
)

// makeDeleteStrategy is a factory method
func makeDeleteStrategy(
	sdex *SDEX,
	assetBase *hProtocol.Asset,
	assetQuote *hProtocol.Asset,
) api.Strategy {
	return makeComposeStrategy(
		assetBase,
		assetQuote,
		makeDeleteSideStrategy(sdex, assetQuote, assetBase), // switch sides of base/quote here for the buy side
		makeDeleteSideStrategy(sdex, assetBase, assetQuote),
	)
}
