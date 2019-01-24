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
	SubmitModeBoth
)

// ParseSubmitMode converts a string to the SubmitMode constant
func ParseSubmitMode(submitMode string) SubmitMode {
	if submitMode == "maker_only" {
		return SubmitModeMakerOnly
	}

	return SubmitModeBoth
}

func (s *SubmitMode) String() string {
	if *s == SubmitModeMakerOnly {
		return "maker_only"
	}

	return "both"
}

// submitFilter allows you to filter out operations before submitting to the network
type submitFilter interface {
	apply(ops []build.TransactionMutator) ([]build.TransactionMutator, error)
}

type sdexMakerFilter struct {
	tradingPair *model.TradingPair
	sdex        *plugins.SDEX
}

var _ submitFilter = &sdexMakerFilter{}

// makeSubmitFilter makes a submit filter based on the passed in submitMode
func makeSubmitFilter(submitMode SubmitMode, sdex *plugins.SDEX, tradingPair *model.TradingPair) submitFilter {
	if submitMode == SubmitModeMakerOnly {
		return &sdexMakerFilter{
			tradingPair: tradingPair,
			sdex:        sdex,
		}
	}
	return nil
}

func (f *sdexMakerFilter) apply(ops []build.TransactionMutator) ([]build.TransactionMutator, error) {
	ob := &model.OrderBook{}
	// we only want the top bid and ask values so use a maxCount of 1
	// ob, e := f.sdex.GetOrderBook(f.tradingPair, 1)
	// if e != nil {
	// 	return nil, fmt.Errorf("could not fetch SDEX orderbook: %s", e)
	// }

	var e error
	ops, e = f.filterOps(ops, ob)
	if e != nil {
		return nil, fmt.Errorf("could not apply filter: %s", e)
	}
	return ops, nil
}

func (f *sdexMakerFilter) filterOps(ops []build.TransactionMutator, ob *model.OrderBook) ([]build.TransactionMutator, error) {
	baseAsset, quoteAsset, e := f.sdex.Assets()
	if e != nil {
		return nil, fmt.Errorf("could not get sdex assets: %s", e)
	}
	topBid := ob.TopBid()
	topAsk := ob.TopAsk()

	numDropped := 0
	filteredOps := []build.TransactionMutator{}
	for _, op := range ops {
		var keep bool
		switch o := op.(type) {
		case *build.ManageOfferBuilder:
			keep, e = f.keepOfferMakerMode(baseAsset, quoteAsset, topBid, topAsk, o)
			if e != nil {
				return nil, fmt.Errorf("could not transform offer (pointer case): %s", e)
			}
		case build.ManageOfferBuilder:
			keep, e = f.keepOfferMakerMode(baseAsset, quoteAsset, topBid, topAsk, &o)
			if e != nil {
				return nil, fmt.Errorf("could not check transform offer (non-pointer case): %s", e)
			}
		default:
			keep = true
		}

		if keep {
			filteredOps = append(filteredOps, op)
		} else {
			numDropped++
		}
	}
	log.Printf("dropped %d operations in sdexMakerFilter\n", numDropped)
	return nil, nil
}

func (f *sdexMakerFilter) keepOfferMakerMode(
	baseAsset horizon.Asset,
	quoteAsset horizon.Asset,
	topBid *model.Order,
	topAsk *model.Order,
	op *build.ManageOfferBuilder,
) (bool, error) {
	isSell, e := isSelling(baseAsset, quoteAsset, op.MO.Selling, op.MO.Buying)
	if e != nil {
		return false, fmt.Errorf("error when running the isSelling check: %s", e)
	}

	// TODO test pricing mechanism here manually
	sellPrice := float64(op.MO.Price.N) / float64(op.MO.Price.D)
	var keep bool
	if !isSell && topAsk != nil {
		// invert price when buying
		keep = 1/sellPrice < topAsk.Price.AsFloat()
	} else if isSell && topBid != nil {
		keep = sellPrice > topBid.Price.AsFloat()
	} else {
		keep = true
	}
	return keep, nil
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
