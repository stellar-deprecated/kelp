package trader

import (
	"fmt"
	"log"

	"github.com/interstellar/kelp/model"
	"github.com/interstellar/kelp/plugins"
	"github.com/interstellar/kelp/support/utils"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/xdr"
)

// SubmitMode is the type of mode to be used when submitting orders to the trader bot
type SubmitMode uint8

// constants for the SubmitMode
const (
	SubmitModeMakerOnly SubmitMode = iota
	SubmitModeTakerOnly
	SubmitModeBoth
)

// ParseSubmitMode converts a string to the SubmitMode constant
func ParseSubmitMode(submitMode string) SubmitMode {
	if submitMode == "maker_only" {
		return SubmitModeMakerOnly
	}

	if submitMode == "taker_only" {
		return SubmitModeTakerOnly
	}

	return SubmitModeBoth
}

func (s *SubmitMode) String() string {
	if *s == SubmitModeMakerOnly {
		return "maker_only"
	}

	if *s == SubmitModeTakerOnly {
		return "taker_only"
	}

	return "both"
}

// submitFilter allows you to filter out operations before submitting to the network
type submitFilter interface {
	apply(ops []build.TransactionMutator) ([]build.TransactionMutator, error)
}

type sdexFilter struct {
	tradingPair    *model.TradingPair
	sdex           *plugins.SDEX
	submitMode     SubmitMode
	transformOffer func(baseAsset horizon.Asset, quoteAsset horizon.Asset, ob *model.OrderBook, op *build.ManageOfferBuilder) (*build.ManageOfferBuilder, error)
}

var _ submitFilter = &sdexFilter{}

// makeSdexFilter makes a submit filter based on the passed in submitMode
func makeSdexFilter(submitMode SubmitMode, sdex *plugins.SDEX, tradingPair *model.TradingPair) submitFilter {
	if submitMode == SubmitModeMakerOnly {
		return &sdexFilter{
			tradingPair:    tradingPair,
			sdex:           sdex,
			submitMode:     submitMode,
			transformOffer: transformOfferMakerMode,
		}
	} else if submitMode == SubmitModeTakerOnly {
		return &sdexFilter{
			tradingPair:    tradingPair,
			sdex:           sdex,
			submitMode:     submitMode,
			transformOffer: transformOfferTakerMode,
		}
	}
	return nil
}

func (f *sdexFilter) apply(ops []build.TransactionMutator) ([]build.TransactionMutator, error) {
	ob := &model.OrderBook{}
	// we only want the top bid and ask values so use a maxCount of 1
	// ob, e := f.sdex.GetOrderBook(f.tradingPair, 1)
	// if e != nil {
	// 	return nil, fmt.Errorf("could not fetch SDEX orderbook: %s", e)
	// }

	var e error
	ops, e = f.filter(ops, ob)
	if e != nil {
		return nil, fmt.Errorf("could not apply filter (submitMode=%s): %s", f.submitMode.String(), e)
	}
	return ops, nil
}

func (f *sdexFilter) filter(ops []build.TransactionMutator, ob *model.OrderBook) ([]build.TransactionMutator, error) {
	baseAsset, quoteAsset, e := f.sdex.Assets()
	if e != nil {
		return nil, fmt.Errorf("could not get sdex assets: %s", e)
	}

	numDropped := 0
	filteredOps := []build.TransactionMutator{}
	for _, op := range ops {
		switch o := op.(type) {
		case *build.ManageOfferBuilder:
			newOp, e := f.transformOffer(baseAsset, quoteAsset, ob, o)
			if e != nil {
				return nil, fmt.Errorf("could not transform offer (pointer case): %s", e)
			}

			if newOp != nil {
				filteredOps = append(filteredOps, newOp)
			} else {
				numDropped++
			}
		case build.ManageOfferBuilder:
			newOp, e := f.transformOffer(baseAsset, quoteAsset, ob, &o)
			if e != nil {
				return nil, fmt.Errorf("could not check transform offer (non-pointer case): %s", e)
			}

			if newOp != nil {
				filteredOps = append(filteredOps, newOp)
			} else {
				numDropped++
			}
		default:
			filteredOps = append(filteredOps, o)
		}
	}
	log.Printf("dropped %d operations in the maker filter", numDropped)
	return nil, nil
}

func transformOfferMakerMode(
	baseAsset horizon.Asset,
	quoteAsset horizon.Asset,
	ob *model.OrderBook,
	op *build.ManageOfferBuilder,
) (*build.ManageOfferBuilder, error) {
	isSell, e := isSelling(baseAsset, quoteAsset, op.MO.Selling, op.MO.Buying)
	if e != nil {
		return nil, fmt.Errorf("error when running the isSelling check: %s", e)
	}

	// TODO test pricing mechanism here manually
	sellPrice := float64(op.MO.Price.N) / float64(op.MO.Price.D)
	topBid := ob.TopBid()
	topAsk := ob.TopAsk()

	var keep bool
	if !isSell && topAsk != nil {
		// invert price when buying
		keep = 1/sellPrice < topAsk.Price.AsFloat()
	} else if isSell && topBid != nil {
		keep = sellPrice > topBid.Price.AsFloat()
	} else {
		keep = true
	}

	if keep {
		return op, nil
	}
	return nil, nil
}

func transformOfferTakerMode(
	baseAsset horizon.Asset,
	quoteAsset horizon.Asset,
	ob *model.OrderBook,
	op *build.ManageOfferBuilder,
) (*build.ManageOfferBuilder, error) {
	// TODO
	return nil, nil
}

func isSelling(sdexBase horizon.Asset, sdexQuote horizon.Asset, selling xdr.Asset, buying xdr.Asset) (bool, error) {
	sellingBase, e := utils.AssetEqualsXDR(sdexBase, selling)
	if e != nil {
		return false, fmt.Errorf("error comparing sdexBase with selling asset")
	}
	buyingQuote, e := utils.AssetEqualsXDR(sdexQuote, buying)
	if e != nil {
		return false, fmt.Errorf("error comparing sdexQuote with buying asset")
	}
	if sellingBase && buyingQuote {
		return true, nil
	}

	sellingQuote, e := utils.AssetEqualsXDR(sdexQuote, selling)
	if e != nil {
		return false, fmt.Errorf("error comparing sdexQuote with selling asset")
	}
	buyingBase, e := utils.AssetEqualsXDR(sdexBase, buying)
	if e != nil {
		return false, fmt.Errorf("error comparing sdexBase with buying asset")
	}
	if sellingQuote && buyingBase {
		return false, nil
	}

	return false, fmt.Errorf("invalid assets, there are more than 2 distinct assets: sdexBase=%s, sdexQuote=%s, selling=%s, buying=%s", sdexBase, sdexQuote, selling, buying)
}
