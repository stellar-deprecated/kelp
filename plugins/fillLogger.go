package plugins

import (
	"log"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

// FillLogger is a FillHandler that logs fills
type FillLogger struct{}

var _ api.FillHandler = &FillLogger{}

// MakeFillLogger is a factory method
func MakeFillLogger() api.FillHandler {
	return &FillLogger{}
}

// HandleFill impl.
func (f *FillLogger) HandleFill(trade model.Trade) error {
	log.Printf("received fill: %s\n", trade)
	return nil
}
