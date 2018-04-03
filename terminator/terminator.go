package main

import (
	"fmt"
	"time"

	"github.com/lightyeario/kelp/support"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/log"
)

// Terminator contains the logic to terminate offers
type Terminator struct {
	api                 *horizon.Client
	txb                 *kelp.TxButler
	tradingAccount      string
	tickIntervalSeconds int32
}

// MakeTerminator is a factory method to make a Terminator
func MakeTerminator(
	api *horizon.Client,
	txb *kelp.TxButler,
	tradingAccount string,
	tickIntervalSeconds int32,
) *Terminator {
	return &Terminator{
		api:                 api,
		txb:                 txb,
		tradingAccount:      tradingAccount,
		tickIntervalSeconds: tickIntervalSeconds,
	}
}

// StartService starts the Terminator service
func (t *Terminator) StartService() {
	for {
		t.run()
		log.Info(fmt.Sprintf("sleeping for %d seconds...", t.tickIntervalSeconds))
		time.Sleep(time.Duration(t.tickIntervalSeconds) * time.Second)
	}
}

func (t *Terminator) run() {
	log.Info("In run method")
}
