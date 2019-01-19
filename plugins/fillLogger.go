package plugins

import (
	"log"

	"github.com/interstellar/kelp/support/logger"

	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/model"
)

// FillLogger is a FillHandler that logs fills
type FillLogger struct {
	l logger.Logger
}

var _ api.FillHandler = &FillLogger{}

// MakeFillLogger is a factory method
func MakeFillLogger() api.FillHandler {
	l := logger.MakeBasicLogger()
	return &FillLogger{
		l,
	}
}

// HandleFill impl.
func (f *FillLogger) HandleFill(trade model.Trade) error {
	log.Printf("received fill: %s\n", trade)
	return nil
}
