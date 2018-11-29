package plugins

import (
	"fmt"

	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/model"
	"github.com/nikhilsaraf/go-tools/multithreading"
)

// FillTracker tracks fills
type FillTracker struct {
	pair          *model.TradingPair
	threadTracker *multithreading.ThreadTracker
	tradeFetcher  api.TradeFetcher

	// uninitialized
	handlers []api.FillHandler
}

// enforce FillTracker implementing api.FillTracker
var _ api.FillTracker = &FillTracker{}

// MakeFillTracker impl.
func MakeFillTracker(pair *model.TradingPair, threadTracker *multithreading.ThreadTracker, tradeFetcher api.TradeFetcher) api.FillTracker {
	return &FillTracker{
		pair:          pair,
		threadTracker: threadTracker,
		tradeFetcher:  tradeFetcher,
	}
}

// GetPair impl
func (f *FillTracker) GetPair() (pair *model.TradingPair) {
	return f.pair
}

// TrackFills impl
func (f *FillTracker) TrackFills() error {
	// use a separate bool to track if we're in the first iteration because we could be tracking an account that
	// has no trades and so we cannot depend on the lastCursor alone
	isFirstIteration := true
	var lastCursor interface{}

	for {
		tradeHistoryResult, e := f.tradeFetcher.GetTradeHistory(lastCursor, nil)
		if e != nil {
			return fmt.Errorf("error when fetching trades: %s", e)
		}

		lastCursor = tradeHistoryResult.Cursor
		if isFirstIteration {
			isFirstIteration = false
			continue
		}

		for _, t := range tradeHistoryResult.Trades {
			for _, h := range f.handlers {
				f.threadTracker.TriggerGoroutine(func(inputs []interface{}) {
					h := inputs[0].(api.FillHandler)
					t := inputs[1].(model.Trade)
					h.HandleFill(t)
				}, []interface{}{h, t})
			}
		}
	}
}

// RegisterHandler impl
func (f *FillTracker) RegisterHandler(handler api.FillHandler) {
	f.handlers = append(f.handlers, handler)
}
