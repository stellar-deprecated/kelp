package plugins

import (
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/support/logger"
)

// makeDeleteStrategy is a factory method
func makeDeleteStrategy(
	sdex *SDEX,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	l logger.Logger,
) api.Strategy {
	return makeComposeStrategy(
		assetBase,
		assetQuote,
		makeDeleteSideStrategy(sdex, assetQuote, assetBase, l), // switch sides of base/quote here for the buy side
		makeDeleteSideStrategy(sdex, assetBase, assetQuote, l),
	)
}
