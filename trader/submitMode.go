package trader

import (
	"github.com/interstellar/kelp/model"
	"github.com/interstellar/kelp/plugins"
	"github.com/stellar/go/build"
)

const maxCountSubmitFilter int32 = 10

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

// submitFilter allows you to filter out operations before submitting to the network
type submitFilter interface {
	apply(ops []build.TransactionMutator) ([]build.TransactionMutator, error)
}

// makeSubmitFilter makes a submit filter based on the passed in submitMode
func makeSubmitFilter(submitMode SubmitMode, sdex *plugins.SDEX, tradingPair *model.TradingPair) submitFilter {
	if submitMode == SubmitModeMakerOnly || submitMode == SubmitModeTakerOnly {
		return &sdexFilter{
			tradingPair: tradingPair,
			sdex:        sdex,
			submitMode:  submitMode,
			maxCount:    maxCountSubmitFilter,
		}
	}
	return nil
}

type sdexFilter struct {
	tradingPair *model.TradingPair
	sdex        *plugins.SDEX
	submitMode  SubmitMode
	maxCount    int32
}

var _ submitFilter = &sdexFilter{}

func (f *sdexFilter) apply(ops []build.TransactionMutator) ([]build.TransactionMutator, error) {
	ob := &model.OrderBook{}
	// ob, e := f.sdex.GetOrderBook(f.tradingPair, f.maxCount)
	// if e != nil {
	// 	return nil, fmt.Errorf("could not fetch SDEX orderbook: %s", e)
	// }

	// TODO find intersection of orderbook and ops
	/*
		1. get top bid and top ask in OB
		2. for each op remove or keep op if it is before/after top bid/ask depending on the mode we're in
	*/

	return nil, nil
}
