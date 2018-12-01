package plugins

import (
	"fmt"
	"log"
	"time"

	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/model"
	"github.com/nikhilsaraf/go-tools/multithreading"
)

// FillTracker tracks fills
type FillTracker struct {
	pair                   *model.TradingPair
	threadTracker          *multithreading.ThreadTracker
	fillTrackable          api.FillTrackable
	fillTrackerSleepMillis uint32

	// uninitialized
	handlers []api.FillHandler
}

// enforce FillTracker implementing api.FillTracker
var _ api.FillTracker = &FillTracker{}

// MakeFillTracker impl.
func MakeFillTracker(
	pair *model.TradingPair,
	threadTracker *multithreading.ThreadTracker,
	fillTrackable api.FillTrackable,
	fillTrackerSleepMillis uint32,
) api.FillTracker {
	return &FillTracker{
		pair:                   pair,
		threadTracker:          threadTracker,
		fillTrackable:          fillTrackable,
		fillTrackerSleepMillis: fillTrackerSleepMillis,
	}
}

// GetPair impl
func (f *FillTracker) GetPair() (pair *model.TradingPair) {
	return f.pair
}

// TrackFills impl
func (f *FillTracker) TrackFills() error {
	// get the last cursor so we only start querying from the current position
	lastCursor, e := f.fillTrackable.GetLatestTradeCursor()
	if e != nil {
		return fmt.Errorf("error while getting last trade: %s", e)
	}
	log.Printf("got latest trade cursor from where to start tracking fills: %v\n", lastCursor)

	for {
		tradeHistoryResult, e := f.fillTrackable.GetTradeHistory(lastCursor, nil)
		if e != nil {
			return fmt.Errorf("error when fetching trades: %s", e)
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

		lastCursor = tradeHistoryResult.Cursor
		time.Sleep(time.Duration(f.fillTrackerSleepMillis) * time.Millisecond)
	}
}

// RegisterHandler impl
func (f *FillTracker) RegisterHandler(handler api.FillHandler) {
	f.handlers = append(f.handlers, handler)
}
