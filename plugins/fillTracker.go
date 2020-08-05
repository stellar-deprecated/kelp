package plugins

import (
	"fmt"
	"log"
	"runtime/debug"
	"sync"
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
	lastCursor                       interface{}

	// initialized runtime vars
	fillTrackerDeleteCycles int64
	lockFill                *sync.Mutex
	isRunningInBackground   bool

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
	lastCursor interface{},
) api.FillTracker {
	return &FillTracker{
		pair:                             pair,
		threadTracker:                    threadTracker,
		fillTrackable:                    fillTrackable,
		fillTrackerSleepMillis:           fillTrackerSleepMillis,
		fillTrackerDeleteCyclesThreshold: fillTrackerDeleteCyclesThreshold,
		lastCursor:                       lastCursor,
		// initialized runtime vars
		fillTrackerDeleteCycles: 0,
		lockFill:                &sync.Mutex{},
		isRunningInBackground:   false,
	}
}

// IsRunningInBackground impl
func (f *FillTracker) IsRunningInBackground() bool {
	return f.isRunningInBackground
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
	f.isRunningInBackground = true
	defer func() {
		f.isRunningInBackground = false
	}()

	for {
		_, e := f.FillTrackSingleIteration()
		if e != nil {
			eMsg := fmt.Sprintf("error when running an iteration of fill tracker: %s", e)
			if f.countError() {
				return fmt.Errorf(eMsg)
			}
			log.Printf("%s\n", eMsg)
		}

		f.sleep()
	}
}

// FillTrackSingleIteration is a single run of a call to track fills and to handle the results
func (f *FillTracker) FillTrackSingleIteration() ([]model.Trade, error) {
	// first take the lock
	f.lockFill.Lock()
	defer f.lockFill.Unlock()

	tradeHistoryResult, e := f.fillTrackable.GetTradeHistory(*f.GetPair(), f.lastCursor, nil)
	if e != nil {
		return nil, fmt.Errorf("error when fetching trades: %s", e)
	}

	if len(tradeHistoryResult.Trades) > 0 {
		// create channel with which we can collect errors within goroutines
		ech := make(chan error, len(f.handlers))

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

		// need to wait for fill handlers to finish
		f.threadTracker.Wait()

		// now check for errors in triggering the goroutines
		if e != nil {
			return nil, fmt.Errorf("error spawning fill handler: %s", e)
		}

		// check result of goroutine calls
		select {
		case e := <-ech:
			// always return an error if any of the fill handlers returns an error
			return nil, fmt.Errorf("caught an error when tracking fills: %s", e)
		default:
			// do nothing
		}

		// only update lastCursor if there were trades
		f.lastCursor = tradeHistoryResult.Cursor
		log.Printf("updated lastCursor value to %v\n", f.lastCursor)
	} else {
		log.Printf("there were no trades, leaving lastCursor value as %v\n", f.lastCursor)
	}

	f.fillTrackerDeleteCycles = 0
	return tradeHistoryResult.Trades, nil
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
