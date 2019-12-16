package plugins

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/nikhilsaraf/go-tools/multithreading"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

// FillTracker tracks fills
type FillTracker struct {
	pair                             *model.TradingPair
	threadTracker                    *multithreading.ThreadTracker
	fillTrackable                    api.FillTrackable
	fillTrackerSleepMillis           uint32
	fillTrackerDeleteCyclesThreshold int64

	// initialized runtime vars
	fillTrackerDeleteCycles int64

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
	fillTrackerDeleteCyclesThreshold int64,
) api.FillTracker {
	return &FillTracker{
		pair:                             pair,
		threadTracker:                    threadTracker,
		fillTrackable:                    fillTrackable,
		fillTrackerSleepMillis:           fillTrackerSleepMillis,
		fillTrackerDeleteCyclesThreshold: fillTrackerDeleteCyclesThreshold,
		// initialized runtime vars
		fillTrackerDeleteCycles: 0,
	}
}

// GetPair impl
func (f *FillTracker) GetPair() (pair *model.TradingPair) {
	return f.pair
}

// countError updates the error count and returns true if the error limit has been exceeded
func (f *FillTracker) countError() bool {
	if f.fillTrackerDeleteCyclesThreshold < 0 {
		log.Printf("not deleting any offers because fillTrackerDeleteCyclesThreshold is negative\n")
		return false
	}

	f.fillTrackerDeleteCycles++
	if f.fillTrackerDeleteCycles <= f.fillTrackerDeleteCyclesThreshold {
		log.Printf("not deleting any offers, fillTrackerDeleteCycles (=%d) needs to exceed fillTrackerDeleteCyclesThreshold (=%d)\n", f.fillTrackerDeleteCycles, f.fillTrackerDeleteCyclesThreshold)
		return false
	}

	log.Printf("deleting all offers, num. continuous fill tracking cycles with errors (including this one): %d; (fillTrackerDeleteCyclesThreshold to be exceeded=%d)\n", f.fillTrackerDeleteCycles, f.fillTrackerDeleteCyclesThreshold)
	return true
}

// TrackFills impl
func (f *FillTracker) TrackFills() error {
	// get the last cursor so we only start querying from the current position
	lastCursor, e := f.fillTrackable.GetLatestTradeCursor()
	if e != nil {
		return fmt.Errorf("error while getting last trade: %s", e)
	}
	log.Printf("got latest trade cursor from where to start tracking fills: %v\n", lastCursor)

	ech := make(chan error, len(f.handlers))
	for {
		select {
		case e := <-ech:
			// always return an error if any of the fill handlers return an eror
			return fmt.Errorf("caught an error when tracking fills: %s", e)
		default:
			// do nothing
		}

		tradeHistoryResult, e := f.fillTrackable.GetTradeHistory(*f.GetPair(), lastCursor, nil)
		if e != nil {
			eMsg := fmt.Sprintf("error when fetching trades: %s", e)
			if f.countError() {
				return fmt.Errorf(eMsg)
			}
			log.Printf("%s\n", eMsg)
			f.sleep()
			continue
		}

		if len(tradeHistoryResult.Trades) > 0 {
			// use a single goroutine so we handle trades sequentially and also respect the handler sequence
			e = f.threadTracker.TriggerGoroutine(func(inputs []interface{}) {
				ech := inputs[0].(chan error)
				defer handlePanic(ech)

				handlers := inputs[1].([]api.FillHandler)
				trades := inputs[2].([]model.Trade)
				for _, t := range trades {
					for _, h := range handlers {
						e := h.HandleFill(t)
						if e != nil {
							ech <- fmt.Errorf("error in a fill handler: %s", e)
							// we do NOT want to exit from the goroutine immediately after encountering an error
							// because we want to give all handlers a chance to get called for each trade
						}
					}
				}
			}, []interface{}{ech, f.handlers, tradeHistoryResult.Trades})
			if e != nil {
				eMsg := fmt.Sprintf("error spawning fill handler: %s", e)
				if f.countError() {
					return fmt.Errorf(eMsg)
				}
				log.Printf("%s\n", eMsg)
				f.sleep()
				continue
			}
		}

		lastCursor = tradeHistoryResult.Cursor
		log.Printf("updated lastCursor value to %v\n", lastCursor)
		f.fillTrackerDeleteCycles = 0
		f.sleep()
	}
}

func (f *FillTracker) sleep() {
	time.Sleep(time.Duration(f.fillTrackerSleepMillis) * time.Millisecond)
}

func handlePanic(ech chan error) {
	if r := recover(); r != nil {
		e := r.(error)

		log.Printf("handling panic by passing onto error channel: %s\n%s", e, string(debug.Stack()))
		ech <- e
	}
}

// RegisterHandler impl
func (f *FillTracker) RegisterHandler(handler api.FillHandler) {
	f.handlers = append(f.handlers, handler)
}

// NumHandlers impl
func (f *FillTracker) NumHandlers() uint8 {
	return uint8(len(f.handlers))
}
