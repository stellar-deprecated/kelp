package trader

import (
	"github.com/interstellar/kelp/plugins"
	"github.com/stellar/go/build"
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

// submitFilter allows you to filter out operations before submitting to the network
type submitFilter interface {
	apply(ops []build.TransactionMutator) ([]build.TransactionMutator, error)
}

// makeSubmitFilter makes a submit filter based on the passed in submitMode
func makeSubmitFilter(submitMode SubmitMode, sdex *plugins.SDEX) submitFilter {
	if submitMode == SubmitModeMakerOnly || submitMode == SubmitModeTakerOnly {
		return &sdexFilter{
			sdex:       sdex,
			submitMode: submitMode,
		}
	}
	return nil
}

type sdexFilter struct {
	sdex       *plugins.SDEX
	submitMode SubmitMode
}

var _ submitFilter = &sdexFilter{}

func (f *sdexFilter) apply(ops []build.TransactionMutator) ([]build.TransactionMutator, error) {
	// TODO
	return nil, nil
}
